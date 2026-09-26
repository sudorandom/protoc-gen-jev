package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	incidentv1 "github.com/sudorandom/protoc-gen-jev/examples/gen/jev/incident/v1"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("  protoc-gen-jev: Go Client End-to-End Example   ")
	fmt.Println("==================================================")

	apiKey := os.Getenv("TYPESAFE_API_KEY")
	var container testcontainers.Container

	client := incidentv1.NewIncidentTriageJevClient(apiKey)

	if apiKey == "" {
		ctx := context.Background()

		absSchemaDir, err := filepath.Abs("testdata/openapi")
		if err != nil {
			log.Fatalf("failed to locate openapi schema: %v", err)
		}
		fmt.Println("\n[INFO] TYPESAFE_API_KEY not set in environment.")
		fmt.Println("[INFO] Starting FauxRPC Testcontainer with Jev OpenAPI spec...")

		req := testcontainers.ContainerRequest{
			Image:        "docker.io/sudorandom/fauxrpc:v0.29.1",
			ExposedPorts: []string{"6660/tcp"},
			Cmd: []string{
				"run",
				"--schema=/openapi/typesafe-jev.yaml",
				"--stubs=/openapi/stubs.jev.yaml",
				"--addr=0.0.0.0:6660",
			},
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      filepath.Join(absSchemaDir, "typesafe-jev.yaml"),
					ContainerFilePath: "/openapi/typesafe-jev.yaml",
					FileMode:          0644,
				},
				{
					HostFilePath:      filepath.Join(absSchemaDir, "stubs.jev.yaml"),
					ContainerFilePath: "/openapi/stubs.jev.yaml",
					FileMode:          0644,
				},
			},
			WaitingFor: wait.ForHTTP("/fauxrpc/openapi-docs/").WithPort("6660/tcp").WithStartupTimeout(20 * time.Second),
		}

		c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			log.Fatalf("failed to start fauxrpc testcontainer: %v", err)
		}
		container = c
		endpoint, err := container.PortEndpoint(ctx, "6660/tcp", "http")
		if err != nil {
			log.Fatalf("failed to get fauxrpc container endpoint: %v", err)
		}
		fmt.Printf("[INFO] FauxRPC mock server running at %s\n", endpoint)
		client.Endpoint = fmt.Sprintf("%s/v1/systemone", endpoint)
	} else {
		fmt.Println("\n[INFO] Using live TypeSafe AI Jev API: https://api.typesafe.ai/v1/systemone")
	}

	defer func() {
		if container != nil {
			_ = container.Terminate(context.Background())
		}
	}()

	// 1. Inspect generated questions
	questions := client.BuildQuestions()
	qJSON, _ := json.MarshalIndent(questions, "", "  ")
	fmt.Printf("\n1. Built Jev Questions (%d total):\n%s\n", len(questions), string(qJSON))

	// 2. Evaluate single input state
	state := map[string]any{
		"incident_id": "INC-8891",
		"title":       "Database connection pool exhausted",
		"description": "API latency increased to 4500ms and 500 errors spike to 12%",
		"raw_logs":    "Connection refused on port 5432 after 100 pool max connections",
	}

	fmt.Println("\n2. Evaluating Single Incident State...")
	ctx := context.Background()
	decision, err := client.Evaluate(ctx, state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error evaluating state: %v\n", err)
		os.Exit(1)
	}

	decJSON, _ := json.MarshalIndent(decision, "", "  ")
	fmt.Printf("✔ Decision received:\n%s\n", string(decJSON))

	// 3. Batch evaluation
	fmt.Println("\n3. Batch Evaluating 3 Incident States...")
	batchStates := []any{
		"Batch Item 1: Ingress 502 bad gateway spikes across region us-east-1",
		"Batch Item 2: Low-priority deprecation warning logged in analytics service",
		"Batch Item 3: Routine memory compaction completed without customer impact",
	}

	batchDecisions, err := client.BatchEvaluate(ctx, batchStates)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in BatchEvaluate: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✔ Successfully evaluated %d batch items.\n", len(batchDecisions))
	for i, d := range batchDecisions {
		fmt.Printf("  - Item [%d]: RoutingTarget=%s, Paging=%t, Priority=%s, Urgency=%.0f, BlastRadius=%.1f%%\n",
			i+1, d.RoutingTarget, d.RequiresImmediatePaging, d.Priority, d.UrgencyRating, d.BlastRadiusPercentage)
	}

	fmt.Println("\n✔ Go End-to-End Test PASSED successfully!")
}
