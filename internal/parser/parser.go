package parser

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/sudorandom/protoc-gen-jev/internal/model"
	jevv1 "github.com/sudorandom/protoc-gen-jev/pkg/jev/v1"
)

// ProcessMessage parses a single Protobuf message into a Jev MessageSpec.
func ProcessMessage(msg *protogen.Message) (model.MessageSpec, error) {
	fileBase := strings.TrimSuffix(filepath.Base(string(msg.Desc.ParentFile().Path())), ".proto")
	spec := model.MessageSpec{
		MessageName: string(msg.Desc.Name()),
		Package:     string(msg.Desc.ParentFile().Package()),
		FileBase:    fileBase,
		Questions:   make(map[string]model.Question),
	}

	oneofsProcessed := make(map[string]bool)

	for _, field := range msg.Fields {
		desc := field.Desc

		// 1. Handle oneof mutual exclusion as a Choice question
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			oneof := field.Oneof
			oneofName := string(oneof.Desc.Name())
			if !oneofsProcessed[oneofName] {
				oneofsProcessed[oneofName] = true

				instructions := CleanComments(oneof.Comments.Leading.String())
				if instructions == "" {
					instructions = fmt.Sprintf("Select %s variant", oneofName)
				}

				criteria := make(map[string]any)
				oneofCases := make(map[string]string)
				oneofCaseKinds := make(map[string]string)
				for _, f := range oneof.Fields {
					fName := string(f.Desc.Name())
					fDoc := CleanComments(f.Comments.Leading.String())
					if fDoc == "" {
						fDoc = fName
					}
					criteria[fName] = fDoc
					oneofCases[fName] = ToPascalCase(fName)
					oneofCaseKinds[fName] = f.Desc.Kind().String()
				}

				jsonName := string(oneof.Desc.Name())
				goFieldName := ToPascalCase(jsonName)
				spec.Questions[jsonName] = model.Question{
					Type:           model.TypeChoice,
					Instructions:   instructions,
					Criteria:       criteria,
					ProtoField:     jsonName,
					JSONField:      jsonName,
					GoField:        goFieldName,
					PyField:        jsonName,
					FieldType:      "oneof",
					IsOneof:        true,
					OneofName:      oneofName,
					OneofCases:     oneofCases,
					OneofCaseKinds: oneofCaseKinds,
				}
				spec.Order = append(spec.Order, jsonName)
			}
			continue
		}

		// 2. Check for custom jev.v1.field extension options
		var customInstructions string
		var skipField bool
		var jevOpt *jevv1.FieldOptions
		if proto.HasExtension(desc.Options(), jevv1.E_Field) {
			ext := proto.GetExtension(desc.Options(), jevv1.E_Field)
			if opt, ok := ext.(*jevv1.FieldOptions); ok && opt != nil {
				jevOpt = opt
				skipField = opt.GetSkip()
				customInstructions = opt.GetInstructions()
			}
		}

		if skipField {
			continue
		}

		// 3. Strict type checking: fail fast on unsupported Protobuf structures when annotated for Jev
		if jevOpt != nil {
			if desc.IsList() {
				return spec, fmt.Errorf("field %q: repeated fields are not yet supported by protoc-gen-jev", desc.FullName())
			}
			if desc.IsMap() {
				return spec, fmt.Errorf("field %q: map fields are not yet supported by protoc-gen-jev", desc.FullName())
			}
			if desc.Kind() == protoreflect.MessageKind {
				return spec, fmt.Errorf("field %q: nested message fields are not yet supported by protoc-gen-jev", desc.FullName())
			}
			if desc.Kind() == protoreflect.BytesKind {
				return spec, fmt.Errorf("field %q: bytes fields are not yet supported by protoc-gen-jev", desc.FullName())
			}
		} else {
			// Unannotated complex fields are ignored as non-decision fields
			if desc.IsList() || desc.IsMap() || desc.Kind() == protoreflect.MessageKind || desc.Kind() == protoreflect.BytesKind {
				continue
			}
		}

		// 4. Resolve instructions: custom option takes priority, then doc comment, then fallback
		instructions := CleanComments(field.Comments.Leading.String())
		if customInstructions != "" {
			instructions = customInstructions
		} else if instructions == "" {
			instructions = fmt.Sprintf("Evaluate %s", desc.JSONName())
		}

		goFieldName := ToPascalCase(desc.JSONName())

		// 5. Explicit primitive overrides or natural Protobuf kind mapping
		var choiceRule *jevv1.ChoiceRules
		var scoreRule *jevv1.ScoreRules
		var noulRule *jevv1.NoulRules
		if jevOpt != nil {
			choiceRule = jevOpt.GetChoice()
			scoreRule = jevOpt.GetScore()
			noulRule = jevOpt.GetNoul()
		}

		// Sanity Check: Ensure no conflicting primitive configurations
		primitiveCount := 0
		if choiceRule != nil {
			primitiveCount++
		}
		if scoreRule != nil {
			primitiveCount++
		}
		if noulRule != nil {
			primitiveCount++
		}
		if primitiveCount > 1 {
			return spec, fmt.Errorf("field %q: cannot specify more than one Jev primitive (choice, score, noul)", desc.FullName())
		}

		if choiceRule != nil {
			criteria := resolveExplicitChoiceCriteria(desc, choiceRule)
			if len(criteria) == 0 {
				return spec, fmt.Errorf("field %q: choice question has no valid choices", desc.FullName())
			}
			fieldType := "string"
			isEnum := desc.Kind() == protoreflect.EnumKind
			enumName := ""
			if isEnum {
				fieldType = "enum"
				enumName = string(desc.Enum().Name())
			}
			spec.Questions[desc.JSONName()] = model.Question{
				Type:         model.TypeChoice,
				Instructions: instructions,
				Criteria:     criteria,
				ProtoField:   string(desc.Name()),
				JSONField:    desc.JSONName(),
				GoField:      goFieldName,
				PyField:      string(desc.Name()),
				FieldType:    fieldType,
				IsEnum:       isEnum,
				EnumTypeName: enumName,
			}
			spec.Order = append(spec.Order, desc.JSONName())
			continue
		}

		if scoreRule != nil {
			levels, criteria, err := resolveScoreLevels(desc, scoreRule)
			if err != nil {
				return spec, err
			}
			fieldType := resolveNumberFieldType(desc.Kind())
			spec.Questions[desc.JSONName()] = model.Question{
				Type:         model.TypeScore,
				Instructions: instructions,
				Criteria:     criteria,
				ScoreLevels:  levels,
				ProtoField:   string(desc.Name()),
				JSONField:    desc.JSONName(),
				GoField:      goFieldName,
				PyField:      string(desc.Name()),
				FieldType:    fieldType,
			}
			spec.Order = append(spec.Order, desc.JSONName())
			continue
		}

		if noulRule != nil || desc.Kind() == protoreflect.BoolKind {
			threshold := 0.5
			if noulRule != nil && noulRule.Threshold > 0 {
				threshold = float64(noulRule.Threshold)
			}
			spec.Questions[desc.JSONName()] = model.Question{
				Type:         model.TypeNoul,
				Instructions: instructions,
				Threshold:    threshold,
				ProtoField:   string(desc.Name()),
				JSONField:    desc.JSONName(),
				GoField:      goFieldName,
				PyField:      string(desc.Name()),
				FieldType:    "bool",
			}
			spec.Order = append(spec.Order, desc.JSONName())
			continue
		}

		// 6. Default mapping based on Protobuf kind
		switch desc.Kind() {
		case protoreflect.EnumKind:
			criteria := resolveEnumCriteria(desc.Enum(), nil)
			if len(criteria) > 0 {
				spec.Questions[desc.JSONName()] = model.Question{
					Type:         model.TypeChoice,
					Instructions: instructions,
					Criteria:     criteria,
					ProtoField:   string(desc.Name()),
					JSONField:    desc.JSONName(),
					GoField:      goFieldName,
					PyField:      string(desc.Name()),
					FieldType:    "enum",
					IsEnum:       true,
					EnumTypeName: string(desc.Enum().Name()),
				}
				spec.Order = append(spec.Order, desc.JSONName())
			}

		case protoreflect.FloatKind, protoreflect.DoubleKind:
			levels, criteria, err := resolveScoreLevels(desc, nil)
			if err != nil {
				return spec, err
			}
			fType := "float64"
			if desc.Kind() == protoreflect.FloatKind {
				fType = "float32"
			}
			spec.Questions[desc.JSONName()] = model.Question{
				Type:         model.TypeScore,
				Instructions: instructions,
				Criteria:     criteria,
				ScoreLevels:  levels,
				ProtoField:   string(desc.Name()),
				JSONField:    desc.JSONName(),
				GoField:      goFieldName,
				PyField:      string(desc.Name()),
				FieldType:    fType,
			}
			spec.Order = append(spec.Order, desc.JSONName())

		case protoreflect.Int32Kind, protoreflect.Int64Kind,
			protoreflect.Sint32Kind, protoreflect.Sint64Kind,
			protoreflect.Uint32Kind, protoreflect.Uint64Kind:
			levels, criteria, err := resolveScoreLevels(desc, nil)
			if err != nil {
				return spec, err
			}
			iType := "int32"
			if desc.Kind() == protoreflect.Int64Kind || desc.Kind() == protoreflect.Sint64Kind || desc.Kind() == protoreflect.Uint64Kind {
				iType = "int64"
			}
			spec.Questions[desc.JSONName()] = model.Question{
				Type:         model.TypeScore,
				Instructions: instructions,
				Criteria:     criteria,
				ScoreLevels:  levels,
				ProtoField:   string(desc.Name()),
				JSONField:    desc.JSONName(),
				GoField:      goFieldName,
				PyField:      string(desc.Name()),
				FieldType:    iType,
			}
			spec.Order = append(spec.Order, desc.JSONName())
		case protoreflect.StringKind:
			// Plain string without choice rule is not a decision question; ignore
		default:
			return spec, fmt.Errorf("field %q: unsupported Protobuf kind %v", desc.FullName(), desc.Kind())
		}
	}

	return spec, nil
}

func resolveNumberFieldType(k protoreflect.Kind) string {
	switch k {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	default:
		return "float64"
	}
}

func isIntegerKind(k protoreflect.Kind) bool {
	switch k {
	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		return true
	default:
		return false
	}
}

// resolveEnumCriteria respects choice.choices, choice.not_in, and choice.criteria options, preserving descriptions.
func resolveEnumCriteria(enumDesc protoreflect.EnumDescriptor, rule *jevv1.ChoiceRules) map[string]any {
	var choicesMap, notInMap map[string]bool
	if rule != nil {
		if len(rule.GetChoices()) > 0 {
			choicesMap = make(map[string]bool)
			for _, v := range rule.GetChoices() {
				choicesMap[v] = true
			}
		}
		if len(rule.GetNotIn()) > 0 {
			notInMap = make(map[string]bool)
			for _, v := range rule.GetNotIn() {
				notInMap[v] = true
			}
		}
	}

	criteria := make(map[string]any)
	for i := 0; i < enumDesc.Values().Len(); i++ {
		val := enumDesc.Values().Get(i)
		valNum := int32(val.Number())
		valName := string(val.Name())

		if valNum == 0 || strings.HasSuffix(valName, "_UNSPECIFIED") {
			continue
		}
		if len(choicesMap) > 0 && !choicesMap[valName] {
			continue
		}
		if len(notInMap) > 0 && notInMap[valName] {
			continue
		}
		criteria[valName] = ""
	}

	// Jev options criteria overrides or supplements (preserve descriptions)
	if rule != nil && len(rule.GetCriteria()) > 0 {
		for k, v := range rule.GetCriteria() {
			criteria[k] = v
		}
	}

	return criteria
}

// resolveStringCriteria maps string fields to Choice when discrete values or criteria are configured, preserving descriptions.
func resolveStringCriteria(desc protoreflect.FieldDescriptor, rule *jevv1.ChoiceRules) map[string]any {
	criteria := make(map[string]any)
	if rule == nil {
		return criteria
	}

	for _, val := range rule.GetChoices() {
		criteria[val] = ""
	}
	for k, v := range rule.GetCriteria() {
		criteria[k] = v
	}

	return criteria
}

// resolveExplicitChoiceCriteria handles fields explicitly configured with choice: { ... }.
func resolveExplicitChoiceCriteria(desc protoreflect.FieldDescriptor, rule *jevv1.ChoiceRules) map[string]any {
	if desc.Kind() == protoreflect.EnumKind {
		return resolveEnumCriteria(desc.Enum(), rule)
	}
	return resolveStringCriteria(desc, rule)
}

// resolveScoreLevels extracts explicit score levels or provides defaults if rule is nil.
func resolveScoreLevels(desc protoreflect.FieldDescriptor, rule *jevv1.ScoreRules) ([]model.ScoreLevel, []string, error) {
	if rule != nil {
		if len(rule.GetLevels()) < 2 {
			fName := "unknown"
			if desc != nil {
				fName = string(desc.FullName())
			}
			return nil, nil, fmt.Errorf("field %q: score question must specify at least 2 levels", fName)
		}
		var levels []model.ScoreLevel
		var criteria []string
		for _, l := range rule.GetLevels() {
			levels = append(levels, model.ScoreLevel{
				Value:       l.GetValue(),
				Description: l.GetDescription(),
			})
			descText := l.GetDescription()
			if descText == "" {
				descText = formatFloat(l.GetValue())
			}
			criteria = append(criteria, descText)
		}
		return levels, criteria, nil
	}

	// Default fallback when field has no explicit score rules
	var levels []model.ScoreLevel
	var criteria []string
	if isIntegerKind(desc.Kind()) {
		for v := int64(1); v <= 5; v++ {
			s := strconv.FormatInt(v, 10)
			levels = append(levels, model.ScoreLevel{Value: float64(v), Description: s})
			criteria = append(criteria, s)
		}
	} else {
		for _, v := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
			s := formatFloat(v)
			levels = append(levels, model.ScoreLevel{Value: v, Description: s})
			criteria = append(criteria, s)
		}
	}
	return levels, criteria, nil
}

func formatFloat(v float64) string {
	// Round to 4 decimal places to prevent floating-point inaccuracies
	v = math.Round(v*10000) / 10000
	str := strconv.FormatFloat(v, 'f', -1, 64)
	if !strings.Contains(str, ".") {
		str += ".0"
	}
	return str
}

// CleanComments strips slashes, spaces, and linebreaks from Protobuf comments.
func CleanComments(s string) string {
	lines := strings.Split(s, "\n")
	var cleaned []string
	for _, l := range lines {
		l = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "//"))
		if l != "" {
			cleaned = append(cleaned, l)
		}
	}
	return strings.Join(cleaned, " ")
}

// ToPascalCase converts camelCase or snake_case to PascalCase for Go identifiers.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	var result string
	for _, p := range parts {
		if len(p) > 0 {
			result += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return result
}

// ProcessService parses a Protobuf service into a Jev ServiceSpec.
func ProcessService(svc *protogen.Service) (model.ServiceSpec, error) {
	fileBase := strings.TrimSuffix(filepath.Base(string(svc.Desc.ParentFile().Path())), ".proto")
	spec := model.ServiceSpec{
		ServiceName: string(svc.Desc.Name()),
		Package:     string(svc.Desc.ParentFile().Package()),
		FileBase:    fileBase,
	}

	// Check service-level options
	var serviceExplicitEnabled *bool
	if proto.HasExtension(svc.Desc.Options(), jevv1.E_Service) {
		if opt, ok := proto.GetExtension(svc.Desc.Options(), jevv1.E_Service).(*jevv1.ServiceOptions); ok && opt != nil && opt.Enabled != nil {
			if !*opt.Enabled {
				// Explicitly disabled at service level
				return spec, nil
			}
			serviceExplicitEnabled = opt.Enabled
		}
	}

	for _, method := range svc.Methods {
		var methodExplicitEnabled *bool
		if proto.HasExtension(method.Desc.Options(), jevv1.E_Method) {
			if opt, ok := proto.GetExtension(method.Desc.Options(), jevv1.E_Method).(*jevv1.MethodOptions); ok && opt != nil && opt.Enabled != nil {
				if !*opt.Enabled {
					// Explicitly disabled at method level
					continue
				}
				methodExplicitEnabled = opt.Enabled
			}
		}

		isJevMethod := false
		if methodExplicitEnabled != nil {
			isJevMethod = *methodExplicitEnabled
		} else if serviceExplicitEnabled != nil {
			isJevMethod = *serviceExplicitEnabled
		} else {
			// Auto-discovery: check if either request or response message has Jev field options
			if HasJevFieldOptions(method.Input) || HasJevFieldOptions(method.Output) {
				isJevMethod = true
			}
		}

		if !isJevMethod {
			continue
		}

		// Sanity Check 1: Request must have at least one input field
		if len(method.Input.Fields) == 0 {
			return spec, fmt.Errorf("method %q: request message %q must have at least one input field to evaluate", method.Desc.FullName(), method.Input.Desc.Name())
		}

		questionSpec, err := ProcessMessage(method.Output)
		if err != nil {
			return spec, fmt.Errorf("method %q: %w", method.Desc.FullName(), err)
		}

		// Sanity Check 2: Response must have at least one decision question
		if len(questionSpec.Questions) == 0 {
			return spec, fmt.Errorf("method %q: response message %q must contain at least one decision field (Choice, Score, or Noul)", method.Desc.FullName(), method.Output.Desc.Name())
		}

		var inputFields []model.FieldSpec
		for _, f := range method.Input.Fields {
			if f.Desc.IsList() {
				return spec, fmt.Errorf("method %q: input field %q: repeated fields are not yet supported by protoc-gen-jev", method.Desc.FullName(), f.Desc.FullName())
			}
			if f.Desc.IsMap() {
				return spec, fmt.Errorf("method %q: input field %q: map fields are not yet supported by protoc-gen-jev", method.Desc.FullName(), f.Desc.FullName())
			}
			if f.Desc.Kind() == protoreflect.MessageKind {
				return spec, fmt.Errorf("method %q: input field %q: nested message fields are not yet supported by protoc-gen-jev", method.Desc.FullName(), f.Desc.FullName())
			}
			if f.Desc.Kind() == protoreflect.BytesKind {
				return spec, fmt.Errorf("method %q: input field %q: bytes fields are not yet supported by protoc-gen-jev", method.Desc.FullName(), f.Desc.FullName())
			}

			var t string
			switch f.Desc.Kind() {
			case protoreflect.BoolKind:
				t = "bool"
			case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
				t = "int32"
			case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
				t = "int64"
			case protoreflect.FloatKind:
				t = "float32"
			case protoreflect.DoubleKind:
				t = "float64"
			case protoreflect.StringKind:
				t = "string"
			case protoreflect.EnumKind:
				t = "enum"
			default:
				return spec, fmt.Errorf("method %q: input field %q: unsupported type %v", method.Desc.FullName(), f.Desc.FullName(), f.Desc.Kind())
			}
			jsonName := string(f.Desc.Name())
			inputFields = append(inputFields, model.FieldSpec{
				Name:     string(f.Desc.Name()),
				JSONName: jsonName,
				GoName:   ToPascalCase(jsonName),
				Type:     t,
			})
		}

		mSpec := model.MethodSpec{
			Name:         string(method.Desc.Name()),
			InputType:    string(method.Input.Desc.Name()),
			InputFields:  inputFields,
			OutputType:   string(method.Output.Desc.Name()),
			QuestionSpec: questionSpec,
		}
		spec.Methods = append(spec.Methods, mSpec)
	}

	return spec, nil
}

// HasJevFieldOptions returns true if any field in msg has jev.v1.field extension.
func HasJevFieldOptions(msg *protogen.Message) bool {
	if msg == nil {
		return false
	}
	for _, f := range msg.Fields {
		if proto.HasExtension(f.Desc.Options(), jevv1.E_Field) {
			return true
		}
	}
	return false
}
