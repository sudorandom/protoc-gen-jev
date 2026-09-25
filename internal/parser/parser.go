package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"protoc-gen-jev/internal/model"
	jevv1 "protoc-gen-jev/pkg/jev/v1"
)

// ProcessMessage parses a single Protobuf message into a Jev MessageSpec.
func ProcessMessage(msg *protogen.Message) model.MessageSpec {
	spec := model.MessageSpec{
		MessageName: string(msg.Desc.Name()),
		Package:     string(msg.Desc.ParentFile().Package()),
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
				for _, f := range oneof.Fields {
					criteria[string(f.Desc.Name())] = nil
				}

				jsonName := string(oneof.Desc.Name())
				goFieldName := ToPascalCase(jsonName)
				spec.Questions[jsonName] = model.Question{
					Type:         model.TypeChoice,
					Instructions: instructions,
					Criteria:     criteria,
					ProtoField:   jsonName,
					JSONField:    jsonName,
					GoField:      goFieldName,
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

		// 3. Check for protovalidate field rules
		var fieldRule *validate.FieldRules
		if proto.HasExtension(desc.Options(), validate.E_Field) {
			if rule, ok := proto.GetExtension(desc.Options(), validate.E_Field).(*validate.FieldRules); ok && rule != nil {
				fieldRule = rule
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

		// 5. Map Protobuf kind and protovalidate rules to Jev primitive
		switch desc.Kind() {
		case protoreflect.EnumKind:
			criteria := resolveEnumCriteria(desc.Enum(), fieldRule, jevOpt)
			if len(criteria) > 0 {
				spec.Questions[desc.JSONName()] = model.Question{
					Type:         model.TypeChoice,
					Instructions: instructions,
					Criteria:     criteria,
					ProtoField:   string(desc.Name()),
					JSONField:    desc.JSONName(),
					GoField:      goFieldName,
				}
				spec.Order = append(spec.Order, desc.JSONName())
			}

		case protoreflect.StringKind:
			criteria := resolveStringCriteria(desc, fieldRule, jevOpt)
			if len(criteria) > 0 {
				spec.Questions[desc.JSONName()] = model.Question{
					Type:         model.TypeChoice,
					Instructions: instructions,
					Criteria:     criteria,
					ProtoField:   string(desc.Name()),
					JSONField:    desc.JSONName(),
					GoField:      goFieldName,
				}
				spec.Order = append(spec.Order, desc.JSONName())
			}

		case protoreflect.BoolKind:
			spec.Questions[desc.JSONName()] = model.Question{
				Type:         model.TypeNoul,
				Instructions: instructions,
				ProtoField:   string(desc.Name()),
				JSONField:    desc.JSONName(),
				GoField:      goFieldName,
			}
			spec.Order = append(spec.Order, desc.JSONName())

		case protoreflect.FloatKind, protoreflect.DoubleKind:
			criteria := resolveFloatCriteria(desc, fieldRule, jevOpt)
			spec.Questions[desc.JSONName()] = model.Question{
				Type:         model.TypeScore,
				Instructions: instructions,
				Criteria:     criteria,
				ProtoField:   string(desc.Name()),
				JSONField:    desc.JSONName(),
				GoField:      goFieldName,
			}
			spec.Order = append(spec.Order, desc.JSONName())

		case protoreflect.Int32Kind, protoreflect.Int64Kind,
			protoreflect.Sint32Kind, protoreflect.Sint64Kind,
			protoreflect.Uint32Kind, protoreflect.Uint64Kind:
			criteria := resolveIntCriteria(desc, fieldRule, jevOpt)
			spec.Questions[desc.JSONName()] = model.Question{
				Type:         model.TypeScore,
				Instructions: instructions,
				Criteria:     criteria,
				ProtoField:   string(desc.Name()),
				JSONField:    desc.JSONName(),
				GoField:      goFieldName,
			}
			spec.Order = append(spec.Order, desc.JSONName())
		}
	}

	return spec
}

// resolveEnumCriteria respects protovalidate enum.in, enum.not_in, and custom criteria.
func resolveEnumCriteria(enumDesc protoreflect.EnumDescriptor, rule *validate.FieldRules, opt *jevv1.FieldOptions) map[string]any {
	var inMap, notInMap map[int32]bool
	if rule != nil && rule.GetEnum() != nil {
		er := rule.GetEnum()
		if len(er.In) > 0 {
			inMap = make(map[int32]bool)
			for _, v := range er.In {
				inMap[v] = true
			}
		}
		if len(er.NotIn) > 0 {
			notInMap = make(map[int32]bool)
			for _, v := range er.NotIn {
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
		if len(inMap) > 0 && !inMap[valNum] {
			continue
		}
		if len(notInMap) > 0 && notInMap[valNum] {
			continue
		}
		criteria[valName] = nil
	}

	// Jev options criteria overrides or supplements
	if opt != nil && len(opt.GetCriteria()) > 0 {
		for k := range opt.GetCriteria() {
			criteria[k] = nil
		}
	}

	return criteria
}

// resolveStringCriteria maps string fields to Choice when discrete values are enforced.
func resolveStringCriteria(desc protoreflect.FieldDescriptor, rule *validate.FieldRules, opt *jevv1.FieldOptions) map[string]any {
	criteria := make(map[string]any)

	// 1. protovalidate string.in
	if rule != nil && rule.GetString_() != nil {
		sr := rule.GetString_()
		for _, val := range sr.In {
			criteria[val] = nil
		}
	}

	// 2. Jev options criteria
	if opt != nil && len(opt.GetCriteria()) > 0 {
		for k := range opt.GetCriteria() {
			criteria[k] = nil
		}
	}

	return criteria
}

// resolveIntCriteria calculates dynamic rubrics from int.in, int.gte/lte, and jev.v1.field min/max.
func resolveIntCriteria(desc protoreflect.FieldDescriptor, rule *validate.FieldRules, opt *jevv1.FieldOptions) []string {
	var inValues []int64
	var hasMin, hasMax bool
	var minVal, maxVal int64

	// 1. protovalidate rules
	if rule != nil {
		if r := rule.GetInt32(); r != nil {
			for _, v := range r.In {
				inValues = append(inValues, int64(v))
			}
			if r.HasGte() {
				hasMin = true
				minVal = int64(r.GetGte())
			} else if r.HasGt() {
				hasMin = true
				minVal = int64(r.GetGt()) + 1
			}
			if r.HasLte() {
				hasMax = true
				maxVal = int64(r.GetLte())
			} else if r.HasLt() {
				hasMax = true
				maxVal = int64(r.GetLt()) - 1
			}
		} else if r := rule.GetInt64(); r != nil {
			for _, v := range r.In {
				inValues = append(inValues, v)
			}
			if r.HasGte() {
				hasMin = true
				minVal = r.GetGte()
			} else if r.HasGt() {
				hasMin = true
				minVal = r.GetGt() + 1
			}
			if r.HasLte() {
				hasMax = true
				maxVal = r.GetLte()
			} else if r.HasLt() {
				hasMax = true
				maxVal = r.GetLt() - 1
			}
		}
	}

	// 2. Custom Jev options min/max override
	if opt != nil && (opt.Min != 0 || opt.Max != 0) {
		hasMin = true
		hasMax = true
		minVal = int64(opt.Min)
		maxVal = int64(opt.Max)
	}

	// 3. Explicit in list
	if len(inValues) > 0 {
		var res []string
		for _, v := range inValues {
			res = append(res, strconv.FormatInt(v, 10))
		}
		return res
	}

	// 4. Bound interpolation
	if hasMin && hasMax && maxVal >= minVal {
		count := maxVal - minVal + 1
		if count <= 10 {
			var res []string
			for v := minVal; v <= maxVal; v++ {
				res = append(res, strconv.FormatInt(v, 10))
			}
			return res
		}
		step := float64(maxVal-minVal) / 4.0
		return []string{
			strconv.FormatInt(minVal, 10),
			strconv.FormatInt(minVal+int64(math.Round(step*1)), 10),
			strconv.FormatInt(minVal+int64(math.Round(step*2)), 10),
			strconv.FormatInt(minVal+int64(math.Round(step*3)), 10),
			strconv.FormatInt(maxVal, 10),
		}
	}

	return []string{"1", "2", "3", "4", "5"}
}

// resolveFloatCriteria calculates dynamic continuous rubrics from float.in, float.gte/lte, and jev.v1.field min/max.
func resolveFloatCriteria(desc protoreflect.FieldDescriptor, rule *validate.FieldRules, opt *jevv1.FieldOptions) []string {
	var inValues []float64
	var hasMin, hasMax bool
	var minVal, maxVal float64

	// 1. protovalidate rules
	if rule != nil {
		if r := rule.GetFloat(); r != nil {
			for _, v := range r.In {
				inValues = append(inValues, float64(v))
			}
			if r.HasGte() {
				hasMin = true
				minVal = float64(r.GetGte())
			} else if r.HasGt() {
				hasMin = true
				minVal = float64(r.GetGt())
			}
			if r.HasLte() {
				hasMax = true
				maxVal = float64(r.GetLte())
			} else if r.HasLt() {
				hasMax = true
				maxVal = float64(r.GetLt())
			}
		} else if r := rule.GetDouble(); r != nil {
			for _, v := range r.In {
				inValues = append(inValues, v)
			}
			if r.HasGte() {
				hasMin = true
				minVal = r.GetGte()
			} else if r.HasGt() {
				hasMin = true
				minVal = r.GetGt()
			}
			if r.HasLte() {
				hasMax = true
				maxVal = r.GetLte()
			} else if r.HasLt() {
				hasMax = true
				maxVal = r.GetLt()
			}
		}
	}

	// 2. Custom Jev options min/max override
	if opt != nil && (opt.Min != 0 || opt.Max != 0) {
		hasMin = true
		hasMax = true
		minVal = float64(opt.Min)
		maxVal = float64(opt.Max)
	}

	// 3. Explicit in list
	if len(inValues) > 0 {
		var res []string
		for _, v := range inValues {
			res = append(res, formatFloat(v))
		}
		return res
	}

	// 4. Bound interpolation
	if hasMin && hasMax && maxVal > minVal {
		step := (maxVal - minVal) / 4.0
		return []string{
			formatFloat(minVal),
			formatFloat(minVal + step*1),
			formatFloat(minVal + step*2),
			formatFloat(minVal + step*3),
			formatFloat(maxVal),
		}
	}

	return []string{"0.0", "0.25", "0.5", "0.75", "1.0"}
}

func formatFloat(v float64) string {
	s := strconv.FormatFloat(v, 'g', 6, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
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
