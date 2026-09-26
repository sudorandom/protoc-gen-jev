#!/usr/bin/env python3
"""protoc-gen-jev: Python Client End-to-End Example."""

import json
import os
import sys
from pathlib import Path

# Add examples/gen/jev to module path
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "gen" / "jev"))

from incident.v1.incident_jev import IncidentTriageJevClient
from typesafe_sdk import TypeSafeClient


class MockTypeSafeClient:
    """Offline mock client simulating Jev System One responses when Docker is unavailable."""

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
    container = None

    if api_key:
        print("\n[INFO] Using live TypeSafe AI Jev SDK with TYPESAFE_API_KEY")
        client = IncidentTriageJevClient(api_key=api_key)
    else:
        # Try starting FauxRPC Testcontainer with Jev OpenAPI spec
        abs_openapi = str((Path(__file__).resolve().parents[2] / "testdata" / "openapi").resolve())
        try:
            if "DOCKER_HOST" not in os.environ:
                home = os.environ.get("HOME", "")
                colima_sock = Path(home) / ".colima/default/docker.sock"
                if colima_sock.exists():
                    os.environ["DOCKER_HOST"] = f"unix://{colima_sock}"
            if "TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE" not in os.environ:
                os.environ["TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE"] = "/var/run/docker.sock"

            from testcontainers.core.config import testcontainers_config
            # Guard against invalid tc.host pointing to non-existent local socket
            tc_host = testcontainers_config.tc_properties.get("tc.host")
            if tc_host and tc_host.startswith("unix://") and not Path(tc_host[7:]).exists():
                testcontainers_config.tc_properties.pop("tc.host", None)

            from testcontainers.core.container import DockerContainer

            print("\n[INFO] Starting FauxRPC Testcontainer with Jev OpenAPI spec...")
            container = (
                DockerContainer("docker.io/sudorandom/fauxrpc:v0.29.1")
                .with_volume_mapping(abs_openapi, "/openapi")
                .with_command(
                    "run --schema=/openapi/typesafe-jev.yaml --stubs=/openapi/stubs.jev.yaml --addr=0.0.0.0:6660"
                )
                .with_exposed_ports(6660)
            )
            container.start()
            port = container.get_exposed_port(6660)
            print(f"[INFO] FauxRPC mock server running on port {port}")

            sdk_client = TypeSafeClient(api_key="mock", base_url=f"http://localhost:{port}")
            client = IncidentTriageJevClient(client=sdk_client)
        except Exception as err:
            print(f"\n[INFO] Testcontainers unavailable ({err}), falling back to in-memory mock...")
            client = IncidentTriageJevClient(client=MockTypeSafeClient())

    try:
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
    finally:
        if container:
            container.stop()


if __name__ == "__main__":
    main()
