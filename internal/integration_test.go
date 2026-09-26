package integration_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	incidentv1 "github.com/sudorandom/protoc-gen-jev/examples/gen/jev/incident/v1"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIncidentJevClient_WithFauxRPC(t *testing.T) {
	ctx := context.Background()

	// Locate OpenAPI schema and stub files
	absSchemaDir, err := filepath.Abs("../testdata/openapi")
	require.NoError(t, err)

	// Spin up FauxRPC container with Testcontainers
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
		WaitingFor: wait.ForHTTP("/fauxrpc/openapi-docs/").WithPort("6660/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start fauxrpc testcontainer")
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	endpoint, err := container.PortEndpoint(ctx, "6660/tcp", "http")
	require.NoError(t, err)

	// Point the generated Jev client to the FauxRPC container
	client := incidentv1.NewIncidentTriageJevClient("mock-api-key")
	client.Endpoint = fmt.Sprintf("%s/v1/systemone", endpoint)

	// 1. Test BuildQuestions
	questions := client.BuildQuestions()
	assert.Len(t, questions, 6)
	assert.Contains(t, questions, "routing_target")
	assert.Contains(t, questions, "requiresImmediatePaging")
	assert.Contains(t, questions, "priority")
	assert.Contains(t, questions, "urgencyRating")
	assert.Contains(t, questions, "blastRadiusPercentage")
	assert.Contains(t, questions, "complianceClassification")

	// 2. Test Evaluate against FauxRPC HTTP OpenAPI endpoint
	state := map[string]any{
		"incident_id": "INC-9912",
		"title":       "Database connection pool exhausted",
		"description": "API latency increased to 4500ms and 500 errors spike to 12%",
	}

	decision, err := client.Evaluate(ctx, state)
	require.NoError(t, err)
	require.NotNil(t, decision)

	// Validate typed responses from FauxRPC OpenAPI stubs
	assert.Equal(t, "oncall_engineer", decision.RoutingTarget)
	assert.True(t, decision.RequiresImmediatePaging)
	assert.Equal(t, "PRIORITY_LEVEL_HIGH", decision.Priority)
	assert.InDelta(t, 4.0, decision.UrgencyRating, 0.001)
	assert.InDelta(t, 45.0, decision.BlastRadiusPercentage, 0.001)
	assert.Equal(t, "INTERNAL_CONFIDENTIAL", decision.ComplianceClassification)

	// 3. Test BatchEvaluate against FauxRPC
	batchStates := []any{
		map[string]any{"incident_id": "INC-1", "title": "Crash 1"},
		map[string]any{"incident_id": "INC-2", "title": "Crash 2"},
		map[string]any{"incident_id": "INC-3", "title": "Crash 3"},
	}

	decisions, err := client.BatchEvaluate(ctx, batchStates)
	require.NoError(t, err)
	require.Len(t, decisions, 3)

	for i, d := range decisions {
		assert.Equal(t, "oncall_engineer", d.RoutingTarget, "batch item %d", i)
		assert.True(t, d.RequiresImmediatePaging, "batch item %d", i)
		assert.Equal(t, "PRIORITY_LEVEL_HIGH", d.Priority, "batch item %d", i)
	}

	// 4. Failing test case: Calling an invalid endpoint returns error status
	invalidClient := incidentv1.NewIncidentTriageJevClient("mock-api-key")
	invalidClient.Endpoint = fmt.Sprintf("%s/v1/nonexistent", endpoint)
	_, err = invalidClient.Evaluate(ctx, state)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
