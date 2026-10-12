#!/usr/bin/env python3
"""Select existing independent consumer files; shared or unknown inputs stay full."""

import json
import os
import re
import selectors
import stat
import subprocess
import sys
import time

DIRECTORY = "release/compatibility-consumer/"
# The public consumer stages its own manifests and never imports the Tools module.
# Root dependency changes still receive Tools checks, not unrelated compositions.
SHARED = {".go-version", ".github/workflows/ci.yml",
          ".github/scripts/select-compatibility-consumer.py",
          ".github/scripts/run-compatibility-consumer.sh"}


def git(*arguments):
    process = subprocess.Popen(["git", *arguments], stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    chunks, size, deadline = [], 0, time.monotonic() + 20
    try:
        with selectors.DefaultSelector() as ready:
            ready.register(process.stdout, selectors.EVENT_READ)
            while ready.get_map():
                remaining = deadline - time.monotonic()
                if remaining <= 0:
                    raise ValueError("Git deadline exceeded")
                for key, _ in ready.select(min(remaining, 1)):
                    chunk = os.read(key.fileobj.fileno(), 65536)
                    if not chunk:
                        ready.unregister(key.fileobj)
                        continue
                    size += len(chunk)
                    if size > 1024 * 1024:
                        raise ValueError("Git path budget exceeded")
                    chunks.append(chunk)
        if process.wait(timeout=max(0.01, deadline-time.monotonic())) != 0:
            raise ValueError("Git scope refused")
        return b"".join(chunks).decode("utf-8", errors="strict")
    finally:
        if process.poll() is None:
            process.kill()
            process.wait()
        process.stdout.close()


try:
    base = os.environ["BASE_SHA"]
    publication = os.environ["PUBLICATION_REQUIRED"]
    if not re.fullmatch(r"[0-9a-f]{40}", base) or publication not in {"true", "false"}:
        raise ValueError("invalid immutable scope")
    git("cat-file", "-e", base + "^{commit}")
    git("merge-base", "--is-ancestor", base, "HEAD")
    tracked = [path for path in git("ls-files", "-z", "--", DIRECTORY).split("\0") if path]
    files = sorted(path[len(DIRECTORY):] for path in tracked
                   if re.fullmatch(r"[a-z0-9_]+\.go", path[len(DIRECTORY):]))
    if not files or not any(name.endswith("_test.go") for name in files) or len(files) > 64:
        raise ValueError("consumer sources missing or excessive")
    for name in files:
        if not stat.S_ISREG(os.lstat(DIRECTORY + name).st_mode):
            raise ValueError("unsafe consumer source")
    changes = git("diff", "--no-ext-diff", "--no-textconv", "--no-renames", "--name-status", "-z", base, "HEAD", "--").split("\0")
    if changes[-1] != "" or (len(changes) - 1) % 2:
        raise ValueError("incomplete Git changes")
    full = publication == "true"
    selected = set()
    for status, path in zip(changes[:-1:2], changes[1:-1:2]):
        if path in SHARED or path.startswith("tools/compatibility/"):
            full = True
        if path.startswith(DIRECTORY):
            name = path[len(DIRECTORY):]
            # New/deleted/retargeted/unknown inputs may alter cross-file ownership.
            # Only a modification to an existing independent file narrows scope.
            if status == "M" and name in files and name.endswith("_test.go"):
                selected.add(name)
            else:
                full = True
    if full:
        if any(path.endswith(".go") and path[len(DIRECTORY):] not in files for path in tracked):
            raise ValueError("nested or unknown Go sources require explicit consumer ownership")
        selected = set(files)
    with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as output:
        output.write(f"required={str(bool(selected)).lower()}\n")
        output.write("files=" + json.dumps(sorted(selected), separators=(",", ":")) + "\n")
except (KeyError, ValueError, OSError, subprocess.SubprocessError):
    print("Cannot establish affected compatibility consumer inputs.", file=sys.stderr)
    sys.exit(1)
