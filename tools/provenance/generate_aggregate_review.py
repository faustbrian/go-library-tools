#!/usr/bin/env python3
"""Generate the non-authorizing aggregate review for the v2 oracle index."""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

from check_v2 import ROOT, SCHEMA_PATHS

ORACLE_REL = "testdata/cohesion/schema-provenance-v2-multiplexed-oracle.json"


def sha(path: Path) -> str:
    return "sha256:" + hashlib.sha256(path.read_bytes()).hexdigest()


def build() -> dict:
    oracle = ROOT / ORACLE_REL
    value = json.loads(oracle.read_text(encoding="utf-8"))
    entries = [{"schema_path": e["schema_path"], "schema_bytes_sha256": e["schema_bytes_sha256"]} for e in value["entries"]]
    missing = value["missing_case_report"]["missing"]
    return {
        "format": "golib-cohesion-schema-provenance-v2-aggregate-review",
        "schema_id": "urn:golib:cohesion:schema-provenance:v2",
        "schema_version": 2,
        "status": "incomplete-non-authorizing",
        "authorization": {"authorized": False, "reason": "17 schema paths lack authoritative accepted and rejected semantic fixtures"},
        "oracle": {"path": ORACLE_REL, "bytes_sha256": sha(oracle), "schema_count": value["schema_count"], "case_count": sum(e["case_count"] for e in value["entries"])},
        "schema_inventory": {"entry_count": len(entries), "entries": entries},
        "missing_case_report": {"missing_count": len(missing), "missing": missing},
        "dimensions": [
            {"name": "coverage", "authored_by": "codex-coverage", "outcome": "incomplete", "basis": "oracle entries are complete only where released semantic fixtures exist", "covered_schema_count": len(SCHEMA_PATHS) - len(missing), "missing_schema_count": len(missing)},
            {"name": "integrity", "authored_by": "codex-integrity", "outcome": "pass", "basis": "the oracle verifier binds every listed schema, source, and case digest"},
            {"name": "reachability", "authored_by": "codex-reachability", "outcome": "incomplete", "basis": "17 schema paths have no accepted and rejected fixture pair to exercise"},
        ],
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
