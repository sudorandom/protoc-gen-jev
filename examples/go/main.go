package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	rulesv1 "protoc-gen-jev/gen/jev/ai/rules/v1"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("  protoc-gen-jev: Go Client End-to-End Example   ")
	fmt.Println("==================================================")

	apiKey := os.Getenv("TYPESAFE_API_KEY")
	var mockServer *httptest.Server

	client := rulesv1.NewRuleTestRecordJevClient(apiKey)

	if apiKey == "" {
		fmt.Println("\n[INFO] TYPESAFE_API_KEY not set in environment.")
		fmt.Println("[INFO] Starting local simulated Jev System One server for offline demo...")

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]any{
				"choices": map[string]any{
					"delivery_method":   map[string]any{"choice": "push_notification"},
					"executionMode":     map[string]any{"choice": "MODE_FAST"},
					"filteredMode":      map[string]any{"choice": "MODE_BALANCED"},
					"securityClearance": map[string]any{"choice": "TOP_SECRET"},
					"decisionFlag":      map[string]any{"choice": "APPROVE"},
				},
				"scores": map[string]any{
					"ratingSmall":        map[string]any{"score": 5},
					"ratingStrict":       map[string]any{"score": 3},
					"discreteCode":       map[string]any{"score": 50},
					"largeScale":         map[string]any{"score": 500},
					"temperature":        map[string]any{"score": 25.0},
					"discreteRatio":      map[string]any{"score": 0.8},
					"customBoundedScore": map[string]any{"score": 30.0},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer mockServer.Close()

		client.Endpoint = mockServer.URL
		client.HTTPClient = mockServer.Client()
	} else {
		fmt.Println("\n[INFO] Using live TypeSafe AI Jev API: https://api.typesafe.ai/v1/systemone")
	}

	// 1. Inspect generated questions
	questions := client.BuildQuestions()
	qJSON, _ := json.MarshalIndent(questions, "", "  ")
	fmt.Printf("\n1. Built Jev Questions (%d total):\n%s\n", len(questions), string(qJSON))

	// 2. Evaluate single input state
	state := map[string]any{
		"user_id":      "usr_10492",
		"request_type": "expedited_transfer",
		"priority":     "high",
		"text":         "Urgent security request: please authorize immediately and send via push notification",
	}

	fmt.Println("\n2. Evaluating Single State...")
	ctx := context.Background()
	decision, err := client.Evaluate(ctx, state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error evaluating state: %v\n", err)
		os.Exit(1)
	}

	decJSON, _ := json.MarshalIndent(decision, "", "  ")
	fmt.Printf("✔ Decision received:\n%s\n", string(decJSON))

	// 3. Batch evaluation
	fmt.Println("\n3. Batch Evaluating 3 States...")
	batchStates := []any{
		"Batch Item 1: Standard alert",
		"Batch Item 2: High security alert",
		"Batch Item 3: Operational check",
	}

	batchDecisions, err := client.BatchEvaluate(ctx, batchStates)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in BatchEvaluate: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✔ Successfully evaluated %d batch items.\n", len(batchDecisions))
	for i, d := range batchDecisions {
		fmt.Printf("  - Item [%d]: DeliveryMethod=%s, RatingSmall=%.0f, Clearance=%s\n",
			i+1, d.DeliveryMethod, d.RatingSmall, d.SecurityClearance)
	}

	fmt.Println("\n✔ Go End-to-End Test PASSED successfully!")
}
