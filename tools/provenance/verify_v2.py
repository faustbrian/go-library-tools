#!/usr/bin/env python3
"""Verify the shape and exact schema binding of a provenance-v2 inventory."""
from __future__ import annotations

import argparse
import json
from pathlib import Path

from check_v2 import ROOT, SCHEMA_PATHS


def verify(path: Path) -> None:
    value = json.loads(path.read_text(encoding="utf-8"))
    if value.get("format") != "golib-cohesion-schema-provenance-v2-input-inventory":
        raise ValueError("wrong format")
    if value.get("schema_id") != "urn:golib:cohesion:schema-provenance:v2":
        raise ValueError("wrong schema id")
    if value.get("status") != "inputs-inventory":
        raise ValueError("inventory must not claim semantic completion")
    if value.get("entry_count") != len(SCHEMA_PATHS):
        raise ValueError("entry count mismatch")
    entries = value.get("entries")
    if not isinstance(entries, list) or [e.get("schema_path") for e in entries] != list(SCHEMA_PATHS):
        raise ValueError("schema paths are not the frozen ordered set")
    for entry in entries:
        if set(entry) != {"schema_path", "schema_bytes_sha256"}:
            raise ValueError("unexpected semantic claim in entry")
        if entry["schema_bytes_sha256"] != __import__("check_v2").digest(ROOT / entry["schema_path"]):
            raise ValueError("schema digest mismatch")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("path", type=Path)
    args = parser.parse_args()
    verify(args.path)
    print("provenance-v2 inventory verified")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
