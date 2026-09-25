package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jevv1 "github.com/sudorandom/protoc-gen-jev/pkg/jev/v1"
)

func TestResolveIntCriteria(t *testing.T) {
	tests := []struct {
		name     string
		opt      *jevv1.FieldOptions
		expected []string
	}{
		{
			name:     "default fallback without options",
			opt:      nil,
			expected: []string{"1", "2", "3", "4", "5"},
		},
		{
			name: "small range exact sequence (1..5)",
			opt: &jevv1.FieldOptions{
				Min: 1,
				Max: 5,
			},
			expected: []string{"1", "2", "3", "4", "5"},
		},
		{
			name: "explicit discrete scale takes priority",
			opt: &jevv1.FieldOptions{
				Scale: []float32{10, 20, 50, 100},
				Min:   1,
				Max:   100,
			},
			expected: []string{"10", "20", "50", "100"},
		},
		{
			name: "large integer range 5-tier interpolated rubric (0..1000)",
			opt: &jevv1.FieldOptions{
				Min: 0,
				Max: 1000,
			},
			expected: []string{"0", "250", "500", "750", "1000"},
		},
		{
			name: "custom Jev options min/max override (10..50)",
			opt: &jevv1.FieldOptions{
				Min: 10,
				Max: 50,
			},
			expected: []string{"10", "20", "30", "40", "50"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveIntCriteria(nil, tt.opt)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestResolveFloatCriteria(t *testing.T) {
	tests := []struct {
		name     string
		opt      *jevv1.FieldOptions
		expected []string
	}{
		{
			name:     "default fallback continuous rubric (0.0..1.0)",
			opt:      nil,
			expected: []string{"0.0", "0.25", "0.5", "0.75", "1.0"},
		},
		{
			name: "continuous range interpolation (-40.0..60.0)",
			opt: &jevv1.FieldOptions{
				Min: -40.0,
				Max: 60.0,
			},
			expected: []string{"-40.0", "-15.0", "10.0", "35.0", "60.0"},
		},
		{
			name: "discrete float scale takes priority",
			opt: &jevv1.FieldOptions{
				Scale: []float32{0.2, 0.5, 0.8},
				Min:   0.0,
				Max:   1.0,
			},
			expected: []string{"0.2", "0.5", "0.8"},
		},
		{
			name: "custom Jev options min/max bounds (0.0..50.0)",
			opt: &jevv1.FieldOptions{
				Min: 0.0,
				Max: 50.0,
			},
			expected: []string{"0.0", "12.5", "25.0", "37.5", "50.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveFloatCriteria(nil, tt.opt)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestResolveStringCriteria(t *testing.T) {
	tests := []struct {
		name     string
		opt      *jevv1.FieldOptions
		expected map[string]any
	}{
		{
			name:     "empty without options",
			opt:      nil,
			expected: map[string]any{},
		},
		{
			name: "jev choices whitelist",
			opt: &jevv1.FieldOptions{
				Choices: []string{"PUBLIC", "INTERNAL_CONFIDENTIAL", "RESTRICTED_PII"},
			},
			expected: map[string]any{
				"PUBLIC":                nil,
				"INTERNAL_CONFIDENTIAL": nil,
				"RESTRICTED_PII":        nil,
			},
		},
		{
			name: "custom jev criteria options",
			opt: &jevv1.FieldOptions{
				Criteria: map[string]string{
					"APPROVE": "Request approved",
					"REJECT":  "Request rejected",
				},
			},
			expected: map[string]any{
				"APPROVE": nil,
				"REJECT":  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveStringCriteria(nil, tt.opt)
			assert.Equal(t, tt.expected, got)
		})
	}
}
