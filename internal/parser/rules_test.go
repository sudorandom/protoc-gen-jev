package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"

	jevv1 "github.com/sudorandom/protoc-gen-jev/pkg/jev/v1"
)

func TestResolveIntCriteria(t *testing.T) {
	tests := []struct {
		name     string
		rule     *validate.FieldRules
		opt      *jevv1.FieldOptions
		expected []string
	}{
		{
			name:     "default fallback without rules",
			rule:     nil,
			opt:      nil,
			expected: []string{"1", "2", "3", "4", "5"},
		},
		{
			name: "small range exact sequence (1..5)",
			rule: &validate.FieldRules{
				Type: &validate.FieldRules_Int32{
					Int32: &validate.Int32Rules{
						GreaterThan: &validate.Int32Rules_Gte{Gte: 1},
						LessThan:    &validate.Int32Rules_Lte{Lte: 5},
					},
				},
			},
			expected: []string{"1", "2", "3", "4", "5"},
		},
		{
			name: "strict bounds (gt 0, lt 4)",
			rule: &validate.FieldRules{
				Type: &validate.FieldRules_Int32{
					Int32: &validate.Int32Rules{
						GreaterThan: &validate.Int32Rules_Gt{Gt: 0},
						LessThan:    &validate.Int32Rules_Lt{Lt: 4},
					},
				},
			},
			expected: []string{"1", "2", "3"},
		},
		{
			name: "discrete allowed in list takes priority",
			rule: &validate.FieldRules{
				Type: &validate.FieldRules_Int32{
					Int32: &validate.Int32Rules{
						In:          []int32{10, 20, 50, 100},
						GreaterThan: &validate.Int32Rules_Gte{Gte: 1},
						LessThan:    &validate.Int32Rules_Lte{Lte: 100},
					},
				},
			},
			expected: []string{"10", "20", "50", "100"},
		},
		{
			name: "large int64 range 5-tier interpolated rubric (0..1000)",
			rule: &validate.FieldRules{
				Type: &validate.FieldRules_Int64{
					Int64: &validate.Int64Rules{
						GreaterThan: &validate.Int64Rules_Gte{Gte: 0},
						LessThan:    &validate.Int64Rules_Lte{Lte: 1000},
					},
				},
			},
			expected: []string{"0", "250", "500", "750", "1000"},
		},
		{
			name: "custom Jev options min/max override",
			rule: nil,
			opt: &jevv1.FieldOptions{
				Min: 10,
				Max: 50,
			},
			expected: []string{"10", "20", "30", "40", "50"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveIntCriteria(nil, tt.rule, tt.opt)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestResolveFloatCriteria(t *testing.T) {
	tests := []struct {
		name     string
		rule     *validate.FieldRules
		opt      *jevv1.FieldOptions
		expected []string
	}{
		{
			name:     "default fallback continuous rubric (0.0..1.0)",
			rule:     nil,
			opt:      nil,
			expected: []string{"0.0", "0.25", "0.5", "0.75", "1.0"},
		},
		{
			name: "continuous range interpolation (-40.0..60.0)",
			rule: &validate.FieldRules{
				Type: &validate.FieldRules_Float{
					Float: &validate.FloatRules{
						GreaterThan: &validate.FloatRules_Gte{Gte: -40.0},
						LessThan:    &validate.FloatRules_Lte{Lte: 60.0},
					},
				},
			},
			expected: []string{"-40.0", "-15.0", "10.0", "35.0", "60.0"},
		},
		{
			name: "discrete float in list",
			rule: &validate.FieldRules{
				Type: &validate.FieldRules_Float{
					Float: &validate.FloatRules{
						In: []float32{0.2, 0.5, 0.8},
					},
				},
			},
			expected: []string{"0.2", "0.5", "0.8"},
		},
		{
			name: "double continuous bounds (0.0..100.0)",
			rule: &validate.FieldRules{
				Type: &validate.FieldRules_Double{
					Double: &validate.DoubleRules{
						GreaterThan: &validate.DoubleRules_Gte{Gte: 0.0},
						LessThan:    &validate.DoubleRules_Lte{Lte: 100.0},
					},
				},
			},
			expected: []string{"0.0", "25.0", "50.0", "75.0", "100.0"},
		},
		{
			name: "custom Jev options min/max bounds (0.0..50.0)",
			rule: nil,
			opt: &jevv1.FieldOptions{
				Min: 0,
				Max: 50,
			},
			expected: []string{"0.0", "12.5", "25.0", "37.5", "50.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveFloatCriteria(nil, tt.rule, tt.opt)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestResolveStringCriteria(t *testing.T) {
	t.Run("empty without rules", func(t *testing.T) {
		got := resolveStringCriteria(nil, nil, nil)
		assert.Empty(t, got)
	})

	t.Run("protovalidate string.in discrete values", func(t *testing.T) {
		rule := &validate.FieldRules{
			Type: &validate.FieldRules_String_{
				String_: &validate.StringRules{
					In: []string{"ALPHA", "BETA", "GAMMA"},
				},
			},
		}
		got := resolveStringCriteria(nil, rule, nil)
		require.Len(t, got, 3)
		assert.Contains(t, got, "ALPHA")
		assert.Contains(t, got, "BETA")
		assert.Contains(t, got, "GAMMA")
	})

	t.Run("custom jev criteria options", func(t *testing.T) {
		opt := &jevv1.FieldOptions{
			Criteria: map[string]string{
				"APPROVE": "Request approved",
				"REJECT":  "Request rejected",
			},
		}
		got := resolveStringCriteria(nil, nil, opt)
		require.Len(t, got, 2)
		assert.Contains(t, got, "APPROVE")
		assert.Contains(t, got, "REJECT")
	})
}
