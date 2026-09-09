import importlib.util
import unittest
from pathlib import Path

spec = importlib.util.spec_from_file_location("generate", Path(__file__).with_name("generate.py"))
generate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(generate)

class BindingPolicyTest(unittest.TestCase):
    def test_only_unreleased_noninstallable_unverified_is_portable(self):
        base = {"publication_status": "unreleased", "installable": False,
                "evidence": {"observation": "source is unverified"}}
        self.assertTrue(generate.allows_unverified_local_binding(base))
        for item in ({**base, "installable": True},
                     {**base, "publication_status": "published"},
                     {**base, "evidence": {"observation": "pending"}}):
            self.assertFalse(generate.allows_unverified_local_binding(item))

if __name__ == "__main__":
    unittest.main()
