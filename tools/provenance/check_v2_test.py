import importlib.util
import unittest
from pathlib import Path


def load_checker():
    path = Path(__file__).with_name("check_v2.py")
    spec = importlib.util.spec_from_file_location("check_v2", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class CheckerTest(unittest.TestCase):
    def test_v2_inventory_uses_one_oracle_and_one_review(self):
        report = load_checker().inspect()
        self.assertEqual(report["format"], "golib-cohesion-schema-provenance-input-report-v2")
        self.assertTrue(report["oracle"]["path"].endswith("schema-provenance-v2-multiplexed-oracle.json"))
        self.assertTrue(report["review"]["path"].endswith("schema-provenance-v2-aggregate-review.json"))
        self.assertEqual(report["entry_count"], 25)
        self.assertLessEqual(len(report["missing"]), 2)


if __name__ == "__main__":
    unittest.main()
