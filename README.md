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

Add `protoc-gen-jev` to your `buf.gen.yaml`:

```yaml
version: v2
plugins:
  - local: protoc-gen-jev
    out: gen/jev
    opt:
      # Options: go, ts, py, json, or all (comma or plus separated)
      - targets=all
```

## Custom Field Options

Import `jev/v1/options.proto` to customize how fields map to Jev questions:

```protobuf
syntax = "proto3";

package example.v1;

import "jev/v1/options.proto";

message UserIntent {
  // Override prompt instructions sent to Jev
  string churn_risk = 1 [(jev.v1.field).instructions = "Assess customer cancellation risk"];

  // Skip evaluating this field with Jev
  string raw_metadata = 2 [(jev.v1.field).skip = true];
}
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

