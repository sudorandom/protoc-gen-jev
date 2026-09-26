# protoc-gen-jev

[![Go CI](https://img.shields.io/github/actions/workflow/status/sudorandom/protoc-gen-jev/go.yml?label=Go%20CI)](https://github.com/sudorandom/protoc-gen-jev/actions/workflows/go.yml)

> **Experimental:** generated APIs and schema options may change. Pin the generator and runtime versions together.

`protoc-gen-jev` turns Protobuf decision contracts into Jev question schemas and clients for Go, TypeScript, and Python. Keep request types, response types, instructions, and scoring rubrics in one versioned `.proto` definition.

TypeSafe's official SDKs already provide typed questions and answers. This plugin adds shared Protobuf contracts and mappings to native application messages. It does not guarantee model accuracy or eliminate the need for runtime validation.

## What it generates

| Target | Output | Runtime dependencies |
| --- | --- | --- |
| Go | `*_jev.pb.go` | Protobuf Go runtime and `github.com/sudorandom/protoc-gen-jev/pkg/jev` |
| TypeScript | `*_jev.ts` | `@bufbuild/protobuf` and `@typesafe-ai/sdk` |
| Python | `*_jev.py` | `protobuf` and `typesafe-sdk`; `.pyi` bindings for static checking |
| JSON | `*.jev.json` | None; an intermediate specification containing questions and mapping metadata |

The plugin generates wrappers around native Protobuf messages. Run the corresponding Protobuf generators as well. JSON specifications are not directly usable as Jev HTTP request bodies; `BuildQuestions`/`buildQuestions`/`build_questions` produces the API question map.

Protobuf services describe request-to-decision methods. These clients call the Jev HTTP endpoint; they are not gRPC servers or clients. Standalone message clients accept unstructured state and return the corresponding decision message.

## Installation and generation

```sh
go install github.com/sudorandom/protoc-gen-jev@latest
protoc-gen-jev --version
```

Go 1.27 or newer is required. For reproducible builds, replace `latest` with a release tag and use the same version for the Go runtime dependency.

Add the options schema to `buf.yaml`:

```yaml
version: v2
deps:
  - buf.build/sudo-random/protoc-gen-jev
```

Run `buf dep update`. A complete multi-language `buf.gen.yaml` needs companion generators:

```yaml
version: v2
plugins:
  - local: protoc-gen-go
    out: gen
    opt: paths=source_relative
  - local: protoc-gen-es
    out: gen
    opt: target=ts
  - remote: buf.build/protocolbuffers/python:v29.3
    out: gen
  - local: protoc-gen-jev
    out: gen
    opt:
      - paths=source_relative
      - targets=all
```

Install the local Go and protobuf-es plugins separately. Set each schema's `go_package` to the actual import path of its generated Go package, including your module prefix. Imported messages must also have their native bindings generated. This repository's complete configurations are [buf.gen.yaml](buf.gen.yaml) and [examples/buf.gen.yaml](examples/buf.gen.yaml).

Targets are `go`, `ts`/`typescript`, `python`/`py`, `json`, or `all`. Unknown targets are errors. JSON specifications accompany language output. Because protoc separates options on commas, repeat the `targets` option to combine targets (for example `targets=go,targets=ts`), or use `all`.

Python typing stubs are generated with `protoc --pyi_out=...` alongside `--python_out=...`. The repository uses a descriptor set and [generate-python-stubs.py](scripts/generate-python-stubs.py) to do this reproducibly. Put the generated root on `PYTHONPATH` when importing clients.

## Decision schema

```protobuf
syntax = "proto3";
package triage.v1;
option go_package = "example.com/myapp/gen/triage/v1;triagev1";
import "jev/v1/options.proto";

message TriageRequest {
  string description = 1;
}

message TriageResponse {
  optional bool page = 1 [
    (jev.v1.field).instructions = "Does this incident require immediate paging?",
    (jev.v1.field).noul = { threshold: 0.85 }
  ];
  int32 urgency = 2 [
    (jev.v1.field).instructions = "How urgently does this incident need attention?",
    (jev.v1.field).score = {
      levels: [
        { value: 1, description: "Minor; can wait for routine maintenance" },
        { value: 3, description: "Degraded service; a workaround exists" },
        { value: 5, description: "Active outage; immediate intervention required" }
      ]
    }
  ];
}

service TriageService {
  rpc Triage(TriageRequest) returns (TriageResponse);
}
```

The [incident example](examples/proto/incident/v1/incident.proto) also demonstrates enum criteria, string choices, and routing oneofs.

### Options

| Option | Behavior |
| --- | --- |
| `instructions` | Overrides the field comment; otherwise a generated prompt is used |
| `skip` | Excludes the field or oneof member from decision questions |
| `choice.choices` | Allowed string labels or enum names |
| `choice.not_in` | Excludes enum names |
| `choice.criteria` | Descriptions for allowed labels; cannot bypass filtering |
| `score.levels` | 2–10 ordered `{ value, description }` rubric levels |
| `noul.threshold` | Optional probability threshold in `[0, 1]`; unset defaults to `0.5`, explicit zero is preserved |

`score.min`, `score.max`, and `score.scale` are not supported. Level values must be finite and strictly increasing. Integer fields require integral level values within the supported numeric domain. Provide domain-specific descriptions; numerical labels alone make weak rubrics.

An RPC is discovered when its request or response has a Jev field annotation. Set `(jev.v1.service).enabled = true` or `(jev.v1.method).enabled = true` to opt in without field annotations. Explicit false disables the service or method. Standalone message clients are independently generated for messages containing naturally mapped decision fields, including nested message declarations; disabling an RPC does not disable standalone clients.

## Supported schema boundaries

| Feature | Support |
| --- | --- |
| Multiple files per package; imported request/response types | Supported |
| Nested message declarations and imported enums | Supported |
| Optional scalar decisions and custom JSON names | Supported; native Protobuf conversion preserves presence |
| Bool | Noul |
| Enum | Choice; zero/`_UNSPECIFIED` values excluded |
| String | Choice when explicitly configured; otherwise ignored in responses |
| Signed, unsigned, fixed and floating point numbers | Score |
| Decision oneof | String/bytes members only; selected label stored as the member value |
| Other oneof member kinds | Rejected during generation |
| Request fields | Scalars, enums and optional scalar fields |
| Repeated, map, message or bytes request fields | Rejected |
| Complex response fields | Ignored when unannotated; rejected when annotated unless skipped |
| Streaming RPCs | Rejected |

Integer score level values for 64-bit fields are restricted to `±(2^53−1)` (unsigned: `0..2^53−1`) so interpolation is consistent across languages. Native 64-bit responses still use Go integers, Python integers and TypeScript `bigint`. Request integers retain their full Protobuf range.

A decision oneof selects a route; it does not generate arbitrary member content. Prefer an enum if only a routing label is needed.

## Client use

Complete, compiled examples:

- [Go](examples/go/main.go)
- [TypeScript](examples/typescript/index.ts)
- [Python](examples/python/main.py)

Normal methods return native Protobuf messages. Additional detailed methods retain provider metadata:

| Operation | Go | TypeScript | Python |
| --- | --- | --- | --- |
| Service call | `Triage` | `triage` | `triage` |
| With metadata | `TriageDetailed` | `triageDetailed` | `triage_detailed` |
| Standalone call | `Evaluate` | `evaluate` | `evaluate` |
| With metadata | `EvaluateDetailed` | `evaluateDetailed` | `evaluate_detailed` |

Detailed results contain `Value`/`value` and `Response`/`response`. Go's response contains model, usage, canonical answers and the original raw JSON; TypeScript and Python retain the complete SDK response. Probabilities and confidence remain accessible through the answers. Each result belongs to its call; clients do not store a shared last response.

TypeScript requests are created with `create(TriageRequestSchema, {...})`, imported from the native `_pb` module. Python returns Protobuf messages, not dataclasses. Only the detailed evaluation envelope is a dataclass.

### Conversion and error contract

Every requested decision must be present. Unknown choice labels, wrong answer types, missing/null values, and out-of-range scores or probabilities cause an error. Canonical `answers` require the correct `type` discriminator. Legacy `choices`/`scores`/`nouls` groups are accepted when `answers` is absent; an invalid canonical answer never falls back to a legacy group. Boolean Noul values are accepted for compatibility with backends that have already thresholded them.

Scores are zero-based rubric positions. A fractional position is linearly interpolated between adjacent domain values. Integer results round halfway **away from zero** in all languages. This conversion is not the same as computing an expected domain value from the full distribution when levels are unevenly spaced; use detailed probabilities if that is what your application needs.

Protobuf requests use standard Protobuf JSON names in every language. Standalone raw JSON is passed as supplied. Generated batch methods execute sequentially, preserve order, and stop on the first failure. They do not return partial results or start later requests after failure.

Go supports context cancellation, an injectable HTTP client, a default 30-second timeout, and a 16 MiB response limit. Set `Endpoint` and `Model` on the client. Go does not automatically retry; TypeScript/Python use their SDK's retry configuration. Inject a configured SDK client to control endpoint, model, timeouts, retry policy and lifetime.

## Validation and evaluation

```sh
just test          # generation, native compilation, strict TS, Python typing,
                   # shared response fixtures, golden comparison, Go + Docker tests
just lint
just run-examples  # all three languages against FauxRPC unless a backend is configured
```

The shared [response fixtures](testdata/behavior/responses.json) exercise the same success and failure cases in all three languages. Generation starts with clean output directories, and golden verification checks both unexpected and missing files. Dependency versions used for Python checks are locked in `requirements-dev.txt`; generator versions are pinned in `mise.toml`.

Mocks establish transport and mapping behavior. They do not measure model quality. To evaluate a real Jev or compatible Laya backend using a labelled dataset:

```sh
# Uses TYPESAFE_API_KEY for the hosted service; or pass --endpoint for a local server.
.venv/bin/python scripts/evaluate.py --dataset testdata/evaluation/incidents.json --output evaluation.json
```

The harness records exact-match accuracy against expected fields, per-field accuracy, latency, model identity, token usage and raw answers. The bundled labels are illustrative, not a validated benchmark. Use representative held-out examples, inspect uncertain answers, and compare reports before changing rubrics, thresholds or model versions. No live quality or backend-compatibility result is implied by a passing mock suite.

For local Laya, use `just start-laya`, then run the evaluation with `--endpoint http://127.0.0.1:8000`. SDK endpoints are base URLs; the Go client uses the full `/v1/systemone` URL. [TypeSafe's scoring guidance](https://docs.typesafe.ai/primitives/score) explains rubric design and confidence.

## Development

`just generate` refreshes clients and Python stubs. `just generate-options` refreshes the options Go binding. `just update-golden` intentionally replaces expected output. `just update-python-lock` resolves development dependencies after an intentional dependency update.

MIT licensed; see [LICENSE](LICENSE).
