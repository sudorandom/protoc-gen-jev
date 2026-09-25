package model

// QuestionType represents Jev's primitive question types.
type QuestionType string

const (
	TypeChoice QuestionType = "Choice"
	TypeNoul   QuestionType = "Noul"
	TypeScore  QuestionType = "Score"
)

// Question represents a single Jev decision question mapped from a Protobuf field.
type Question struct {
	Type         QuestionType   `json:"type"`
	Instructions string         `json:"instructions"`
	Criteria     any            `json:"criteria,omitempty"`
	ProtoField   string         `json:"proto_field"`
	JSONField    string         `json:"json_field"`
	GoField      string         `json:"go_field"`
}

// MessageSpec represents all Jev questions extracted from a single Protobuf message.
type MessageSpec struct {
	MessageName string              `json:"message_name"`
	Package     string              `json:"package"`
	Questions   map[string]Question `json:"questions"`
	Order       []string            `json:"order,omitempty"`
}

// TargetOptions controls which code generation targets to produce.
type TargetOptions struct {
	GenerateJSON       bool
	GenerateGo         bool
	GenerateTypeScript bool
	GeneratePython     bool
}

// ParseTargets parses target flags like "go+ts+python+json" or "all".
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

	for _, t := range splitTargets(targetOpt) {
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

func splitTargets(s string) []string {
	var res []string
	for _, part := range []string{s} {
		for _, sub := range []string{"+", ","} {
			part = replaceAll(part, sub, " ")
		}
		for _, token := range fields(part) {
			if token != "" {
				res = append(res, token)
			}
		}
	}
	return res
}

func replaceAll(s, old, newStr string) string {
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			return s
		}
		s = s[:idx] + newStr + s[idx+len(old):]
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func fields(s string) []string {
	var res []string
	var cur []byte
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' || s[i] == '\n' {
			if len(cur) > 0 {
				res = append(res, string(cur))
				cur = nil
			}
		} else {
			cur = append(cur, s[i])
		}
	}
	if len(cur) > 0 {
		res = append(res, string(cur))
	}
	return res
}
