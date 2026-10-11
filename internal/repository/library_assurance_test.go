package repository_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// The fixture executes the actual workflow command. Only public CLI discovery
// and the slow native Go boundary are doubled; selection and resource ownership
// remain the production runner's behavior.
const libraryAssuranceFixture = `
import json, os, pathlib, signal, subprocess, sys, tempfile, time

with tempfile.TemporaryDirectory(prefix="golib-assurance-fixture-") as tmp, tempfile.TemporaryDirectory(prefix="golib-assurance-outside-") as outside_tmp:
    root = pathlib.Path(tmp)
    bindir = root / "bin"
    bindir.mkdir()
    cache = root / "temp"
    cache.mkdir()
    owner = os.environ.get("FIXTURE_OWNER", "go-clock")
    package = "./manual" if owner == "go-clock" else "."
    targets = ["^FuzzAlpha$", "^FuzzBeta$"]
    policy = {"operations": [{"module": ".", "gate": "fuzz", "steps": [
        {"type": "go-test", "packages": [package], "fuzz": target, "budget": "10000x", "timeout": "2m"}
        for target in targets]}]}
    inventory = {"repository": "github.com/faustbrian/" + owner, "modules": [{"directory": ".", "gates": {"fuzz": True}, "test_tags": [], "required_services": [], "packages": [{"directory": package.removeprefix("./")}]}]}
    if owner == "go-http-client": inventory["modules"][0]["external_runtime_dependencies"] = ["go.uber.org/goleak", "golang.org/x/net", "golang.org/x/oauth2"]
    case = os.environ.get("FIXTURE_CASE", "schedule")
    if case == "malformed": policy["operations"][0]["steps"][0]["fuzz"] = "^FuzzAlpha$; touch escaped"
    if case == "services": inventory["modules"][0]["required_services"] = ["postgres"]
    if case == "multiple-packages": policy["operations"][0]["steps"][0]["packages"] = [".", "./manual"]
    if case == "duplicate": policy["operations"][0]["steps"][1]["fuzz"] = "^FuzzAlpha$"
    if case == "unknown-package": policy["operations"][0]["steps"][0]["packages"] = ["./unowned"]
    if case == "runtime": policy["runtimes"] = {"deno": "2.6.4"}
    if case == "tags": inventory["modules"][0]["test_tags"] = ["integration"]
    if case == "identity": inventory["repository"] = "github.com/faustbrian/go-other"
    if case == "disabled": inventory["modules"][0]["gates"]["fuzz"] = False
    if case == "race":
        for step in policy["operations"][0]["steps"]: step["race"] = True
    if case == "nested-module":
        policy["operations"][0]["module"] = "integration/owned"
        inventory["modules"][0]["directory"] = "integration/owned"
        (root / "integration" / "owned").mkdir(parents=True)
        inventory["modules"][0]["packages"][0]["directory"] = str(pathlib.Path("integration/owned") / package)
    module_root = root / inventory["modules"][0]["directory"]
    (module_root / "go.mod").write_text("module example\n\ngo 1.27.0\n")
    package_root = module_root / package
    package_root.mkdir(parents=True, exist_ok=True)
    (package_root / "fixture.go").write_text("package fixture\n")
    (root / "policy.json").write_text(json.dumps(policy))
    if case == "json": (root / "policy.json").write_text("{malformed")
    (root / "inventory.json").write_text(json.dumps(inventory))
    (root / ".golib.yaml").write_text("fixture policy is supplied by strict CLI boundary\n")
    sha = "a" * 40
    (root / "log").write_text("")
    if case == "external-corpus":
        outside = pathlib.Path(outside_tmp)
        base = root / package
        base.mkdir(exist_ok=True)
        (base / "testdata").symlink_to(outside, target_is_directory=True)
    programs = {
        "git": '#!/bin/sh\nprintf "%s\\n" ' + sha + '\n',
        "golib": '''#!/usr/bin/env python3
import os, pathlib, sys
root = pathlib.Path(os.environ["FIXTURE_ROOT"])
if sys.argv[1:] == ["--help"]:
    print("golib check [--local]\\ngolib check --local --module <directory> [--module <directory> ...]")
elif sys.argv[1:] == ["config", "show", "--json"]: print((root / "policy.json").read_text())
elif sys.argv[1:] == ["inventory", "--json"]: print((root / "inventory.json").read_text())
elif sys.argv[1:2] == ["check"]: pass
else: sys.exit(2)
''',
        "go": '''#!/usr/bin/env python3
import json, os, pathlib, sys, time
root = pathlib.Path(os.environ["FIXTURE_ROOT"])
args = sys.argv[1:]
with (root / "log").open("a") as log: log.write(json.dumps({"args": args, "pid": os.getpid(), "cache": os.environ.get("GOCACHE"), "cpu": os.environ.get("GOMAXPROCS"), "flags": os.environ.get("GOFLAGS")}) + "\\n")
if any(a.startswith("-list=") for a in args):
    if os.environ["FIXTURE_CASE"] != "missing": print("FuzzAlpha\\nFuzzBeta")
elif any(a.startswith("-fuzz=") for a in args):
    if os.environ["FIXTURE_CASE"] == "failure": sys.exit(7)
    if os.environ["FIXTURE_CASE"] == "output": print("x" * (2 << 20))
    if os.environ["FIXTURE_CASE"] == "cancel": time.sleep(30)
    if os.environ["FIXTURE_CASE"] == "corpus":
        selector = next(a[6:] for a in args if a.startswith("-fuzz="))
        directory = pathlib.Path(args[-1]) / "testdata" / "fuzz" / selector[1:-1]
        directory.mkdir(parents=True)
        (directory / ("b" * 16)).write_text("go test fuzz v1\\nint(1)\\n")
        sys.exit(7)
    if os.environ["FIXTURE_CASE"] == "memory": time.sleep(0.25)
'''}
    if case == "memory":
        programs["ps"] = '''#!/usr/bin/env python3
import json, pathlib
log = pathlib.Path(ROOT) / "log"
calls = log.read_text().splitlines()
if calls: print(str(json.loads(calls[-1])["pid"]) + " 2097153")
'''.replace("ROOT", repr(str(root)))
    for name, body in programs.items():
        path = bindir / name
        body = body.replace('os.environ["FIXTURE_ROOT"]', repr(str(root))).replace('os.environ["FIXTURE_CASE"]', repr(case))
        path.write_text(body)
        path.chmod(0o700)
    env = dict(os.environ, PATH=str(bindir) + os.pathsep + os.environ["PATH"],
        RUNNER_TEMP=str(cache), FIXTURE_ROOT=str(root), FIXTURE_CASE=case,
        GITHUB_EVENT_NAME="schedule", EVENT_NAME="schedule", MODULE_MATRIX='[{"directory":"."}]',
        GITHUB_REF="refs/heads/main", GITHUB_SHA=sha, GITHUB_REPOSITORY="faustbrian/" + owner,
        ASSURANCE_TARGET="", GITHUB_RUN_ID="1234", GITHUB_RUN_ATTEMPT="1")
    if case in ("manual", "empty", "unknown-target"):
        env["GITHUB_EVENT_NAME"] = "workflow_dispatch"
        env["ASSURANCE_TARGET"] = "^FuzzBeta$" if case == "manual" else ("" if case == "empty" else "^FuzzUnknown$")
    if case == "pull-request": env["GITHUB_EVENT_NAME"] = "pull_request"
    if case == "branch": env["GITHUB_REF"] = "refs/heads/feature/untrusted"
    if case == "unknown-owner": env["GITHUB_REPOSITORY"] = "faustbrian/unknown"
    if case == "source": env["GITHUB_SHA"] = "b" * 40
    if case == "unknown-module": env["ASSURANCE_MODULE"] = "unowned"
    if case == "nested-module": env["ASSURANCE_MODULE"] = "integration/owned"
    if case == "cancel":
        running = subprocess.Popen(["bash", "-c", "exec " + os.environ["ASSURANCE_RUN"]], cwd=root, env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        deadline = time.monotonic() + 5
        while not any("-fuzz=" in line for line in (root / "log").read_text().splitlines()):
            assert running.poll() is None and time.monotonic() < deadline, "cancellation fixture never entered campaign"
            time.sleep(0.01)
        try:
            running.send_signal(signal.SIGTERM)
            stdout, stderr = running.communicate(timeout=3)
            completed = subprocess.CompletedProcess(running.args, running.returncode, stdout, stderr)
        finally:
            if running.poll() is None:
                running.kill()
                running.wait(timeout=2)
    else:
        completed = subprocess.run(["bash", "-c", os.environ["ASSURANCE_RUN"]], cwd=root, env=env, capture_output=True, timeout=15)
    calls = [json.loads(line) for line in (root / "log").read_text().splitlines()]
    campaigns = [call for call in calls if any(arg.startswith("-fuzz=") for arg in call["args"])]
    valid = case in ("schedule", "manual", "nested-module", "race")
    if valid:
        assert completed.returncode == 0, (completed.returncode, completed.stderr.decode())
        assert len(campaigns) == 1, "selected active fuzz campaign absent from actual workflow execution"
        call = campaigns[0]
        assert "-fuzztime=30s" in call["args"] and "-parallel=2" in call["args"]
        assert "-run=^$" in call["args"] and "-count=1" in call["args"]
        assert call["args"][-1] == package and call["cpu"] == "2" and call["flags"] == "-p=1"
        assert call["cache"] and str(cache) in call["cache"]
        if case == "manual": assert "-fuzz=^FuzzBeta$" in call["args"]
        assert ("-race" in call["args"]) == (case == "race"), "declared native race selection lost or manufactured"
    else:
        assert completed.returncode != 0, "invalid scope or failed native campaign passed"
        if case not in ("failure", "output", "cancel", "corpus"): assert not campaigns, "invalid scope entered campaign"
        if case == "cancel": assert b"cancellation" in completed.stderr
        if case == "output": assert b"output budget" in completed.stderr
        if case == "memory": assert b"memory budget" in completed.stderr, completed.stderr.decode()
    assert not (root / "escaped").exists(), "selector was interpreted as shell"
    remaining = list(cache.iterdir())
    if case == "corpus":
        assert len(remaining) == 1 and remaining[0].name == "golib-assurance-failure-1234-1"
        assert [file.name for file in remaining[0].iterdir()] == ["b" * 16]
    else: assert not remaining, "private cache survived selected campaign"
`

func TestLibraryAssuranceExecutesSelectedNativeRiskAndRefusesInvalidScope(t *testing.T) {
	var workflow workflowDocument
	if err := yaml.Unmarshal([]byte(readProjectFile(t, ".github/workflows/library-assurance.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	var command string
	for _, step := range workflow.Jobs["assurance"].Steps {
		if step.Name == "Execute selected native assurance" {
			command = step.Run
		}
	}
	if command == "" {
		t.Fatal("missing selected native execution")
	}
	// Replace only the checkout location, as hosted CI places this immutable
	// source at .golib-tooling in the owner checkout.
	command = strings.ReplaceAll(command, ".golib-tooling/tools/selected-assurance.py", filepath.Join(projectRoot(t), "tools/selected-assurance.py"))
	for _, owner := range []string{"go-clock", "go-concurrency-limit", "go-http-client"} {
		for _, test := range []string{"schedule", "manual", "nested-module", "race", "empty", "unknown-target", "pull-request", "branch", "unknown-owner", "unknown-module", "unknown-package", "source", "identity", "disabled", "malformed", "json", "duplicate", "services", "tags", "runtime", "multiple-packages", "missing", "external-corpus", "failure", "output", "cancel", "memory", "corpus"} {
			t.Run(owner+"/"+test, func(t *testing.T) {
				cmd := exec.CommandContext(t.Context(), "python3", "-c", libraryAssuranceFixture)
				cmd.Env = append(os.Environ(), "ASSURANCE_RUN="+command, "FIXTURE_CASE="+test, "FIXTURE_OWNER="+owner)
				if data, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("actual selected execution: %v: %s", err, data)
				}
			})
		}
	}
}

func TestLibraryAssurancePreservesPackageTargetIdentityAndOwnedTestOnlyPackages(t *testing.T) {
	command := exec.CommandContext(t.Context(), "python3", "-B", filepath.Join(projectRoot(t), "tools/test_selected_assurance.py"))
	if data, err := command.CombinedOutput(); err != nil {
		t.Fatalf("actual assurance identity and ownership producer: %v: %s", err, data)
	}
}

func TestLibraryAssuranceRefusesExpiredBudgetAndCleansReadonlyCache(t *testing.T) {
	const fixture = `
import pathlib, sys, tempfile, time, stat
source=pathlib.Path(sys.argv[1]).read_text()
module={"__name":"assurance_fixture"}
exec(compile(source,sys.argv[1],"exec"),module)
with tempfile.TemporaryDirectory(prefix="golib-assurance-budget-fixture-") as tmp:
    cache=pathlib.Path(tmp)/"owned"
    with module["private_cache"](cache):
        session=module["Session"](cache)
        session.deadline=time.monotonic()-1
        try: session.run(["deliberately-absent-command"],"expired combined budget")
        except module["Refusal"] as error: assert str(error)=="combined time budget"
        else: raise AssertionError("expired budget executed a command")
        nested=cache/"mod"/"readonly"
        nested.mkdir()
        (nested/"cached").write_text("task owned cache")
        nested.chmod(stat.S_IRUSR|stat.S_IXUSR)
    assert not cache.exists(), "read-only cache survived cleanup"
`
	command := exec.CommandContext(t.Context(), "python3", "-c", fixture, filepath.Join(projectRoot(t), "tools/selected-assurance.py"))
	if data, err := command.CombinedOutput(); err != nil {
		t.Fatalf("deadline and cache ownership: %v: %s", err, data)
	}
}

func TestLibraryAssuranceCancellationDuringAdmissionOwnsCleanup(t *testing.T) {
	// Exercise the real runner and signal handler at process admission, while
	// doubling OS process effects so no process-control test runs locally.
	const fixture = `
import io, pathlib, signal, sys, tempfile
source=pathlib.Path(sys.argv[1]).read_text()
module={"__name":"assurance_admission_fixture"}
exec(compile(source,sys.argv[1],"exec"),module)
class NativeProcess:
    pid=123456789
    stdout=io.BytesIO()
    running=True
    def poll(self): return None if self.running else -9
    def wait(self, timeout=None):
        assert not self.running, "native process was waited without termination"
        return -9
child=NativeProcess()
def launch(*args, **kwargs):
    assert kwargs["start_new_session"] is True
    # The process exists before Popen returns to its caller. Delivery here
    # reproduces the same admission window without signaling any OS process.
    module["canceled"](signal.SIGTERM, None)
    return child
def terminate(pid, delivered):
    assert pid==child.pid and delivered==signal.SIGKILL
    child.running=False
module["subprocess"].Popen=launch
module["os"].killpg=terminate
with tempfile.TemporaryDirectory(prefix="golib-assurance-admission-") as tmp:
    cache=pathlib.Path(tmp)/"owned"
    with module["private_cache"](cache):
        session=module["Session"](cache)
        try: session.run(["controlled-native-boundary"], "admission")
        except module["Refusal"] as error: assert str(error)=="cancellation"
        else: raise AssertionError("canceled admission became successful")
        assert not child.running, "cancellation orphaned the admitted native process"
    assert not cache.exists(), "canceled admission retained its owned cache"
`
	command := exec.CommandContext(t.Context(), "python3", "-c", fixture, filepath.Join(projectRoot(t), "tools/selected-assurance.py"))
	if data, err := command.CombinedOutput(); err != nil {
		t.Fatalf("native admission cancellation: %v: %s", err, data)
	}
}
