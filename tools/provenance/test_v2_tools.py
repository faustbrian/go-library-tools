import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


class V2ToolsTest(unittest.TestCase):
    def test_generator_and_verifier_are_non_claiming(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "inventory.json"
            subprocess.run([sys.executable, str(ROOT / "tools/provenance/generate_v2.py"), "--output", str(output)], check=True)
            value = json.loads(output.read_text())
            self.assertEqual(value["status"], "inputs-inventory")
            self.assertEqual(value["entry_count"], 25)
            subprocess.run([sys.executable, str(ROOT / "tools/provenance/verify_v2.py"), str(output)], check=True)

    def test_multiplexed_oracle_is_bound_and_reports_gaps(self):
        oracle = ROOT / "testdata/cohesion/schema-provenance-v2-multiplexed-oracle.json"
        subprocess.run([sys.executable, str(ROOT / "tools/provenance/verify_multiplexed_oracle.py"), str(oracle)], check=True)
        value = json.loads(oracle.read_text())
        self.assertEqual(value["schema_count"], 25)
        self.assertGreater(sum(entry["case_count"] for entry in value["entries"]), 0)
        self.assertGreater(value["missing_case_report"]["missing_count"], 0)


if __name__ == "__main__":
    unittest.main()
