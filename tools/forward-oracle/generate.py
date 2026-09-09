#!/usr/bin/env python3
"""Clean-room diagnostic forward-oracle generator.

This intentionally does not import the production cohesion package.  It emits
the two frozen fixtures and the six schema-level diagnostic mutations from the
schema-v3 decision matrix, using only Python's independent JSON encoder and
SHA-256 implementation.
"""
import base64
import copy
import hashlib
import json
from pathlib import Path


def canonical(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()


def digest(value):
    return "sha256:" + hashlib.sha256(value).hexdigest()


def main():
    root = Path(__file__).resolve().parents[2]
    output = root / "testdata/cohesion/forward-oracles/cohesion-diagnostic-v1-forward-oracle.json"
    minimum = {"schema_id": "urn:golib:cohesion:diagnostic:v1", "schema_version": 1,
               "status": "failed", "diagnostics": [{"code": "failure", "path": "", "message": "failed"}]}
    rich = {"schema_id": "urn:golib:cohesion:diagnostic:v1", "schema_version": 1,
            "status": "failed", "diagnostics": [{"code": "failure", "path": "/x",
            "message": "failed", "causes": [{"code": "cause", "path": "", "message": "cause"}]}]}
    fixtures = [("base.canonical-minimum", canonical(minimum)), ("base.canonical-rich", canonical(rich))]
    rows = []
    rows.append(("base.canonical-minimum", fixtures[0][1], "accepted", digest(fixtures[0][1]), None))
    rows.append(("base.canonical-rich", fixtures[1][1], "accepted", digest(fixtures[1][1]), None))
    mutations = [
        ("diagnostic.invalid.code", lambda v: v["diagnostics"][0].update(code="BAD"), "schema-pattern"),
        ("diagnostic.invalid.empty", lambda v: v.update(diagnostics=[]), "schema-range"),
        ("diagnostic.invalid.empty-causes", lambda v: v["diagnostics"][0].update(causes=[]), "schema-range"),
        ("diagnostic.invalid.message", lambda v: v["diagnostics"][0].update(message=""), "schema-range"),
        ("diagnostic.invalid.path", lambda v: v["diagnostics"][0].update(path="not-pointer"), "schema-pattern"),
        ("diagnostic.invalid.unknown-cause", lambda v: v["diagnostics"][0].update(extra=True), "schema-unknown-member"),
    ]
    for case_id, mutate, code in mutations:
        value = copy.deepcopy(rich)
        mutate(value)
        rows.append((case_id, canonical(value), "rejected", None, code))
    encoded_fixtures = [{"fixture_id": name, "bytes_base64": base64.b64encode(data).decode(), "bytes_sha256": digest(data)} for name, data in fixtures]
    cases = []
    for index, (case_id, data, outcome, normalized, error) in enumerate(rows):
        if index < 2:
            source = {"kind": "fixture", "fixture_id": fixtures[index][0]}
        else:
            source = {"kind": "splice", "fixture_id": "base.canonical-rich", "splices": [{"offset": 0, "delete_count": len(fixtures[1][1]), "insert_base64": base64.b64encode(data).decode()}]}
        cases.append({"case_id": case_id, "input": source, "outcome": outcome,
                      "normalized_value_sha256": normalized, "error_code": error})
    payload = {"format": "golib-forward-oracle-v2", "fixture_count": 2, "fixtures": encoded_fixtures,
               "case_count": len(cases), "cases": cases}
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_bytes(canonical(payload))


if __name__ == "__main__":
    main()
