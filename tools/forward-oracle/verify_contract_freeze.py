#!/usr/bin/env python3
"""Independent structural and mutation-semantic checker for the oracle asset."""
import base64, hashlib, json
from pathlib import Path

def canonical(value): return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()
def sha(value): return "sha256:" + hashlib.sha256(value).hexdigest()

def changed_paths(before, after, prefix=""):
    if isinstance(before, dict) and isinstance(after, dict):
        paths = []
        for key in sorted(set(before) | set(after)):
            path = f"{prefix}/{key}"
            if key not in before or key not in after:
                paths.append(path)
            else:
                paths.extend(changed_paths(before[key], after[key], path))
        return paths
    return [] if before == after else [prefix]

def check_candidate(case_id, candidate, base):
    """Check the frozen contract shape and each declared mutation directly."""
    if not isinstance(candidate, dict): raise SystemExit("candidate is not an object")
    required = {"schema_id", "goal_id", "contract", "review", "created_at"}
    if case_id in {"base.canonical-minimum", "base.canonical-rich", "contract-freeze.valid"}:
        if set(candidate) != required: raise SystemExit("accepted candidate shape mismatch")
        return
    if case_id.endswith("unknown-member"):
        if "unknown" not in candidate: raise SystemExit("unknown-member mutation missing")
    elif case_id.endswith("missing-review"):
        if "review" in candidate: raise SystemExit("missing-review mutation retained review")
    else:
        if not required.issubset(candidate): raise SystemExit("rejected candidate missing required field")
    if case_id.endswith(".schema") and candidate.get("schema_id") == "urn:golib:cohesion:contract-freeze:v1": raise SystemExit("schema mutation was not applied")
    if case_id.endswith(".goal") and candidate.get("goal_id") == "golib-cohesion-v1": raise SystemExit("goal mutation was not applied")
    if case_id.endswith(".contract") and candidate.get("contract", {}).get("tag") == "v1.5.5": raise SystemExit("contract mutation was not applied")
    if case_id.endswith("review-path") and candidate.get("review", {}).get("record_path") == ".ai/cohesion/phase3/contract-releases/v1.5.5/CONTRACT_REVIEW.json": raise SystemExit("review-path mutation was not applied")
    if case_id.endswith("timestamp") and candidate.get("created_at") != "2026-02-31T03:32:14Z": raise SystemExit("timestamp mutation was not applied")
    expected_path = {
        "contract-freeze.invalid.schema": "/schema_id",
        "contract-freeze.invalid.goal": "/goal_id",
        "contract-freeze.invalid.contract": "/contract/tag",
        "contract-freeze.invalid.review-path": "/review/record_path",
        "contract-freeze.invalid.timestamp": "/created_at",
        "contract-freeze.invalid.missing-review": "/review",
        "contract-freeze.invalid.unknown-member": "/unknown",
    }.get(case_id)
    if expected_path and changed_paths(base, candidate) != [expected_path]:
        raise SystemExit(f"mutation changed unexpected paths for {case_id}")

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
        if base64.b64encode(raw).decode() != fixture["bytes_base64"]: raise SystemExit("fixture base64 is not canonical")
        if raw != canonical(json.loads(raw.decode())) or fixture.get("bytes_sha256") != sha(raw):
            raise SystemExit("fixture canonical bytes or digest mismatch")
        fixture_bytes[fixture["fixture_id"]] = raw
    expected = ["base.canonical-minimum", "base.canonical-rich", "contract-freeze.invalid.contract", "contract-freeze.invalid.goal", "contract-freeze.invalid.missing-review", "contract-freeze.invalid.review-path", "contract-freeze.invalid.schema", "contract-freeze.invalid.timestamp", "contract-freeze.invalid.unknown-member", "contract-freeze.valid"]
    if value.get("case_count") != len(expected) or [x.get("case_id") for x in value.get("cases", [])] != expected:
        raise SystemExit("case roster mismatch")
    errors = {"contract-freeze.invalid.schema": "schema-constant", "contract-freeze.invalid.goal": "schema-constant", "contract-freeze.invalid.contract": "semantic-external-identity", "contract-freeze.invalid.review-path": "semantic-external-identity", "contract-freeze.invalid.timestamp": "semantic-cross-field", "contract-freeze.invalid.missing-review": "schema-required-member", "contract-freeze.invalid.unknown-member": "schema-unknown-member"}
    for row in value["cases"]:
        if set(row) != {"case_id", "input", "outcome", "normalized_value_sha256", "error_code"}: raise SystemExit("case row shape mismatch")
        if not isinstance(row["outcome"], str) or row["outcome"] not in {"accepted", "rejected"}: raise SystemExit("unknown outcome")
        if row["error_code"] is not None and not isinstance(row["error_code"], str): raise SystemExit("error code must be string or null")
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
        try: candidate_value = json.loads(candidate.decode())
        except Exception:
            candidate_value = None
        check_candidate(row["case_id"], candidate_value, json.loads(fixture_bytes["base.canonical-rich"].decode()))
        if row["outcome"] == "accepted":
            if row["error_code"] is not None or row["normalized_value_sha256"] != sha(candidate): raise SystemExit("accepted result mismatch")
        elif row["error_code"] != errors.get(row["case_id"]) or row["normalized_value_sha256"] is not None: raise SystemExit("rejected result mismatch")
        else: continue
        if row["case_id"] in errors: raise SystemExit("unexpected accepted result")
    print("contract-freeze forward oracle verified")

if __name__ == "__main__": main()
