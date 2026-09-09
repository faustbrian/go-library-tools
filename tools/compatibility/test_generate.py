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
        generate.ROOT = self.root
        generate.SOURCE = self.ecosystem / "compatibility-sets.json"
        generate.OUTPUT = self.ecosystem / "compatibility-sets.md"

    def tearDown(self):
        generate.ROOT = self.old_root
        generate.SOURCE = self.old_source
        generate.OUTPUT = self.old_output
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
        self.assertIn("### External versions", rendered)
        self.assertIn("PostgreSQL: `18`", rendered)
        self.assertIn("### Upgrade", rendered)
        self.assertIn("### Rollback", rendered)
        self.assertIn("### Exclusions", rendered)


if __name__ == "__main__":
    unittest.main()
