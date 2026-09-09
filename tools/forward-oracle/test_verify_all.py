import importlib.util
import json
import tempfile
import unittest
from pathlib import Path

HERE = Path(__file__).parent
spec = importlib.util.spec_from_file_location("verify_all", HERE / "verify_all.py")
verify_all = importlib.util.module_from_spec(spec)
spec.loader.exec_module(verify_all)


class VerifyAllTest(unittest.TestCase):
    def test_generic_rejects_one_case_lane(self):
        value = {
            "format": "golib-forward-oracle-v2",
            "fixture_count": 2,
            "fixtures": [{"fixture_id": "base.canonical-minimum"}, {"fixture_id": "base.canonical-rich"}],
            "case_count": 1,
            "cases": [{"case_id": "only", "input": {"fixture_id": "base.canonical-minimum"}, "outcome": "accepted", "normalized_value_sha256": None, "error_code": None}],
        }
        with tempfile.NamedTemporaryFile(mode="w", suffix=".json") as f:
            json.dump(value, f)
            f.flush()
            with self.assertRaises(ValueError):
                verify_all.generic(Path(f.name))

    def test_lane_manifest_is_fixed(self):
        self.assertEqual(len(verify_all.LANES), 7)
        self.assertNotIn("exploratory", verify_all.LANES.values())


if __name__ == "__main__":
    unittest.main()
