#!/usr/bin/env python3
"""Execute one policy-owned in-process fuzz risk, not a full contract or CLI API."""

import datetime
from contextlib import contextmanager
import json
import os
from pathlib import Path
import re
import selectors
import shutil
import signal
import stat
import subprocess
import sys
import time


TOTAL_SECONDS = 300
OUTPUT_BYTES = 1 << 20
MEMORY_KIB = 2 << 20
CORPUS_BYTES = 2 << 20
CANCELED = False


class Refusal(Exception):
    pass


def require(condition, message):
    if not condition:
        raise Refusal(message)


def canceled(_signal, _frame):
    # Never raise across Popen's admission boundary: it may have created a
    # child before returning ownership to Session.run. Cleanup checks this
    # latched request after admission and remains interruption-safe itself.
    global CANCELED
    CANCELED = True


@contextmanager
def private_cache(directory):
    # A run/attempt-owned slot also permits the workflow's always-cleanup step
    # to resolve the exact resource after a hard runner interruption.
    directory.mkdir()
    try:
        yield directory
    finally:
        for root, _, _ in os.walk(directory):
            os.chmod(root, os.stat(root).st_mode | stat.S_IWUSR)
        shutil.rmtree(directory)


class Session:
    def __init__(self, cache):
        self.deadline = time.monotonic() + TOTAL_SECONDS
        self.output = 0
        self.stage = "scope"
        # No credentials or owner environment overrides enter native execution.
        self.env = {key: os.environ[key] for key in ("PATH", "HOME", "LANG") if key in os.environ}
        self.env.update(GOCACHE=str(cache / "build"), GOMODCACHE=str(cache / "mod"),
                        GOTMPDIR=str(cache / "tmp"), TMPDIR=str(cache / "tmp"),
                        GOENV="off", GOWORK="off", GOTOOLCHAIN="local", GOFLAGS="-p=1",
                        GOMAXPROCS="2", GOPROXY="https://proxy.golang.org", GOSUMDB="sum.golang.org",
                        GOPRIVATE="", GONOPROXY="", GONOSUMDB="")
        for key in ("GOCACHE", "GOMODCACHE", "GOTMPDIR"):
            Path(self.env[key]).mkdir()

    def run(self, args, stage):
        self.stage = stage
        require(time.monotonic() < self.deadline, "combined time budget")
        require(not CANCELED, "cancellation")
        process = None
        poller = None
        try:
            process = subprocess.Popen(args, env=self.env, stdout=subprocess.PIPE,
                                       stderr=subprocess.STDOUT, start_new_session=True)
            require(not CANCELED, "cancellation")
            data = bytearray()
            poller = selectors.DefaultSelector()
            poller.register(process.stdout, selectors.EVENT_READ)
            while poller.get_map():
                require(not CANCELED, "cancellation")
                require(time.monotonic() < self.deadline, "combined time budget")
                # Aggregate resident memory of this owned process group. Probe
                # metadata is bounded and never contains commands or secrets.
                snapshot = subprocess.run(["ps", "-eo", "pgid=,rss="], env=self.env,
                                          capture_output=True, timeout=2, check=True)
                require(len(snapshot.stdout) <= 1 << 20, "process metadata output budget")
                rss = 0
                for row in snapshot.stdout.splitlines():
                    fields = row.split()
                    if len(fields) == 2 and fields[0] == str(process.pid).encode():
                        rss += int(fields[1])
                require(rss <= MEMORY_KIB, "owned process memory budget")
                for key, _ in poller.select(0.05):
                    chunk = os.read(key.fd, 65536)
                    if not chunk:
                        poller.unregister(key.fileobj)
                        continue
                    self.output += len(chunk)
                    require(self.output <= OUTPUT_BYTES, "combined output budget")
                    data.extend(chunk)
            status = process.wait(timeout=max(0.01, min(2, self.deadline - time.monotonic())))
            require(not CANCELED, "cancellation")
            require(status == 0, "native command failed (exit %d)" % status)
            return bytes(data)
        finally:
            if poller is not None:
                poller.close()
            if process is not None:
                process.stdout.close()
                # Only this session's process group is addressed, including
                # children whose output descriptor outlives their parent.
                process.poll()
                try:
                    os.killpg(process.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
                process.wait(timeout=2)


def relative_path(value):
    require(isinstance(value, str) and len(value) <= 512 and (value == "." or re.fullmatch(r"(?:\./)?[A-Za-z0-9_-]+(?:/[A-Za-z0-9_-]+)*", value)), "canonical owned path")
    return value.removeprefix("./")


def targets_from(policy, inventory, repository, selected_module):
    require(isinstance(policy, dict) and isinstance(inventory, dict), "policy and inventory objects")
    require(inventory.get("repository") == "github.com/" + repository, "repository inventory identity")
    relative_path(selected_module)
    modules = inventory.get("modules")
    require(isinstance(modules, list) and 0 < len(modules) <= 64, "module inventory")
    require(all(isinstance(module, dict) for module in modules), "module objects")
    roots = [module for module in modules if module.get("directory") == selected_module]
    require(len(roots) == 1, "exact selected module")
    require(isinstance(roots[0].get("gates"), dict) and roots[0]["gates"].get("fuzz") is True, "enabled module fuzz contract")
    # This selected route is deliberately in-process. An evolved
    # service/tag contract needs risk review, not implicit live certification.
    require(not roots[0].get("test_tags") and not roots[0].get("required_services"), "in-process module scope")
    # Inventory external_runtime_dependencies includes Go libraries (e.g.
    # goleak), not an executable requirement. Typed policy owns executables.
    runtimes = policy.get("runtimes", {})
    require(isinstance(runtimes, dict) and not any(runtimes.values()), "external executable scope")
    declared = roots[0].get("packages")
    require(isinstance(declared, list) and len(declared) <= 4096, "owned package inventory")
    require(all(isinstance(package, dict) for package in declared), "package objects")
    package_directories = {relative_path(package.get("directory")) for package in declared}
    operations = policy.get("operations")
    require(isinstance(operations, list) and len(operations) <= 256, "typed operation policy")
    targets = {}
    for operation in operations:
        require(isinstance(operation, dict), "operation object")
        if operation.get("module") != selected_module or operation.get("gate") != "fuzz":
            continue
        steps = operation.get("steps")
        require(isinstance(steps, list) and 0 < len(steps) <= 32, "typed fuzz steps")
        for step in steps:
            require(isinstance(step, dict), "fuzz step object")
            target = step.get("fuzz", "")
            packages = step.get("packages")
            require(step.get("type") == "go-test" and isinstance(packages, list) and len(packages) == 1, "one native fuzz package")
            require(isinstance(packages[0], str) and (packages[0] == "." or packages[0].startswith("./")), "relative explicit native package")
            package = relative_path(packages[0])
            require(package in package_directories, "declared owned fuzz package")
            require(isinstance(target, str) and len(target) <= 128 and re.fullmatch(r"\^Fuzz[A-Za-z0-9_]+\$", target), "anchored single fuzz target")
            require(target not in targets, "ambiguous fuzz target")
            race = step.get("race", False)
            require(type(race) is bool, "native race selection")
            targets[target] = (packages[0], race)
    require(0 < len(targets) <= 32, "declared fuzz targets")
    return targets


def save_new_corpus(directory, before, destination):
    files = []
    total = 0
    if not directory.exists():
        return
    require(not directory.is_symlink(), "symbolic failure corpus")
    for path in directory.iterdir():
        if path.name in before:
            continue
        require(not path.is_symlink() and path.is_file() and re.fullmatch(r"(?:[a-f0-9]{16}|[a-f0-9]{64})", path.name), "native corpus identity")
        size = path.stat().st_size
        total += size
        require(len(files) < 32 and total <= CORPUS_BYTES, "failure corpus budget")
        files.append(path)
    if files:
        destination.mkdir()
        for path in files:
            shutil.copyfile(path, destination / path.name)


def main():
    require(len(sys.argv) == 1, "unexpected arguments")
    repository = os.environ.get("GITHUB_REPOSITORY", "")
    require(re.fullmatch(r"faustbrian/go-[a-z0-9]+(?:-[a-z0-9]+)*", repository), "repository owner")
    owner = repository.removeprefix("faustbrian/")
    require(os.environ.get("GITHUB_REF") == "refs/heads/main", "main-only scope")
    event = os.environ.get("GITHUB_EVENT_NAME")
    require(event in ("schedule", "workflow_dispatch"), "selected assurance event")
    sha = os.environ.get("GITHUB_SHA", "")
    require(re.fullmatch(r"[a-f0-9]{40}", sha), "immutable owner source")
    signal.signal(signal.SIGTERM, canceled)
    signal.signal(signal.SIGINT, canceled)
    # Hosted Linux descendants inherit an exact two-CPU affinity. GOMAXPROCS,
    # worker and compiler admission also apply on local diagnostic platforms.
    if hasattr(os, "sched_getaffinity"):
        cpus = sorted(os.sched_getaffinity(0))
        require(len(cpus) >= 2, "two-CPU execution capability")
        os.sched_setaffinity(0, cpus[:2])
    temp_root = Path(os.environ.get("RUNNER_TEMP", ""))
    require(temp_root.is_dir() and not temp_root.is_symlink(), "private resource root")
    run = os.environ.get("GITHUB_RUN_ID", "")
    attempt = os.environ.get("GITHUB_RUN_ATTEMPT", "")
    require(re.fullmatch(r"[0-9]{1,20}", run) and re.fullmatch(r"[0-9]{1,4}", attempt), "diagnostic run identity")
    with private_cache(temp_root / ("golib-selected-assurance-%s-%s" % (run, attempt))) as temp:
        session = Session(temp)
        target = "unselected"
        corpus = None
        before = set()
        module = os.environ.get("ASSURANCE_MODULE", ".")
        try:
            actual = session.run(["git", "rev-parse", "HEAD"], "source identity").decode().strip()
            require(actual == sha, "source identity mismatch")
            policy = json.loads(session.run(["golib", "config", "show", "--json"], "strict policy discovery"))
            inventory = json.loads(session.run(["golib", "inventory", "--json"], "strict module discovery"))
            targets = targets_from(policy, inventory, repository, module)
            selected = os.environ.get("ASSURANCE_TARGET", "")
            if event == "workflow_dispatch":
                require(selected in targets, "explicit declared diagnostic target")
                target = selected
            else:
                require(selected == "", "scheduled selector override")
                week = (datetime.datetime.now(datetime.timezone.utc).date() - datetime.date(1970, 1, 5)).days // 7
                target = sorted(targets)[week % len(targets)]
            package, race = targets[target]
            module_path = Path(module)
            require(module_path.is_dir() and module_path.resolve().is_relative_to(Path.cwd()), "owned module directory")
            os.chdir(module_path)
            require(Path(package).resolve().is_relative_to(Path.cwd()), "owned native package path")
            print("selected assurance owner=%s source=%s module=%s package=%s target=%s" % (owner, sha, module, package, target), flush=True)
            native = ["go", "test", "-mod=readonly"] + (["-race"] if race else [])
            names = session.run(native + ["-list=" + target, package], "native target discovery").decode().splitlines()
            require(names.count(target[1:-1]) == 1, "native target missing or ambiguous")
            corpus = Path(package) / "testdata" / "fuzz" / target[1:-1]
            require(not corpus.is_symlink() and corpus.resolve().is_relative_to(Path.cwd()), "symbolic or external native corpus")
            if corpus.exists():
                for path in corpus.iterdir():
                    require(len(before) < 4096 and len(path.name) <= 256, "existing corpus inventory budget")
                    before.add(path.name)
            session.run(native + ["-run=^$", "-fuzz=" + target, "-fuzztime=30s", "-parallel=2", "-count=1", "-timeout=2m", package], "active fuzz campaign")
        except (Refusal, OSError, ValueError, subprocess.SubprocessError) as error:
            # Never log untrusted native payloads. A newly emitted native corpus
            # is the bounded reproducer; command failure remains authoritative.
            print("selected assurance failed owner=%s source=%s stage=%s target=%s: %s" %
                  (owner, sha, session.stage, target, error if isinstance(error, Refusal) else type(error).__name__), file=sys.stderr)
            if corpus is not None:
                destination = temp_root / ("golib-assurance-failure-%s-%s" % (run, attempt))
                save_new_corpus(corpus, before, destination)
            raise
    require(not CANCELED, "cancellation")
    print("selected assurance passed owner=%s source=%s target=%s" % (owner, sha, target))


if __name__ == "__main__":
    try:
        main()
    except (Refusal, OSError, ValueError, subprocess.SubprocessError) as error:
        repository = os.environ.get("GITHUB_REPOSITORY", "")
        source = os.environ.get("GITHUB_SHA", "")
        owner = repository if re.fullmatch(r"faustbrian/go-[a-z0-9]+(?:-[a-z0-9]+)*", repository) else "invalid"
        sha = source if re.fullmatch(r"[a-f0-9]{40}", source) else "invalid"
        print("selected assurance refused owner=%s source=%s: %s" %
              (owner, sha, str(error) if isinstance(error, Refusal) else type(error).__name__), file=sys.stderr)
        sys.exit(1)
