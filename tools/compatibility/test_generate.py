import copy
import hashlib
import importlib.util
import json
import subprocess
import tempfile
import unittest
from unittest import mock
from pathlib import Path


spec = importlib.util.spec_from_file_location(
    "generate", Path(__file__).with_name("generate.py")
)
generate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(generate)


class CompatibilitySetTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name) / "go-library-tools"
        self.ecosystem = self.root / "docs" / "ecosystem"
        self.ecosystem.mkdir(parents=True)
        self.old_root = generate.ROOT
        self.old_source = generate.SOURCE
        self.old_output = generate.OUTPUT
        self.old_consumer_directory = generate.CONSUMER_DIRECTORY
        self.old_residuals = generate.RESIDUALS
        generate.ROOT = self.root
        generate.SOURCE = self.ecosystem / "compatibility-sets.json"
        generate.OUTPUT = self.ecosystem / "compatibility-sets.md"
        generate.CONSUMER_DIRECTORY = self.root / "release/compatibility-consumer"
        generate.RESIDUALS = self.root / "release/cohesion-residuals.json"

    def tearDown(self):
        generate.ROOT = self.old_root
        generate.SOURCE = self.old_source
        generate.OUTPUT = self.old_output
        generate.CONSUMER_DIRECTORY = self.old_consumer_directory
        generate.RESIDUALS = self.old_residuals
        self.temp.cleanup()

    def create_repository(self, name="go-parent"):
        repository = self.root.parent / name
        repository.mkdir()
        subprocess.run(["git", "init", "-q", str(repository)], check=True)
        subprocess.run(
            ["git", "-C", str(repository), "config", "user.name", "Test"],
            check=True,
        )
        subprocess.run(
            ["git", "-C", str(repository), "config", "user.email", "test@example.com"],
            check=True,
        )
        (repository / "README.md").write_text("fixture\n")
        subprocess.run(["git", "-C", str(repository), "add", "README.md"], check=True)
        subprocess.run(
            ["git", "-C", str(repository), "commit", "-q", "-m", "fixture"],
            check=True,
        )
        revision = subprocess.run(
            ["git", "-C", str(repository), "rev-parse", "HEAD"],
            capture_output=True,
            check=True,
            text=True,
        ).stdout.strip()
        return repository, revision

    def write_catalogs(self, modules):
        consumer_modules = []
        for module in modules:
            consumer_modules.append(
                {
                    key: module[key]
                    for key in ("repository", "directory", "module_path", "version")
                }
            )
        (self.ecosystem / "catalog-consumer.json").write_text(
            json.dumps({"modules": consumer_modules})
        )
        (self.ecosystem / "catalog-engineering.json").write_text(
            json.dumps({"modules": modules})
        )

    def base_set(self, revision, **overrides):
        item = {
            "set_id": "golib-compat-v1-20260909.1",
            "publication_status": "published",
            "installable": True,
            "go": {"version": "1.26.6", "os_arch": ["darwin/arm64"]},
            "modules": [
                {
                    "module_path": "github.com/faustbrian/go-parent/adapters/foo",
                    "version": "v1.2.3",
                    "source_revision": revision,
                }
            ],
            "scenarios": ["exercise the public adapter"],
            "observed_at": "2026-09-09T00:00:00Z",
            "evidence": {"observation": "verified public resolution"},
            "caveats": ["No production acceptance is implied."],
            "source_references": ["docs/ecosystem/design-language.md"],
        }
        item.update(overrides)
        item["evidence"] = dict(item["evidence"])
        item["evidence"]["content_sha256"] = self.fingerprint(item)
        return item

    @staticmethod
    def fingerprint(item):
        probe = copy.deepcopy(item)
        probe["evidence"].pop("content_sha256", None)
        payload = json.dumps(
            probe, sort_keys=True, separators=(",", ":"), ensure_ascii=False
        ).encode()
        return "sha256:" + hashlib.sha256(payload).hexdigest()

    def write_set(self, item):
        (self.ecosystem / "design-language.md").write_text("# Design language\n")
        generate.SOURCE.write_text(
            json.dumps({"format": "golib-compatibility-sets-v1", "sets": [item]})
        )

    def load_published(self, revision, remote_tag_lookup=None, public_version_lookup=None):
        if remote_tag_lookup is None:
            remote_tag_lookup = lambda _repository, _tag: revision
        if public_version_lookup is None:
            public_version_lookup = lambda _module_path, _version: True
        return generate.load(
            remote_tag_lookup=remote_tag_lookup,
            public_version_lookup=public_version_lookup,
        )

    def catalog_module(self):
        return {
            "repository": "github.com/faustbrian/go-parent",
            "directory": "adapters/foo",
            "module_path": "github.com/faustbrian/go-parent/adapters/foo",
            "version": "1.2.3",
            "tag_prefix": "adapters/foo/v",
            "releasable": True,
        }

    def test_published_nested_module_uses_repository_and_tag_prefix(self):
        revision = "a" * 40
        self.write_catalogs([self.catalog_module()])
        self.write_set(self.base_set(revision))

        value = self.load_published(revision)

        self.assertEqual(value["sets"][0]["modules"][0]["version"], "v1.2.3")

    def test_published_set_requires_exact_expected_tag(self):
        revision = "a" * 40
        self.write_catalogs([self.catalog_module()])
        self.write_set(self.base_set(revision))

        def exact_tag_only(_repository, tag):
            if tag == "adapters/foo/v1.2.3":
                return revision
            return None

        self.load_published(revision, exact_tag_only)
        module = self.catalog_module()
        module["tag_prefix"] = "v"
        self.write_catalogs([module])
        with self.assertRaisesRegex(ValueError, "matching remote tag"):
            self.load_published(revision, exact_tag_only)

    def test_publication_state_and_identifier_are_frozen(self):
        repository, revision = self.create_repository()
        subprocess.run(
            ["git", "-C", str(repository), "tag", "adapters/foo/v1.2.3"],
            check=True,
        )
        self.write_catalogs([self.catalog_module()])
        invalid = (
            {"publication_status": "published", "installable": False},
            {"publication_status": "candidate", "installable": False},
            {"set_id": "golib-core-2026-09"},
            {
                "set_id": "draft-20260909.1",
                "publication_status": "published",
                "installable": True,
            },
        )
        for changes in invalid:
            with self.subTest(changes=changes):
                item = self.base_set(revision, **changes)
                self.write_set(item)
                with self.assertRaises(ValueError):
                    self.load_published(revision)

    def test_unreleased_draft_may_defer_local_public_binding(self):
        revision = "a" * 40
        self.write_catalogs([self.catalog_module()])
        item = self.base_set(
            revision,
            set_id="draft-20260909.1",
            publication_status="unreleased",
            installable=False,
            evidence={"observation": "pending public release"},
        )
        self.write_set(item)

        generate.load(
            remote_tag_lookup=lambda _repository, _tag: self.fail(
                "draft performed remote tag verification"
            ),
            public_version_lookup=lambda _module_path, _version: self.fail(
                "draft performed public proxy verification"
            ),
        )

    def test_published_set_requires_remote_tag_and_public_proxy_identity(self):
        repository, revision = self.create_repository()
        subprocess.run(
            ["git", "-C", str(repository), "tag", "adapters/foo/v1.2.3"],
            check=True,
        )
        self.write_catalogs([self.catalog_module()])
        self.write_set(self.base_set(revision))
        calls = []

        def remote_tag_lookup(repository_path, tag):
            calls.append(("remote", repository_path, tag))
            return revision

        def public_version_lookup(module_path, version):
            calls.append(("proxy", module_path, version))
            return True

        self.load_published(revision, remote_tag_lookup, public_version_lookup)

        self.assertEqual(
            calls,
            [
                (
                    "remote",
                    "github.com/faustbrian/go-parent",
                    "adapters/foo/v1.2.3",
                ),
                (
                    "proxy",
                    "github.com/faustbrian/go-parent/adapters/foo",
                    "v1.2.3",
                ),
            ],
        )
        with self.assertRaisesRegex(ValueError, "remote tag"):
            self.load_published(revision, lambda _repository, _tag: "b" * 40)
        with self.assertRaisesRegex(ValueError, "public proxy"):
            self.load_published(
                revision,
                public_version_lookup=lambda _module_path, _version: False,
            )

    def test_published_set_rejects_repository_outside_ecosystem(self):
        revision = "a" * 40
        module = self.catalog_module()
        module["repository"] = "example.com/untrusted/go-parent"
        self.write_catalogs([module])
        self.write_set(self.base_set(revision))

        with self.assertRaisesRegex(ValueError, "invalid catalog repository"):
            self.load_published(revision)

    def test_remote_tag_revision_prefers_annotated_tag_peel(self):
        result = subprocess.CompletedProcess(
            args=[],
            returncode=0,
            stdout=(
                "b" * 40 + "\trefs/tags/adapters/foo/v1.2.3\n"
                + "a" * 40 + "\trefs/tags/adapters/foo/v1.2.3^{}\n"
            ),
        )
        with mock.patch.object(generate.subprocess, "run", return_value=result) as run:
            revision = generate.remote_tag_revision(
                "github.com/faustbrian/go-parent", "adapters/foo/v1.2.3"
            )

        self.assertEqual(revision, "a" * 40)
        self.assertEqual(
            run.call_args.args[0][3], "https://github.com/faustbrian/go-parent.git"
        )

    def test_public_module_version_checks_bounded_proxy_metadata(self):
        class Response:
            def __enter__(self):
                return self

            def __exit__(self, *_args):
                return False

            def read(self, _limit):
                return b'{"Version":"v1.2.3-RC1"}'

        with mock.patch.object(
            generate.urllib.request, "urlopen", return_value=Response()
        ) as urlopen:
            available = generate.public_module_version(
                "github.com/FaustBrian/go-parent/adapters/foo", "v1.2.3-RC1"
            )

        self.assertTrue(available)
        self.assertIn(
            "github.com/!faust!brian/go-parent/adapters/foo/@v/v1.2.3-!r!c1.info",
            urlopen.call_args.args[0],
        )

    def test_optional_structured_guidance_is_accepted_and_rendered(self):
        repository, revision = self.create_repository()
        subprocess.run(
            ["git", "-C", str(repository), "tag", "adapters/foo/v1.2.3"],
            check=True,
        )
        self.write_catalogs([self.catalog_module()])
        item = self.base_set(
            revision,
            recipes=[
                {
                    "id": "queue-worker",
                    "status": "verified",
                    "source_reference": "docs/ecosystem/design-language.md",
                }
            ],
            external_versions=[{"name": "PostgreSQL", "version": "18"}],
            upgrade={"notes": ["Upgrade modules independently."]},
            rollback={"notes": ["Restore the previous compatible set."]},
            exclusions=[
                {"subject": "production deployment", "reason": "not exercised"}
            ],
        )
        self.write_set(item)

        rendered = generate.render(self.load_published(revision))

        self.assertIn("### Recipes", rendered)
        self.assertIn("`queue-worker` (verified)", rendered)
        self.assertIn("[source](design-language.md)", rendered)
        self.assertIn("### Sources", rendered)
        self.assertIn("[docs/ecosystem/design-language.md](design-language.md)", rendered)
        self.assertIn("### External versions", rendered)
        self.assertIn("PostgreSQL: `18`", rendered)
        self.assertIn("### Upgrade", rendered)
        self.assertIn("### Rollback", rendered)
        self.assertIn("### Exclusions", rendered)

    def test_candidate_generation_selects_complete_sorted_active_roster(self):
        modules = []
        revisions = {}
        for index in range(128):
            module = self.catalog_module()
            module_path = f"github.com/faustbrian/go-parent/adapters/a{index:03d}"
            module.update(
                {
                    "directory": f"adapters/a{index:03d}",
                    "module_path": module_path,
                    "version": "1.2.3",
                    "tag_prefix": f"adapters/a{index:03d}/v",
                    "cohesion": {"lifecycle_status": "active"},
                }
            )
            modules.append(module)
            revisions[(module["repository"], module["tag_prefix"] + "1.2.3")] = (
                f"{index + 1:040x}"
            )
        for lifecycle in ("deprecated", "planned"):
            module = self.catalog_module()
            module["module_path"] += "/" + lifecycle
            module["cohesion"] = {"lifecycle_status": lifecycle}
            modules.append(module)

        item = self.base_set("a" * 40)
        item["modules"] = []
        item["roster"] = {"selection": "active-public", "module_count": 128}
        item = generate.build_candidate(
            item,
            list(reversed(modules)),
            remote_tag_lookup=lambda repository, tag: revisions.get((repository, tag)),
            public_version_lookup=lambda _module_path, _version: True,
        )

        self.assertEqual(len(item["modules"]), 128)
        self.assertEqual(
            [module["module_path"] for module in item["modules"]],
            sorted(module["module_path"] for module in modules[:128]),
        )
        self.assertEqual(item["evidence"]["content_sha256"], self.fingerprint(item))

    def test_candidate_generation_rejects_unpublished_active_module(self):
        module = self.catalog_module()
        module["cohesion"] = {"lifecycle_status": "active"}
        item = self.base_set("a" * 40)
        item["modules"] = []
        item["roster"] = {"selection": "active-public", "module_count": 1}

        with self.assertRaisesRegex(ValueError, "remote release tag"):
            generate.build_candidate(
                item,
                [module],
                remote_tag_lookup=lambda _repository, _tag: None,
                public_version_lookup=lambda _module_path, _version: True,
            )
        with self.assertRaisesRegex(ValueError, "public proxy"):
            generate.build_candidate(
                item,
                [module],
                remote_tag_lookup=lambda _repository, _tag: "a" * 40,
                public_version_lookup=lambda _module_path, _version: False,
            )

    def test_candidate_generation_accepts_deliberate_public_version_override(self):
        module = self.catalog_module()
        module["version"] = "1.0.0"
        module["cohesion"] = {"lifecycle_status": "active"}
        item = self.base_set("a" * 40)
        item["modules"] = []
        item["roster"] = {"selection": "active-public", "module_count": 1}
        tags = []

        candidate = generate.build_candidate(
            item,
            [module],
            version_overrides={module["module_path"]: "v1.1.0"},
            remote_tag_lookup=lambda _repository, tag: tags.append(tag) or "b" * 40,
            public_version_lookup=lambda _module_path, _version: True,
        )

        self.assertEqual(candidate["modules"][0]["version"], "v1.1.0")
        self.assertEqual(tags, ["adapters/foo/v1.1.0"])

    def test_candidate_generation_raises_go_floor_to_selected_module_requirement(self):
        module = self.catalog_module()
        module["go_version"] = "1.27.0"
        module["cohesion"] = {"lifecycle_status": "active"}
        item = self.base_set("a" * 40)
        item["modules"] = []
        item["roster"] = {"selection": "active-public", "module_count": 1}

        candidate = generate.build_candidate(
            item,
            [module],
            remote_tag_lookup=lambda _repository, _tag: "b" * 40,
            public_version_lookup=lambda _module_path, _version: True,
        )

        self.assertEqual(candidate["go"]["version"], "1.27.0")

    def test_candidate_generation_rejects_unknown_version_override(self):
        module = self.catalog_module()
        module["cohesion"] = {"lifecycle_status": "active"}
        item = self.base_set("a" * 40)
        item["modules"] = []
        item["roster"] = {"selection": "active-public", "module_count": 1}

        with self.assertRaisesRegex(ValueError, "unknown module"):
            generate.build_candidate(
                item,
                [module],
                version_overrides={"github.com/faustbrian/go-unknown": "v1.1.0"},
                remote_tag_lookup=lambda _repository, _tag: "a" * 40,
                public_version_lookup=lambda _module_path, _version: True,
            )

    def test_complete_roster_rejects_missing_and_deprecated_modules(self):
        active = self.catalog_module()
        active["cohesion"] = {"lifecycle_status": "active"}
        deprecated = dict(active)
        deprecated["module_path"] += "/legacy"
        deprecated["cohesion"] = {"lifecycle_status": "deprecated"}
        item = self.base_set("a" * 40)
        item["modules"] = []
        item["roster"] = {"selection": "active-public", "module_count": 1}

        with self.assertRaisesRegex(ValueError, "roster mismatch"):
            generate.validate_complete_roster(item, [active, deprecated])
        item["modules"] = [
            {
                "module_path": deprecated["module_path"],
                "version": "v1.2.3",
                "source_revision": "a" * 40,
            }
        ]
        with self.assertRaisesRegex(ValueError, "roster mismatch"):
            generate.validate_complete_roster(item, [active, deprecated])

    def test_clean_consumer_covers_modules_and_importable_entry_points(self):
        library = self.catalog_module()
        library["cohesion"] = {
            "lifecycle_status": "active",
            "primary_entry_packages": [library["module_path"]],
        }
        library["packages"] = [
            {"import_path": library["module_path"], "name": "foo"}
        ]
        command = dict(library)
        command["module_path"] = "github.com/faustbrian/go-parent/cmd/tool"
        command["cohesion"] = {
            "lifecycle_status": "active",
            "primary_entry_packages": [command["module_path"]],
        }
        command["packages"] = [
            {"import_path": command["module_path"], "name": "main"}
        ]
        item = self.base_set("a" * 40)
        item["modules"] = [
            {
                "module_path": library["module_path"],
                "version": "v1.2.3",
                "source_revision": "a" * 40,
            },
            {
                "module_path": command["module_path"],
                "version": "v1.2.3",
                "source_revision": "b" * 40,
            },
        ]

        go_mod, test_source = generate.render_clean_consumer(
            item, [library, command]
        )

        self.assertIn(library["module_path"] + " v1.2.3", go_mod)
        self.assertIn(command["module_path"] + " v1.2.3", go_mod)
        self.assertIn('_ "' + library["module_path"] + '"', test_source)
        self.assertNotIn('_ "' + command["module_path"] + '"', test_source)

    def test_ephemeral_rebase_updates_only_existing_candidate_requirements(self):
        item = self.base_set("a" * 40)
        item["modules"].append(
            {
                "module_path": "github.com/faustbrian/go-other",
                "version": "v2.0.0",
                "source_revision": "b" * 40,
            }
        )
        go_mod = {
            "Require": [
                {
                    "Path": "github.com/faustbrian/go-parent/adapters/foo",
                    "Version": "v1.0.0",
                    "Indirect": False,
                },
                {
                    "Path": "example.com/external",
                    "Version": "v1.0.0",
                    "Indirect": True,
                },
            ],
            "Replace": None,
        }

        arguments = generate.ephemeral_rebase_arguments(item, go_mod)

        self.assertEqual(
            arguments,
            [
                "-require=github.com/faustbrian/go-parent/adapters/foo@v1.2.3"
            ],
        )

    def test_ephemeral_rebase_rejects_local_replacements(self):
        item = self.base_set("a" * 40)
        with self.assertRaisesRegex(ValueError, "replace directives"):
            generate.ephemeral_rebase_arguments(
                item,
                {
                    "Require": [],
                    "Replace": [{"Old": {"Path": "example.com/old"}}],
                },
            )

    def test_rebase_command_updates_a_task_checkout(self):
        revision = "a" * 40
        self.write_catalogs([self.catalog_module()])
        item = self.base_set(
            revision,
            set_id="draft-20260910.1",
            publication_status="unreleased",
            installable=False,
            evidence={"observation": "pending final receipts"},
        )
        self.write_set(item)
        checkout = Path(self.temp.name) / "receipt-checkout"
        checkout.mkdir()
        go_mod = checkout / "go.mod"
        go_mod.write_text(
            "module example.com/receipt\n\n"
            "go 1.26.6\n\n"
            "require github.com/faustbrian/go-parent/adapters/foo v1.0.0\n"
        )

        self.assertEqual(
            generate.main(
                [
                    "rebase",
                    "--set-id",
                    item["set_id"],
                    "--go-mod",
                    str(go_mod),
                ]
            ),
            0,
        )

        self.assertIn(
            "github.com/faustbrian/go-parent/adapters/foo v1.2.3",
            go_mod.read_text(),
        )

    def test_candidate_artifacts_are_generated_and_checked_together(self):
        module = self.catalog_module()
        module["cohesion"] = {
            "lifecycle_status": "active",
            "primary_entry_packages": [module["module_path"]],
        }
        module["packages"] = [
            {"import_path": module["module_path"], "name": "foo"}
        ]
        item = self.base_set("a" * 40)
        item["roster"] = {"selection": "active-public", "module_count": 1}
        item["evidence"]["content_sha256"] = self.fingerprint(item)
        value = {"format": "golib-compatibility-sets-v1", "sets": [item]}

        generate.write_candidate_artifacts(value, item, [module])
        generate.check_clean_consumer(item, [module])

        go_mod = generate.CONSUMER_DIRECTORY / "go.mod"
        go_mod.write_text(
            go_mod.read_text()
            + "\nrequire (\n\texample.com/transitive v1.0.0 // indirect\n)\n"
        )
        generate.check_clean_consumer(item, [module])

        go_mod.write_text(
            go_mod.read_text()
            + "\nrequire example.com/unexpected v1.0.0\n"
        )
        with self.assertRaisesRegex(ValueError, "consumer is stale"):
            generate.check_clean_consumer(item, [module])

        self.assertEqual(
            json.loads(generate.SOURCE.read_text())["sets"][0]["roster"]["module_count"],
            1,
        )
        self.assertEqual(generate.OUTPUT.read_text(), generate.render(value))
        self.assertTrue((generate.CONSUMER_DIRECTORY / "go.mod").is_file())
        self.assertTrue(
            (generate.CONSUMER_DIRECTORY / "consumer_test.go").is_file()
        )
        (generate.CONSUMER_DIRECTORY / "consumer_test.go").write_text("stale\n")
        with self.assertRaisesRegex(ValueError, "consumer is stale"):
            generate.check_clean_consumer(item, [module])

    def test_candidate_write_preserves_published_sets(self):
        module = self.catalog_module()
        module["cohesion"] = {
            "lifecycle_status": "active",
            "primary_entry_packages": [module["module_path"]],
        }
        module["packages"] = [
            {"import_path": module["module_path"], "name": "foo"}
        ]
        historical = self.base_set("b" * 40)
        historical["set_id"] = "golib-compat-v1-20260909.1"
        candidate = self.base_set(
            "a" * 40,
            set_id="draft-20260910.1",
            publication_status="unreleased",
            installable=False,
            evidence={"observation": "pending final receipts"},
        )
        candidate["roster"] = {
            "selection": "active-public",
            "module_count": 1,
        }
        candidate["evidence"]["content_sha256"] = self.fingerprint(candidate)
        value = {
            "format": "golib-compatibility-sets-v1",
            "sets": [historical, candidate],
        }

        generate.write_candidate_artifacts(value, candidate, [module])

        written = json.loads(generate.SOURCE.read_text())
        self.assertEqual(
            [item["set_id"] for item in written["sets"]],
            [historical["set_id"], candidate["set_id"]],
        )

    def test_published_roster_remains_valid_after_catalog_changes(self):
        revision = "a" * 40
        historical_module = self.catalog_module()
        historical_module["version"] = "1.3.0"
        historical_module["cohesion"] = {"lifecycle_status": "deprecated"}
        active_module = self.catalog_module()
        active_module.update(
            {
                "directory": "adapters/current",
                "module_path": "github.com/faustbrian/go-parent/adapters/current",
                "tag_prefix": "adapters/current/v",
                "cohesion": {"lifecycle_status": "active"},
            }
        )
        self.write_catalogs([historical_module, active_module])
        historical = self.base_set(revision)
        historical["roster"] = {
            "selection": "active-public",
            "module_count": 1,
        }
        historical["evidence"]["content_sha256"] = self.fingerprint(historical)
        current = self.base_set(
            revision,
            set_id="golib-compat-v1-20260910.1",
        )
        current["modules"][0]["module_path"] = active_module["module_path"]
        current["roster"] = {
            "selection": "active-public",
            "module_count": 1,
        }
        current["evidence"]["content_sha256"] = self.fingerprint(current)
        self.write_set(historical)
        generate.SOURCE.write_text(
            json.dumps(
                {
                    "format": "golib-compatibility-sets-v1",
                    "sets": [historical, current],
                }
            )
        )
        tags = []

        self.load_published(
            revision,
            remote_tag_lookup=lambda _repository, tag: tags.append(tag) or revision,
        )

        self.assertEqual(
            tags,
            ["adapters/foo/v1.2.3", "adapters/current/v1.2.3"],
        )

    def test_current_published_roster_must_match_active_consumer_catalog(self):
        revision = "a" * 40
        active = self.catalog_module()
        additional = self.catalog_module()
        additional.update(
            {
                "directory": "adapters/current",
                "module_path": "github.com/faustbrian/go-parent/adapters/current",
                "tag_prefix": "adapters/current/v",
                "cohesion": {"lifecycle_status": "active"},
            }
        )
        self.write_catalogs([active, additional])
        current = self.base_set(revision)
        current["roster"] = {"selection": "active-public", "module_count": 1}
        current["evidence"]["content_sha256"] = self.fingerprint(current)
        self.write_set(current)

        with self.assertRaisesRegex(ValueError, "roster mismatch"):
            self.load_published(revision)

    def test_clean_consumer_selects_current_candidate_over_history(self):
        historical = self.base_set("a" * 40)
        historical["roster"] = {
            "selection": "active-public",
            "module_count": 1,
        }
        candidate = self.base_set(
            "b" * 40,
            set_id="draft-20260910.1",
            publication_status="unreleased",
            installable=False,
            evidence={"observation": "pending final receipts"},
        )
        candidate["roster"] = {
            "selection": "active-public",
            "module_count": 1,
        }

        selected = generate.select_clean_consumer_set(
            {"sets": [historical, candidate]}
        )

        self.assertIs(selected, candidate)

    def test_catalog_membership_projects_exact_published_module_versions(self):
        item = self.base_set("a" * 40)
        catalog = {
            "modules": [
                {
                    "module_path": "github.com/faustbrian/go-parent/adapters/foo",
                    "version": "1.2.3",
                    "cohesion": {"known_good_compatibility_sets": []},
                },
                {
                    "module_path": "github.com/faustbrian/go-other",
                    "version": "v1.0.0",
                    "cohesion": {
                        "known_good_compatibility_sets": ["stale-set"]
                    },
                },
            ]
        }

        projected = generate.project_catalog_membership(
            {"sets": [item]}, catalog
        )

        self.assertEqual(
            projected["modules"][0]["cohesion"][
                "known_good_compatibility_sets"
            ],
            ["golib-compat-v1-20260909.1"],
        )
        self.assertEqual(
            projected["modules"][1]["cohesion"][
                "known_good_compatibility_sets"
            ],
            [],
        )

    def test_catalog_membership_ignores_unreleased_sets(self):
        item = self.base_set(
            "a" * 40,
            set_id="draft-20260910.1",
            publication_status="unreleased",
            installable=False,
            evidence={"observation": "pending final receipts"},
        )
        catalog = {
            "modules": [
                {
                    "module_path": "github.com/faustbrian/go-parent/adapters/foo",
                    "version": "1.2.3",
                    "cohesion": {"known_good_compatibility_sets": []},
                }
            ]
        }

        projected = generate.project_catalog_membership(
            {"sets": [item]}, catalog
        )

        self.assertEqual(
            projected["modules"][0]["cohesion"][
                "known_good_compatibility_sets"
            ],
            [],
        )

    def test_consumer_roster_loader_excludes_engineering_only_modules(self):
        product = self.catalog_module()
        tooling = copy.deepcopy(product)
        tooling.update(
            {
                "repository": "github.com/faustbrian/go-tooling",
                "module_path": "github.com/faustbrian/go-tooling",
                "directory": ".",
            }
        )
        self.write_catalogs([product, tooling])
        consumer = json.loads(
            (self.ecosystem / "catalog-consumer.json").read_text()
        )
        consumer["modules"] = [consumer["modules"][0]]
        (self.ecosystem / "catalog-consumer.json").write_text(
            json.dumps(consumer)
        )

        modules = generate.load_catalog_modules(consumer_only=True)

        self.assertEqual(
            [module["module_path"] for module in modules],
            ["github.com/faustbrian/go-parent/adapters/foo"],
        )

    def test_residual_register_requires_actionable_removal_conditions(self):
        residuals = {
            "format": "golib-cohesion-residuals-v1",
            "status": "current",
            "compatibility_set": "golib-compat-v1-20260910.1",
            "residuals": [
                {
                    "id": "planned-boundary",
                    "scope": ["github.com/faustbrian/go-parent"],
                    "classification": "planned-product-boundary",
                    "consumer_impact": "No installable module exists.",
                    "owner": "future goal",
                    "removal_condition": "Publish the selected module.",
                }
            ],
        }

        generate.validate_residuals(
            residuals,
            {
                "sets": [
                    {
                        "set_id": "golib-compat-v1-20260910.1",
                        "publication_status": "published",
                        "installable": True,
                    }
                ]
            },
        )
        del residuals["residuals"][0]["removal_condition"]
        with self.assertRaisesRegex(ValueError, "residual fields mismatch"):
            generate.validate_residuals(
                residuals,
                {
                    "sets": [
                        {
                            "set_id": "golib-compat-v1-20260910.1",
                            "publication_status": "published",
                            "installable": True,
                        }
                    ]
                },
            )

    def test_residual_register_rejects_non_string_identifier(self):
        residuals = {
            "format": "golib-cohesion-residuals-v1",
            "status": "current",
            "compatibility_set": "golib-compat-v1-20260910.1",
            "residuals": [
                {
                    "id": None,
                    "scope": ["github.com/faustbrian/go-parent"],
                    "classification": "planned-product-boundary",
                    "consumer_impact": "No installable module exists.",
                    "owner": "future goal",
                    "removal_condition": "Publish the selected module.",
                }
            ],
        }

        with self.assertRaisesRegex(
            ValueError, "residual id must use lowercase kebab case"
        ):
            generate.validate_residuals(
                residuals,
                {
                    "sets": [
                        {
                            "set_id": "golib-compat-v1-20260910.1",
                            "publication_status": "published",
                            "installable": True,
                        }
                    ]
                },
            )

    def test_residual_register_rejects_unpublished_compatibility_set(self):
        residuals = {
            "format": "golib-cohesion-residuals-v1",
            "status": "current",
            "compatibility_set": "draft-20260910.1",
            "residuals": [
                {
                    "id": "planned-boundary",
                    "scope": ["github.com/faustbrian/go-parent"],
                    "classification": "planned-product-boundary",
                    "consumer_impact": "No installable module exists.",
                    "owner": "future goal",
                    "removal_condition": "Publish the selected module.",
                }
            ],
        }

        with self.assertRaisesRegex(
            ValueError, "residual register compatibility set is not published"
        ):
            generate.validate_residuals(
                residuals,
                {
                    "sets": [
                        {
                            "set_id": "draft-20260910.1",
                            "publication_status": "unreleased",
                            "installable": False,
                        }
                    ]
                },
            )

    def test_candidate_command_rejects_invalid_identity_before_resolution(self):
        revision = "a" * 40
        module = self.catalog_module()
        module["cohesion"] = {"lifecycle_status": "active"}
        self.write_catalogs([module])
        item = self.base_set(
            revision,
            set_id="draft-20260910.1",
            publication_status="unreleased",
            installable=False,
            evidence={"observation": "pending final receipts"},
        )
        self.write_set(item)

        with self.assertRaisesRegex(ValueError, "draft form"):
            generate.main(
                [
                    "candidate",
                    "--set-id",
                    "invalid",
                    "--observed-at",
                    "2026-09-10T00:00:00Z",
                    "--module-count",
                    "1",
                ]
            )
        with self.assertRaisesRegex(ValueError, "observation time"):
            generate.main(
                [
                    "candidate",
                    "--set-id",
                    "draft-20260910.2",
                    "--observed-at",
                    "now",
                    "--module-count",
                    "1",
                ]
            )

    def test_candidate_command_replaces_a_stale_unreleased_template(self):
        revision = "a" * 40
        module = self.catalog_module()
        module["version"] = "1.3.0"
        module["cohesion"] = {"lifecycle_status": "active"}
        self.write_catalogs([module])
        stale = self.base_set(
            revision,
            set_id="draft-20260909.1",
            publication_status="unreleased",
            installable=False,
            evidence={"observation": "pending catalog refresh"},
        )
        self.write_set(stale)

        with mock.patch.object(
            generate,
            "build_candidate",
            side_effect=lambda template, _catalogs, **_options: template,
        ) as build_candidate, mock.patch("builtins.print"):
            result = generate.main(
                [
                    "candidate",
                    "--set-id",
                    "draft-20260910.2",
                    "--observed-at",
                    "2026-09-10T00:00:00Z",
                    "--module-count",
                    "1",
                ]
            )

        self.assertEqual(result, 0)
        build_candidate.assert_called_once()


if __name__ == "__main__":
    unittest.main()
