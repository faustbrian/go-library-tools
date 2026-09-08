#!/usr/bin/env python3
"""Independent verifier for the Go-toolchains schema probe."""
import base64, hashlib, json
from pathlib import Path
def canonical(v): return json.dumps(v, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()
def digest(b): return "sha256:" + hashlib.sha256(b).hexdigest()
def main():
    root = Path(__file__).resolve().parents[2]; path = root / "testdata/cohesion/forward-oracles/cohesion-go-toolchains-v1-forward-oracle.json"
    data = path.read_bytes(); value = json.loads(data)
    if data != canonical(value) or value.get("format") != "golib-forward-oracle-v2": raise SystemExit("noncanonical or wrong format")
    fixtures = value["fixtures"]
    if [f["fixture_id"] for f in fixtures] != ["base.canonical-minimum", "base.canonical-rich"]: raise SystemExit("fixture roster mismatch")
    raw = {}
    for f in fixtures:
        b = base64.b64decode(f["bytes_base64"], validate=True)
        if f["bytes_sha256"] != digest(b) or b != canonical(json.loads(b)): raise SystemExit("fixture digest mismatch")
        raw[f["fixture_id"]] = b
    expected = ["base.canonical-minimum", "base.canonical-rich", "go-toolchains.invalid.missing-entries", "go-toolchains.invalid.unknown-member", "go-toolchains.valid"]
    if [r["case_id"] for r in value["cases"]] != expected: raise SystemExit("case roster mismatch")
    errors = {expected[2]: "schema-required-member", expected[3]: "schema-unknown-member"}
    for row in value["cases"]:
        inp = row["input"]
        if inp["kind"] == "fixture": candidate = raw[inp["fixture_id"]]
        else:
            s = inp["splices"][0]; src = raw[inp["fixture_id"]]; candidate = base64.b64decode(s["insert_base64"], validate=True)
            if s["offset"] != 0 or s["delete_count"] != len(src): raise SystemExit("splice bounds")
        obj = json.loads(candidate)
        if row["outcome"] == "accepted":
            if set(obj) != {"schema_id", "schema_version", "official_index", "entry_count", "entries"} or row["normalized_value_sha256"] != digest(candidate) or row["error_code"] is not None: raise SystemExit("accepted mismatch")
            if obj["entry_count"] != len(obj["entries"]): raise SystemExit("entry count mismatch")
            for e in obj["entries"]:
                if e["distribution_url"] != "https://go.dev/dl/" + e["official_filename"]: raise SystemExit("URL mismatch")
        else:
            if row["error_code"] != errors[row["case_id"]] or row["normalized_value_sha256"] is not None: raise SystemExit("rejected mismatch")
    print("go-toolchains forward oracle verified")
if __name__ == "__main__": main()
