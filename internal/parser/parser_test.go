package parser_test

import (
	"testing"

	"protoc-gen-jev/internal/model"
	"protoc-gen-jev/internal/parser"
)

func TestCleanComments(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "// Single line comment",
			expected: "Single line comment",
		},
		{
			input:    "// First line\n// Second line",
			expected: "First line Second line",
		},
		{
			input:    "   //   Padded line   ",
			expected: "Padded line",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		got := parser.CleanComments(tt.input)
		if got != tt.expected {
			t.Errorf("CleanComments(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"confidence_score", "ConfidenceScore"},
		{"status", "Status"},
		{"is_enabled", "IsEnabled"},
		{"requiresImmediateAction", "RequiresImmediateAction"},
		{"", ""},
	}

	for _, tt := range tests {
		got := parser.ToPascalCase(tt.input)
		if got != tt.expected {
			t.Errorf("ToPascalCase(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseTargets(t *testing.T) {
	allOpts := model.ParseTargets("all")
	if !allOpts.GenerateGo || !allOpts.GenerateTypeScript || !allOpts.GeneratePython || !allOpts.GenerateJSON {
		t.Errorf("ParseTargets('all') should enable all targets")
	}

	goOnly := model.ParseTargets("go")
	if !goOnly.GenerateGo || goOnly.GenerateTypeScript || goOnly.GeneratePython {
		t.Errorf("ParseTargets('go') should only enable Go (+ json)")
	}

	tsGo := model.ParseTargets("ts+go")
	if !tsGo.GenerateGo || !tsGo.GenerateTypeScript || tsGo.GeneratePython {
		t.Errorf("ParseTargets('ts+go') should enable TS and Go")
	}
}
