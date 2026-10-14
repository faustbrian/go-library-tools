"""Exercise bounded runner metadata without invoking or controlling processes."""

import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

spec = importlib.util.spec_from_file_location("observer", Path(__file__).with_name("observe-runner.py"))
observer = importlib.util.module_from_spec(spec)
spec.loader.exec_module(observer)


class ObservationTests(unittest.TestCase):
    def test_executable_roles_never_return_arbitrary_names(self):
        with tempfile.TemporaryDirectory() as directory:
            process = Path(directory)
            executable = process / "exe"
            for target, expected in [("/opt/go/bin/go", "go"), ("/tmp/compile", "compile"),
                                     ("/tmp/private-fixture.test", "go-test"),
                                     ("/tmp/private-value", "unknown")]:
                executable.symlink_to(target)
                self.assertEqual(observer.process_role(process), expected)
                executable.unlink()
            self.assertEqual(observer.process_role(process), "unknown")

    def test_cgroup_events_and_pressure_are_numeric_metadata(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "memory.events").write_text("low 1\nhigh 2\nmax 3\noom 4\noom_kill 5\noom_group_kill 6\nprivate-value 99\n")
            (root / "memory.current").write_text("1234\n")
            (root / "memory.max").write_text("max\n")
            (root / "memory.pressure").write_text("some avg10=1.25 avg60=2.0 avg300=3.0 total=456\nfull avg10=0.0 avg60=0.5 avg300=1.0 total=789\n")
            result = observer.cgroup_memory(root)
            self.assertEqual(result["events"]["oom_kill"], 5)
            self.assertEqual(result["current_bytes"], 1234)
            self.assertIsNone(result["limit_bytes"])
            self.assertEqual(result["pressure"]["some"]["avg10"], 1.25)
            self.assertEqual(result["pressure"]["full"]["total"], 789)
            self.assertNotIn("private-value", json.dumps(result))

    def test_unavailable_and_invalid_values_do_not_imply_no_oom(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.assertIsNone(observer.cgroup_memory(root)["events"]["oom_kill"])
            (root / "memory.events").write_text("oom_kill -1\n")
            (root / "memory.pressure").write_text("some avg10=nan avg60=-1 avg300=inf total=bad\n")
            result = observer.cgroup_memory(root)
            self.assertIsNone(result["events"]["oom_kill"])
            self.assertTrue(all(value is None for value in result["pressure"]["some"].values()))

    def test_descendant_and_record_caps_remain_bounded(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for pid in range(1, 40):
                process = root / str(pid)
                process.mkdir()
                (process / "status").write_text(f"PPid: {0 if pid == 1 else 1}\nVmRSS: 10 kB\nThreads: 1\n")
                (process / "exe").symlink_to("/tmp/private-value")
            result = observer.sample(1, proc=root, disk=root, cgroup=root)
            self.assertEqual(len(result["processes"]), observer.MAX_PROCESSES)
            self.assertTrue(result["processes_truncated"])
            self.assertLessEqual(len(json.dumps(result).encode()), observer.MAX_RECORD_BYTES)
            self.assertNotIn("private-value", json.dumps(result))


if __name__ == "__main__":
    unittest.main()
