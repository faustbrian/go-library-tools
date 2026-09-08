#!/usr/bin/env python3
"""Validate and render the canonical ecosystem compatibility sets."""
from __future__ import annotations

import json
import hashlib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / "docs/ecosystem/compatibility-sets.json"
OUTPUT = ROOT / "docs/ecosystem/compatibility-sets.md"


def load() -> dict:
    value = json.loads(SOURCE.read_text())
    if value.get("format") != "golib-compatibility-sets-v1":
        raise ValueError("unexpected compatibility-set format")
    sets = value.get("sets")
    if not isinstance(sets, list) or not sets:
        raise ValueError("sets must be a non-empty array")
    seen = set()
    for item in sets:
        required = {"set_id", "publication_status", "installable", "go", "modules", "scenarios", "observed_at", "evidence", "caveats", "source_references"}
        if set(item) != required:
            raise ValueError(f"set fields mismatch: {item.get('set_id')}")
        if item["set_id"] in seen:
            raise ValueError(f"duplicate set id: {item['set_id']}")
        seen.add(item["set_id"])
        if item["publication_status"] == "unreleased" and item["installable"]:
            raise ValueError("unreleased sets must be non-installable")
        if not item["modules"] or not item["scenarios"]:
            raise ValueError(f"set {item['set_id']} lacks modules or scenarios")
        module_ids = set()
        for module in item["modules"]:
            if set(module) != {"module_path", "version", "source_revision"}:
                raise ValueError(f"module fields mismatch in {item['set_id']}")
            if module["module_path"] in module_ids:
                raise ValueError(f"duplicate module in {item['set_id']}")
            module_ids.add(module["module_path"])
            if module["version"].startswith("v") is False:
                raise ValueError("module versions must be semantic-version strings")
            if len(module["source_revision"]) != 40:
                raise ValueError("source revisions must be full Git identities")
        if not item["evidence"].get("content_sha256", "").startswith("sha256:"):
            raise ValueError("evidence requires a content fingerprint")
        probe = dict(item)
        probe["evidence"] = dict(item["evidence"])
        expected = "sha256:" + hashlib.sha256(json.dumps({k: probe[k] for k in sorted(probe) if k != "evidence"} | {"evidence": {k: v for k, v in probe["evidence"].items() if k != "content_sha256"}}, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
        if item["evidence"]["content_sha256"] != expected:
            raise ValueError(f"content fingerprint mismatch: {item['set_id']}")
        for reference in item["source_references"]:
            if not (ROOT / reference).is_file():
                raise ValueError(f"missing source reference: {reference}")
    return value


def render(value: dict) -> str:
    lines = ["# Golib Compatibility Sets", "", "Generated from `compatibility-sets.json`; edit the JSON source only.", ""]
    for item in value["sets"]:
        status = item["publication_status"]
        install = "installable" if item["installable"] else "non-installable"
        lines += [f"## `{item['set_id']}`", "", f"**Status:** {status}; **{install}**. Observed `{item['observed_at']}`.", "", "### Modules", ""]
        for module in item["modules"]:
            lines.append(f"- `{module['module_path']}@{module['version']}` (source `{module['source_revision']}`)")
        lines += ["", "### Covered scenarios", ""]
        lines += [f"- {scenario}" for scenario in item["scenarios"]]
        lines += ["", f"**Go:** `{item['go']['version']}` on {', '.join(item['go']['os_arch'])}.", "", "### Caveats", ""]
        lines += [f"- {caveat}" for caveat in item["caveats"]]
        lines += ["", "### Evidence", "", f"- Content fingerprint: `{item['evidence']['content_sha256']}`", f"- Observation: `{item['evidence']['observation']}`", ""]
    return "\n".join(lines)


if __name__ == "__main__":
    data = load()
    rendered = render(data)
    if OUTPUT.exists() and OUTPUT.read_text() != rendered:
        raise SystemExit("compatibility-sets.md is stale; regenerate it")
    print(f"validated {len(data['sets'])} compatibility set(s)")
