#!/usr/bin/env python3
"""Generate the clean-room cohesion contract-review forward oracle."""
import base64, copy, hashlib, json
from pathlib import Path

def canonical(value): return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode()
def sha(data): return "sha256:" + hashlib.sha256(data).hexdigest()

def common_cases(rich):
    prefix=b'"schema_id":'; start=rich.index(prefix)+len(prefix); old=canonical(json.loads(rich)["schema_id"]); close=len(rich)-1
    without=copy.deepcopy(json.loads(rich)); del without["schema_id"]
    return {
      "json.noncanonical": rich[:1]+b" "+rich[1:], "json.bom": b"\xef\xbb\xbf"+rich,
      "json.invalid-utf8": rich[:start+1]+b"\xff"+rich[start+1:], "json.trailing-value": rich+b"\nnull",
      "json.duplicate-key": rich[:close]+b',"schema_id":'+old+rich[close:],
      "json.lone-surrogate": rich[:start]+b'"\\uD800"'+rich[start+len(old):], "json.noncharacter": rich[:start]+b'"\\uFFFF"'+rich[start+len(old):],
      "json.negative-zero": rich[:close]+b',"__oracle_number__":-0'+rich[close:], "json.fraction": rich[:close]+b',"__oracle_number__":0.5'+rich[close:],
      "json.exponent": rich[:close]+b',"__oracle_number__":1e0'+rich[close:], "json.integer-overflow": rich[:close]+b',"__oracle_number__":9007199254740993'+rich[close:],
      "json.depth-65": rich[:close]+b',"__oracle_depth__":'+b"["*64+b"null"+b"]"*64+rich[close:],
      "schema.missing-required": canonical(without), "schema.unknown-member": rich[:close]+b',"__oracle_unknown__":null'+rich[close:],
      "schema.wrong-type": rich[:start]+b"null"+rich[start+len(old):],
    }

def main():
    root=Path(__file__).resolve().parents[2]; out=root/"testdata/cohesion/forward-oracles/cohesion-contract-review-v1-forward-oracle.json"
    contract={"repository":"github.com/faustbrian/go-library-tools","path":"docs/ecosystem/goals/cohesion-v1.md","tag":"v1.5.5","tag_object_sha":"6e54f464376865a60c618cce8f07fe70bcfabed0","peeled_commit":"66d2874dc98afe43dd7dad379d6f5e7d1b613111","bytes_sha256":"sha256:6e5ac948c2cf1ec439fdf90c53c0648befb74a65c3eca2eb595737ee8211891c"}
    base={"schema_id":"urn:golib:cohesion:contract-review:v1","goal_id":"golib-cohesion-v1","contract":contract,"review":{"reviewer_id":"/root/tooling_v154_delivery/v155_release_review","reviewed_commit":contract["peeled_commit"],"outcome":"accepted-no-findings","findings":[]},"release":{"workflow_run_id":34008713395,"workflow_outcome":"success","url":"https://github.com/faustbrian/go-library-tools/releases/tag/v1.5.5"},"created_at":"2026-09-06T03:32:14Z"}
    rich=canonical(base); cases={"base.canonical-minimum":(rich,"accepted",sha(rich),None),"base.canonical-rich":(rich,"accepted",sha(rich),None),"contract-review.valid.accepted":(rich,"accepted",sha(rich),None)}
    rejected=copy.deepcopy(base); rejected["review"].update(outcome="rejected",findings=["contract requires review"]); rb=canonical(rejected); cases["contract-review.valid.rejected"]=(rb,"accepted",sha(rb),None)
    mutations=[("contract-review.invalid.goal",lambda v:v.update(goal_id="wrong-goal"),"schema-constant"),("contract-review.invalid.contract",lambda v:v["contract"].update(tag="v1.5.4"),"semantic-external-identity"),("contract-review.invalid.reviewed-commit",lambda v:v["review"].update(reviewed_commit="0"*40),"semantic-external-identity"),("contract-review.invalid.accepted-findings",lambda v:v["review"].update(findings=["unexpected finding"]),"schema-union"),("contract-review.invalid.rejected-findings",lambda v:v["review"].update(outcome="rejected",findings=[]),"schema-union"),("contract-review.invalid.workflow-id",lambda v:v["release"].update(workflow_run_id=0),"schema-range"),("contract-review.invalid.workflow-outcome",lambda v:v["release"].update(workflow_outcome="failure"),"schema-constant"),("contract-review.invalid.release-url",lambda v:v["release"].update(url="not-a-release"),"schema-pattern")]
    for name,mutate,error in mutations: value=copy.deepcopy(base); mutate(value); cases[name]=(canonical(value),"rejected",None,error)
    common_errors={"json.bom":"json-bom","json.invalid-utf8":"json-invalid-utf8","json.trailing-value":"json-trailing-value","json.duplicate-key":"json-duplicate-key","json.lone-surrogate":"json-lone-surrogate","json.noncharacter":"json-noncharacter","json.negative-zero":"json-negative-zero","json.fraction":"json-noninteger-number","json.exponent":"json-noninteger-number","json.integer-overflow":"json-integer-overflow","json.depth-65":"limit-depth","schema.missing-required":"schema-required-member","schema.unknown-member":"schema-unknown-member","schema.wrong-type":"schema-type"}
    for name,data in common_cases(rich).items(): cases[name]=(data,"accepted",sha(rich),None) if name=="json.noncanonical" else (data,"rejected",None,common_errors[name])
    fixtures=[("base.canonical-minimum",rich),("base.canonical-rich",rich)]; rows=[]
    for name in sorted(cases):
        data,outcome,norm,error=cases[name]; inp={"kind":"fixture","fixture_id":name} if name.startswith("base.canonical-") else {"kind":"splice","fixture_id":"base.canonical-rich","splices":[{"offset":0,"delete_count":len(rich),"insert_base64":base64.b64encode(data).decode()}]}; rows.append({"case_id":name,"input":inp,"outcome":outcome,"normalized_value_sha256":norm,"error_code":error})
    payload={"format":"golib-forward-oracle-v2","fixture_count":2,"fixtures":[{"fixture_id":n,"bytes_base64":base64.b64encode(d).decode(),"bytes_sha256":sha(d)} for n,d in fixtures],"case_count":len(rows),"cases":rows}; out.parent.mkdir(parents=True,exist_ok=True); out.write_bytes(canonical(payload))
if __name__=="__main__": main()
