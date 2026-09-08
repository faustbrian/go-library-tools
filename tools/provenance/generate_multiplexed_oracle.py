#!/usr/bin/env python3
"""Build the evidence-indexed provenance-v2 multiplexed oracle.

The oracle is an index, not a semantic oracle implementation: every case is
copied by identity from an existing released corpus or checked forward-oracle
asset.  Schemas without such an asset are listed in ``missing_cases`` rather
than being assigned invented outcomes.
"""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

from check_v2 import ROOT, SCHEMA_PATHS

SOURCES = {
    "schema/modules-v1.schema.json": ("release/modules-v1-accepted-corpus.json", "release/modules-v1-rejected-corpus.json"),
    "schema/modules-v2.schema.json": ("release/modules-v2-accepted-corpus.json", "release/modules-v2-rejected-corpus.json"),
    "schema/cohesion-catalog-v1.schema.json": ("release/cohesion-catalog-v1-accepted-corpus.json", "release/cohesion-catalog-v1-rejected-corpus.json"),
    "schema/cohesion-inputs-v1.schema.json": ("release/cohesion-inputs-v1-accepted-corpus.json", "release/cohesion-inputs-v1-rejected-corpus.json"),
    "schema/cohesion-sources-v1.schema.json": ("release/cohesion-sources-v1-accepted-corpus.json", "release/cohesion-sources-v1-rejected-corpus.json"),
    "schema/cohesion-contract-freeze-v1.schema.json": ("testdata/cohesion/forward-oracles/cohesion-contract-freeze-v1-forward-oracle.json",),
    "schema/cohesion-contract-review-v1.schema.json": ("testdata/cohesion/forward-oracles/cohesion-contract-review-v1-forward-oracle.json",),
    "schema/cohesion-diagnostic-v1.schema.json": ("testdata/cohesion/forward-oracles/cohesion-diagnostic-v1-forward-oracle.json",),
}


def sha(path: Path) -> str:
    return "sha256:" + hashlib.sha256(path.read_bytes()).hexdigest()


def source_cases(path: Path) -> list[dict]:
    value = json.loads(path.read_text(encoding="utf-8"))
    if path.name.endswith("-accepted-corpus.json"):
        return [{"case_id": row["case_id"], "outcome": "accepted"} for row in value]
    if path.name.endswith("-rejected-corpus.json"):
        return [{"case_id": row["case_id"], "outcome": "rejected"} for row in value]
    return [{"case_id": row["case_id"], "outcome": row["outcome"]} for row in value["cases"]]


def build() -> dict:
    entries = []
    missing = []
    for schema_path in SCHEMA_PATHS:
        schema = ROOT / schema_path
        sources = SOURCES.get(schema_path, ())
        cases = []
        source_rows = []
        for relative in sources:
            path = ROOT / relative
            if not path.is_file():
                continue
            source_rows.append({"path": relative, "bytes_sha256": sha(path)})
            cases.extend({"source_path": relative, **case} for case in source_cases(path))
        cases.sort(key=lambda row: (row["case_id"], row["source_path"]))
        outcomes = {case["outcome"] for case in cases}
        absent = sorted({"accepted", "rejected"} - outcomes)
        if absent:
            missing.append({
                "schema_path": schema_path,
                "missing_outcomes": absent,
                "reason": "no authoritative semantic fixture is available for this outcome",
            })
        entries.append({
            "schema_path": schema_path,
            "schema_bytes_sha256": sha(schema),
            "source_count": len(source_rows),
            "sources": source_rows,
            "case_count": len(cases),
            "cases": cases,
        })
    return {
        "format": "golib-cohesion-schema-provenance-v2-multiplexed-oracle",
        "schema_id": "urn:golib:cohesion:schema-provenance:v2",
        "schema_version": 2,
        "schema_count": len(entries),
        "entries": entries,
        "missing_case_report": {
            "format": "golib-cohesion-schema-provenance-v2-missing-case-report",
            "missing_count": len(missing),
            "missing": missing,
        },
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(build(), ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
