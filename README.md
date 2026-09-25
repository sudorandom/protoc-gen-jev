# protoc-gen-jev

> [!WARNING]
> **Experimental**: `protoc-gen-jev` is currently experimental and under active development. APIs, question schema formats, and generated code templates are subject to change without notice.

A Protobuf compiler plugin for **TypeSafe AI's Jev** (System One fast cognitive model).

`protoc-gen-jev` inspects Protobuf messages and automatically generates typed Jev client code, question schemas, and decision structures across multiple languages.

## Generated Targets

| Target | File Suffix | Purpose |
| :--- | :--- | :--- |
| **Go** | `*_jev.pb.go` | Native typed Go client with HTTP execution against the Jev API |
| **TypeScript** | `*_jev.ts` | Typed TypeScript client wrapping `@typesafe-ai/sdk` |
| **Python** | `*_jev.py` | Python SDK client wrapping `typesafe-sdk` |
| **JSON** | `*_<Message>.jev.json` | Language-agnostic question schema definitions |

## Usage with Buf

### 1. Depend on `jev/v1/options.proto`

To import `jev/v1/options.proto` into your schemas, add the dependency to your `buf.yaml` (see [`examples/buf.yaml`](examples/buf.yaml)):

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

## Schema Definition & Field Options

Import `jev/v1/options.proto` to configure how Protobuf fields map to Jev cognitive decisions. `protoc-gen-jev` is completely self-contained and has zero external dependencies (no `protovalidate` required).

### Comprehensive Example

The example below models an **Incident Triage & Routing** system demonstrating all supported primitives and options (see [`examples/proto/incident/v1/incident.proto`](examples/proto/incident/v1/incident.proto)):

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

message IncidentTriage {
  // 1. Oneof mutual exclusion -> Jev Choice question
  // Jev selects the single most appropriate routing target based on input context.
  oneof routing_target {
    string automated_runbook = 1;
    string oncall_engineer = 2;
    string incident_commander = 3;
  }

  // 2. Boolean field -> Jev Noul question
  // Evaluates binary intent with a calibrated confidence threshold (0.0 - 1.0).
  bool requires_immediate_paging = 4 [
    (jev.v1.field).instructions = "Does this incident indicate active user-facing outage requiring paging?",
    (jev.v1.field).noul = { threshold: 0.85 }
  ];

  // 3. Enum field with criteria guidance -> Jev Choice question
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

  // 4. Bounded integer range (<= 10) -> Jev Score with exact sequence rubric [1, 2, 3, 4, 5]
  int32 urgency_rating = 6 [
    (jev.v1.field).instructions = "How urgently does this issue need to be resolved?",
    (jev.v1.field).score = { min: 1, max: 5 }
  ];

  // 5. Continuous numeric scale -> Jev Score with 5-tier interpolated rubric [0.0, 25.0, 50.0, 75.0, 100.0]
  float blast_radius_percentage = 7 [
    (jev.v1.field).instructions = "Estimated percentage of production infrastructure impacted",
    (jev.v1.field).score = { min: 0.0, max: 100.0 }
  ];

  // 6. String with discrete allowed choices -> Jev Choice question
  string compliance_classification = 8 [
    (jev.v1.field).instructions = "Categorize any sensitive or regulated data exposed in the incident report",
    (jev.v1.field).choice = {
      choices: ["PUBLIC", "INTERNAL_CONFIDENTIAL", "RESTRICTED_PII", "PCI_DSS"]
    }
  ];

  // 7. Skipped fields -> Excluded from Jev decision payload
  // Ideal for freeform logs, identifiers, or downstream output fields.
  string incident_id = 9 [(jev.v1.field).skip = true];
  string raw_stack_trace = 10 [(jev.v1.field).skip = true];
}
```

### Options Reference (`(jev.v1.field)`)

Options are grouped by Jev's cognitive primitives (`choice`, `score`, `noul`):

| Option | Type | Description |
| :--- | :--- | :--- |
| `instructions` | `string` | Custom instructions / prompt sent to Jev. Defaults to the field's Protobuf doc comment if omitted. |
| `skip` | `bool` | When `true`, completely skips this field during Jev evaluation (e.g. for raw text, IDs, or timestamps). |
| **`choice`** | `ChoiceRules` | Configures discrete selection choices. |
| `choice.choices` | `repeated string` | Explicit whitelist of allowed choice values (for strings or enums). |
| `choice.not_in` | `repeated string` | Blacklist of enum value names to exclude from choices. |
| `choice.criteria` | `map<string, string>` | Guidance or descriptive rubric criteria attached to specific options (labels $\to$ descriptions). When placed on a `string` field without `choices`, map keys define the choices. |
| **`score`** | `ScoreRules` | Configures continuous or rubric score evaluations. |
| `score.min` | `float` | Lower bound of rubric scale (default: `0.0`). |
| `score.max` | `float` | Upper bound of rubric scale (default: `100.0`). |
| `score.scale` | `repeated float` | Explicit discrete rubric tiers (e.g. `[1, 2, 3]` or `[10, 20, 50, 100]`). |
| **`noul`** | `NoulRules` | Configures binary yes/no intent questions. |
| `noul.threshold` | `float` | Confidence threshold `[0.0 - 1.0]` required to evaluate as `true`. |

### Type & Primitive Mapping

| Protobuf Feature | Configured Option | Jev Primitive | Behavior |
| :--- | :--- | :--- | :--- |
| `oneof` | *(automatic)* | **`Choice`** | Field names become the selectable choice options. |
| `bool` | *(optional)* `noul` | **`Noul`** | Binary yes/no classification with calibrated probability and optional threshold. |
| `enum` | *(automatic)* | **`Choice`** | All defined enum values (excluding `0` / `_UNSPECIFIED`) become choices. |
| `enum` | `choice.choices` | **`Choice`** | Restricts choices to the specified whitelist of enum value names. |
| `enum` | `choice.not_in` | **`Choice`** | Excludes specified enum value names from the choices. |
| `string` | `choice.choices` | **`Choice`** | Discrete allowed strings become the choices. |
| `string` | `choice.criteria` | **`Choice`** | Criteria map keys become choices with accompanying descriptive guidance. |
| Numeric | `score.scale` | **`Score`** | Discrete numbers become the ordered rubric tiers. |
| Numeric | `score.min` / `max` | **`Score`** | Small integer ranges ($\le 10$) expand to exact sequence `[min..max]`; large or continuous ranges interpolate into 5 tiers. |

### Generated Client Usage

#### Go
```go
import incidentv1 "your_project/gen/jev/incident/v1"

client := incidentv1.NewIncidentTriageJevClient(os.Getenv("TYPESAFE_API_KEY"))

// Single evaluation
decision, err := client.Evaluate(ctx, map[string]any{
    "title": "Database connection pool exhausted",
    "description": "API latency increased to 4500ms and 500 errors spike to 12%",
})
fmt.Printf("Route To: %s, Priority: %s, Urgency: %v\n", 
    decision.RoutingTarget, decision.Priority, decision.UrgencyRating)

// Batch evaluation
decisions, err := client.BatchEvaluate(ctx, []any{"alert 1", "alert 2"})
```

#### TypeScript
```typescript
import { IncidentTriageJevClient } from "./gen/jev/incident/v1/incident_jev";

const client = new IncidentTriageJevClient();

const decision = await client.evaluate({
  title: "Database connection pool exhausted",
  description: "API latency increased to 4500ms",
});
console.log(`Route To: ${decision.routing_target}, Priority: ${decision.priority}`);
```

#### Python
```python
from gen.jev.incident.v1.incident_jev import IncidentTriageJevClient

client = IncidentTriageJevClient()

decision = client.evaluate({
    "title": "Database connection pool exhausted",
    "description": "API latency increased to 4500ms",
})
print(f"Route To: {decision['routing_target']}, Priority: {decision['priority']}")
```

## Development & Testing

Manage the project using `just`:

```bash
# Build binary
just build

# Generate code from protos
just generate

# Run unit tests and golden file verification
just test

# Update golden files after intentional changes
just update-golden

# Run linters (buf lint, go vet, golangci-lint)
just lint
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

