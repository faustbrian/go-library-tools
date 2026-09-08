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


if __name__ == "__main__":
    unittest.main()
