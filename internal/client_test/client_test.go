package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	rulesv1 "protoc-gen-jev/gen/jev/ai/rules/v1"
)

func TestGeneratedClient_BuildQuestions(t *testing.T) {
	client := rulesv1.NewRuleTestRecordJevClient("test-key")
	questions := client.BuildQuestions()

	// 1. Oneof delivery_method -> Choice
	dm, ok := questions["delivery_method"].(map[string]any)
	if !ok {
		t.Fatalf("expected delivery_method question")
	}
	if dm["type"] != "choice" {
		t.Errorf("delivery_method type = %v, want choice", dm["type"])
	}
	crit, _ := dm["criteria"].(map[string]any)
	if _, ok := crit["email"]; !ok {
		t.Errorf("missing email in delivery_method criteria")
	}
	if _, ok := crit["sms"]; !ok {
		t.Errorf("missing sms in delivery_method criteria")
	}
	if _, ok := crit["push_notification"]; !ok {
		t.Errorf("missing push_notification in delivery_method criteria")
	}

	// 2. Enum.in executionMode -> Choice with only MODE_FAST and MODE_BALANCED
	em, ok := questions["executionMode"].(map[string]any)
	if !ok {
		t.Fatalf("expected executionMode question")
	}
	emCrit, _ := em["criteria"].(map[string]any)
	if len(emCrit) != 2 {
		t.Errorf("executionMode criteria count = %d, want 2 (only MODE_FAST and MODE_BALANCED)", len(emCrit))
	}

	// 3. Enum.not_in filteredMode -> MODE_DEBUG must NOT be in criteria
	fm, ok := questions["filteredMode"].(map[string]any)
	if !ok {
		t.Fatalf("expected filteredMode question")
	}
	fmCrit, _ := fm["criteria"].(map[string]any)
	if _, ok := fmCrit["MODE_DEBUG"]; ok {
		t.Errorf("MODE_DEBUG should have been excluded by not_in rule")
	}

	// 4. Integer small range ratingSmall -> 1..5
	rs, ok := questions["ratingSmall"].(map[string]any)
	if !ok {
		t.Fatalf("expected ratingSmall question")
	}
	rsCrit, _ := rs["criteria"].([]string)
	wantRs := []string{"1", "2", "3", "4", "5"}
	if len(rsCrit) != len(wantRs) {
		t.Errorf("ratingSmall criteria = %v, want %v", rsCrit, wantRs)
	}

	// 5. Integer discrete allowed values discreteCode -> 10, 20, 50, 100
	dc, ok := questions["discreteCode"].(map[string]any)
	if !ok {
		t.Fatalf("expected discreteCode question")
	}
	dcCrit, _ := dc["criteria"].([]string)
	wantDc := []string{"10", "20", "50", "100"}
	if len(dcCrit) != len(wantDc) {
		t.Errorf("discreteCode criteria = %v, want %v", dcCrit, wantDc)
	}

	// 6. Large scale integer largeScale -> [0, 250, 500, 750, 1000]
	ls, ok := questions["largeScale"].(map[string]any)
	if !ok {
		t.Fatalf("expected largeScale question")
	}
	lsCrit, _ := ls["criteria"].([]string)
	wantLs := []string{"0", "250", "500", "750", "1000"}
	if len(lsCrit) != len(wantLs) {
		t.Errorf("largeScale criteria = %v, want %v", lsCrit, wantLs)
	}

	// 7. Float continuous temperature -> [-40.0, -15.0, 10.0, 35.0, 60.0]
	temp, ok := questions["temperature"].(map[string]any)
	if !ok {
		t.Fatalf("expected temperature question")
	}
	tempCrit, _ := temp["criteria"].([]string)
	wantTemp := []string{"-40.0", "-15.0", "10.0", "35.0", "60.0"}
	if len(tempCrit) != len(wantTemp) {
		t.Errorf("temperature criteria = %v, want %v", tempCrit, wantTemp)
	}

	// 8. String securityClearance -> Choice with PUBLIC, SECRET, TOP_SECRET
	sc, ok := questions["securityClearance"].(map[string]any)
	if !ok {
		t.Fatalf("expected securityClearance question")
	}
	if sc["type"] != "choice" {
		t.Errorf("securityClearance type = %v, want choice", sc["type"])
	}
	scCrit, _ := sc["criteria"].(map[string]any)
	if len(scCrit) != 3 {
		t.Errorf("securityClearance criteria count = %d, want 3", len(scCrit))
	}

	// 9. Custom Jev criteria decisionFlag -> Choice with APPROVE, REJECT
	df, ok := questions["decisionFlag"].(map[string]any)
	if !ok {
		t.Fatalf("expected decisionFlag question")
	}
	dfCrit, _ := df["criteria"].(map[string]any)
	if _, ok := dfCrit["APPROVE"]; !ok {
		t.Errorf("missing APPROVE in decisionFlag criteria")
	}

	// 10. Custom Jev min/max customBoundedScore -> [10.0, 20.0, 30.0, 40.0, 50.0]
	cbs, ok := questions["customBoundedScore"].(map[string]any)
	if !ok {
		t.Fatalf("expected customBoundedScore question")
	}
	cbsCrit, _ := cbs["criteria"].([]string)
	wantCbs := []string{"10.0", "20.0", "30.0", "40.0", "50.0"}
	if len(cbsCrit) != len(wantCbs) {
		t.Errorf("customBoundedScore criteria = %v, want %v", cbsCrit, wantCbs)
	}

	// 11. Skipped & freeform fields MUST NOT be present
	if _, ok := questions["secretToken"]; ok {
		t.Errorf("secretToken was marked skip=true and must not be in questions")
	}
	if _, ok := questions["freeformDescription"]; ok {
		t.Errorf("freeformDescription has no decision rules and must not be in questions")
	}
	if _, ok := questions["optionalNote"]; ok {
		t.Errorf("optionalNote is a proto3 optional (synthetic oneof) and must not be in questions")
	}
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
	if err != nil {
		t.Fatalf("client.Evaluate error: %v", err)
	}

	if decisions.DeliveryMethod != "email" {
		t.Errorf("DeliveryMethod = %q, want email", decisions.DeliveryMethod)
	}
	if decisions.ExecutionMode != "MODE_FAST" {
		t.Errorf("ExecutionMode = %q, want MODE_FAST", decisions.ExecutionMode)
	}
	if decisions.RatingSmall != 5 {
		t.Errorf("RatingSmall = %v, want 5", decisions.RatingSmall)
	}
	if decisions.SecurityClearance != "TOP_SECRET" {
		t.Errorf("SecurityClearance = %q, want TOP_SECRET", decisions.SecurityClearance)
	}
	if decisions.CustomBoundedScore != 40.0 {
		t.Errorf("CustomBoundedScore = %v, want 40.0", decisions.CustomBoundedScore)
	}
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
	if err != nil {
		t.Fatalf("BatchEvaluate error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if callCount != 3 {
		t.Errorf("expected 3 server calls, got %d", callCount)
	}
	for i, dec := range results {
		if dec.DeliveryMethod != "sms" {
			t.Errorf("result[%d].DeliveryMethod = %q, want sms", i, dec.DeliveryMethod)
		}
		if dec.RatingSmall != 4 {
			t.Errorf("result[%d].RatingSmall = %v, want 4", i, dec.RatingSmall)
		}
	}
}
