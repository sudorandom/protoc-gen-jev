package model

import "strings"

// QuestionType represents Jev's primitive question types.
type QuestionType string

const (
	TypeChoice QuestionType = "Choice"
	TypeNoul   QuestionType = "Noul"
	TypeScore  QuestionType = "Score"
)

// Question represents a single Jev decision question mapped from a Protobuf field.
type Question struct {
	Type         QuestionType `json:"type"`
	Instructions string       `json:"instructions"`
	Criteria     any          `json:"criteria,omitempty"`
	ProtoField   string       `json:"proto_field"`
	JSONField    string       `json:"json_field"`
	GoField      string       `json:"go_field"`
}

// MessageSpec represents all Jev questions extracted from a single Protobuf message.
type MessageSpec struct {
	MessageName string              `json:"message_name"`
	Package     string              `json:"package"`
	Questions   map[string]Question `json:"questions"`
	Order       []string            `json:"order,omitempty"`
}

// FieldSpec describes a field in an input request message.
type FieldSpec struct {
	Name     string `json:"name"`
	JSONName string `json:"json_name"`
	GoName   string `json:"go_name"`
	Type     string `json:"type"`
}

// MethodSpec represents an RPC method on a Service.
type MethodSpec struct {
	Name         string      `json:"name"`
	InputType    string      `json:"input_type"`
	InputFields  []FieldSpec `json:"input_fields"`
	OutputType   string      `json:"output_type"`
	QuestionSpec MessageSpec `json:"question_spec"`
}

// ServiceSpec represents a Protobuf service containing Jev RPC methods.
type ServiceSpec struct {
	ServiceName string       `json:"service_name"`
	Package     string       `json:"package"`
	Methods     []MethodSpec `json:"methods"`
}

// TargetOptions controls which code generation targets to produce.
type TargetOptions struct {
	GenerateJSON       bool
	GenerateGo         bool
	GenerateTypeScript bool
	GeneratePython     bool
}

// ParseTargets parses comma-separated target flags like "go,ts,python,json" or "all".
func ParseTargets(targetOpt string) TargetOptions {
	if targetOpt == "all" || targetOpt == "" {
		return TargetOptions{
			GenerateJSON:       true,
			GenerateGo:         true,
			GenerateTypeScript: true,
			GeneratePython:     true,
		}
	}

	var opts TargetOptions
	opts.GenerateJSON = true // always generate declarative JSON spec

	for _, t := range strings.Split(targetOpt, ",") {
		t = strings.TrimSpace(t)
		switch t {
		case "go":
			opts.GenerateGo = true
		case "ts", "typescript":
			opts.GenerateTypeScript = true
		case "py", "python":
			opts.GeneratePython = true
		case "json":
			opts.GenerateJSON = true
		case "all":
			opts.GenerateGo = true
			opts.GenerateTypeScript = true
			opts.GeneratePython = true
			opts.GenerateJSON = true
		}
	}
	return opts
}
