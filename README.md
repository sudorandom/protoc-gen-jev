# protoc-gen-jev

[![Latest Release](https://img.shields.io/github/v/release/sudorandom/protoc-gen-jev?logo=github&style=flat-square)](https://github.com/sudorandom/protoc-gen-jev/releases/latest)
[![Go CI](https://img.shields.io/github/actions/workflow/status/sudorandom/protoc-gen-jev/go.yml?label=Go%20CI&style=flat-square)](https://github.com/sudorandom/protoc-gen-jev/actions/workflows/go.yml)
[![Buf CI](https://img.shields.io/github/actions/workflow/status/sudorandom/protoc-gen-jev/buf.yml?label=Buf%20CI&style=flat-square)](https://github.com/sudorandom/protoc-gen-jev/actions/workflows/buf.yml)
[![Release](https://img.shields.io/github/actions/workflow/status/sudorandom/protoc-gen-jev/publish.yml?label=Release&style=flat-square)](https://github.com/sudorandom/protoc-gen-jev/actions/workflows/publish.yml)

> [!WARNING]
> **Experimental**: `protoc-gen-jev` is currently experimental and under active development. APIs, question schema formats, and generated code templates are subject to change without notice.

A Protobuf compiler plugin for **TypeSafe AI's Jev** and **Laya** (System One fast cognitive decision models).

`protoc-gen-jev` inspects Protobuf services and messages to generate strictly typed Jev client code, question schemas, and decision structures across Go, TypeScript, and Python.

---

## 💡 Why `protoc-gen-jev`?

Jev and related cognitive decision engines (like [Laya](https://github.com/typesafe-ai)) rely on structured decisions evaluated against unstructured state. However, official client SDKs avoid strict typing—forcing developers to construct loose JSON dictionaries and parse untyped `dict` / `Record<string, unknown>` maps by hand.

**`protoc-gen-jev` brings true compile-time safety to TypeSafe AI.**

By defining your decision models as standard Protobuf services, your application and Jev share an immutable, strictly typed contract:

- **No Possibility of Type Mistakes**: Eliminate runtime typos (`res["prioriy"]`) and type mismatches. Your compiler and type checker (`go build`, `tsc`, `mypy`) catch errors before code ever deploys.
- **Contract-First Architecture**: Define your input context (`TriageRequest`) and decision structure (`TriageResponse`) once in `.proto`; get native, idiomatic clients across Go, Python, and TypeScript for free.
- **Zero Handwritten Schema Boilerplate**: Question rubrics, doc comments, and scoring scales live in your schema, eliminating manual, out-of-band JSON question authoring.
- **Built-in Sanity Checks**: The plugin validates schemas at compile time, guaranteeing that RPC requests define input state, response messages contain valid decision questions, and choice/score rubrics are structurally valid.

---

## Installation

### Via `go install` (Recommended)

Install the latest plugin binary directly using Go:

```bash
go install github.com/sudorandom/protoc-gen-jev@latest
```

Make sure `$GOPATH/bin` (typically `~/go/bin`) is in your system `PATH`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

### Pre-built Binaries (GitHub Releases)

Pre-compiled binaries for Linux, macOS, and Windows are available on the [GitHub Releases](https://github.com/sudorandom/protoc-gen-jev/releases/latest) page.

Verify that the plugin is available:

```bash
protoc-gen-jev -version
```

---

## Generated Targets

| Target | File Suffix | Purpose |
| :--- | :--- | :--- |
| **Go** | `*_jev.pb.go` | Native typed Go client with strongly typed `Request` $\to$ `Response` methods |
| **TypeScript** | `*_jev.ts` | Typed TypeScript client wrapping `@typesafe-ai/sdk` with strict interfaces |
| **Python** | `*_jev.py` | Typed Python client wrapping `typesafe-sdk` with strict dataclasses |
| **JSON** | `*.jev.json` | Language-agnostic Jev question schema definitions |

---

## Usage with Buf

### 1. Depend on `jev/v1/options.proto`

Add the dependency to your `buf.yaml` (see [`examples/buf.yaml.example`](examples/buf.yaml.example)):

```yaml
version: v2
deps:
  - buf.build/sudo-random/protoc-gen-jev
```

Then update your dependencies:

```bash
buf dep update
```

### 2. Configure Code Generation

Add `protoc-gen-jev` to your `buf.gen.yaml` (see [`examples/buf.gen.yaml`](examples/buf.gen.yaml)):

```yaml
version: v2
plugins:
  - local: protoc-gen-jev
    out: gen/jev
    opt:
      # Options: go, ts, python, json, or all (comma-separated)
      - targets=all
```

---

## Schema Definition & Service Architecture

Define your request context, decision response, and Jev service in Protobuf. `protoc-gen-jev` is completely self-contained with zero external runtime dependencies.

### Comprehensive Example

The example below models an **Incident Triage & Routing Service** (see [`examples/proto/incident/v1/incident.proto`](examples/proto/incident/v1/incident.proto)):

```protobuf
syntax = "proto3";

package incident.v1;

import "jev/v1/options.proto";

enum PriorityLevel {
  PRIORITY_LEVEL_UNSPECIFIED = 0;
  PRIORITY_LEVEL_LOW = 1;
  PRIORITY_LEVEL_MEDIUM = 2;
  PRIORITY_LEVEL_HIGH = 3;
  PRIORITY_LEVEL_CRITICAL = 4;
}

// 1. Request Context: Unstructured input state evaluated by Jev
message TriageRequest {
  string incident_id = 1;
  string title = 2;
  string description = 3;
  string raw_logs = 4;
}

// 2. Decision Response: Structured cognitive decisions returned by Jev
message TriageResponse {
  // Oneof mutual exclusion -> Jev Choice question
  oneof routing_target {
    string automated_runbook = 1;
    string oncall_engineer = 2;
    string incident_commander = 3;
  }

  // Boolean field -> Jev Noul question with calibrated threshold
  bool requires_immediate_paging = 4 [
    (jev.v1.field).instructions = "Does this incident indicate active user-facing outage requiring paging?",
    (jev.v1.field).noul = { threshold: 0.85 }
  ];

  // Enum field with criteria guidance -> Jev Choice question
  PriorityLevel priority = 5 [
    (jev.v1.field).instructions = "Assess the operational severity and customer blast radius",
    (jev.v1.field).choice = {
      criteria: {
        key: "PRIORITY_LEVEL_CRITICAL",
        value: "Complete service outage affecting >10% of traffic"
      },
      criteria: {
        key: "PRIORITY_LEVEL_HIGH",
        value: "Significant latency spike or core feature degradation"
      },
      criteria: {
        key: "PRIORITY_LEVEL_MEDIUM",
        value: "Isolated component failure with working fallback"
      },
      criteria: {
        key: "PRIORITY_LEVEL_LOW",
        value: "Minor cosmetic or non-customer-impacting bug"
      }
    }
  ];

  // Bounded integer range -> Jev Score with discrete rubric [1, 2, 3, 4, 5]
  int32 urgency_rating = 6 [
    (jev.v1.field).instructions = "How urgently does this issue need to be resolved?",
    (jev.v1.field).score = { min: 1, max: 5 }
  ];

  // Continuous numeric scale -> Jev Score with 5-tier interpolated rubric [0.0..100.0]
  float blast_radius_percentage = 7 [
    (jev.v1.field).instructions = "Estimated percentage of production infrastructure impacted",
    (jev.v1.field).score = { min: 0.0, max: 100.0 }
  ];

  // String with discrete allowed choices -> Jev Choice question
  string compliance_classification = 8 [
    (jev.v1.field).instructions = "Categorize sensitive or regulated data exposed in the incident report",
    (jev.v1.field).choice = {
      choices: ["PUBLIC", "INTERNAL_CONFIDENTIAL", "RESTRICTED_PII", "PCI_DSS"]
    }
  ];
}

// 3. Service Definition: Strongly-typed RPC interface for Jev
service IncidentTriageService {
  rpc Triage(TriageRequest) returns (TriageResponse);
}
```

---

## Service & Field Options Reference

### Service & Method Options (`(jev.v1.service)`, `(jev.v1.method)`)

`protoc-gen-jev` features **smart auto-discovery**: any RPC whose request or response references `(jev.v1.field)` is automatically generated as a Jev RPC. Non-Jev services are safely ignored.

You can also explicitly control code generation:

```protobuf
service CustomService {
  // Force-enable Jev generation (activates natural type mappings even without field options)
  option (jev.v1.service).enabled = true;

  rpc Evaluate(CustomRequest) returns (CustomResponse);

  rpc StandardRpc(StandardRequest) returns (StandardResponse) {
    // Explicitly exclude a non-Jev RPC
    option (jev.v1.method).enabled = false;
  }
}
```

### Field Options Reference (`(jev.v1.field)`)

Options are grouped by Jev's cognitive primitives (`choice`, `score`, `noul`):

| Option | Type | Description |
| :--- | :--- | :--- |
| `instructions` | `string` | Custom instructions / prompt sent to Jev. Defaults to the field's Protobuf doc comment if omitted. |
| `skip` | `bool` | When `true`, completely skips this field during evaluation. |
| **`choice`** | `ChoiceRules` | Configures discrete selection choices. |
| `choice.choices` | `repeated string` | Explicit whitelist of allowed choice values (for strings or enums). |
| `choice.not_in` | `repeated string` | Blacklist of enum value names to exclude from choices. |
| `choice.criteria` | `map<string, string>` | Descriptive rubric criteria attached to specific options (labels $\to$ descriptions). |
| **`score`** | `ScoreRules` | Configures continuous or rubric score evaluations. |
| `score.min` | `float` | Lower bound of rubric scale (default: `0.0`). |
| `score.max` | `float` | Upper bound of rubric scale (default: `100.0`). |
| `score.scale` | `repeated float` | Explicit discrete rubric tiers (e.g. `[1, 2, 3]` or `[10, 20, 50, 100]`). |
| **`noul`** | `NoulRules` | Configures binary yes/no intent questions. |
| `noul.threshold` | `float` | Confidence threshold `[0.0 - 1.0]` required to evaluate as `true`. |

---

## Strictly Typed Client Usage

### Go

```go
package main

import (
	"context"
	"fmt"
	"os"

	incidentv1 "your_project/gen/jev/incident/v1"
)

func main() {
	client := incidentv1.NewIncidentTriageServiceClient(os.Getenv("TYPESAFE_API_KEY"))

	// Single evaluation: strongly typed request -> strongly typed response
	res, err := client.Triage(context.Background(), &incidentv1.TriageRequest{
		IncidentId:  "INC-1042",
		Title:       "Database connection pool exhausted",
		Description: "API latency increased to 4500ms and 500 errors spike to 12%",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Route: %s, Priority: %s, Urgency: %d\n",
		res.RoutingTarget, res.Priority, res.UrgencyRating)

	// Batch evaluation: slice of typed requests -> slice of typed responses
	batch, err := client.BatchTriage(context.Background(), []*incidentv1.TriageRequest{
		{IncidentId: "INC-1043", Title: "Memory leak in worker pool"},
		{IncidentId: "INC-1044", Title: "Elevated cache miss rate"},
	})
}
```

### TypeScript

```typescript
import {
  IncidentTriageServiceClient,
  TriageRequest,
} from "./gen/jev/incident/v1/incident_jev";

const client = new IncidentTriageServiceClient();

// Single evaluation
const res = await client.triage({
  incident_id: "INC-1042",
  title: "Database connection pool exhausted",
  description: "API latency increased to 4500ms and 500 errors spike to 12%",
  raw_logs: "pq: remaining connection slots are reserved",
});

console.log(`Route: ${res.routing_target}, Priority: ${res.priority}`);

// Batch evaluation
const batch = await client.batchTriage([
  { incident_id: "INC-1043", title: "Memory leak in worker pool" },
  { incident_id: "INC-1044", title: "Elevated cache miss rate" },
]);
```

### Python

```python
from gen.jev.incident.v1.incident_jev import (
    IncidentTriageServiceClient,
    TriageRequest,
)

client = IncidentTriageServiceClient()

# Single evaluation
res = client.triage(TriageRequest(
    incident_id="INC-1042",
    title="Database connection pool exhausted",
    description="API latency increased to 4500ms and 500 errors spike to 12%",
    raw_logs="pq: remaining connection slots are reserved",
))

print(f"Route: {res.routing_target}, Priority: {res.priority}, Urgency: {res.urgency_rating}")

# Batch evaluation
batch = client.batch_triage([
    TriageRequest(incident_id="INC-1043", title="Memory leak in worker pool"),
    TriageRequest(incident_id="INC-1044", title="Elevated cache miss rate"),
])
```

---

## Development & Testing

Manage the project using `just`:

```bash
# Build binary
just build

# Generate code from test protos and examples
just generate

# Re-generate Go bindings for jev/v1 options
just generate-options

# Run all unit tests, syntax checks, and golden file verification
just test

# Run end-to-end examples with live FauxRPC mock container
just run-examples

# Run linters (buf lint, go vet, golangci-lint)
just lint
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
