"""Run the real selection/resource producer, doubling only native CLI commands.

These finite Python fixtures do not compile Go or exercise OS process control.
Native target behavior and hosted resource isolation remain separate proofs.
"""

import contextlib
import copy
import datetime
import importlib.util
import io
import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest import mock


spec = importlib.util.spec_from_file_location(
    "selected_assurance", Path(__file__).with_name("selected-assurance.py")
)
assurance = importlib.util.module_from_spec(spec)
spec.loader.exec_module(assurance)


class SelectedAssuranceTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="golib-assurance-identity-")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name) / "owner"
        self.root.mkdir()
        self.root = self.root.resolve()
        self.cache = Path(self.temp.name) / "cache"
        self.cache.mkdir()
        self.module = "."
        self.policy = {"operations": [{"module": ".", "gate": "fuzz", "steps": []}]}
        self.inventory = {"repository": "github.com/faustbrian/go-wire", "modules": [{
            "directory": ".", "gates": {"fuzz": True}, "packages": [],
            "test_tags": [], "required_services": [],
        }]}
        (self.root / "go.mod").write_text("module example\n\ngo 1.27.0\n")
        self.names = {}
        self.calls = []

    def declare(self, package, selector, race=False, production=True):
        directory = self.root / self.module / package
        directory.mkdir(parents=True, exist_ok=True)
        (directory / "owned_test.go").write_text("package fixture\n")
        if production:
            (directory / "owned.go").write_text("package fixture\n")
            record = {"directory": str(Path(self.module) / package)}
            if record not in self.inventory["modules"][0]["packages"]:
                self.inventory["modules"][0]["packages"].append(record)
        self.policy["operations"][0]["steps"].append({
            "type": "go-test", "packages": [package], "fuzz": selector,
            "race": race, "budget": "10000x", "timeout": "20m",
        })
        self.names.setdefault(package, []).append(selector[1:-1])

    def execute(self, selected="", week=0, event="schedule", failure=None):
        self.calls = []
        sha = "a" * 40
        environment = {
            "GITHUB_REPOSITORY": "faustbrian/go-wire", "GITHUB_REF": "refs/heads/main",
            "GITHUB_EVENT_NAME": event, "GITHUB_SHA": sha,
            "GITHUB_RUN_ID": "1234", "GITHUB_RUN_ATTEMPT": "1",
            "RUNNER_TEMP": str(self.cache), "ASSURANCE_MODULE": self.module,
            "ASSURANCE_TARGET": selected,
        }

        def native(session, args, stage):
            session.stage = stage
            if args == ["git", "rev-parse", "HEAD"]:
                return (sha + "\n").encode()
            if args == ["golib", "config", "show", "--json"]:
                return json.dumps(self.policy).encode()
            if args == ["golib", "inventory", "--json"]:
                return json.dumps(self.inventory).encode()
            self.assertEqual(args[:3], ["go", "test", "-mod=readonly"])
            self.calls.append((args, stage, Path.cwd(), session.env.copy()))
            if failure == stage:
                raise assurance.Refusal("native command failed (exit 23)")
            if stage == "native target discovery":
                return ("\n".join(self.names.get(args[-1], [])) + "\n").encode()
            self.assertEqual(stage, "active fuzz campaign")
            return b""

        instant = datetime.datetime(1970, 1, 5, tzinfo=datetime.timezone.utc) + datetime.timedelta(weeks=week)
        stdout, stderr = io.StringIO(), io.StringIO()
        previous = Path.cwd()
        try:
            os.chdir(self.root)
            with mock.patch.dict(os.environ, environment), \
                    mock.patch.object(assurance.sys, "argv", ["selected-assurance.py"]), \
                    mock.patch.object(assurance.Session, "run", native), \
                    mock.patch.object(assurance.signal, "signal"), \
                    mock.patch.object(assurance.os, "sched_getaffinity", return_value={0, 1}, create=True), \
                    mock.patch.object(assurance.os, "sched_setaffinity", create=True), \
                    mock.patch.object(assurance.datetime, "datetime") as clock, \
                    contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr):
                clock.now.return_value = instant
                assurance.CANCELED = False
                assurance.main()
        finally:
            os.chdir(previous)
            self.assertEqual(list(self.cache.iterdir()), [], "owned cache survived producer outcome")
        return stdout.getvalue()

    def assert_selected(self, package, selector, race=False):
        self.assertEqual(len(self.calls), 2, "selection must discover and execute only one pair")
        native = ["go", "test", "-mod=readonly"] + (["-race"] if race else [])
        self.assertEqual(self.calls[0][0], native + ["-list=" + selector, package])
        self.assertEqual(self.calls[1][0], native + [
            "-run=^$", "-fuzz=" + selector, "-fuzztime=30s", "-parallel=2",
            "-count=1", "-timeout=2m", package,
        ])
        for _, _, directory, environment in self.calls:
            self.assertEqual(directory, self.root / self.module)
            self.assertTrue(Path(environment["GOCACHE"]).is_relative_to(self.cache))
            self.assertEqual(environment["GOFLAGS"], "-p=1")
            self.assertEqual(environment["GOMAXPROCS"], "2")
            self.assertNotIn("GITHUB_TOKEN", environment)

    def test_same_name_in_distinct_packages_rotates_each_declared_pair(self):
        self.declare(".", "^FuzzDecode$")
        self.declare("./jsonwire", "^FuzzDecode$", race=True)
        for week, (package, race) in enumerate([(".", False), ("./jsonwire", True), (".", False)]):
            with self.subTest(week=week):
                self.execute(week=week)
                self.assert_selected(package, "^FuzzDecode$", race)

    def test_declared_test_only_root_is_not_a_missing_production_package(self):
        self.declare(".", "^FuzzPayloadEncoders$", production=False)
        self.execute(selected="^FuzzPayloadEncoders$", event="workflow_dispatch")
        self.assert_selected(".", "^FuzzPayloadEncoders$")

    def test_manual_qualified_pair_and_unique_legacy_selector(self):
        self.declare(".", "^FuzzDecode$")
        self.declare("./jsonwire", "^FuzzDecode$", race=True)
        self.declare("./jsonwire", "^FuzzEncode$")
        for selected, package, selector, race in [
            (".::^FuzzDecode$", ".", "^FuzzDecode$", False),
            ("./jsonwire::^FuzzDecode$", "./jsonwire", "^FuzzDecode$", True),
            ("^FuzzEncode$", "./jsonwire", "^FuzzEncode$", False),
        ]:
            with self.subTest(selected=selected):
                self.execute(selected=selected, event="workflow_dispatch")
                self.assert_selected(package, selector, race)

    def test_manual_ambiguous_unknown_and_unchecked_combined_selectors_refuse(self):
        self.declare(".", "^FuzzDecode$")
        self.declare("./jsonwire", "^FuzzDecode$")
        for selected in [
            "", "^FuzzDecode$", "^FuzzUnknown$", "./unknown::^FuzzDecode$",
            "./jsonwire::^FuzzUnknown$", "jsonwire::^FuzzDecode$", "../jsonwire::^FuzzDecode$",
            "./../jsonwire::^FuzzDecode$", "./jsonwire/...::^FuzzDecode$",
            "./jsonwire::FuzzDecode", "./jsonwire::^FuzzDecode$::extra",
            "./jsonwire::^FuzzDecode$; touch escaped", "x" * 645,
        ]:
            with self.subTest(selected=selected):
                with self.assertRaises(assurance.Refusal):
                    self.execute(selected=selected, event="workflow_dispatch")
                self.assertEqual(self.calls, [], "unchecked manual scope reached Go")

    def test_repeated_pair_and_conflicting_race_refuse_before_native_execution(self):
        self.declare("./jsonwire", "^FuzzDecode$")
        step = self.policy["operations"][0]["steps"][0]
        for race in (False, True):
            with self.subTest(race=race):
                self.policy["operations"][0]["steps"] = [step, dict(step, race=race)]
                with self.assertRaisesRegex(assurance.Refusal, "repeated package fuzz identity"):
                    self.execute()
                self.assertEqual(self.calls, [])

    def test_rotation_retains_all_32_pairs_and_refuses_the_33rd(self):
        for number in range(32):
            self.declare("./p%02d" % number, "^FuzzDecode$")
        for week in range(32):
            self.execute(week=week)
            self.assert_selected("./p%02d" % week, "^FuzzDecode$")
        self.declare("./extra", "^FuzzDecode$")
        last = self.policy["operations"][0]["steps"].pop()
        self.policy["operations"].append({"module": ".", "gate": "fuzz", "steps": [last]})
        with self.assertRaisesRegex(assurance.Refusal, "declared fuzz targets"):
            self.execute()
        self.assertEqual(self.calls, [])

    def test_native_discovery_zero_multiple_and_failure_remain_fatal(self):
        self.declare(".", "^FuzzDecode$", production=False)
        for names in ([], ["FuzzDecode", "FuzzDecode"]):
            with self.subTest(names=names):
                self.names["."] = names
                with self.assertRaisesRegex(assurance.Refusal, "native target missing or ambiguous"):
                    self.execute()
                self.assertEqual(len(self.calls), 1)
        self.names["."] = ["FuzzDecode"]
        for stage, calls in [("native target discovery", 1), ("active fuzz campaign", 2)]:
            with self.subTest(stage=stage):
                with self.assertRaisesRegex(assurance.Refusal, "exit 23"):
                    self.execute(failure=stage)
                self.assertEqual(len(self.calls), calls)

    def test_test_only_subpackage_and_selected_nested_module_preserve_owned_scope(self):
        self.module = "integration/owned"
        directory = self.root / self.module
        directory.mkdir(parents=True)
        (directory / "go.mod").write_text("module example/owned\n\ngo 1.27.0\n")
        self.inventory["modules"][0]["directory"] = self.module
        self.policy["operations"][0]["module"] = self.module
        self.declare(".", "^FuzzRoot$", production=True)
        self.declare("./tests", "^FuzzDecode$", production=False)
        self.execute(selected="./tests::^FuzzDecode$", event="workflow_dispatch")
        self.assert_selected("./tests", "^FuzzDecode$")
        self.execute(selected="^FuzzRoot$", event="workflow_dispatch")
        self.assert_selected(".", "^FuzzRoot$")

    def test_unlisted_production_missing_and_empty_packages_do_not_use_test_only_fallback(self):
        for kind in ("production", "missing", "empty"):
            with self.subTest(kind=kind):
                self.policy["operations"][0]["steps"] = []
                self.inventory["modules"][0]["packages"] = []
                self.declare("./" + kind, "^FuzzDecode$", production=False)
                directory = self.root / kind
                if kind == "production":
                    (directory / "source.go").write_text("package fixture\n")
                elif kind == "missing":
                    directory.rename(self.root / "moved")
                else:
                    (directory / "owned_test.go").unlink()
                with self.assertRaises(assurance.Refusal):
                    self.execute()
                self.assertEqual(self.calls, [])

    def test_symlink_and_nested_module_packages_refuse_before_native_execution(self):
        for kind in ("outside", "inside", "source", "nested", "nested-parent"):
            with self.subTest(kind=kind):
                self.policy["operations"][0]["steps"] = []
                self.inventory["modules"][0]["packages"] = []
                self.declare("./" + kind + "/tests", "^FuzzDecode$", production=False)
                directory = self.root / kind / "tests"
                if kind in ("outside", "inside"):
                    directory.rename(self.root / (kind + "-moved"))
                    destination = self.cache if kind == "outside" else self.root / (kind + "-moved")
                    directory.symlink_to(destination, target_is_directory=True)
                elif kind == "source":
                    source = directory / "owned_test.go"
                    source.unlink()
                    source.symlink_to(self.root / "go.mod")
                else:
                    boundary = directory if kind == "nested" else directory.parent
                    (boundary / "go.mod").write_text("module unselected\n")
                with self.assertRaises(assurance.Refusal):
                    self.execute()
                self.assertEqual(self.calls, [])

    def test_test_only_enumeration_is_bounded(self):
        self.declare("./tests", "^FuzzDecode$", production=False)
        for number in range(4096):
            (self.root / "tests" / ("file%04d" % number)).touch()
        with self.assertRaisesRegex(assurance.Refusal, "file inventory budget"):
            self.execute()
        self.assertEqual(self.calls, [])

    def test_existing_contract_guards_remain_fatal_before_native_execution(self):
        self.declare(".", "^FuzzDecode$")
        policy, inventory = copy.deepcopy(self.policy), copy.deepcopy(self.inventory)
        for kind in ("services", "tags", "runtime", "race", "identity", "disabled", "package", "selector"):
            with self.subTest(kind=kind):
                self.policy, self.inventory = copy.deepcopy(policy), copy.deepcopy(inventory)
                module = self.inventory["modules"][0]
                step = self.policy["operations"][0]["steps"][0]
                if kind == "services": module["required_services"] = ["postgresql"]
                if kind == "tags": module["test_tags"] = ["integration"]
                if kind == "runtime": self.policy["runtimes"] = {"deno": "2.6.4"}
                if kind == "race": step["race"] = "true"
                if kind == "identity": self.inventory["repository"] = "github.com/faustbrian/go-other"
                if kind == "disabled": module["gates"]["fuzz"] = False
                if kind == "package": step["packages"] = ["./..."]
                if kind == "selector": step["fuzz"] = "FuzzDecode"
                with self.assertRaises(assurance.Refusal):
                    self.execute()
                self.assertEqual(self.calls, [])


if __name__ == "__main__":
    unittest.main()
