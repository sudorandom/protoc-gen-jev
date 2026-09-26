"""Shared response contract checks against generated Python clients."""
import json
import sys
from pathlib import Path
from typing import Any, cast

sys.path.insert(0, str(Path("gen").resolve()))
from typesafe_sdk import TypeSafeClient
from jev.ai.contract.v1.service_jev import ContractServiceClient
from jev.ai.contract.v1.contract_pb import EvaluateRequest, Container
from jev.ai.shared.v1.shared_pb import ExternalRequest

class Fake:
    def __init__(self, response: dict[str, Any], check_state: bool = True):
        self.response = response
        self.check_state = check_state

    def system_one(self, **kwargs: Any) -> dict[str, Any]:
        if self.check_state:
            assert kwargs["state"] == {"inputText":"hello", "inputCount":"5"}
        return self.response

cases = json.loads(Path("testdata/behavior/responses.json").read_text())
for tc in cases:
    client = ContractServiceClient(client=cast(TypeSafeClient, Fake(tc["response"])))
    try:
        result = client.evaluate_detailed(EvaluateRequest(input_text="hello", input_count=5))
    except (ValueError, TypeError):
        assert tc.get("error"), tc["name"]
        continue
    assert not tc.get("error"), tc["name"]
    assert json.loads(result.value.to_json()) == tc["expected"], (tc["name"], result.value)
    assert result.value.has_field("flag") and result.value.has_field("rating")
    assert result.response["model"] == "test-model"
    assert result.response["usage"] == tc["response"]["usage"]

client = ContractServiceClient(client=cast(TypeSafeClient, Fake({"answers":{"accepted":{"type":"noul", "noul":0.8}}}, False)))
assert client.external(ExternalRequest(input_text="hello")).accepted
assert client.nested(Container.NestedRequest(input_text="hello")).accepted
print(f"Python: {len(cases)} response fixtures and imported/nested RPCs passed")
