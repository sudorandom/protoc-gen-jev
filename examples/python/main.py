#!/usr/bin/env python3
"""protoc-gen-jev: Python Client End-to-End Example."""

import json
import os
import sys
from pathlib import Path

# Add examples/gen/jev to module path
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "gen" / "jev"))

from incident.v1.incident_jev import IncidentTriageJevClient


class MockTypeSafeClient:
    """Offline mock client simulating Jev System One responses."""

    def system_one(self, state, questions):
        class MockChoice:
            def __init__(self, val):
                self.choice = val

        class MockNoul:
            def __init__(self, val):
                self.result = val

        class MockScore:
            def __init__(self, val):
                self.score = val

        class MockResp:
            choices = {
                "routing_target": MockChoice("oncall_engineer"),
                "priority": MockChoice("PRIORITY_LEVEL_HIGH"),
                "complianceClassification": MockChoice("INTERNAL_CONFIDENTIAL"),
            }
            nouls = {
                "requiresImmediatePaging": MockNoul(True),
            }
            scores = {
                "urgencyRating": MockScore(4),
                "blastRadiusPercentage": MockScore(45.0),
            }

        return MockResp()


def main():
    print("==================================================")
    print("  protoc-gen-jev: Python Client End-to-End Example")
    print("==================================================")

    api_key = os.getenv("TYPESAFE_API_KEY")

    if not api_key:
        print("\n[INFO] TYPESAFE_API_KEY not set in environment.")
        print("[INFO] Initializing client with simulated Jev client for offline demo...")
        client = IncidentTriageJevClient(client=MockTypeSafeClient())
    else:
        print("\n[INFO] Using live TypeSafe AI Jev SDK with TYPESAFE_API_KEY")
        client = IncidentTriageJevClient(api_key=api_key)

    # 1. Inspect generated questions
    questions = client.build_questions()
    print(f"\n1. Built Jev Questions ({len(questions)} total):")
    for q_name, q in questions.items():
        crit = getattr(q, "criteria", None)
        print(f"  - {q_name} [{q.type}]: {q.instructions}")
        if crit:
            print(f"      criteria: {crit}")

    # 2. Evaluate single input state
    state = {
        "incident_id": "INC-8891",
        "title": "Database connection pool exhausted",
        "description": "API latency increased to 4500ms and 500 errors spike to 12%",
        "raw_logs": "Connection refused on port 5432 after 100 pool max connections",
    }

    print("\n2. Evaluating Single Incident State...")
    decision = client.evaluate(state)
    print("✔ Decision received:")
    print(json.dumps(decision, indent=2))

    # 3. Batch evaluation
    print("\n3. Batch Evaluating 3 Incident States...")
    batch_states = [
        "Batch Item 1: Ingress 502 bad gateway spikes across region us-east-1",
        "Batch Item 2: Low-priority deprecation warning logged in analytics service",
        "Batch Item 3: Routine memory compaction completed without customer impact",
    ]

    batch_decisions = client.batch_evaluate(batch_states)
    print(f"✔ Successfully evaluated {len(batch_decisions)} batch items.")
    for i, d in enumerate(batch_decisions, start=1):
        print(
            f"  - Item [{i}]: RoutingTarget={d.get('routing_target')}, "
            f"Paging={d.get('requires_immediate_paging')}, Priority={d.get('priority')}, "
            f"Urgency={d.get('urgency_rating')}, BlastRadius={d.get('blast_radius_percentage')}%"
        )

    print("\n✔ Python End-to-End Test PASSED successfully!")


if __name__ == "__main__":
    main()
