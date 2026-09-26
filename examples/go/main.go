package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protojson"

	incidentv1 "github.com/sudorandom/protoc-gen-jev/examples/gen/jev/incident/v1"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("  protoc-gen-jev: Go Client End-to-End Example   ")
	fmt.Println("==================================================")
	apiKey := os.Getenv("TYPESAFE_API_KEY")
	customEndpoint := os.Getenv("JEV_ENDPOINT")

	client := incidentv1.NewIncidentTriageServiceClient(apiKey)
	if customEndpoint != "" {
		client.Endpoint = customEndpoint
		fmt.Printf("\n[INFO] Using custom Jev/Laya endpoint: %s\n", customEndpoint)
	} else if apiKey != "" {
		fmt.Println("\n[INFO] Using live TypeSafe AI Jev API: https://api.typesafe.ai/v1/systemone")
	} else {
		fmt.Printf("\n[INFO] Using default endpoint: %s\n", client.Endpoint)
	}

	// 1. Inspect generated questions
	questions := client.BuildTriageQuestions()
	qJSON, _ := json.MarshalIndent(questions, "", "  ")
	fmt.Printf("\n1. Built Jev Questions (%d total):\n%s\n", len(questions), string(qJSON))

	// 2. Evaluate single input state
	req := &incidentv1.TriageRequest{
		IncidentId:  "INC-8891",
		Title:       "Database connection pool exhausted",
		Description: "API latency increased to 4500ms and 500 errors spike to 12%",
		RawLogs:     "Connection refused on port 5432 after 100 pool max connections",
	}

	fmt.Println("\n2. Evaluating Single Incident State...")
	ctx := context.Background()
	decision, err := client.Triage(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error evaluating state: %v\n", err)
		os.Exit(1)
	}

	decBytes, _ := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(decision)
	fmt.Printf("✔ Decision received:\n%s\n", string(decBytes))

	// 3. Batch evaluation
	fmt.Println("\n3. Batch Evaluating 3 Incident States...")
	batchReqs := []*incidentv1.TriageRequest{
		{
			IncidentId:  "INC-8892",
			Title:       "Ingress 502 bad gateway spikes across region us-east-1",
			Description: "Edge proxy reports connection reset by peer from upstream cluster",
			RawLogs:     "HTTP 502 Bad Gateway - upstream connect error or disconnect/reset before headers",
		},
		{
			IncidentId:  "INC-8893",
			Title:       "Low-priority deprecation warning logged in analytics service",
			Description: "Client library using deprecated v1 query endpoint; scheduled for removal in Q3",
			RawLogs:     "WARN [analytics-worker] Endpoint /v1/query is deprecated, migrate to /v2/query",
		},
		{
			IncidentId:  "INC-8894",
			Title:       "Routine memory compaction completed without customer impact",
			Description: "Background compaction cycle reclaimed 4.2GB memory; latency within SLO",
			RawLogs:     "INFO [compactor] Compaction cycle finished in 45s, 0 errors, 4200MB reclaimed",
		},
	}

	batchDecisions, err := client.BatchTriage(ctx, batchReqs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in BatchTriage: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✔ Successfully evaluated %d batch items.\n", len(batchDecisions))
	for i, d := range batchDecisions {
		var target string
		switch x := d.GetRoutingTarget().(type) {
		case *incidentv1.TriageResponse_AutomatedRunbook:
			target = "automated_runbook:" + x.AutomatedRunbook
		case *incidentv1.TriageResponse_OncallEngineer:
			target = "oncall_engineer:" + x.OncallEngineer
		case *incidentv1.TriageResponse_IncidentCommander:
			target = "incident_commander:" + x.IncidentCommander
		}
		fmt.Printf("  - Item [%d]: RoutingTarget=%s, Paging=%t, Priority=%s, Urgency=%d, BlastRadius=%.1f%%\n",
			i+1, target, d.RequiresImmediatePaging, d.Priority.String(), d.UrgencyRating, d.BlastRadiusPercentage)
	}

	fmt.Println("\n✔ Go End-to-End Test PASSED successfully!")
}
