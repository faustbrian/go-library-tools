#!/usr/bin/env python3
"""Avoid repeating unchanged scheduled attempts; native records own assurance."""

import json
import os
import re
import selectors
import subprocess
import sys
import time

DEADLINE = time.monotonic() + 90
PACKAGES = {"internal/config", "internal/inventory", "internal/repositoryfile"}
PROSE = {
    "README.md", "CONTRIBUTING.md", "AGENTS.md", "CHANGELOG.md", "SECURITY.md",
    "SUPPORT.md", "CODE_OF_CONDUCT.md", "LICENSE",
}
OUTPUTS = (".verification/by-input/mutation/", ".verification/mutation/reports/",
           ".verification/mutation/failed-reports/")


def capture(arguments, limit=1024 * 1024):
    """Bound command time and captured bytes before accepting their contents."""
    process = subprocess.Popen(arguments, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    chunks, size = [], 0
    try:
        with selectors.DefaultSelector() as ready:
            ready.register(process.stdout, selectors.EVENT_READ)
            while ready.get_map():
                remaining = DEADLINE - time.monotonic()
                if remaining <= 0:
                    raise ValueError
                for key, _ in ready.select(min(remaining, 1)):
                    chunk = os.read(key.fileobj.fileno(), 65536)
                    if not chunk:
                        ready.unregister(key.fileobj)
                        continue
                    size += len(chunk)
                    if size > limit:
                        raise ValueError
                    chunks.append(chunk)
        if process.wait(timeout=max(0.01, DEADLINE - time.monotonic())) != 0:
            raise ValueError
        return b"".join(chunks).decode("utf-8", errors="strict")
    finally:
        if process.poll() is None:
            process.kill()
            process.wait()
        process.stdout.close()


def api(path, limit=1024 * 1024):
    arguments = ["gh", "api"]
    # Native job output may contain ANSI escapes. Permit them only in bounded
    # in-memory log capture, never terminal output or JSON API responses.
    if path.endswith("/logs"):
        arguments.append("--allow-escape-sequences")
    return capture(arguments + [path], limit)


def positive(value):
    if type(value) is not int or value <= 0:
        raise ValueError
    return value


def commit(value):
    if not isinstance(value, str) or not re.fullmatch(r"[0-9a-f]{40}", value):
        raise ValueError
    capture(["git", "cat-file", "-e", value + "^{commit}"])
    return value


def material(path):
    return not (path in PROSE or
                (path.startswith("docs/") and path.endswith(".md")) or
                any(path.startswith(prefix) for prefix in OUTPUTS))


def decide():
    selected = os.environ["SELECTED_PACKAGE"]
    event = os.environ["EVENT_NAME"]
    if selected not in PACKAGES or event not in {"schedule", "workflow_dispatch"}:
        raise ValueError
    if event == "workflow_dispatch":
        return True, None
    if os.environ["GITHUB_REPOSITORY"] != "faustbrian/go-library-tools":
        raise ValueError
    current = positive(int(os.environ["GITHUB_RUN_ID"]))
    head = commit(os.environ["GITHUB_SHA"])
    if capture(["git", "rev-parse", "HEAD"]).strip() != head:
        raise ValueError
    capture(["git", "diff", "--quiet", "HEAD", "--", "."])
    version = open(".go-version", encoding="utf-8").read().strip()
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", version):
        raise ValueError
    image = os.environ["ImageOS"] + "/" + os.environ["ImageVersion"]
    arch = os.environ["RUNNER_ARCH"]
    if not re.fullmatch(r"[A-Za-z0-9._/-]{1,160}", image) or arch not in {"X64", "ARM64"}:
        raise ValueError
    repository = "/repos/faustbrian/go-library-tools/actions"
    history = json.loads(api(repository + "/workflows/mutation-assurance.yml/runs?per_page=100"))
    runs, total = history["workflow_runs"], history["total_count"]
    if not isinstance(runs, list) or len(runs) > 100 or type(total) is not int or total < len(runs):
        raise ValueError
    inspected = 0
    for run in sorted(runs, key=lambda item: positive(item["id"]), reverse=True):
        run_id = positive(run["id"])
        if run_id == current or run["event"] not in {"schedule", "workflow_dispatch"} or run["head_branch"] != "main":
            continue
        # A live predecessor cannot be treated as completed or safely duplicated.
        if run["status"] != "completed":
            raise ValueError
        prior = commit(run["head_sha"])
        capture(["git", "merge-base", "--is-ancestor", prior, head])
        paths = capture(["git", "diff", "--no-ext-diff", "--no-renames", "--name-only", "-z", prior, head, "--"], 4 * 1024 * 1024).split("\0")
        if any(material(path) for path in paths if path):
            continue
        inspected += 1
        if inspected > 20:
            raise ValueError
        jobs = json.loads(api(repository + f"/runs/{run_id}/jobs?filter=latest&per_page=100"))
        if jobs["total_count"] != len(jobs["jobs"]) or len(jobs["jobs"]) > 100:
            raise ValueError
        owned = [job for job in jobs["jobs"] if job["name"] == "Selected mutation assurance"]
        if len(owned) != 1:
            raise ValueError
        job = owned[0]
        if job["run_id"] != run_id or job["head_sha"] != prior or job["status"] != "completed":
            raise ValueError
        campaigns = [step for step in job["steps"] if step["name"] == "Run bounded package campaign"]
        if len(campaigns) != 1:
            raise ValueError
        campaign = campaigns[0]
        if campaign["conclusion"] == "skipped":
            continue
        if campaign["status"] != "completed" or campaign["conclusion"] not in {"success", "failure", "cancelled", "timed_out"} or not campaign["started_at"]:
            raise ValueError
        log = api(repository + f"/jobs/{positive(job['id'])}/logs", 8 * 1024 * 1024)
        markers = re.findall(
            r"^(?:\d{4}-\d{2}-\d{2}T[0-9:.]+Z\s+)?golib-mutation-inputs "
            r"package=([a-z/]+) image=([A-Za-z0-9._/-]+) arch=(X64|ARM64) go=(go[0-9.]+)\s*$",
            log, re.MULTILINE)
        candidates = re.findall(
            r"^(?:\d{4}-\d{2}-\d{2}T[0-9:.]+Z\s+)?golib-mutation-inputs .*$", log, re.MULTILINE)
        if not markers:
            # Build/capability failures never reached the actual campaign.
            if not candidates:
                continue
            raise ValueError
        if len(markers) != 1 or len(candidates) != 1:
            raise ValueError
        if markers[0] == (selected, image, arch, "go" + version):
            return False, (run_id, campaign["conclusion"])
    if total > len(runs):
        raise ValueError
    return True, None


try:
    execute, prior = decide()
    if prior:
        run_id, result = prior
        with open(os.environ["GITHUB_STEP_SUMMARY"], "a", encoding="utf-8") as summary:
            summary.write("## Unchanged scheduled mutation attempt\n\n"
                          f"Owner: faustbrian. Prior actual campaign: {result}. "
                          f"[Run {run_id}](https://github.com/faustbrian/go-library-tools/actions/runs/{run_id}).\n\n"
                          "Automatic rerun suppressed; this is not passing assurance. "
                          "Unresolved findings remain unresolved; use manual dispatch for an intentional retry.\n")
    with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as output:
        output.write(f"run={str(execute).lower()}\n")
except (KeyError, ValueError, TypeError, OSError, subprocess.SubprocessError):
    print("Cannot establish previous mutation attempt; automatic execution refused.", file=sys.stderr)
    sys.exit(1)
