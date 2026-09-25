#!/usr/bin/env python3
"""protoc-gen-jev: Python Client End-to-End Example."""

import json
import os
import sys
from pathlib import Path

# Add gen/jev to module path
sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "gen" / "jev"))

from ai.rules.v1.rules_jev import RuleTestRecordJevClient


class MockTypeSafeClient:
    """Offline mock client simulating Jev System One responses."""

    def system_one(self, state, questions):
        class MockChoice:
            def __init__(self, val):
                self.choice = val

        class MockScore:
            def __init__(self, val):
                self.score = val

        class MockResp:
            choices = {
                "delivery_method": MockChoice("push_notification"),
                "executionMode": MockChoice("MODE_FAST"),
                "filteredMode": MockChoice("MODE_BALANCED"),
                "securityClearance": MockChoice("TOP_SECRET"),
                "decisionFlag": MockChoice("APPROVE"),
            }
            nouls = {}
            scores = {
                "ratingSmall": MockScore(5),
                "ratingStrict": MockScore(3),
                "discreteCode": MockScore(50),
                "largeScale": MockScore(500),
                "temperature": MockScore(25.0),
                "discreteRatio": MockScore(0.8),
                "customBoundedScore": MockScore(30.0),
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
        client = RuleTestRecordJevClient(client=MockTypeSafeClient())
    else:
        print("\n[INFO] Using live TypeSafe AI Jev SDK with TYPESAFE_API_KEY")
        client = RuleTestRecordJevClient(api_key=api_key)

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
        "user_id": "usr_10492",
        "request_type": "expedited_transfer",
        "priority": "high",
        "text": "Urgent security request: please authorize immediately and send via push notification",
    }

    print("\n2. Evaluating Single State...")
    decision = client.evaluate(state)
    print("✔ Decision received:")
    print(json.dumps(decision, indent=2))

    # 3. Batch evaluation
    print("\n3. Batch Evaluating 3 States...")
    batch_states = [
        "Batch Item 1: Standard alert",
        "Batch Item 2: High security alert",
        "Batch Item 3: Operational check",
    ]

    batch_decisions = client.batch_evaluate(batch_states)
    print(f"✔ Successfully evaluated {len(batch_decisions)} batch items.")
    for i, d in enumerate(batch_decisions, start=1):
        print(
            f"  - Item [{i}]: DeliveryMethod={d.get('delivery_method')}, "
            f"RatingSmall={d.get('rating_small')}, Clearance={d.get('security_clearance')}"
        )

    print("\n✔ Python End-to-End Test PASSED successfully!")


if __name__ == "__main__":
    main()
