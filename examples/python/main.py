#!/usr/bin/env python3
"""protoc-gen-jev: Python Client End-to-End Example."""

import json
import os
import sys
from dataclasses import asdict
from pathlib import Path

# Add examples/gen/jev to module path
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "gen" / "jev"))

from incident.v1.incident_jev import (
    IncidentTriageServiceClient,
    TriageRequest,
    TriageResponse,
)
from typesafe_sdk import TypeSafeClient


def main():
    print("==================================================")
    print("  protoc-gen-jev: Python Client End-to-End Example")
    print("==================================================")

    api_key = os.getenv("TYPESAFE_API_KEY")
    custom_endpoint = os.getenv("JEV_ENDPOINT")

    if custom_endpoint:
        base_url = custom_endpoint.removesuffix("/v1/systemone").removesuffix("/")
        print(f"\n[INFO] Using custom Jev/Laya endpoint: {base_url}")
        sdk_client = TypeSafeClient(api_key=api_key or "local", base_url=base_url)
        client = IncidentTriageServiceClient(client=sdk_client)
    elif api_key:
        print("\n[INFO] Using live TypeSafe AI Jev SDK with TYPESAFE_API_KEY")
        client = IncidentTriageServiceClient(api_key=api_key)
    else:
        print("\n[INFO] Using default TypeSafe AI Jev SDK client")
        client = IncidentTriageServiceClient()

    # 1. Inspect generated questions
    questions = client.build_triage_questions()
    print(f"\n1. Built Jev Questions ({len(questions)} total):")
    for q_name, q in questions.items():
        crit = getattr(q, "criteria", None)
        print(f"  - {q_name} [{q.type}]: {q.instructions}")
        if crit:
            print(f"      criteria: {crit}")

    # 2. Evaluate single input state
    req = TriageRequest(
        incident_id="INC-8891",
        title="Database connection pool exhausted",
        description="API latency increased to 4500ms and 500 errors spike to 12%",
        raw_logs="Connection refused on port 5432 after 100 pool max connections",
    )

    print("\n2. Evaluating Single Incident State...")
    decision = client.triage(req)
    print("✔ Decision received:")
    print(json.dumps(asdict(decision), indent=2))

    # 3. Batch evaluation
    print("\n3. Batch Evaluating 3 Incident States...")
    batch_reqs = [
        TriageRequest(
            incident_id="INC-8892",
            title="Ingress 502 bad gateway spikes across region us-east-1",
            description="Edge proxy reports connection reset by peer from upstream cluster",
            raw_logs="HTTP 502 Bad Gateway - upstream connect error or disconnect/reset before headers",
        ),
        TriageRequest(
            incident_id="INC-8893",
            title="Low-priority deprecation warning logged in analytics service",
            description="Client library using deprecated v1 query endpoint; scheduled for removal in Q3",
            raw_logs="WARN [analytics-worker] Endpoint /v1/query is deprecated, migrate to /v2/query",
        ),
        TriageRequest(
            incident_id="INC-8894",
            title="Routine memory compaction completed without customer impact",
            description="Background compaction cycle reclaimed 4.2GB memory; latency within SLO",
            raw_logs="INFO [compactor] Compaction cycle finished in 45s, 0 errors, 4200MB reclaimed",
        ),
    ]

    batch_decisions = client.batch_triage(batch_reqs)
    print(f"✔ Successfully evaluated {len(batch_decisions)} batch items.")
    for i, d in enumerate(batch_decisions, start=1):
        print(
            f"  - Item [{i}]: RoutingTarget={d.routing_target}, "
            f"Paging={d.requires_immediate_paging}, Priority={d.priority}, "
            f"Urgency={d.urgency_rating}, BlastRadius={d.blast_radius_percentage}%"
        )

    print("\n✔ Python End-to-End Test PASSED successfully!")


if __name__ == "__main__":
    main()
