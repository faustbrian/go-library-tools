#!/usr/bin/env python3
"""Bound native mutation cache/artifact output without replacing its validator."""

import os
import hashlib
import json
import shutil
import stat
import sys

PATHS = (
    ".verification/by-input/mutation",
    ".verification/mutation/reports",
    ".verification/mutation/failed-reports",
)


def check():
    count = size = 0
    files = []
    for path in PATHS:
        current = ""
        missing = False
        for component in path.split("/"):
            current = os.path.join(current, component)
            try:
                info = os.lstat(current)
            except FileNotFoundError:
                missing = True
                break
            if not stat.S_ISDIR(info.st_mode):
                raise ValueError
        if missing:
            continue
        with os.scandir(path) as entries:
            for entry in entries:
                count += 1
                info = entry.stat(follow_symlinks=False)
                if not stat.S_ISREG(info.st_mode) or not entry.name.endswith(".json"):
                    raise ValueError
                size += info.st_size
                if count > 2048 or size > 32 * 1024 * 1024:
                    raise ValueError
                files.append(os.path.join(path, entry.name))
    return files


def identities(files):
    result = {}
    size = 0
    for path in files:
        digest = hashlib.sha256()
        with open(path, "rb") as source:
            while chunk := source.read(65536):
                size += len(chunk)
                if size > 32 * 1024 * 1024:
                    raise ValueError
                digest.update(chunk)
        result[path] = digest.hexdigest()
    return result


def fresh_output(baseline, destination, files):
    info = os.lstat(baseline)
    if not stat.S_ISREG(info.st_mode) or info.st_size > 1024 * 1024:
        raise ValueError
    with open(baseline, encoding="utf-8") as source:
        before = json.load(source)
    if not isinstance(before, dict) or len(before) > 2048:
        raise ValueError
    info = os.lstat(destination)
    if not stat.S_ISDIR(info.st_mode) or os.listdir(destination):
        raise ValueError
    fresh = False
    for path, digest in identities(files).items():
        if before.get(path) == digest:
            continue
        target = os.path.join(destination, path)
        os.makedirs(os.path.dirname(target), mode=0o700, exist_ok=True)
        shutil.copyfile(path, target, follow_symlinks=False)
        fresh = True
    print(f"fresh={str(fresh).lower()}")


try:
    files = check()
    if len(sys.argv) == 3 and sys.argv[1] == "--snapshot":
        with open(sys.argv[2], "x", encoding="utf-8") as destination:
            json.dump(identities(files), destination)
    elif len(sys.argv) == 4 and sys.argv[1] == "--fresh":
        fresh_output(sys.argv[2], sys.argv[3], files)
    elif len(sys.argv) != 1:
        raise ValueError
except (OSError, ValueError):
    print("mutation output is not confined flat JSON within 2048 files/32 MiB", file=sys.stderr)
    sys.exit(1)
