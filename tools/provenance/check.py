#!/usr/bin/env python3
"""Inspect the inputs required to build cohesion-schema-provenance-v1.

This tool is intentionally an inventory/checker, not a provenance generator.
It never synthesizes oracle, review, commit, or release evidence.  A report can
therefore be used while the evidence set is being assembled without creating a
control that falsely claims completion.
"""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
SCHEMA_PATHS = [
    "schema/modules-v1.schema.json",
    "schema/modules-v2.schema.json",
    "schema/modules-v3.schema.json",
    "schema/cohesion-catalog-v1.schema.json",
    "schema/cohesion-catalog-v2.schema.json",
    "schema/cohesion-inputs-v1.schema.json",
    "schema/cohesion-inputs-v2.schema.json",
    "schema/cohesion-sources-v1.schema.json",
    "schema/cohesion-sources-v2.schema.json",
    "schema/cohesion-delivery-evidence-v1.schema.json",
    "schema/cohesion-authorization-registry-v1.schema.json",
    "schema/cohesion-authorization-record-v1.schema.json",
    "schema/cohesion-residual-inventory-v1.schema.json",
    "schema/cohesion-residual-register-v1.schema.json",
    "schema/cohesion-gate-policy-v1.schema.json",
    "schema/cohesion-go-toolchains-v1.schema.json",
    "schema/cohesion-go-official-downloads-oracle-v1.schema.json",
    "schema/cohesion-go-official-downloads-oracle-review-v1.schema.json",
    "schema/cohesion-engineering-identities-v1.schema.json",
    "schema/cohesion-schema-provenance-v1.schema.json",
    "schema/cohesion-schema-v3-decision-review-v1.schema.json",
    "schema/cohesion-schema-v3-decision-freeze-v1.schema.json",
    "schema/cohesion-contract-freeze-v1.schema.json",
    "schema/cohesion-contract-review-v1.schema.json",
    "schema/cohesion-diagnostic-v1.schema.json",
]
HISTORICAL = {"modules-v1", "modules-v2", "cohesion-catalog-v1", "cohesion-inputs-v1", "cohesion-sources-v1"}


def digest(path: Path) -> str:
    return "sha256:" + hashlib.sha256(path.read_bytes()).hexdigest()


def expected_assets(schema_path: str) -> list[str]:
    stem = Path(schema_path).name.removesuffix(".schema.json")
    if stem in HISTORICAL:
        return [
            f"release/{stem}-historical-baseline.json",
            f"release/{stem}-accepted-corpus.json",
            f"release/{stem}-rejected-corpus.json",
        ]
    return [
        f"testdata/cohesion/forward-oracles/{stem}-forward-oracle.json",
        f"testdata/cohesion/forward-oracle-reviews/{stem}-forward-oracle-review.json",
    ]


def inspect() -> dict:
    rows = []
    missing: list[str] = []
    for schema_path in SCHEMA_PATHS:
        required = [f"release/{Path(schema_path).name.removesuffix('.schema.json')}-resolved-reference-graph.json"]
        required.extend(expected_assets(schema_path))
        assets = []
        for relative in required:
            path = ROOT / relative
            item = {"path": relative, "present": path.is_file()}
            if path.is_file():
                item["bytes_sha256"] = digest(path)
            else:
                missing.append(relative)
            assets.append(item)
        rows.append({"schema_path": schema_path, "schema_bytes_sha256": digest(ROOT / schema_path), "assets": assets})
    return {"format": "golib-cohesion-provenance-input-report-v1", "entry_count": len(SCHEMA_PATHS), "missing_count": len(missing), "missing": missing, "entries": rows}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--strict", action="store_true", help="return failure when any required input is missing")
    args = parser.parse_args()
    report = inspect()
    print(json.dumps(report, indent=2, sort_keys=True) + "\n")
    return 1 if args.strict and report["missing_count"] else 0


if __name__ == "__main__":
    raise SystemExit(main())
