#!/usr/bin/env python3
"""Verify the multiplexed oracle's exact schema and evidence bindings."""
from __future__ import annotations

import argparse
import json
from pathlib import Path

from check_v2 import ROOT, SCHEMA_PATHS, digest
from generate_multiplexed_oracle import BASE_EXPECTED_CASES, DECISION_EXPECTED_CASES, SOURCES, source_cases


def verify(path: Path) -> None:
    value = json.loads(path.read_text(encoding="utf-8"))
    if value.get("format") != "golib-cohesion-schema-provenance-v2-multiplexed-oracle":
        raise ValueError("wrong oracle format")
    if value.get("schema_id") != "urn:golib:cohesion:schema-provenance:v2" or value.get("schema_version") != 2:
        raise ValueError("wrong schema identity")
    entries = value.get("entries")
    if value.get("schema_count") != len(SCHEMA_PATHS) or not isinstance(entries, list) or [e.get("schema_path") for e in entries] != list(SCHEMA_PATHS):
        raise ValueError("schema roster mismatch")
    expected_missing = []
    for entry in entries:
        schema_path = entry["schema_path"]
        schema = ROOT / schema_path
        if entry.get("schema_bytes_sha256") != digest(schema):
            raise ValueError(f"schema digest mismatch: {schema_path}")
        source_paths = [source["path"] for source in entry.get("sources", [])]
        if source_paths != sorted(source_paths) or len(source_paths) != len(set(source_paths)):
            raise ValueError(f"source roster is not unique/sorted: {schema_path}")
        expected_sources = list(SOURCES.get(schema_path, ()))
        available_sources = [relative for relative in expected_sources if (ROOT / relative).is_file()]
        if source_paths != available_sources:
            raise ValueError(f"source binding mismatch: {schema_path}")
        all_cases = []
        for source in entry["sources"]:
            source_path = ROOT / source["path"]
            if source.get("bytes_sha256") != digest(source_path):
                raise ValueError(f"source digest mismatch: {source['path']}")
            all_cases.extend({"source_path": source["path"], **case} for case in source_cases(source_path))
        expected_cases = sorted(all_cases, key=lambda row: (row["case_id"], row["source_path"]))
        if entry.get("cases") != expected_cases or entry.get("case_count") != len(expected_cases):
            raise ValueError(f"case binding mismatch: {schema_path}")
        expected_ids = DECISION_EXPECTED_CASES if "schema-v3-decision-" in schema_path else BASE_EXPECTED_CASES
        missing_required = sorted(set(expected_ids) - {case["case_id"] for case in expected_cases})
        completeness = "complete" if not missing_required else ("missing" if not expected_cases else "partial")
        if entry.get("expected_case_ids") != list(expected_ids) or entry.get("missing_required_cases") != missing_required or entry.get("completeness") != completeness:
            raise ValueError(f"completeness metadata mismatch: {schema_path}")
        outcomes = {case["outcome"] for case in expected_cases}
        absent = sorted({"accepted", "rejected"} - outcomes)
        if absent:
            expected_missing.append({"schema_path": schema_path, "missing_outcomes": absent, "reason": "no authoritative semantic fixture is available for this outcome"})
    report = value.get("missing_case_report")
    if report != {"format": "golib-cohesion-schema-provenance-v2-missing-case-report", "missing_count": len(expected_missing), "missing": expected_missing}:
        raise ValueError("missing-case report does not match bound evidence")
    print(f"multiplexed oracle verified: {len(entries)} schemas, {sum(e['case_count'] for e in entries)} cases, {len(expected_missing)} missing reports")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("path", type=Path)
    args = parser.parse_args()
    verify(args.path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
