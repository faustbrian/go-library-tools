#!/usr/bin/env python3
"""Generate an evidence-backed catalog-v2 schema probe.

The accepted fixture is the smallest repository-scoped preview envelope whose
cross-field values are fixed by the published catalog-v2 schema.  The policy
object is copied from the released gate-policy corpus; rejected cases are
bounded required-member and unknown-member mutations of that fixture.
"""
import argparse, base64, copy, hashlib, json
from pathlib import Path

def canonical(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()
def digest(data): return "sha256:" + hashlib.sha256(data).hexdigest()

def main():
    ap = argparse.ArgumentParser(); ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args(); root = Path(__file__).resolve().parents[2]
    policy_row = json.loads((root / "release/cohesion-gate-policy-v1-accepted-corpus.json").read_bytes())[0]
    policy = json.loads(base64.b64decode(policy_row["input_base64"], validate=True))
    base = {
        "schema_id": "urn:golib:cohesion:catalog:v2", "schema_version": 2,
        "manifest_schema_version": 2, "view": "engineering", "scope": "repository",
        "repository": "github.com/faustbrian/go-example", "source_revision": "0" * 40,
        "design_language": {"version": "1.0", "repository": "github.com/faustbrian/go-library-tools", "path": "docs/ecosystem/design-language.md", "bytes_sha256": "sha256:d767f6d4d5eb3ba73861d898255d42691fac49101ea55edb079af3b1969357c8", "release": "v1.5.3", "tag_object_sha": "c116e5d313ce7acedcb5d459f7737a85ef3cd34e", "peeled_commit": "b1f8313166388c69a5b646686d19617bf75b53ca", "goal_contract": {"goal_id": "golib-cohesion-v1", "requirements_sha256": "sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c", "repository": "github.com/faustbrian/go-library-tools", "path": "docs/ecosystem/goals/cohesion-v1.md", "release": "v1.5.5", "tag_object_sha": "6e54f464376865a60c618cce8f07fe70bcfabed0", "peeled_commit": "66d2874dc98afe43dd7dad379d6f5e7d1b613111"}},
        "tooling": {"kind": "source-build", "version": "dev", "source_commit": "0" * 40, "executable_sha256": "sha256:" + "0" * 64},
        "publication_status": "preview", "policies": [policy],
        "input_manifest_sha256": "sha256:" + "1" * 64, "source_lock": None,
        "planned_identities": [], "modules": [],
    }
    raw = canonical(base)
    cases = [("base.canonical-minimum", base, "accepted", digest(raw), None), ("base.canonical-rich", base, "accepted", digest(raw), None)]
    missing = copy.deepcopy(base); missing.pop("modules")
    unknown = copy.deepcopy(base); unknown["__unknown__"] = None
    final_input = copy.deepcopy(base); final_input["publication_status"] = "final-input"
    final_input["tooling"] = {"kind": "release", "repository": "github.com/faustbrian/go-library-tools", "release": "v1.6.0", "tag_object_sha": "0" * 40, "peeled_commit": "0" * 40, "goos": "darwin", "goarch": "arm64", "executable_asset": "tool", "executable_url": "https://github.com/faustbrian/go-library-tools/releases/download/v1.6.0/tool", "platform_artifact_sha256": "sha256:" + "0" * 64, "checksums_asset": "checksums", "checksums_url": "https://github.com/faustbrian/go-library-tools/releases/download/v1.6.0/checksums", "checksums_sha256": "sha256:" + "0" * 64, "release_manifest_asset": "manifest", "release_manifest_url": "https://github.com/faustbrian/go-library-tools/releases/download/v1.6.0/manifest", "release_manifest_sha256": "sha256:" + "0" * 64}
    cases += [("catalog-v2.invalid.final-input-manifest-version", final_input, "rejected", None, "semantic-cross-field"), ("catalog-v2.invalid.missing-modules", missing, "rejected", None, "schema-required-member"), ("catalog-v2.invalid.unknown-member", unknown, "rejected", None, "schema-unknown-member")]
    rows=[]
    for case_id, value, outcome, normalized, error in sorted(cases):
        candidate = canonical(value)
        inp = {"kind":"fixture", "fixture_id":case_id} if case_id.startswith("base.") else {"kind":"splice", "fixture_id":"base.canonical-rich", "splices":[{"offset":0,"delete_count":len(raw),"insert_base64":base64.b64encode(candidate).decode()}]}
        rows.append({"case_id":case_id,"input":inp,"outcome":outcome,"normalized_value_sha256":normalized,"error_code":error})
    payload={"format":"golib-forward-oracle-v2","fixture_count":2,"fixtures":[{"fixture_id":n,"bytes_base64":base64.b64encode(raw).decode(),"bytes_sha256":digest(raw)} for n in ("base.canonical-minimum","base.canonical-rich")],"case_count":len(rows),"cases":rows}
    args.output.parent.mkdir(parents=True, exist_ok=True); args.output.write_bytes(canonical(payload))
if __name__ == "__main__": main()
