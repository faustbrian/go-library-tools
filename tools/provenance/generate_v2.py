#!/usr/bin/env python3
"""Emit a non-claiming inventory envelope for provenance-v2 inputs.

The output intentionally contains no accepted/rejected outcomes. It records
only exact schema bytes so an eventual multiplexed oracle can be assembled
without pretending that missing semantic evidence exists.
"""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

from check_v2 import ROOT, SCHEMA_PATHS


def digest(path: Path) -> str:
    return "sha256:" + hashlib.sha256(path.read_bytes()).hexdigest()


def build() -> dict:
    entries = []
    for relative in SCHEMA_PATHS:
        path = ROOT / relative
        if not path.is_file():
            raise FileNotFoundError(relative)
        entries.append({"schema_path": relative, "schema_bytes_sha256": digest(path)})
    return {
        "format": "golib-cohesion-schema-provenance-v2-input-inventory",
        "schema_id": "urn:golib:cohesion:schema-provenance:v2",
        "status": "inputs-inventory",
        "oracle": {"path": "testdata/cohesion/schema-provenance-v2-multiplexed-oracle.json"},
        "review": {"path": "testdata/cohesion/schema-provenance-v2-aggregate-review.json"},
        "entry_count": len(entries),
        "entries": entries,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(build(), indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
