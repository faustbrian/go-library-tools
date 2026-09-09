#!/usr/bin/env python3
"""Verify the complete, fixed forward-oracle lane set."""
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
ORACLES = ROOT / "testdata/cohesion/forward-oracles"
LANES = {
    "cohesion-catalog-v2-forward-oracle.json": "generic",
    "cohesion-contract-freeze-v1-forward-oracle.json": "verify_contract_freeze.py",
    "cohesion-contract-review-v1-forward-oracle.json": "verify_contract_review.py",
    "cohesion-diagnostic-v1-forward-oracle.json": "verify.py",
    "cohesion-go-toolchains-v1-forward-oracle.json": "verify_go_toolchains.py",
    "cohesion-schema-v3-decision-freeze-v1-forward-oracle.json": "generic",
    "cohesion-schema-v3-decision-review-v1-forward-oracle.json": "generic",
}


def generic(path: Path) -> None:
    value = json.loads(path.read_bytes())
    if value.get("format") != "golib-forward-oracle-v2":
        raise ValueError("format mismatch")
    fixtures = value.get("fixtures")
    if value.get("fixture_count") != 2 or not isinstance(fixtures, list):
        raise ValueError("fixture declaration mismatch")
    if [x.get("fixture_id") for x in fixtures] != ["base.canonical-minimum", "base.canonical-rich"]:
        raise ValueError("fixture roster mismatch")
    cases = value.get("cases")
    if not isinstance(cases, list) or len(cases) < 2 or value.get("case_count") != len(cases):
        raise ValueError("case count must match and contain at least two cases")
    ids = [x.get("case_id") for x in cases]
    if any(not isinstance(x, str) or not x for x in ids) or len(ids) != len(set(ids)):
        raise ValueError("case IDs must be non-empty and unique")
    for row in cases:
        if set(row) != {"case_id", "input", "outcome", "normalized_value_sha256", "error_code"}:
            raise ValueError("case shape mismatch")
        if row["outcome"] not in {"accepted", "rejected"}:
            raise ValueError("invalid case outcome")
        if not isinstance(row["input"], dict) or row["input"].get("fixture_id") not in {x.get("fixture_id") for x in fixtures}:
            raise ValueError("invalid fixture reference")


def main() -> None:
    actual = {p.name for p in ORACLES.glob("*-forward-oracle.json")}
    expected = set(LANES)
    if actual != expected:
        raise SystemExit(f"forward-oracle lane set mismatch: actual={sorted(actual)} expected={sorted(expected)}")
    for name, checker in LANES.items():
        path = ORACLES / name
        if checker == "generic":
            generic(path)
        else:
            subprocess.run([sys.executable, str(Path(__file__).with_name(checker))], check=True)
    print(f"verified {len(LANES)} forward-oracle lanes")


if __name__ == "__main__":
    main()
