package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sudorandom/protoc-gen-jev/internal/model"
	"github.com/sudorandom/protoc-gen-jev/internal/parser"
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
		assert.Equal(t, tt.expected, got, "CleanComments(%q)", tt.input)
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
		assert.Equal(t, tt.expected, got, "ToPascalCase(%q)", tt.input)
	}
}

func TestParseTargets(t *testing.T) {
	allOpts := model.ParseTargets("all")
	require.True(t, allOpts.GenerateGo && allOpts.GenerateTypeScript && allOpts.GeneratePython && allOpts.GenerateJSON, "ParseTargets('all') should enable all targets")

	goOnly := model.ParseTargets("go")
	assert.True(t, goOnly.GenerateGo, "ParseTargets('go') should enable Go")
	assert.True(t, goOnly.GenerateJSON, "ParseTargets('go') should enable JSON")
	assert.False(t, goOnly.GenerateTypeScript, "ParseTargets('go') should not enable TypeScript")
	assert.False(t, goOnly.GeneratePython, "ParseTargets('go') should not enable Python")

	tsGo := model.ParseTargets("ts,go")
	assert.True(t, tsGo.GenerateGo, "ParseTargets('ts,go') should enable Go")
	assert.True(t, tsGo.GenerateTypeScript, "ParseTargets('ts,go') should enable TypeScript")
	assert.True(t, tsGo.GenerateJSON, "ParseTargets('ts,go') should enable JSON")
	assert.False(t, tsGo.GeneratePython, "ParseTargets('ts,go') should not enable Python")
}
