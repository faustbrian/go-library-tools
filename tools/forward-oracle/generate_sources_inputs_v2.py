#!/usr/bin/env python3
"""Generate bounded rejected probes for the v2 source and input contracts.

The fixtures are exact bytes from the released source roster and repository
manifests; production probes establish the rejected outcome and diagnostic.
"""
import argparse, base64, hashlib, json
from pathlib import Path

def canon(v): return json.dumps(v, ensure_ascii=False, sort_keys=True, separators=(",", ":")).encode()
def digest(b): return "sha256:" + hashlib.sha256(b).hexdigest()
def asset(root, name, source):
    raw = (root / source).read_bytes()
    return {"format":"golib-forward-oracle-v2", "fixture_count":1,
            "fixtures":[{"fixture_id":"authoritative."+name,"bytes_base64":base64.b64encode(raw).decode(),"bytes_sha256":digest(raw)}],
            "case_count":1, "cases":[{"case_id":name+".reject-authoritative-v1-input", "input":{"kind":"fixture","fixture_id":"authoritative."+name}, "outcome":"rejected", "normalized_value_sha256":None, "error_code":"schema-required-member"}]}
def main():
    ap=argparse.ArgumentParser(); ap.add_argument("--output-dir", type=Path, required=True); a=ap.parse_args(); root=Path(__file__).resolve().parents[2]
    a.output_dir.mkdir(parents=True, exist_ok=True)
    for name, source in (("sources-v2","release/cohesion-sources.json"),("inputs-v2","modules.json")):
        (a.output_dir/("cohesion-"+name+"-forward-oracle.json")).write_bytes(canon(asset(root,name,source)))
if __name__ == "__main__": main()
