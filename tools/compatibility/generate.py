#!/usr/bin/env python3
"""Validate and render the canonical ecosystem compatibility sets."""
from __future__ import annotations

import json
import hashlib
import re
import subprocess
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / "docs/ecosystem/compatibility-sets.json"
OUTPUT = ROOT / "docs/ecosystem/compatibility-sets.md"
SEMVER = re.compile(r"^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$")
PUBLISHED_SET_ID = re.compile(r"^golib-compat-v1-[0-9]{8}\.[1-9][0-9]*$")
DRAFT_SET_ID = re.compile(r"^draft-[0-9]{8}\.[1-9][0-9]*$")
REQUIRED_SET_FIELDS = {
    "set_id", "publication_status", "installable", "go", "modules",
    "scenarios", "observed_at", "evidence", "caveats", "source_references",
}
OPTIONAL_SET_FIELDS = {
    "recipes", "external_versions", "upgrade", "rollback", "exclusions",
}


def expected_tag(catalog_module: dict, version: str) -> str:
    prefix = catalog_module.get("tag_prefix")
    if not isinstance(prefix, str) or not prefix:
        raise ValueError(
            f"catalog module lacks tag prefix: {catalog_module.get('module_path')}"
        )
    return prefix + version.removeprefix("v")


def repository_identity(catalog_module: dict) -> str:
    repository = catalog_module.get("repository")
    if not isinstance(repository, str) or not re.fullmatch(
        r"github\.com/faustbrian/go-[a-z0-9]+(?:-[a-z0-9]+)*", repository
    ):
        raise ValueError(f"invalid catalog repository: {repository}")
    return repository


def remote_tag_revision(repository: str, tag: str) -> str | None:
    reference = f"refs/tags/{tag}"
    peeled = reference + "^{}"
    try:
        result = subprocess.run(
            [
                "git", "ls-remote", "--tags", f"https://{repository}.git",
                reference, peeled,
            ],
            capture_output=True,
            text=True,
            timeout=15,
        )
    except (OSError, subprocess.TimeoutExpired):
        return None
    if result.returncode != 0:
        return None
    revisions = {}
    for line in result.stdout.splitlines():
        fields = line.split()
        if len(fields) == 2:
            revisions[fields[1]] = fields[0]
    return revisions.get(peeled, revisions.get(reference))


def proxy_escape(value: str) -> str:
    return "".join(
        "!" + character.lower() if character.isupper() else character
        for character in value
    )


def proxy_module_path(module_path: str) -> str:
    return urllib.parse.quote(proxy_escape(module_path), safe="/!")


def public_module_version(module_path: str, version: str) -> bool:
    encoded_version = urllib.parse.quote(proxy_escape(version), safe="!")
    url = (
        f"https://proxy.golang.org/{proxy_module_path(module_path)}"
        f"/@v/{encoded_version}.info"
    )
    try:
        with urllib.request.urlopen(url, timeout=10) as response:
            payload = response.read(65537)
    except (OSError, urllib.error.URLError):
        return False
    if len(payload) > 65536:
        return False
    try:
        metadata = json.loads(payload)
    except (UnicodeDecodeError, json.JSONDecodeError):
        return False
    return metadata.get("Version") == version


def validate_string(value: object, field: str) -> None:
    if not isinstance(value, str) or not value.strip():
        raise ValueError(f"{field} must be a non-empty string")


def validate_string_list(value: object, field: str) -> None:
    if not isinstance(value, list) or not value:
        raise ValueError(f"{field} must be a non-empty array")
    for entry in value:
        validate_string(entry, field)


def validate_optional_fields(item: dict) -> None:
    recipes = item.get("recipes")
    if recipes is not None:
        if not isinstance(recipes, list):
            raise ValueError("recipes must be an array")
        for recipe in recipes:
            if not isinstance(recipe, dict) or set(recipe) != {
                "id", "status", "source_reference"
            }:
                raise ValueError("recipe fields mismatch")
            validate_string(recipe["id"], "recipe id")
            if recipe["status"] not in {"planned", "verified"}:
                raise ValueError("recipe status must be planned or verified")
            validate_string(recipe["source_reference"], "recipe source reference")

    external_versions = item.get("external_versions")
    if external_versions is not None:
        if not isinstance(external_versions, list):
            raise ValueError("external_versions must be an array")
        for external in external_versions:
            if not isinstance(external, dict) or set(external) != {"name", "version"}:
                raise ValueError("external version fields mismatch")
            validate_string(external["name"], "external version name")
            validate_string(external["version"], "external version")

    for field, link_field in (("upgrade", "from_set_id"), ("rollback", "to_set_id")):
        guidance = item.get(field)
        if guidance is None:
            continue
        if not isinstance(guidance, dict) or not set(guidance).issubset(
            {"notes", link_field}
        ) or "notes" not in guidance:
            raise ValueError(f"{field} fields mismatch")
        validate_string_list(guidance["notes"], f"{field} notes")
        if link_field in guidance:
            validate_string(guidance[link_field], f"{field} {link_field}")

    exclusions = item.get("exclusions")
    if exclusions is not None:
        if not isinstance(exclusions, list):
            raise ValueError("exclusions must be an array")
        for exclusion in exclusions:
            if not isinstance(exclusion, dict) or set(exclusion) != {
                "subject", "reason"
            }:
                raise ValueError("exclusion fields mismatch")
            validate_string(exclusion["subject"], "exclusion subject")
            validate_string(exclusion["reason"], "exclusion reason")


def content_fingerprint(item: dict) -> str:
    probe = dict(item)
    probe["evidence"] = dict(item["evidence"])
    probe["evidence"].pop("content_sha256", None)
    encoded = json.dumps(
        probe, sort_keys=True, separators=(",", ":"), ensure_ascii=False
    ).encode()
    return "sha256:" + hashlib.sha256(encoded).hexdigest()


def load(
    *,
    remote_tag_lookup=remote_tag_revision,
    public_version_lookup=public_module_version,
) -> dict:
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
    catalog_modules = {}
    for module in catalogs:
        module_path = module.get("module_path")
        if module_path not in catalog_modules:
            catalog_modules[module_path] = {}
        catalog_modules[module_path].update(module)
    for item in sets:
        if not isinstance(item, dict):
            raise ValueError("sets must contain objects")
        fields = set(item)
        if not REQUIRED_SET_FIELDS.issubset(fields) or not fields.issubset(
            REQUIRED_SET_FIELDS | OPTIONAL_SET_FIELDS
        ):
            raise ValueError(f"set fields mismatch: {item.get('set_id')}")
        if item["set_id"] in seen:
            raise ValueError(f"duplicate set id: {item['set_id']}")
        seen.add(item["set_id"])
        status = item["publication_status"]
        if status not in {"unreleased", "published"}:
            raise ValueError("publication status must be unreleased or published")
        if not isinstance(item["installable"], bool):
            raise ValueError("installable must be a boolean")
        if status == "unreleased":
            if item["installable"]:
                raise ValueError("unreleased sets must be non-installable")
            if not DRAFT_SET_ID.fullmatch(item["set_id"]):
                raise ValueError("unreleased set id must use the draft form")
        else:
            if not item["installable"]:
                raise ValueError("published sets must be installable")
            if not PUBLISHED_SET_ID.fullmatch(item["set_id"]):
                raise ValueError("published set id must use the stable public form")
        if not item["modules"] or not item["scenarios"]:
            raise ValueError(f"set {item['set_id']} lacks modules or scenarios")
        validate_optional_fields(item)
        evidence = item["evidence"]
        if not isinstance(evidence, dict):
            raise ValueError("evidence must be an object")
        observation = evidence.get("observation")
        validate_string(observation, "evidence observation")
        observation = observation.lower()
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
            catalog_module = catalog_modules.get(module["module_path"])
            catalog_version = None if catalog_module is None else catalog_module.get("version")
            if catalog_version is None or "v" + catalog_version.lstrip("v") != module["version"]:
                raise ValueError(f"module identity is absent or version-mismatched: {module['module_path']}")
            tag = expected_tag(catalog_module, module["version"])
            repository = repository_identity(catalog_module)
            if status == "published" and remote_tag_lookup(
                repository, tag
            ) != module["source_revision"]:
                raise ValueError(
                    f"published module revision lacks matching remote tag: {module['module_path']}"
                )
            if status == "published" and not public_version_lookup(
                module["module_path"], module["version"]
            ):
                raise ValueError(
                    f"published module version is unavailable from the public proxy: {module['module_path']}"
                )
        if item["publication_status"] == "unreleased" and "pending" not in observation:
            raise ValueError("unreleased scenarios require an explicitly pending observation")
        if not item["evidence"].get("content_sha256", "").startswith("sha256:"):
            raise ValueError("evidence requires a content fingerprint")
        expected = content_fingerprint(item)
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
        if "recipes" in item:
            lines += ["", "### Recipes", ""]
            lines += [
                f"- `{recipe['id']}` ({recipe['status']}): `{recipe['source_reference']}`"
                for recipe in item["recipes"]
            ]
        if "external_versions" in item:
            lines += ["", "### External versions", ""]
            lines += [
                f"- {external['name']}: `{external['version']}`"
                for external in item["external_versions"]
            ]
        for field, heading, link_field in (
            ("upgrade", "Upgrade", "from_set_id"),
            ("rollback", "Rollback", "to_set_id"),
        ):
            if field in item:
                lines += ["", f"### {heading}", ""]
                if link_field in item[field]:
                    lines.append(f"- Compatibility set: `{item[field][link_field]}`")
                lines += [f"- {note}" for note in item[field]["notes"]]
        if "exclusions" in item:
            lines += ["", "### Exclusions", ""]
            lines += [
                f"- **{exclusion['subject']}:** {exclusion['reason']}"
                for exclusion in item["exclusions"]
            ]
        lines += ["", "### Evidence", "", f"- Content fingerprint: `{item['evidence']['content_sha256']}`", f"- Observation: `{item['evidence']['observation']}`", ""]
    return "\n".join(lines)


if __name__ == "__main__":
    data = load()
    rendered = render(data)
    if OUTPUT.exists() and OUTPUT.read_text() != rendered:
        raise SystemExit("compatibility-sets.md is stale; regenerate it")
    print(f"validated {len(data['sets'])} compatibility set(s)")
