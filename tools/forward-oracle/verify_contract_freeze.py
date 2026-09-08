#!/usr/bin/env python3
"""Independent structural checker for the contract-freeze oracle asset."""
import base64, hashlib, json
from pathlib import Path

def canonical(value): return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()
def sha(value): return "sha256:" + hashlib.sha256(value).hexdigest()

def main():
    path = Path(__file__).resolve().parents[2] / "testdata/cohesion/forward-oracles/cohesion-contract-freeze-v1-forward-oracle.json"
    data = path.read_bytes(); value = json.loads(data)
    if data != canonical(value): raise SystemExit("oracle is not canonical JSON")
    fixtures = value.get("fixtures", [])
    if value.get("fixture_count") != 2 or [x.get("fixture_id") for x in fixtures] != ["base.canonical-minimum", "base.canonical-rich"]:
        raise SystemExit("fixture roster mismatch")
    fixture_bytes = {}
    for fixture in fixtures:
        raw = base64.b64decode(fixture["bytes_base64"], validate=True)
        if raw != canonical(json.loads(raw.decode())) or fixture.get("bytes_sha256") != sha(raw):
            raise SystemExit("fixture canonical bytes or digest mismatch")
        fixture_bytes[fixture["fixture_id"]] = raw
    expected = ["base.canonical-minimum", "base.canonical-rich", "contract-freeze.invalid.contract", "contract-freeze.invalid.goal", "contract-freeze.invalid.missing-review", "contract-freeze.invalid.review-path", "contract-freeze.invalid.schema", "contract-freeze.invalid.timestamp", "contract-freeze.invalid.unknown-member", "contract-freeze.valid"]
    if value.get("case_count") != len(expected) or [x.get("case_id") for x in value.get("cases", [])] != expected:
        raise SystemExit("case roster mismatch")
    errors = {"contract-freeze.invalid.schema": "schema-constant", "contract-freeze.invalid.goal": "schema-constant", "contract-freeze.invalid.contract": "semantic-external-identity", "contract-freeze.invalid.review-path": "semantic-external-identity", "contract-freeze.invalid.timestamp": "semantic-cross-field", "contract-freeze.invalid.missing-review": "schema-required-member", "contract-freeze.invalid.unknown-member": "schema-unknown-member"}
    for row in value["cases"]:
        if set(row) != {"case_id", "input", "outcome", "normalized_value_sha256", "error_code"}: raise SystemExit("case row shape mismatch")
        inp = row["input"]
        if inp.get("fixture_id") not in fixture_bytes: raise SystemExit("unknown fixture")
        if inp.get("kind") == "fixture":
            if row["case_id"] not in {"base.canonical-minimum", "base.canonical-rich"} or inp.get("splices") is not None: raise SystemExit("fixture input mismatch")
            candidate = fixture_bytes[inp["fixture_id"]]
        elif inp.get("kind") == "splice":
            splice = inp.get("splices", [])
            if len(splice) != 1: raise SystemExit("splice count mismatch")
            src = fixture_bytes[inp["fixture_id"]]; item = splice[0]; inserted = base64.b64decode(item["insert_base64"], validate=True)
            if item["offset"] != 0 or item["delete_count"] != len(src): raise SystemExit("splice bounds mismatch")
            candidate = inserted
        else: raise SystemExit("input kind mismatch")
        if row["outcome"] == "accepted":
            if row["error_code"] is not None or row["normalized_value_sha256"] != sha(candidate): raise SystemExit("accepted result mismatch")
        elif row["error_code"] != errors.get(row["case_id"]) or row["normalized_value_sha256"] is not None: raise SystemExit("rejected result mismatch")
        else: continue
        if row["case_id"] in errors: raise SystemExit("unexpected accepted result")
    print("contract-freeze forward oracle verified")

if __name__ == "__main__": main()
