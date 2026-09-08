#!/usr/bin/env python3
"""Generate clean-room semantic forward oracles for the released schema-v3 decisions.

Inputs are the exact released decision documents.  Rejections are explicit
schema mutations (constant, required-member, or unknown-member violations),
so this tool never derives production outcomes or fabricates review evidence.
"""
import argparse, base64, hashlib, json
from pathlib import Path

CASES = {
    "cohesion-schema-v3-decision-review-v1": {
        "source": "release/cohesion-schema-v3-decision-review.json",
        "schema": "schema/cohesion-schema-v3-decision-review-v1.schema.json",
        "mutations": [("missing-decision", "decision", "schema-required-member"),
                      ("wrong-schema", "schema_id", "schema-constant"),
                      ("unknown-member", "__unknown__", "schema-unknown-member")],
    },
    "cohesion-schema-v3-decision-freeze-v1": {
        "source": "release/cohesion-schema-v3-decision-freeze.json",
        "schema": "schema/cohesion-schema-v3-decision-freeze-v1.schema.json",
        "mutations": [("missing-decision-review", "decision_review", "schema-required-member"),
                      ("wrong-schema", "schema_id", "schema-constant"),
                      ("unknown-member", "__unknown__", "schema-unknown-member")],
    },
}

def canonical(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()

def sha(value):
    return "sha256:" + hashlib.sha256(value).hexdigest()

def build(root, identity):
    spec = CASES[identity]
    raw = canonical(json.loads((root / spec["source"]).read_bytes()))
    base = json.loads(raw)
    rows = [("base.canonical-minimum", base, "accepted", sha(raw), None),
            ("base.canonical-rich", base, "accepted", sha(raw), None)]
    for suffix, key, error in spec["mutations"]:
        value = json.loads(raw)
        if key == "__unknown__":
            value[key] = None
        else:
            value.pop(key, None) if suffix.startswith("missing-") else value.update({key: "wrong"})
        rows.append((f"{identity}.{suffix}", value, "rejected", None, error))
    rows.sort(key=lambda row: row[0])
    cases = []
    for case_id, value, outcome, normalized, error in rows:
        candidate = canonical(value)
        source = {"kind": "fixture", "fixture_id": "base.canonical-rich"} if case_id not in {"base.canonical-minimum", "base.canonical-rich"} else {"kind": "fixture", "fixture_id": case_id}
        if case_id not in {"base.canonical-minimum", "base.canonical-rich"}:
            source = {"kind": "splice", "fixture_id": "base.canonical-rich", "splices": [{"offset": 0, "delete_count": len(raw), "insert_base64": base64.b64encode(candidate).decode()}]}
        cases.append({"case_id": case_id, "input": source, "outcome": outcome, "normalized_value_sha256": normalized, "error_code": error})
    return {"format": "golib-forward-oracle-v2", "fixture_count": 2,
            "fixtures": [{"fixture_id": n, "bytes_base64": base64.b64encode(raw).decode(), "bytes_sha256": sha(raw)} for n in ("base.canonical-minimum", "base.canonical-rich")],
            "case_count": len(cases), "cases": cases}

def main():
    parser = argparse.ArgumentParser(); parser.add_argument("--identity", choices=CASES, required=True); parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args(); root = Path(__file__).resolve().parents[2]
    args.output.parent.mkdir(parents=True, exist_ok=True); args.output.write_bytes(canonical(build(root, args.identity)))

if __name__ == "__main__": main()
