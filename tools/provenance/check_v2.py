#!/usr/bin/env python3
"""Inventory and validate inputs for cohesion-schema-provenance-v2.

This is deliberately an inventory gate: it never fabricates an oracle, review,
or provenance envelope. The multiplexed oracle and aggregate review are each
single assets, so adding a schema does not require another review file.
"""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SCHEMA = ROOT / "schema/cohesion-schema-provenance-v2.schema.json"
SCHEMA_PATHS = (
    "schema/modules-v1.schema.json", "schema/modules-v2.schema.json", "schema/modules-v3.schema.json",
    "schema/cohesion-catalog-v1.schema.json", "schema/cohesion-catalog-v2.schema.json",
    "schema/cohesion-inputs-v1.schema.json", "schema/cohesion-inputs-v2.schema.json",
    "schema/cohesion-sources-v1.schema.json", "schema/cohesion-sources-v2.schema.json",
    "schema/cohesion-delivery-evidence-v1.schema.json", "schema/cohesion-authorization-registry-v1.schema.json",
    "schema/cohesion-authorization-record-v1.schema.json", "schema/cohesion-residual-inventory-v1.schema.json",
    "schema/cohesion-residual-register-v1.schema.json", "schema/cohesion-gate-policy-v1.schema.json",
    "schema/cohesion-go-toolchains-v1.schema.json", "schema/cohesion-go-official-downloads-oracle-v1.schema.json",
    "schema/cohesion-go-official-downloads-oracle-review-v1.schema.json", "schema/cohesion-engineering-identities-v1.schema.json",
    "schema/cohesion-schema-provenance-v1.schema.json", "schema/cohesion-schema-v3-decision-review-v1.schema.json",
    "schema/cohesion-schema-v3-decision-freeze-v1.schema.json", "schema/cohesion-contract-freeze-v1.schema.json",
    "schema/cohesion-contract-review-v1.schema.json", "schema/cohesion-diagnostic-v1.schema.json",
)
ORACLE = "testdata/cohesion/schema-provenance-v2-multiplexed-oracle.json"
REVIEW = "testdata/cohesion/schema-provenance-v2-aggregate-review.json"


def digest(path: Path) -> str:
    return "sha256:" + hashlib.sha256(path.read_bytes()).hexdigest()


def inspect() -> dict:
    missing: list[str] = []
    for rel in (ORACLE, REVIEW):
        if not (ROOT / rel).is_file():
            missing.append(rel)
    entries = []
    for rel in SCHEMA_PATHS:
        path = ROOT / rel
        entries.append({"schema_path": rel, "present": path.is_file(), "schema_bytes_sha256": digest(path) if path.is_file() else None})
    return {
        "format": "golib-cohesion-schema-provenance-input-report-v2",
        "schema_id": "urn:golib:cohesion:schema-provenance:v2",
        "entry_count": len(entries),
        "oracle": {"path": ORACLE, "present": (ROOT / ORACLE).is_file()},
        "review": {"path": REVIEW, "present": (ROOT / REVIEW).is_file()},
        "missing_count": len(missing),
        "missing": missing,
        "entries": entries,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--strict", action="store_true")
    args = parser.parse_args()
    report = inspect()
    print(json.dumps(report, indent=2, sort_keys=True) + "\n")
    return 1 if args.strict and report["missing_count"] else 0


if __name__ == "__main__":
    raise SystemExit(main())
