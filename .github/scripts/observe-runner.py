"""Bounded, argument-free observations for one owned rehearsal shell."""

import datetime
import json
import math
import os
from pathlib import Path
import sys
import time

MAX_PROCESSES = 16
MAX_SYSTEM_PROCESSES = 4096
MAX_RECORD_BYTES = 8192
MAX_SAMPLES = 360
INTERVAL_SECONDS = 60
READ_LIMIT = 65536


def read_text(path):
    try:
        with path.open("rb") as stream:
            data = stream.read(READ_LIMIT + 1)
        return data.decode("ascii") if len(data) <= READ_LIMIT else None
    except (OSError, UnicodeError):
        return None


def number(value):
    return int(value) if len(value) <= 20 and value.isascii() and value.isdigit() else None


def fields(path, selected):
    text = read_text(path)
    result = {name: None for name in selected}
    for line in (text or "").splitlines():
        parts = line.split()
        if len(parts) >= 2 and parts[0].rstrip(":") in selected:
            result[parts[0].rstrip(":")] = number(parts[1])
    return result


def sample(parent, proc=Path("/proc"), disk=Path(".")):
    memory = fields(proc / "meminfo", ["MemAvailable"])
    loads = read_text(proc / "loadavg")
    try:
        load = [float(x) for x in (loads or "").split()[:3]]
        if len(load) != 3 or not all(math.isfinite(x) and x >= 0 for x in load):
            load = None
    except ValueError:
        load = None
    try:
        filesystem = os.statvfs(disk)
        free = filesystem.f_bavail * filesystem.f_frsize
    except OSError:
        free = None
    # PPid belongs to the process, including children forked by non-leader
    # threads. Task-local children files cannot provide that view.
    visible, children = {}, {}
    unavailable, scan_truncated = False, False
    try:
        with os.scandir(proc) as entries:
            for entry in entries:
                pid = number(entry.name)
                if pid is None or pid <= 0:
                    continue
                if len(visible) >= MAX_SYSTEM_PROCESSES:
                    scan_truncated = True
                    break
                status = fields(proc / str(pid) / "status", ["PPid", "VmRSS", "Threads"])
                visible[pid] = status
                if status["PPid"] is None:
                    unavailable = True
                else:
                    children.setdefault(status["PPid"], []).append(pid)
    except OSError:
        unavailable = True
    pending, seen, processes = [parent], set(), []
    while pending and len(processes) < MAX_PROCESSES:
        pid = pending.pop(0)
        if pid in seen or pid == os.getpid():
            continue
        seen.add(pid)
        status = visible.get(pid, {"PPid": None, "VmRSS": None, "Threads": None})
        unavailable = unavailable or status["PPid"] is None
        processes.append({"pid": pid, "parent_pid": status["PPid"],
                          "rss_kib": status["VmRSS"], "threads": status["Threads"]})
        pending.extend(child for child in children.get(pid, [])
                       if child not in seen and child != os.getpid())
    return {"event": "runner-observation", "utc": datetime.datetime.now(datetime.timezone.utc).isoformat(),
            "load_averages": load, "logical_cpus": os.cpu_count(),
            "memory_available_kib": memory["MemAvailable"], "disk_available_bytes": free,
            "processes": processes, "processes_truncated": scan_truncated or bool(pending),
            "process_sample_unavailable": unavailable, "process_sample_best_effort": True}


def main():
    if len(sys.argv) != 2 or number(sys.argv[1]) != os.getppid():
        print('{"event":"runner-observation-unavailable"}', flush=True)
        return
    parent = os.getppid()
    disk = Path(os.environ.get("RUNNER_TEMP", "."))
    for sequence in range(MAX_SAMPLES):
        if os.getppid() != parent:
            return
        try:
            record = sample(parent, disk=disk)
            record["sequence"] = sequence
            encoded = json.dumps(record, separators=(",", ":"))
            if len(encoded.encode("utf-8")) > MAX_RECORD_BYTES:
                encoded = '{"event":"runner-observation-unavailable"}'
            print(encoded, flush=True)
        except (OSError, ValueError, OverflowError):
            print('{"event":"runner-observation-unavailable"}', flush=True)
        if sequence + 1 < MAX_SAMPLES:
            time.sleep(INTERVAL_SECONDS)


if __name__ == "__main__":
    main()
