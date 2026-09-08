#!/usr/bin/env python3
import base64, hashlib, json
from pathlib import Path
def canonical(v): return json.dumps(v,ensure_ascii=False,separators=(",",":"),sort_keys=True).encode()
def sha(b): return "sha256:"+hashlib.sha256(b).hexdigest()
def main():
 p=Path(__file__).resolve().parents[2]/"testdata/cohesion/forward-oracles/cohesion-contract-review-v1-forward-oracle.json"; data=p.read_bytes(); v=json.loads(data)
 if data!=canonical(v): raise SystemExit("oracle is not canonical JSON")
 expected_fixtures=["base.canonical-minimum","base.canonical-rich"]
 if v.get("fixture_count") != 2 or len(v.get("fixtures", [])) != 2 or [x.get("fixture_id") for x in v["fixtures"]]!=expected_fixtures: raise SystemExit("fixture count mismatch")
 fs={}
 for fixture in v["fixtures"]:
  fid=fixture["fixture_id"]
  if fid in fs: raise SystemExit("duplicate fixture id")
  raw=base64.b64decode(fixture["bytes_base64"],validate=True)
  if fixture.get("bytes_sha256") != sha(raw): raise SystemExit("fixture digest mismatch")
  fs[fid]=raw
 errors={"contract-review.invalid.goal":"schema-constant","contract-review.invalid.contract":"semantic-external-identity","contract-review.invalid.reviewed-commit":"semantic-external-identity","contract-review.invalid.accepted-findings":"schema-union","contract-review.invalid.rejected-findings":"schema-union","contract-review.invalid.workflow-id":"schema-range","contract-review.invalid.workflow-outcome":"schema-constant","contract-review.invalid.release-url":"schema-pattern","json.bom":"json-bom","json.invalid-utf8":"json-invalid-utf8","json.trailing-value":"json-trailing-value","json.duplicate-key":"json-duplicate-key","json.lone-surrogate":"json-lone-surrogate","json.noncharacter":"json-noncharacter","json.negative-zero":"json-negative-zero","json.fraction":"json-noninteger-number","json.exponent":"json-noninteger-number","json.integer-overflow":"json-integer-overflow","json.depth-65":"limit-depth","schema.missing-required":"schema-required-member","schema.unknown-member":"schema-unknown-member","schema.wrong-type":"schema-type"}
 ids=sorted(["base.canonical-minimum","base.canonical-rich","contract-review.valid.accepted","contract-review.valid.rejected","json.noncanonical",*errors])
 if v["case_count"]!=27 or [x["case_id"] for x in v["cases"]]!=ids: raise SystemExit("case roster mismatch")
 used=set()
 for x in v["cases"]:
  if x["outcome"] not in {"accepted","rejected"}: raise SystemExit("unknown outcome")
  i=x["input"]
  if i["fixture_id"] not in fs: raise SystemExit("unknown fixture reference")
  used.add(i["fixture_id"])
  if i["kind"]=="fixture": candidate=fs[i["fixture_id"]]
  else:
   s=i["splices"][0]; src=fs[i["fixture_id"]]; ins=base64.b64decode(s["insert_base64"],validate=True)
   if s["offset"] < 0 or s["delete_count"] < 0 or s["offset"]+s["delete_count"] > len(src): raise SystemExit("splice bounds mismatch")
   candidate=src[:s["offset"]]+ins+src[s["offset"]+s["delete_count"]:]
  if x["outcome"]=="accepted":
   try: normalized=canonical(json.loads(candidate.decode("utf-8")))
   except Exception as exc: raise SystemExit(f"accepted input is not valid JSON: {exc}")
   if x["normalized_value_sha256"]!=sha(normalized): raise SystemExit("accepted digest mismatch")
  if x["outcome"]=="rejected" and (x["normalized_value_sha256"] is not None or x["error_code"]!=errors[x["case_id"]]): raise SystemExit("rejection mismatch")
 if used != set(fs): raise SystemExit("unreferenced fixture")
 print("contract-review forward oracle verified")
if __name__=="__main__": main()
