package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rulesv1 "github.com/sudorandom/protoc-gen-jev/gen/jev/ai/rules/v1"
)

func TestGeneratedClient_BuildQuestions(t *testing.T) {
	client := rulesv1.NewRuleTestRecordJevClient("test-key")
	questions := client.BuildQuestions()

	// 1. Oneof delivery_method -> Choice
	require.Contains(t, questions, "delivery_method")
	dm, ok := questions["delivery_method"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "choice", dm["type"])
	crit, ok := dm["criteria"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, crit, "email")
	assert.Contains(t, crit, "sms")
	assert.Contains(t, crit, "push_notification")

	// 2. Enum.in executionMode -> Choice with only MODE_FAST and MODE_BALANCED
	require.Contains(t, questions, "executionMode")
	em, ok := questions["executionMode"].(map[string]any)
	require.True(t, ok)
	emCrit, ok := em["criteria"].(map[string]any)
	require.True(t, ok)
	assert.Len(t, emCrit, 2)
	assert.Contains(t, emCrit, "MODE_FAST")
	assert.Contains(t, emCrit, "MODE_BALANCED")

	// 3. Enum.not_in filteredMode -> MODE_DEBUG must NOT be in criteria
	require.Contains(t, questions, "filteredMode")
	fm, ok := questions["filteredMode"].(map[string]any)
	require.True(t, ok)
	fmCrit, ok := fm["criteria"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, fmCrit, "MODE_DEBUG")

	// 4. Integer small range ratingSmall -> 1..5
	require.Contains(t, questions, "ratingSmall")
	rs, ok := questions["ratingSmall"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []string{"1", "2", "3", "4", "5"}, rs["criteria"])

	// 5. Integer discrete allowed values discreteCode -> 10, 20, 50, 100
	require.Contains(t, questions, "discreteCode")
	dc, ok := questions["discreteCode"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []string{"10", "20", "50", "100"}, dc["criteria"])

	// 6. Large scale integer largeScale -> [0, 250, 500, 750, 1000]
	require.Contains(t, questions, "largeScale")
	ls, ok := questions["largeScale"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []string{"0", "250", "500", "750", "1000"}, ls["criteria"])

	// 7. Float continuous temperature -> [-40.0, -15.0, 10.0, 35.0, 60.0]
	require.Contains(t, questions, "temperature")
	temp, ok := questions["temperature"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []string{"-40.0", "-15.0", "10.0", "35.0", "60.0"}, temp["criteria"])

	// 8. String securityClearance -> Choice with PUBLIC, SECRET, TOP_SECRET
	require.Contains(t, questions, "securityClearance")
	sc, ok := questions["securityClearance"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "choice", sc["type"])
	scCrit, ok := sc["criteria"].(map[string]any)
	require.True(t, ok)
	assert.Len(t, scCrit, 3)

	// 9. Custom Jev criteria decisionFlag -> Choice with APPROVE, REJECT
	require.Contains(t, questions, "decisionFlag")
	df, ok := questions["decisionFlag"].(map[string]any)
	require.True(t, ok)
	dfCrit, ok := df["criteria"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, dfCrit, "APPROVE")
	assert.Contains(t, dfCrit, "REJECT")

	// 10. Custom Jev min/max customBoundedScore -> [10.0, 20.0, 30.0, 40.0, 50.0]
	require.Contains(t, questions, "customBoundedScore")
	cbs, ok := questions["customBoundedScore"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []string{"10.0", "20.0", "30.0", "40.0", "50.0"}, cbs["criteria"])

	// 11. Skipped & freeform fields MUST NOT be present
	assert.NotContains(t, questions, "secretToken")
	assert.NotContains(t, questions, "freeformDescription")
	assert.NotContains(t, questions, "optionalNote")
}

func TestGeneratedClient_Evaluate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": map[string]any{
				"delivery_method":   map[string]any{"choice": "email"},
				"executionMode":     map[string]any{"choice": "MODE_FAST"},
				"filteredMode":      map[string]any{"choice": "MODE_BALANCED"},
				"securityClearance": map[string]any{"choice": "TOP_SECRET"},
				"decisionFlag":      map[string]any{"choice": "APPROVE"},
			},
			"scores": map[string]any{
				"ratingSmall":        map[string]any{"score": 5},
				"ratingStrict":       map[string]any{"score": 3},
				"discreteCode":       map[string]any{"score": 100},
				"largeScale":         map[string]any{"score": 750},
				"temperature":        map[string]any{"score": 35.0},
				"discreteRatio":      map[string]any{"score": 0.8},
				"customBoundedScore": map[string]any{"score": 40.0},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := rulesv1.NewRuleTestRecordJevClient("dummy-key")
	client.Endpoint = ts.URL
	client.HTTPClient = ts.Client()

	decisions, err := client.Evaluate(context.Background(), map[string]any{"text": "Sample request"})
	require.NoError(t, err)

	assert.Equal(t, "email", decisions.DeliveryMethod)
	assert.Equal(t, "MODE_FAST", decisions.ExecutionMode)
	assert.InDelta(t, 5.0, decisions.RatingSmall, 0.001)
	assert.Equal(t, "TOP_SECRET", decisions.SecurityClearance)
	assert.InDelta(t, 40.0, decisions.CustomBoundedScore, 0.001)
}

func TestGeneratedClient_BatchEvaluate(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := map[string]any{
			"choices": map[string]any{
				"delivery_method": map[string]any{"choice": "sms"},
			},
			"scores": map[string]any{
				"ratingSmall": map[string]any{"score": 4},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := rulesv1.NewRuleTestRecordJevClient("dummy-key")
	client.Endpoint = ts.URL
	client.HTTPClient = ts.Client()

	states := []any{
		"first text",
		"second text",
		"third text",
	}

	results, err := client.BatchEvaluate(context.Background(), states)
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, 3, callCount)

	for _, dec := range results {
		assert.Equal(t, "sms", dec.DeliveryMethod)
		assert.InDelta(t, 4.0, dec.RatingSmall, 0.001)
	}
}

func TestGeneratedClient_Evaluate_HTTPError400(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "invalid questions payload: missing type"}`))
	}))
	defer ts.Close()

	client := rulesv1.NewRuleTestRecordJevClient("dummy-key")
	client.Endpoint = ts.URL
	client.HTTPClient = ts.Client()

	_, err := client.Evaluate(context.Background(), map[string]any{"text": "Sample request"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "invalid questions payload")
}

func TestGeneratedClient_Evaluate_HTTPError500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`internal server error`))
	}))
	defer ts.Close()

	client := rulesv1.NewRuleTestRecordJevClient("dummy-key")
	client.Endpoint = ts.URL
	client.HTTPClient = ts.Client()

	_, err := client.Evaluate(context.Background(), map[string]any{"text": "Sample request"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestGeneratedClient_Evaluate_MalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not a valid json`))
	}))
	defer ts.Close()

	client := rulesv1.NewRuleTestRecordJevClient("dummy-key")
	client.Endpoint = ts.URL
	client.HTTPClient = ts.Client()

	_, err := client.Evaluate(context.Background(), map[string]any{"text": "Sample request"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode Jev response")
}

func TestGeneratedClient_Evaluate_NetworkError(t *testing.T) {
	client := rulesv1.NewRuleTestRecordJevClient("dummy-key")
	client.Endpoint = "http://127.0.0.1:1" // unreachable port

	_, err := client.Evaluate(context.Background(), map[string]any{"text": "Sample request"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Jev request failed")
}
