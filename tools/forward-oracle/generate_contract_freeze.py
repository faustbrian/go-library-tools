#!/usr/bin/env python3
"""Generate the clean-room contract-freeze forward oracle.

The released contract-freeze document is the authoritative fixture. This
script only canonicalizes that document and applies explicit, reviewable
mutations; it does not import production Go code or create review evidence.
"""
import base64, copy, hashlib, json
from pathlib import Path

def canonical(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()

def sha(value):
    return "sha256:" + hashlib.sha256(value).hexdigest()

def main():
    root = Path(__file__).resolve().parents[2]
    source = root / "release/cohesion-contract-freeze.json"
    output = root / "testdata/cohesion/forward-oracles/cohesion-contract-freeze-v1-forward-oracle.json"
    raw = canonical(json.loads(source.read_bytes()))
    base = json.loads(raw)
    cases = {
        "base.canonical-minimum": (base, "accepted", sha(raw), None),
        "base.canonical-rich": (base, "accepted", sha(raw), None),
        "contract-freeze.valid": (base, "accepted", sha(raw), None),
    }
    mutations = [
        ("contract-freeze.invalid.schema", lambda v: v.update(schema_id="wrong"), "schema-constant"),
        ("contract-freeze.invalid.goal", lambda v: v.update(goal_id="wrong"), "schema-constant"),
        ("contract-freeze.invalid.contract", lambda v: v["contract"].update(tag="v1.5.4"), "semantic-external-identity"),
        ("contract-freeze.invalid.review-path", lambda v: v["review"].update(record_path="CONTRACT_REVIEW.json"), "semantic-external-identity"),
        ("contract-freeze.invalid.timestamp", lambda v: v.update(created_at="2026-02-31T03:32:14Z"), "semantic-cross-field"),
        ("contract-freeze.invalid.missing-review", lambda v: v.pop("review"), "schema-required-member"),
        ("contract-freeze.invalid.unknown-member", lambda v: v.update(unknown=None), "schema-unknown-member"),
    ]
    for case_id, mutate, error in mutations:
        value = copy.deepcopy(base)
        mutate(value)
        cases[case_id] = (value, "rejected", None, error)
    rows = []
    for case_id in sorted(cases):
        value, outcome, normalized, error = cases[case_id]
        candidate = canonical(value)
        input_value = {"kind": "fixture", "fixture_id": case_id} if case_id in {"base.canonical-minimum", "base.canonical-rich"} else {"kind": "splice", "fixture_id": "base.canonical-rich", "splices": [{
                "offset": 0, "delete_count": len(raw), "insert_base64": base64.b64encode(candidate).decode()
            }]}
        rows.append({
            "case_id": case_id,
            "input": input_value,
            "outcome": outcome,
            "normalized_value_sha256": normalized,
            "error_code": error,
        })
    payload = {
        "format": "golib-forward-oracle-v2", "fixture_count": 2,
        "fixtures": [{"fixture_id": n, "bytes_base64": base64.b64encode(raw).decode(), "bytes_sha256": sha(raw)}
                      for n in ("base.canonical-minimum", "base.canonical-rich")],
        "case_count": len(rows), "cases": rows,
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_bytes(canonical(payload))

if __name__ == "__main__":
    main()
