package codegen

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/sudorandom/protoc-gen-jev/internal/model"
	"github.com/sudorandom/protoc-gen-jev/pkg/jev"
)

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func runtimeRules(spec model.MessageSpec) []jev.Question {
	rules := make([]jev.Question, 0, len(spec.Order))
	for _, name := range spec.Order {
		q := spec.Questions[name]
		rule := jev.Question{Name: name, Type: strings.ToLower(string(q.Type)), Instructions: q.Instructions, Field: q.ProtoField, Kind: q.FieldType, Threshold: q.Threshold, Oneof: q.OneofCaseKinds}
		if criteria, ok := q.Criteria.(map[string]any); ok {
			rule.Choices = map[string]string{}
			for k, v := range criteria {
				rule.Choices[k] = v.(string)
			}
		}
		descriptions, _ := q.Criteria.([]string)
		for i, l := range q.ScoreLevels {
			rule.Levels = append(rule.Levels, jev.Level{Value: l.Value, Description: descriptions[i]})
		}
		rules = append(rules, rule)
	}
	return rules
}

func rulesJSON(spec model.MessageSpec) string {
	data, err := json.Marshal(runtimeRules(spec))
	if err != nil {
		panic(err)
	}
	return string(data)
}

func messagePath(msg *protogen.Message) []string {
	var parts []string
	for d := protoreflect.Descriptor(msg.Desc); d != nil; d = d.Parent() {
		if _, ok := d.(protoreflect.MessageDescriptor); ok {
			parts = append([]string{string(d.Name())}, parts...)
		}
	}
	return parts
}

func relativeModule(gen *protogen.Plugin, file *protogen.File, msg *protogen.Message) string {
	target := gen.FilesByPath[msg.Desc.ParentFile().Path()].GeneratedFilenamePrefix + "_pb.js"
	rel, err := filepath.Rel(filepath.Dir(file.GeneratedFilenamePrefix), target)
	if err != nil {
		panic(err)
	}
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, ".") {
		rel = "./" + rel
	}
	return rel
}

func messages(specs []model.MessageSpec, services []model.ServiceSpec) []*protogen.Message {
	seen := map[protoreflect.FullName]bool{}
	var out []*protogen.Message
	add := func(m *protogen.Message) {
		if !seen[m.Desc.FullName()] {
			seen[m.Desc.FullName()] = true
			out = append(out, m)
		}
	}
	for _, s := range specs {
		add(s.Ref)
	}
	for _, s := range services {
		for _, m := range s.Methods {
			add(m.Input)
			add(m.QuestionSpec.Ref)
		}
	}
	return out
}
