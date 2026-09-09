import json
import subprocess
import sys
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


class AggregateReviewTest(unittest.TestCase):
    def test_review_is_bound_and_non_authorizing(self):
        path = ROOT / "testdata/cohesion/schema-provenance-v2-aggregate-review.json"
        subprocess.run([sys.executable, str(ROOT / "tools/provenance/verify_aggregate_review.py"), str(path)], check=True)
        value = json.loads(path.read_text())
        self.assertEqual(value["status"], "incomplete-non-authorizing")
        self.assertFalse(value["authorization"]["authorized"])
        self.assertEqual([d["name"] for d in value["dimensions"]], ["coverage", "integrity", "reachability"])


if __name__ == "__main__":
    unittest.main()
