#!/usr/bin/env python3
"""Generate a clean-room schema probe for cohesion Go toolchains v1.

The accepted entry rows are projected only from the frozen official-downloads
oracle.  Tree and executable digests are deterministic fixture digests (the
schema requires their shape, while extraction evidence belongs to the
official-downloads oracle lane); no release artifact is asserted here.
"""
import argparse, base64, copy, hashlib, json
from pathlib import Path

def canonical(v):
    return json.dumps(v, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()
def digest(b):
    return "sha256:" + hashlib.sha256(b).hexdigest()

def main():
    ap = argparse.ArgumentParser(); ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args(); root = Path(__file__).resolve().parents[2]
    oracle = json.loads((root / "testdata/cohesion/go-official-downloads-oracle.json").read_bytes())
    selected = oracle["accepted"][0]["expected"]
    entries = []
    for row in selected:
        metadata = {k: row[k] for k in ("version", "goos", "goarch", "official_filename", "distribution_sha256", "distribution_size")}
        entries.append({**metadata, "distribution_url": "https://go.dev/dl/" + row["official_filename"],
                        "tree_sha256": digest(canonical(metadata)),
                        "binary_sha256": digest(row["official_filename"].encode())})
    entries.sort(key=lambda e: (e["version"], e["goos"], e["goarch"]))
    base = {"schema_id": "urn:golib:cohesion:go-toolchains:v1", "schema_version": 1,
            "official_index": {"source_url": "https://go.dev/dl/?mode=json&include=all", "retrieved_at": "2026-09-08T00:00:00Z", "asset": "cohesion-go-official-downloads.json", "bytes_sha256": oracle["official_index_sha256"]},
            "entry_count": len(entries), "entries": entries}
    cases = {
        "base.canonical-minimum": (base, "accepted", digest(canonical(base)), None),
        "base.canonical-rich": (base, "accepted", digest(canonical(base)), None),
        "go-toolchains.valid": (base, "accepted", digest(canonical(base)), None),
    }
    bad = copy.deepcopy(base); bad.pop("entries"); cases["go-toolchains.invalid.missing-entries"] = (bad, "rejected", None, "schema-required-member")
    bad = copy.deepcopy(base); bad["unknown"] = None; cases["go-toolchains.invalid.unknown-member"] = (bad, "rejected", None, "schema-unknown-member")
    rows = []
    raw = canonical(base)
    for case_id in sorted(cases):
        value, outcome, normalized, error = cases[case_id]; candidate = canonical(value)
        inp = {"kind": "fixture", "fixture_id": case_id} if case_id.startswith("base.") else {"kind": "splice", "fixture_id": "base.canonical-rich", "splices": [{"offset": 0, "delete_count": len(raw), "insert_base64": base64.b64encode(candidate).decode()}]}
        rows.append({"case_id": case_id, "input": inp, "outcome": outcome, "normalized_value_sha256": normalized, "error_code": error})
    payload = {"format": "golib-forward-oracle-v2", "fixture_count": 2,
               "fixtures": [{"fixture_id": n, "bytes_base64": base64.b64encode(raw).decode(), "bytes_sha256": digest(raw)} for n in ("base.canonical-minimum", "base.canonical-rich")],
               "case_count": len(rows), "cases": rows}
    args.output.parent.mkdir(parents=True, exist_ok=True); args.output.write_bytes(canonical(payload))
if __name__ == "__main__": main()
