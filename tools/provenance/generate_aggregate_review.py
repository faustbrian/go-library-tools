#!/usr/bin/env python3
"""Generate the non-authorizing aggregate review for the v2 oracle index."""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

from check_v2 import ROOT, SCHEMA_PATHS

ORACLE_REL = "testdata/cohesion/schema-provenance-v2-multiplexed-oracle.json"

# These schemas have released reference graphs (structural bindings) but no
# released semantic instances.  Keep the distinction machine-readable so a
# missing accepted/rejected pair cannot be mistaken for an omitted search.
STRUCTURAL_ONLY = {
    "schema/cohesion-delivery-evidence-v1.schema.json": "release/cohesion-delivery-evidence-v1-resolved-reference-graph.json",
    "schema/cohesion-authorization-registry-v1.schema.json": "release/cohesion-authorization-registry-v1-resolved-reference-graph.json",
    "schema/cohesion-authorization-record-v1.schema.json": "release/cohesion-authorization-record-v1-resolved-reference-graph.json",
    "schema/cohesion-residual-inventory-v1.schema.json": "release/cohesion-residual-inventory-v1-resolved-reference-graph.json",
    "schema/cohesion-residual-register-v1.schema.json": "release/cohesion-residual-register-v1-resolved-reference-graph.json",
}


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
        "authorization": {"authorized": False, "reason": f"{len(missing)} schema paths lack authoritative accepted and rejected semantic fixtures"},
        "oracle": {"path": ORACLE_REL, "bytes_sha256": sha(oracle), "schema_count": value["schema_count"], "case_count": sum(e["case_count"] for e in value["entries"])},
        "schema_inventory": {"entry_count": len(entries), "entries": entries},
        "missing_case_report": {"missing_count": len(missing), "missing": missing},
        "unprovable_case_report": {
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
        },
        "dimensions": [
            {"name": "coverage", "authored_by": "codex-coverage", "outcome": "incomplete", "basis": "oracle entries are complete only where released semantic fixtures exist", "covered_schema_count": len(SCHEMA_PATHS) - len(missing), "missing_schema_count": len(missing)},
            {"name": "integrity", "authored_by": "codex-integrity", "outcome": "pass", "basis": "the oracle verifier binds every listed schema, source, and case digest"},
            {"name": "reachability", "authored_by": "codex-reachability", "outcome": "incomplete", "basis": f"{len(missing)} schema paths have no accepted and rejected fixture pair to exercise"},
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
