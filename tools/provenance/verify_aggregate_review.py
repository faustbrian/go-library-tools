#!/usr/bin/env python3
"""Verify aggregate review bindings and its explicitly non-authorizing state."""
from __future__ import annotations

import argparse
import json
from pathlib import Path

from check_v2 import ROOT, SCHEMA_PATHS
from generate_aggregate_review import ORACLE_REL, STRUCTURAL_ONLY, sha


def verify(path: Path) -> None:
    value = json.loads(path.read_text(encoding="utf-8"))
    if value.get("format") != "golib-cohesion-schema-provenance-v2-aggregate-review" or value.get("schema_id") != "urn:golib:cohesion:schema-provenance:v2" or value.get("schema_version") != 2:
        raise ValueError("wrong review identity")
    if value.get("status") != "incomplete-non-authorizing":
        raise ValueError("review must remain explicitly non-authorizing")
    oracle = ROOT / ORACLE_REL
    oracle_value = json.loads(oracle.read_text(encoding="utf-8"))
    expected_entries = [{"schema_path": e["schema_path"], "schema_bytes_sha256": e["schema_bytes_sha256"], "completeness": e["completeness"], "expected_case_ids": e["expected_case_ids"], "missing_required_cases": e["missing_required_cases"]} for e in oracle_value["entries"]]
    if value.get("oracle") != {"path": ORACLE_REL, "bytes_sha256": sha(oracle), "schema_count": 25, "case_count": sum(e["case_count"] for e in oracle_value["entries"])}:
        raise ValueError("oracle binding mismatch")
    counts = {state: sum(e["completeness"] == state for e in oracle_value["entries"]) for state in ("complete", "partial", "missing")}
    if value.get("schema_inventory") != {"entry_count": 25, "entries": expected_entries, "completeness_counts": counts} or [e["schema_path"] for e in expected_entries] != list(SCHEMA_PATHS):
        raise ValueError("schema inventory mismatch")
    report = oracle_value["missing_case_report"]
    expected_authorization = {"authorized": False, "reason": f"{report['missing_count']} schema paths lack authoritative accepted and rejected semantic fixtures"}
    if value.get("authorization") != expected_authorization:
        raise ValueError("authorization boundary mismatch")
    expected_missing = {"missing_count": report["missing_count"], "missing": report["missing"]}
    if value.get("missing_case_report") != expected_missing:
        raise ValueError("missing-case report mismatch")
    expected_unprovable = {
        "schemas": [
            {
                "schema_path": schema_path,
                "available_artifact": artifact,
                "artifact_kind": "resolved-reference-graph",
                "reason": "no released accepted or rejected semantic instance exists; generating one would invent evidence",
            }
            for schema_path, artifact in STRUCTURAL_ONLY.items()
        ],
        "count": len(STRUCTURAL_ONLY),
    }
    if value.get("unprovable_case_report") != expected_unprovable:
        raise ValueError("unprovable-case rationale mismatch")
    dimensions = value.get("dimensions")
    if not isinstance(dimensions, list) or [d.get("name") for d in dimensions] != ["coverage", "integrity", "reachability"]:
        raise ValueError("dimension roster mismatch")
    if any(d.get("authored_by") not in {"codex-coverage", "codex-integrity", "codex-reachability"} for d in dimensions):
        raise ValueError("dimension authorship missing")
    outcomes = {d["name"]: d["outcome"] for d in dimensions}
    if outcomes != {"coverage": "incomplete", "integrity": "pass", "reachability": "incomplete"}:
        raise ValueError("review outcomes overclaim coverage")
    expected_basis = "partial and missing entries still lack one or more required case IDs"
    if next(d for d in dimensions if d["name"] == "reachability").get("basis") != expected_basis:
        raise ValueError("reachability boundary mismatch")
    print(f"aggregate review verified: non-authorizing, {len(expected_entries)} schemas, {report['missing_count']} missing")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("path", type=Path)
    args = parser.parse_args()
    verify(args.path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
