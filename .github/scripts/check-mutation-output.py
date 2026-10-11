#!/usr/bin/env python3
"""Bound native mutation cache/artifact output without replacing its validator."""

import os
import stat
import sys

PATHS = (
    ".verification/by-input/mutation",
    ".verification/mutation/reports",
    ".verification/mutation/failed-reports",
)


def check():
    count = size = 0
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


try:
    check()
except (OSError, ValueError):
    print("mutation output is not confined flat JSON within 2048 files/32 MiB", file=sys.stderr)
    sys.exit(1)
