package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sudorandom/protoc-gen-jev/internal/model"
	jevv1 "github.com/sudorandom/protoc-gen-jev/pkg/jev/v1"
)

func TestResolveScoreLevels(t *testing.T) {
	tests := []struct {
		name             string
		rule             *jevv1.ScoreRules
		expectedLevels   []model.ScoreLevel
		expectedCriteria []string
		expectErr        bool
	}{
		{
			name: "explicit levels",
			rule: &jevv1.ScoreRules{
				Levels: []*jevv1.ScoreLevel{
					{Value: 1, Description: "Minor; can wait"},
					{Value: 2, Description: "Low urgency"},
					{Value: 3, Description: "Needs attention soon"},
					{Value: 4, Description: "Urgent"},
					{Value: 5, Description: "Immediate action required"},
				},
			},
			expectedLevels: []model.ScoreLevel{
				{Value: 1, Description: "Minor; can wait"},
				{Value: 2, Description: "Low urgency"},
				{Value: 3, Description: "Needs attention soon"},
				{Value: 4, Description: "Urgent"},
				{Value: 5, Description: "Immediate action required"},
			},
			expectedCriteria: []string{
				"Minor; can wait",
				"Low urgency",
				"Needs attention soon",
				"Urgent",
				"Immediate action required",
			},
		},
		{
			name: "explicit levels with empty descriptions fallback to formatted float",
			rule: &jevv1.ScoreRules{
				Levels: []*jevv1.ScoreLevel{
					{Value: 0.2},
					{Value: 0.5},
					{Value: 0.8},
				},
			},
			expectedLevels: []model.ScoreLevel{
				{Value: 0.2, Description: ""},
				{Value: 0.5, Description: ""},
				{Value: 0.8, Description: ""},
			},
			expectedCriteria: []string{"0.2", "0.5", "0.8"},
		},
		{
			name: "too few levels error",
			rule: &jevv1.ScoreRules{
				Levels: []*jevv1.ScoreLevel{
					{Value: 1, Description: "Solo"},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			levels, criteria, err := resolveScoreLevels(nil, tt.rule)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLevels, levels)
				assert.Equal(t, tt.expectedCriteria, criteria)
			}
		})
	}
}

func TestResolveStringCriteria(t *testing.T) {
	tests := []struct {
		name     string
		rule     *jevv1.ChoiceRules
		expected map[string]any
	}{
		{
			name:     "empty without choice rules",
			rule:     nil,
			expected: map[string]any{},
		},
		{
			name: "jev choices whitelist",
			rule: &jevv1.ChoiceRules{
				Choices: []string{"PUBLIC", "INTERNAL_CONFIDENTIAL", "RESTRICTED_PII"},
			},
			expected: map[string]any{
				"PUBLIC":                "",
				"INTERNAL_CONFIDENTIAL": "",
				"RESTRICTED_PII":        "",
			},
		},
		{
			name: "custom jev criteria options",
			rule: &jevv1.ChoiceRules{
				Criteria: map[string]string{
					"APPROVE": "Request approved",
					"REJECT":  "Request rejected",
				},
			},
			expected: map[string]any{
				"APPROVE": "Request approved",
				"REJECT":  "Request rejected",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveStringCriteria(nil, tt.rule)
			assert.Equal(t, tt.expected, got)
		})
	}
}
