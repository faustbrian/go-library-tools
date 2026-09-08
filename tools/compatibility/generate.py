#!/usr/bin/env python3
"""Validate and render the canonical ecosystem compatibility sets."""
from __future__ import annotations

import json
import hashlib
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / "docs/ecosystem/compatibility-sets.json"
OUTPUT = ROOT / "docs/ecosystem/compatibility-sets.md"
SEMVER = re.compile(r"^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$")


def load() -> dict:
    value = json.loads(SOURCE.read_text())
    if value.get("format") != "golib-compatibility-sets-v1":
        raise ValueError("unexpected compatibility-set format")
    sets = value.get("sets")
    if not isinstance(sets, list) or not sets:
        raise ValueError("sets must be a non-empty array")
    seen = set()
    catalogs = []
    for catalog_name in ("catalog-consumer.json", "catalog-engineering.json"):
        catalog = json.loads((ROOT / "docs/ecosystem" / catalog_name).read_text())
        catalogs.extend(catalog.get("modules", []))
    catalog_versions = {m.get("module_path"): m.get("version") for m in catalogs}
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
            if not isinstance(module["version"], str) or not SEMVER.fullmatch(module["version"]):
                raise ValueError("module versions must be semantic-version strings")
            if not isinstance(module["source_revision"], str) or not re.fullmatch(r"[0-9a-f]{40}", module["source_revision"]):
                raise ValueError("source revisions must be full Git identities")
            catalog_version = catalog_versions.get(module["module_path"])
            if catalog_version is None or "v" + catalog_version.lstrip("v") != module["version"]:
                raise ValueError(f"module identity is absent or version-mismatched: {module['module_path']}")
            repo_dir = ROOT.parent / module["module_path"].rsplit("/", 1)[-1]
            revision = subprocess.run(["git", "-C", str(repo_dir), "cat-file", "-e", module["source_revision"]], capture_output=True)
            if revision.returncode != 0:
                raise ValueError(f"source revision is not present locally: {module['module_path']}")
            tags = subprocess.run(["git", "-C", str(repo_dir), "tag", "--points-at", module["source_revision"]], capture_output=True, text=True, check=True).stdout.split()
            if module["version"] not in tags and item["installable"]:
                raise ValueError(f"installable module revision lacks matching local tag: {module['module_path']}")
        observation = item["evidence"].get("observation", "").lower()
        if item["publication_status"] == "unreleased" and "pending" not in observation:
            raise ValueError("unreleased scenarios require an explicitly pending observation")
        if item["publication_status"] == "unreleased" and any(
            not subprocess.run(["git", "-C", str(ROOT.parent / m["module_path"].rsplit("/", 1)[-1]), "tag", "--points-at", m["source_revision"]], capture_output=True, text=True, check=True).stdout.split()
            for m in item["modules"]
        ) and "unverified" not in observation:
            raise ValueError("unreleased sets with untagged revisions require an unverified observation")
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
        heading = "### Planned/unverified scenarios" if status == "unreleased" else "### Covered scenarios"
        lines += ["", heading, ""]
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
