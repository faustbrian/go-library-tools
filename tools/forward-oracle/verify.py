#!/usr/bin/env python3
"""Independent structural and digest checker for the diagnostic oracle."""
import base64
import hashlib
import json
from pathlib import Path


def canonical(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()


def sha(data):
    return "sha256:" + hashlib.sha256(data).hexdigest()


def main():
    path = Path(__file__).resolve().parents[2] / "testdata/cohesion/forward-oracles/cohesion-diagnostic-v1-forward-oracle.json"
    data = path.read_bytes()
    if data != canonical(json.loads(data)):
        raise SystemExit("oracle is not canonical JSON")
    value = json.loads(data)
    if value["format"] != "golib-forward-oracle-v2" or value["fixture_count"] != 2:
        raise SystemExit("oracle header mismatch")
    fixtures = {row["fixture_id"]: base64.b64decode(row["bytes_base64"], validate=True) for row in value["fixtures"]}
    for row in value["fixtures"]:
        if sha(fixtures[row["fixture_id"]]) != row["bytes_sha256"]:
            raise SystemExit("fixture digest mismatch")
    expected = ["base.canonical-minimum", "base.canonical-rich", "diagnostic.invalid.code", "diagnostic.invalid.empty", "diagnostic.invalid.empty-causes", "diagnostic.invalid.message", "diagnostic.invalid.path", "diagnostic.invalid.unknown-cause"]
    if value["case_count"] != len(expected) or [row["case_id"] for row in value["cases"]] != expected:
        raise SystemExit("case roster mismatch")
    for row in value["cases"]:
        if row["input"]["kind"] == "fixture":
            candidate = fixtures[row["input"]["fixture_id"]]
        else:
            source = fixtures[row["input"]["fixture_id"]]
            splice = row["input"]["splices"][0]
            insert = base64.b64decode(splice["insert_base64"], validate=True)
            candidate = source[:splice["offset"]] + insert + source[splice["offset"] + splice["delete_count"]:]
        if row["outcome"] == "accepted" and row["normalized_value_sha256"] != sha(candidate):
            raise SystemExit("accepted digest mismatch")
        if row["outcome"] == "rejected" and row["normalized_value_sha256"] is not None:
            raise SystemExit("rejected row carries normalized digest")
    print("diagnostic forward oracle verified")


if __name__ == "__main__":
    main()
