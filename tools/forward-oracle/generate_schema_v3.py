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
        "schema_id": "urn:golib:cohesion:schema-v3-decision-review:v1",
        "decision_digest": "sha256:cfa83030b1b05535148292c4db063edaa35aca861e957ecd5579008090059e2f",
        "coverage": {"accepted": True, "rejected": True, "partial": True},
        "mutations": [("missing-decision", "decision", "schema-required-member"),
                      ("wrong-schema", "schema_id", "schema-constant"),
                      ("unknown-member", "__unknown__", "schema-unknown-member")],
    },
    "cohesion-schema-v3-decision-freeze-v1": {
        "source": "release/cohesion-schema-v3-decision-freeze.json",
        "schema": "schema/cohesion-schema-v3-decision-freeze-v1.schema.json",
        "schema_id": "urn:golib:cohesion:schema-v3-decision-freeze:v1",
        "decision_digest": "sha256:cfa83030b1b05535148292c4db063edaa35aca861e957ecd5579008090059e2f",
        "coverage": {"accepted": True, "rejected": True, "partial": True},
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
    source_path = root / spec["source"]
    source_bytes = source_path.read_bytes()
    source_value = json.loads(source_bytes)
    raw = canonical(source_value)
    # Released records may carry a terminal newline; the fixture itself is
    # canonicalized below, while JSON validity is still required here.
    if source_value.get("schema_id") != spec["schema_id"]:
        raise ValueError(f"{spec['source']} schema identity mismatch")
    if source_value.get("decision", {}).get("bytes_sha256") != spec["decision_digest"]:
        raise ValueError(f"{spec['source']} decision digest mismatch")
    schema_value = json.loads((root / spec["schema"]).read_bytes())
    if schema_value.get("$id", "").rsplit("/", 1)[-1].removesuffix(".schema.json") != spec["schema"].rsplit("/", 1)[-1].removesuffix(".schema.json"):
        raise ValueError(f"{spec['schema']} identity mismatch")
    if not all(spec["coverage"].values()):
        raise ValueError(f"{identity} lane coverage metadata is incomplete")
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
    expected_suffixes = {"base.canonical-minimum", "base.canonical-rich"} | {f"{identity}.{item[0]}" for item in spec["mutations"]}
    if {row[0] for row in rows} != expected_suffixes or not spec["coverage"]["partial"]:
        raise ValueError(f"{identity} lane roster or partial coverage metadata is invalid")
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
