_T = TypeVar("_T", bound=Message)

@dataclass(frozen=True)
class Evaluation(Generic[_T]):
    value: _T
    # Original SDK response: answers, probabilities, confidence, model and usage.
    response: Any


def _questions(rules: list[dict[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for q in rules:
        if q["type"] == "choice":
            result[q["name"]] = Choice(instructions=q["instructions"], criteria=q["choices"])
        elif q["type"] == "score":
            result[q["name"]] = Score(instructions=q["instructions"], criteria=[l["description"] for l in q["levels"]])
        else:
            result[q["name"]] = Noul(instructions=q["instructions"])
    return result


def _get(obj: Any, key: str, default: Any = None) -> Any:
    return obj.get(key, default) if isinstance(obj, dict) else getattr(obj, key, default)


def _map_response(response: Any, rules: list[dict[str, Any]]) -> dict[str, Any]:
    values: dict[str, Any] = {}
    sentinel = object()
    answers = _get(response, "answers", sentinel)
    canonical = answers is not sentinel
    for q in rules:
        group = answers if canonical else _get(response, {"choice":"choices", "noul":"nouls", "score":"scores"}[q["type"]])
        if not isinstance(group, dict) or q["name"] not in group:
            raise ValueError(f"Missing answer for {q['name']}")
        item = group[q["name"]]
        kind = _get(item, "type", sentinel)
        if (canonical or kind is not sentinel) and kind != q["type"]:
            raise ValueError(f"{q['name']}: expected answer type {q['type']}")
        if q["type"] == "choice":
            label = _get(item, "choice")
            if not isinstance(label, str) or label not in q["choices"]:
                raise ValueError(f"{q['name']}: invalid choice")
            if q.get("oneof"):
                values[label] = base64.b64encode(label.encode("ascii")).decode("ascii") if q["oneof"][label] == "bytes" else label
            else:
                values[q["field"]] = label
        elif q["type"] == "noul":
            value = _get(item, "noul", sentinel)
            if value is sentinel:
                value = _get(item, "result")
            if isinstance(value, bool):
                values[q["field"]] = value
            elif isinstance(value, (int, float)) and math.isfinite(value) and 0 <= value <= 1:
                values[q["field"]] = value >= q["threshold"]
            else:
                raise ValueError(f"{q['name']}: invalid noul")
        else:
            pos = _get(item, "score")
            levels = q["levels"]
            if isinstance(pos, bool) or not isinstance(pos, (int, float)) or not math.isfinite(pos) or not 0 <= pos <= len(levels)-1:
                raise ValueError(f"{q['name']}: invalid score")
            i = int(pos)
            value = float(levels[i]["value"])
            if i + 1 < len(levels):
                value = (1 - (pos - i)) * value + (pos - i) * levels[i + 1]["value"]
            if not math.isfinite(value):
                raise ValueError(f"{q['name']}: non-finite domain value")
            if q["kind"].startswith(("int", "uint")):
                whole = math.floor(abs(value))
                value = math.copysign(whole + (1 if abs(value) - whole >= 0.5 else 0), value)
            values[q["field"]] = str(int(value)) if q["kind"] in ("int64", "uint64") else value
    return values
