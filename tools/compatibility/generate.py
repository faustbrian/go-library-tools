#!/usr/bin/env python3
"""Validate and render the canonical ecosystem compatibility sets."""
from __future__ import annotations

import argparse
import copy
import concurrent.futures
import datetime
import json
import hashlib
import os
import re
import subprocess
import sys
import tempfile
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / "docs/ecosystem/compatibility-sets.json"
OUTPUT = ROOT / "docs/ecosystem/compatibility-sets.md"
CONSUMER_DIRECTORY = ROOT / "release/compatibility-consumer"
RESIDUALS = ROOT / "release/cohesion-residuals.json"
SEMVER = re.compile(r"^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$")
PUBLISHED_SET_ID = re.compile(r"^golib-compat-v1-[0-9]{8}\.[1-9][0-9]*$")
DRAFT_SET_ID = re.compile(r"^draft-[0-9]{8}\.[1-9][0-9]*$")
GO_VERSION = re.compile(r"^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$")
REQUIRED_SET_FIELDS = {
    "set_id", "publication_status", "installable", "go", "modules",
    "scenarios", "observed_at", "evidence", "caveats", "source_references",
}
OPTIONAL_SET_FIELDS = {
    "recipes", "external_versions", "upgrade", "rollback", "exclusions",
    "roster",
}


def active_catalog_modules(catalog_modules: list[dict]) -> list[dict]:
    selected = []
    seen = set()
    for module in catalog_modules:
        cohesion = module.get("cohesion") or {}
        if module.get("releasable") is not True:
            continue
        if cohesion.get("lifecycle_status") != "active":
            continue
        module_path = module.get("module_path")
        validate_string(module_path, "catalog module path")
        if module_path in seen:
            raise ValueError(f"duplicate active catalog module: {module_path}")
        seen.add(module_path)
        selected.append(module)
    return sorted(selected, key=lambda module: module["module_path"])


def validate_roster(item: dict) -> None:
    roster = item.get("roster")
    if roster is None:
        return
    if not isinstance(roster, dict) or set(roster) != {"selection", "module_count"}:
        raise ValueError("roster fields mismatch")
    if roster["selection"] != "active-public":
        raise ValueError("roster selection must be active-public")
    if not isinstance(roster["module_count"], int) or roster["module_count"] < 1:
        raise ValueError("roster module count must be a positive integer")
    if len(item.get("modules", [])) != roster["module_count"]:
        raise ValueError("roster mismatch: declared count does not match selected modules")


def validate_complete_roster(item: dict, catalog_modules: list[dict]) -> None:
    roster = item.get("roster")
    if roster is None:
        return
    validate_roster(item)
    expected = [module["module_path"] for module in active_catalog_modules(catalog_modules)]
    actual = [module.get("module_path") for module in item.get("modules", [])]
    if (
        len(expected) != roster["module_count"]
        or len(actual) != roster["module_count"]
        or actual != expected
    ):
        raise ValueError(
            f"roster mismatch: expected {len(expected)} active modules, got {len(actual)}"
        )


def build_candidate(
    template: dict,
    catalog_modules: list[dict],
    *,
    version_overrides=None,
    remote_tag_lookup=None,
    public_version_lookup=None,
) -> dict:
    if version_overrides is None:
        version_overrides = {}
    if remote_tag_lookup is None:
        remote_tag_lookup = remote_tag_revision
    if public_version_lookup is None:
        public_version_lookup = public_module_version
    item = copy.deepcopy(template)
    selected = active_catalog_modules(catalog_modules)
    go_floor = item.get("go", {}).get("version")
    validate_string(go_floor, "compatibility set Go version")
    go_versions = [go_floor]
    for module in selected:
        module_go_version = module.get("go_version")
        if module_go_version is not None:
            validate_string(module_go_version, "catalog module Go version")
            go_versions.append(module_go_version)
    parsed_go_versions = []
    for version in go_versions:
        match = GO_VERSION.fullmatch(version)
        if match is None:
            raise ValueError(f"invalid Go version: {version}")
        parsed_go_versions.append((tuple(map(int, match.groups())), version))
    item["go"]["version"] = max(parsed_go_versions)[1]
    selected_paths = {module["module_path"] for module in selected}
    unknown_overrides = sorted(set(version_overrides) - selected_paths)
    if unknown_overrides:
        raise ValueError(
            f"version override names unknown module: {unknown_overrides[0]}"
        )
    roster = item.get("roster")
    if isinstance(roster, dict) and roster.get("module_count") != len(selected):
        raise ValueError(
            f"roster mismatch: expected {roster.get('module_count')} active modules, "
            f"catalog has {len(selected)}"
        )
    def resolve(catalog_module: dict) -> dict:
        version = version_overrides.get(
            catalog_module["module_path"], catalog_module.get("version")
        )
        validate_string(version, "catalog module version")
        version = "v" + version.removeprefix("v")
        if not SEMVER.fullmatch(version):
            raise ValueError(
                f"catalog module version is not semantic: {catalog_module['module_path']}"
            )
        repository = repository_identity(catalog_module)
        tag = expected_tag(catalog_module, version)
        revision = remote_tag_lookup(repository, tag)
        if revision is None:
            raise ValueError(
                f"active module lacks a remote release tag: {catalog_module['module_path']}"
            )
        if not re.fullmatch(r"[0-9a-f]{40}", revision):
            raise ValueError(
                f"active module has an invalid release revision: {catalog_module['module_path']}"
            )
        if not public_version_lookup(catalog_module["module_path"], version):
            raise ValueError(
                f"active module version is unavailable from the public proxy: {catalog_module['module_path']}"
            )
        return {
            "module_path": catalog_module["module_path"],
            "version": version,
            "source_revision": revision,
        }

    with concurrent.futures.ThreadPoolExecutor(
        max_workers=min(16, max(1, len(selected)))
    ) as executor:
        modules = list(executor.map(resolve, selected))
    item["modules"] = modules
    validate_complete_roster(item, catalog_modules)
    evidence = item.get("evidence")
    if not isinstance(evidence, dict):
        raise ValueError("evidence must be an object")
    evidence.pop("content_sha256", None)
    evidence["content_sha256"] = content_fingerprint(item)
    return item


def render_clean_consumer(
    item: dict, catalog_modules: list[dict]
) -> tuple[str, str]:
    by_path = {module.get("module_path"): module for module in catalog_modules}
    requirements = []
    imports = set()
    for selected in item["modules"]:
        module_path = selected["module_path"]
        catalog_module = by_path.get(module_path)
        if catalog_module is None:
            raise ValueError(f"consumer module is absent from catalog: {module_path}")
        requirements.append((module_path, selected["version"]))
        package_names = {
            package.get("import_path"): package.get("name")
            for package in catalog_module.get("packages", [])
        }
        cohesion = catalog_module.get("cohesion") or {}
        for entry in cohesion.get("primary_entry_packages", []):
            if package_names.get(entry) != "main":
                imports.add(entry)
    requirements.sort()
    go_version = item.get("go", {}).get("version")
    validate_string(go_version, "Go version")
    go_mod = [
        "module github.com/faustbrian/go-library-tools/release/compatibility-consumer",
        "",
        f"go {go_version}",
        "",
        "require (",
    ]
    go_mod.extend(f"\t{module_path} {version}" for module_path, version in requirements)
    go_mod += [")", ""]
    test_source = ["package compatibilityconsumer_test", "", "import ("]
    test_source.extend(f'\t_ "{entry}"' for entry in sorted(imports))
    test_source += [")", ""]
    return "\n".join(go_mod), "\n".join(test_source)


def ephemeral_rebase_arguments(item: dict, go_mod: dict) -> list[str]:
    if go_mod.get("Replace"):
        raise ValueError("ephemeral compatibility rebases reject replace directives")
    versions = {
        module["module_path"]: module["version"] for module in item["modules"]
    }
    arguments = []
    for requirement in sorted(
        go_mod.get("Require") or [], key=lambda value: value.get("Path", "")
    ):
        module_path = requirement.get("Path")
        version = versions.get(module_path)
        if version is not None:
            arguments.append(f"-require={module_path}@{version}")
    return arguments


def load_catalog_modules(*, consumer_only: bool = False) -> list[dict]:
    consumer_path = ROOT / "docs" / "ecosystem" / "catalog-consumer.json"
    consumer_paths = {
        module.get("module_path")
        for module in json.loads(consumer_path.read_text()).get("modules", [])
    }
    catalog_modules = {}
    for catalog_name in ("catalog-consumer.json", "catalog-engineering.json"):
        catalog = json.loads((ROOT / "docs/ecosystem" / catalog_name).read_text())
        for module in catalog.get("modules", []):
            module_path = module.get("module_path")
            if consumer_only and module_path not in consumer_paths:
                continue
            if module_path not in catalog_modules:
                catalog_modules[module_path] = {}
            catalog_modules[module_path].update(module)
    return list(catalog_modules.values())


def project_catalog_membership(value: dict, catalog: dict) -> dict:
    memberships: dict[tuple[str, str], list[str]] = {}
    for item in value.get("sets", []):
        if item.get("publication_status") != "published":
            continue
        for module in item.get("modules", []):
            key = (module["module_path"], module["version"].lstrip("v"))
            memberships.setdefault(key, []).append(item["set_id"])

    projected = copy.deepcopy(catalog)
    for module in projected.get("modules", []):
        cohesion = module.get("cohesion")
        if not isinstance(cohesion, dict):
            continue
        key = (module.get("module_path"), str(module.get("version", "")).lstrip("v"))
        cohesion["known_good_compatibility_sets"] = sorted(
            memberships.get(key, [])
        )
    return projected


def project_catalogs(value: dict, directory: Path, *, write: bool) -> None:
    for name in ("catalog-consumer.json", "catalog-engineering.json"):
        path = directory / name
        projected = project_catalog_membership(value, json.loads(path.read_text()))
        encoded = json.dumps(projected, indent=2, ensure_ascii=False) + "\n"
        if write:
            path.write_text(encoded)
        elif path.read_text() != encoded:
            raise ValueError(f"{name} compatibility-set membership is stale")


def validate_residuals(residuals: dict, compatibility_sets: dict) -> None:
    if set(residuals) != {
        "format", "status", "compatibility_set", "residuals"
    }:
        raise ValueError("residual register fields mismatch")
    if residuals["format"] != "golib-cohesion-residuals-v1":
        raise ValueError("unexpected residual register format")
    if residuals["status"] != "current":
        raise ValueError("residual register status must be current")
    set_ids = {item.get("set_id") for item in compatibility_sets.get("sets", [])}
    if residuals["compatibility_set"] not in set_ids:
        raise ValueError("residual register compatibility set is unknown")
    published_set_ids = {
        item.get("set_id")
        for item in compatibility_sets.get("sets", [])
        if item.get("publication_status") == "published"
        and item.get("installable") is True
    }
    if residuals["compatibility_set"] not in published_set_ids:
        raise ValueError("residual register compatibility set is not published")
    entries = residuals["residuals"]
    if not isinstance(entries, list) or not entries:
        raise ValueError("residual register requires entries")
    expected_fields = {
        "id", "scope", "classification", "consumer_impact", "owner",
        "removal_condition",
    }
    seen = set()
    for entry in entries:
        if not isinstance(entry, dict) or set(entry) != expected_fields:
            raise ValueError("residual fields mismatch")
        if not isinstance(entry["id"], str) or not re.fullmatch(
            r"[a-z0-9]+(?:-[a-z0-9]+)*", entry["id"]
        ):
            raise ValueError("residual id must use lowercase kebab case")
        if entry["id"] in seen:
            raise ValueError(f"duplicate residual id: {entry['id']}")
        seen.add(entry["id"])
        scope = entry["scope"]
        if (
            not isinstance(scope, list) or not scope or
            scope != sorted(set(scope)) or
            any(not isinstance(item, str) or not item.strip() for item in scope)
        ):
            raise ValueError("residual scope must be sorted, unique, and nonempty")
        for field in expected_fields - {"id", "scope"}:
            validate_string(entry[field], f"residual {field}")


def select_set(value: dict, set_id: str | None) -> dict:
    sets = value.get("sets", [])
    if set_id is None:
        if len(sets) != 1:
            raise ValueError("set id is required when multiple compatibility sets exist")
        return sets[0]
    for item in sets:
        if item.get("set_id") == set_id:
            return item
    raise ValueError(f"unknown compatibility set: {set_id}")


def select_candidate_template(value: dict) -> dict:
    candidates = [
        item
        for item in value.get("sets", [])
        if item.get("publication_status") == "unreleased"
    ]
    if len(candidates) != 1:
        raise ValueError("candidate generation requires exactly one unreleased template")
    return candidates[0]


def select_clean_consumer_set(value: dict) -> dict | None:
    roster_sets = [item for item in value.get("sets", []) if "roster" in item]
    candidates = [
        item
        for item in roster_sets
        if item.get("publication_status") == "unreleased"
    ]
    if len(candidates) > 1:
        raise ValueError("multiple unreleased compatibility candidates")
    if candidates:
        return candidates[0]
    published = [
        item
        for item in roster_sets
        if item.get("publication_status") == "published"
    ]
    return published[-1] if published else None


def render_source_reference(reference: str, *, label: str | None = None) -> str:
    github = re.fullmatch(
        r"(github\.com/[^/@]+/[^/@]+)@([0-9a-f]{7,40})(?:/(.+))?", reference
    )
    if github is not None:
        repository, revision, path = github.groups()
        text = label or f"{repository}@{revision}"
        rendered = f"[{text}](https://{repository}/commit/{revision})"
        if path:
            rendered += f" / `{path}`"
        return rendered
    relative = os.path.relpath(ROOT / reference, OUTPUT.parent)
    return f"[{label or reference}]({Path(relative).as_posix()})"


def write_candidate_artifacts(value: dict, item: dict, catalog_modules: list[dict]) -> None:
    candidate_value = copy.deepcopy(value)
    matching = [
        index
        for index, existing in enumerate(candidate_value.get("sets", []))
        if existing.get("set_id") == item["set_id"]
    ]
    if not matching:
        matching = [
            index
            for index, existing in enumerate(candidate_value.get("sets", []))
            if existing.get("publication_status") == "unreleased"
        ]
    if len(matching) > 1:
        raise ValueError("candidate write requires one replaceable compatibility set")
    if matching:
        candidate_value["sets"][matching[0]] = item
    else:
        candidate_value.setdefault("sets", []).append(item)
    go_mod, test_source = render_clean_consumer(item, catalog_modules)
    SOURCE.write_text(json.dumps(candidate_value, indent=2, ensure_ascii=False) + "\n")
    OUTPUT.write_text(render(candidate_value))
    CONSUMER_DIRECTORY.mkdir(parents=True, exist_ok=True)
    (CONSUMER_DIRECTORY / "go.mod").write_text(go_mod)
    (CONSUMER_DIRECTORY / "consumer_test.go").write_text(test_source)


def check_clean_consumer(item: dict, catalog_modules: list[dict]) -> None:
    if item.get("roster") is None:
        return
    go_mod, test_source = render_clean_consumer(item, catalog_modules)
    expected = {
        CONSUMER_DIRECTORY / "go.mod": go_mod,
        CONSUMER_DIRECTORY / "consumer_test.go": test_source,
    }
    for path, content in expected.items():
        if not path.is_file() or path.read_text() != content:
            raise ValueError(f"generated compatibility consumer is stale: {path.name}")


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
    allow_stale_unreleased=False,
) -> dict:
    value = json.loads(SOURCE.read_text())
    if value.get("format") != "golib-compatibility-sets-v1":
        raise ValueError("unexpected compatibility-set format")
    sets = value.get("sets")
    if not isinstance(sets, list) or not sets:
        raise ValueError("sets must be a non-empty array")
    seen = set()
    catalogs = load_catalog_modules()
    roster_catalogs = load_catalog_modules(consumer_only=True)
    catalog_modules = {module.get("module_path"): module for module in catalogs}
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
        validate_roster(item)
        evidence = item["evidence"]
        if not isinstance(evidence, dict):
            raise ValueError("evidence must be an object")
        observation = evidence.get("observation")
        validate_string(observation, "evidence observation")
        observation = observation.lower()
        module_ids = set()
        public_bindings = []
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
            if status == "published" and catalog_module is None:
                raise ValueError(f"module identity is absent or version-mismatched: {module['module_path']}")
            if status == "published":
                tag = expected_tag(catalog_module, module["version"])
                repository = repository_identity(catalog_module)
                public_bindings.append((module, repository, tag))
            elif not allow_stale_unreleased:
                if (
                    catalog_module is None
                    or (
                        item.get("roster") is None
                        and (
                            catalog_version is None
                            or "v" + catalog_version.lstrip("v")
                            != module["version"]
                        )
                    )
                ):
                    raise ValueError(
                        "module identity is absent or version-mismatched: "
                        + module["module_path"]
                    )
                expected_tag(catalog_module, module["version"])
                repository_identity(catalog_module)

        def verify_public_binding(binding: tuple[dict, str, str]) -> None:
            module, repository, tag = binding
            if remote_tag_lookup(repository, tag) != module["source_revision"]:
                raise ValueError(
                    f"published module revision lacks matching remote tag: {module['module_path']}"
                )
            if not public_version_lookup(module["module_path"], module["version"]):
                raise ValueError(
                    f"published module version is unavailable from the public proxy: {module['module_path']}"
                )

        with concurrent.futures.ThreadPoolExecutor(
            max_workers=min(16, max(1, len(public_bindings)))
        ) as executor:
            list(executor.map(verify_public_binding, public_bindings))
        if status == "unreleased" and not allow_stale_unreleased:
            validate_complete_roster(item, roster_catalogs)
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
    current = select_clean_consumer_set(value)
    if current is not None and current.get("publication_status") == "published":
        validate_complete_roster(current, roster_catalogs)
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
        if "roster" in item:
            lines += [
                "",
                "### Roster",
                "",
                f"- Selection: `{item['roster']['selection']}`",
                f"- Modules: `{item['roster']['module_count']}`",
            ]
        if "recipes" in item:
            lines += ["", "### Recipes", ""]
            lines += [
                f"- `{recipe['id']}` ({recipe['status']}): "
                f"{render_source_reference(recipe['source_reference'], label='source')}"
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
        lines += ["", "### Sources", ""]
        lines += [
            f"- {render_source_reference(reference)}"
            for reference in item["source_references"]
        ]
        lines += ["", "### Evidence", "", f"- Content fingerprint: `{item['evidence']['content_sha256']}`", f"- Observation: `{item['evidence']['observation']}`", ""]
    return "\n".join(lines)


def parse_args(arguments: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate and validate Golib compatibility-set artifacts."
    )
    subparsers = parser.add_subparsers(dest="command")
    subparsers.add_parser("check")
    candidate = subparsers.add_parser("candidate")
    candidate.add_argument("--set-id", required=True)
    candidate.add_argument("--observed-at", required=True)
    candidate.add_argument("--module-count", required=True, type=int)
    candidate.add_argument(
        "--version",
        action="append",
        default=[],
        metavar="MODULE@VERSION",
    )
    candidate.add_argument("--write", action="store_true")
    rebase = subparsers.add_parser("rebase")
    rebase.add_argument("--set-id")
    rebase.add_argument("--go-mod", required=True, type=Path)
    catalogs = subparsers.add_parser("catalogs")
    catalogs.add_argument("--directory", required=True, type=Path)
    catalogs.add_argument("--write", action="store_true")
    return parser.parse_args(arguments)


def validate_candidate_request(set_id: str, observed_at: str, module_count: int) -> None:
    if not DRAFT_SET_ID.fullmatch(set_id):
        raise ValueError("unreleased set id must use the draft form")
    try:
        datetime.datetime.strptime(observed_at, "%Y-%m-%dT%H:%M:%SZ")
    except ValueError as error:
        raise ValueError("candidate observation time must use UTC RFC 3339 form") from error
    if module_count < 1:
        raise ValueError("candidate module count must be positive")


def parse_version_overrides(bindings: list[str]) -> dict[str, str]:
    overrides = {}
    for binding in bindings:
        if "@" not in binding:
            raise ValueError("candidate version override must use MODULE@VERSION")
        module_path, version = binding.rsplit("@", 1)
        validate_string(module_path, "candidate version override module")
        if not SEMVER.fullmatch(version):
            raise ValueError("candidate version override must use a semantic version")
        if module_path in overrides:
            raise ValueError(f"duplicate candidate version override: {module_path}")
        overrides[module_path] = version
    return overrides


def main(arguments: list[str] | None = None) -> int:
    options = parse_args([] if arguments is None else arguments)
    command = options.command or "check"
    if command == "candidate":
        validate_candidate_request(
            options.set_id, options.observed_at, options.module_count
        )
        value = load(allow_stale_unreleased=True)
        template = copy.deepcopy(select_candidate_template(value))
        template.update(
            {
                "set_id": options.set_id,
                "publication_status": "unreleased",
                "installable": False,
                "observed_at": options.observed_at,
                "roster": {
                    "selection": "active-public",
                    "module_count": options.module_count,
                },
            }
        )
        template["evidence"] = {
            "observation": "Pending final composition and native clean-consumer receipts."
        }
        catalogs = load_catalog_modules(consumer_only=True)
        item = build_candidate(
            template,
            catalogs,
            version_overrides=parse_version_overrides(options.version),
        )
        if options.write:
            write_candidate_artifacts(value, item, catalogs)
        else:
            print(json.dumps(item, indent=2, ensure_ascii=False))
        return 0
    if command == "rebase":
        data = load()
        item = select_set(data, options.set_id)
        with tempfile.TemporaryDirectory(prefix="golib-compatibility-rebase-") as task:
            environment = dict(os.environ)
            for name, directory in (
                ("GOCACHE", "cache"),
                ("GOMODCACHE", "mod"),
                ("GOTMPDIR", "tmp"),
            ):
                path = Path(task) / directory
                path.mkdir()
                environment[name] = str(path)
            result = subprocess.run(
                ["go", "mod", "edit", "-json", str(options.go_mod)],
                capture_output=True,
                text=True,
                check=True,
                env=environment,
            )
            go_mod = json.loads(result.stdout)
            edit_arguments = ephemeral_rebase_arguments(item, go_mod)
            if edit_arguments:
                subprocess.run(
                    ["go", "mod", "edit", *edit_arguments, str(options.go_mod)],
                    check=True,
                    env=environment,
                )
        print(f"rebased {len(edit_arguments)} compatibility requirement(s)")
        return 0
    if command == "catalogs":
        data = load()
        project_catalogs(data, options.directory, write=options.write)
        print("projected compatibility-set catalog membership")
        return 0
    data = load()
    rendered = render(data)
    if OUTPUT.exists() and OUTPUT.read_text() != rendered:
        raise ValueError("compatibility-sets.md is stale; regenerate it")
    project_catalogs(data, ROOT / "docs" / "ecosystem", write=False)
    if RESIDUALS.exists():
        validate_residuals(json.loads(RESIDUALS.read_text()), data)
    catalogs = load_catalog_modules()
    consumer_set = select_clean_consumer_set(data)
    if consumer_set is not None:
        check_clean_consumer(consumer_set, catalogs)
    print(f"validated {len(data['sets'])} compatibility set(s)")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main(sys.argv[1:]))
    except (OSError, ValueError, subprocess.CalledProcessError) as error:
        raise SystemExit(str(error)) from error
