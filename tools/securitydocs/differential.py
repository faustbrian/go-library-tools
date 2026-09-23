"""Compare security-record CLI decisions with an independent Draft 2020-12 engine.

Requires Python jsonschema 4.17.3. Usage: differential.py REPOSITORY GOLIB_BINARY.
The caller owns the temporary environment, binary, and cleanup.
"""

import copy
import hashlib
import json
import pathlib
import subprocess
import sys
import tempfile

from jsonschema import Draft202012Validator


def main():
    repository = pathlib.Path(sys.argv[1]).resolve()
    binary = pathlib.Path(sys.argv[2]).resolve()
    schemas = {
        "risk-register.json": Draft202012Validator(
            json.loads((repository / "schema/security-risk-register.schema.json").read_text())
        ),
        "security-matrix.json": Draft202012Validator(
            json.loads((repository / "schema/security-matrix.schema.json").read_text())
        ),
    }
    scanner_names = [
        "codeql", "dependency-review", "go-vet", "gosec", "govulncheck", "license",
        "owned-analysis", "secret-current-tree", "secret-history", "staticcheck", "workflow-analysis",
    ]
    risk = {
        "id": "SEC-1", "module": "github.com/acme/example", "severity": "medium",
        "status": "mitigated", "owner": "security@example.com", "rationale": "risk",
        "mitigation": "limits", "review_condition": "on release",
    }
    scanner = {
        "name": "codeql", "tool_version": "v1.0.0", "command": "tool ./...",
        "status": "passed", "completed_at": "2026-09-13T00:00:00Z",
        "result": "artifact://result", "result_sha256": "0123456789abcdef" * 4,
    }
    module = {
        "module": "github.com/acme/example",
        "revision": "0123456789abcdef0123456789abcdef01234567",
        "scanners": [dict(scanner, name=name) for name in scanner_names],
        "release_verdict": {
            "status": "pass", "owner": "security@example.com",
            "decided_at": "2026-09-13T00:00:00Z", "rationale": "reviewed evidence",
            "residual_risks": [],
        },
    }
    full = {
        "risk-register.json": {"schema_version": 1, "risks": [risk]},
        "security-matrix.json": {"schema_version": 1, "modules": [module]},
    }
    cases = [("full", copy.deepcopy(full), True)]
    cases.append(("empty", {
        "risk-register.json": {"schema_version": 1, "risks": []},
        "security-matrix.json": {"schema_version": 1, "modules": []},
    }, True))

    def add(name, filename, path, operation, expected=False):
        documents = copy.deepcopy(full)
        target = documents[filename]
        for key in path[:-1]:
            target = target[key]
        if operation == "missing":
            del target[path[-1]]
        else:
            target[path[-1]] = operation
        cases.append((name, documents, expected))

    for filename, root_array in [("risk-register.json", "risks"), ("security-matrix.json", "modules")]:
        for field in ["schema_version", root_array]:
            add(f"{filename}/{field}/missing", filename, [field], "missing")
            add(f"{filename}/{field}/null", filename, [field], None)
        add(f"{filename}/unknown-root", filename, ["unknown"], 1)
        add(f"{filename}/{root_array}/wrong-item", filename, [root_array, 0], None)
        add(f"{filename}/version-zero", filename, ["schema_version"], 0)
        add(f"{filename}/version-string", filename, ["schema_version"], "1")

    nested = [
        ("risk-register.json", ["risks", 0], list(risk)),
        ("security-matrix.json", ["modules", 0], ["module", "revision", "scanners", "release_verdict"]),
        ("security-matrix.json", ["modules", 0, "scanners", 0], list(scanner)),
        ("security-matrix.json", ["modules", 0, "release_verdict"], list(module["release_verdict"])),
    ]
    for filename, path, fields in nested:
        section = "/".join(map(str, path))
        for field in fields:
            add(f"{section}/{field}/missing", filename, path + [field], "missing")
            add(f"{section}/{field}/null", filename, path + [field], None)
        add(f"{section}/unknown", filename, path + ["unknown"], 1)

    for name, filename, path, value in [
        ("bad-risk-id", "risk-register.json", ["risks", 0, "id"], "invalid"),
        ("bad-severity", "risk-register.json", ["risks", 0, "severity"], "urgent"),
        ("blank-owner", "risk-register.json", ["risks", 0, "owner"], " "),
        ("bad-revision", "security-matrix.json", ["modules", 0, "revision"], "bad"),
        ("bad-scanner-name", "security-matrix.json", ["modules", 0, "scanners", 0, "name"], "unknown"),
        ("bad-scanner-status", "security-matrix.json", ["modules", 0, "scanners", 0, "status"], "failed"),
        ("bad-digest", "security-matrix.json", ["modules", 0, "scanners", 0, "result_sha256"], "bad"),
        ("bad-verdict-status", "security-matrix.json", ["modules", 0, "release_verdict", "status"], "unknown"),
        ("bad-residual-id", "security-matrix.json", ["modules", 0, "release_verdict", "residual_risks"], ["invalid"]),
        ("duplicate-residual", "security-matrix.json", ["modules", 0, "release_verdict", "residual_risks"], ["SEC-1", "SEC-1"]),
        ("too-many-scanners", "security-matrix.json", ["modules", 0, "scanners"], module["scanners"] + [scanner]),
    ]:
        add(name, filename, path, value)

    blocked = copy.deepcopy(full)
    blocked["security-matrix.json"]["modules"][0]["scanners"] = []
    blocked["security-matrix.json"]["modules"][0]["release_verdict"]["status"] = "blocked"
    cases.append(("blocked-empty-scanners", blocked, True))
    for status in ("pass", "fail", "blocked"):
        present = copy.deepcopy(full if status == "pass" else blocked)
        module_row = present["security-matrix.json"]["modules"][0]
        module_row["release_verdict"]["status"] = status
        if status != "pass":
            cases.append((f"{status}-present-empty-scanners", present, True))
        for value in ("missing", None):
            invalid = copy.deepcopy(present)
            row = invalid["security-matrix.json"]["modules"][0]
            if value == "missing":
                del row["scanners"]
            else:
                row["scanners"] = value
            cases.append((f"{status}-{value}-scanners", invalid, False))
        too_many = copy.deepcopy(present)
        rows = too_many["security-matrix.json"]["modules"][0]["scanners"]
        if not rows:
            rows.extend(copy.deepcopy(module["scanners"]))
        rows.append(copy.deepcopy(scanner))
        cases.append((f"{status}-twelve-scanners", too_many, False))
    for status in ("fail", "blocked"):
        present = copy.deepcopy(blocked)
        present["security-matrix.json"]["modules"][0]["release_verdict"]["status"] = status
        cases.append((f"{status}-present-empty-residuals", present, True))
        for value in ("missing", None):
            invalid = copy.deepcopy(present)
            verdict = invalid["security-matrix.json"]["modules"][0]["release_verdict"]
            if value == "missing":
                del verdict["residual_risks"]
            else:
                verdict["residual_risks"] = value
            cases.append((f"{status}-{value}-residuals", invalid, False))
    short_pass = copy.deepcopy(full)
    short_pass["security-matrix.json"]["modules"][0]["scanners"].pop()
    cases.append(("pass-ten-scanners", short_pass, False))
    failed = copy.deepcopy(blocked)
    failed["security-matrix.json"]["modules"][0]["release_verdict"]["status"] = "fail"
    cases.append(("fail-empty-scanners", failed, True))
    for name, version in [("decimal-version", 1.0), ("exponent-version", 1e0)]:
        numeric = copy.deepcopy(full)
        numeric["risk-register.json"]["schema_version"] = version
        numeric["security-matrix.json"]["schema_version"] = version
        cases.append((name, numeric, True))

    accepted = copy.deepcopy(full)
    accepted_risk = accepted["risk-register.json"]["risks"][0]
    accepted_risk.update({
        "status": "accepted", "evidence": "reviewed", "expires_at": "2099-01-01T00:00:00Z",
    })
    accepted["security-matrix.json"]["modules"][0]["release_verdict"]["residual_risks"] = ["SEC-1"]
    cases.append(("accepted-with-evidence-and-expiry", accepted, True))
    for field in ("evidence", "expires_at"):
        missing = copy.deepcopy(accepted)
        del missing["risk-register.json"]["risks"][0][field]
        cases.append((f"accepted-missing-{field}", missing, False))
        null = copy.deepcopy(accepted)
        null["risk-register.json"]["risks"][0][field] = None
        cases.append((f"accepted-null-{field}", null, False))

    for name, documents, expected in cases:
        documents = copy.deepcopy(documents)
        matrix_encoded = json.dumps(documents["security-matrix.json"])
        if name == "exponent-version":
            matrix_encoded = matrix_encoded.replace('"schema_version": 1.0', '"schema_version": 1e0')
        evidence = f"sha256:{hashlib.sha256(matrix_encoded.encode()).hexdigest()}:security-matrix.json"
        for record in documents["risk-register.json"].get("risks") or []:
            if not isinstance(record, dict):
                continue
            if record.get("status") in ("mitigated", "closed") or (
                record.get("status") == "accepted" and isinstance(record.get("evidence"), str)
            ):
                record["evidence"] = evidence
        classification = all(schemas[filename].is_valid(document) for filename, document in documents.items())
        if classification != expected:
            raise AssertionError(f"{name}: fixture classification {classification}, expected {expected}")
        with tempfile.TemporaryDirectory(prefix="security-differential-") as directory:
            for filename, document in documents.items():
                encoded = matrix_encoded if filename == "security-matrix.json" else json.dumps(document)
                if name == "exponent-version":
                    encoded = encoded.replace('"schema_version": 1.0', '"schema_version": 1e0')
                    if '"schema_version": 1e0' not in encoded:
                        raise AssertionError("exponent fixture replacement failed")
                (pathlib.Path(directory) / filename).write_text(encoded)
            result = subprocess.run(
                [str(binary), "security", "validate", "--directory", directory],
                cwd=repository, capture_output=True, text=True, check=False,
            )
            if (result.returncode == 0) != classification:
                raise AssertionError(f"{name}: schema={classification}, CLI exit={result.returncode}")
            if classification and result.stdout != "ecosystem security records valid\n":
                raise AssertionError(f"{name}: incorrect success output")
            if not classification and result.stdout:
                raise AssertionError(f"{name}: failure emitted stdout")
    print(f"independent Draft 2020-12 differential: {len(cases)} cases agree")


if __name__ == "__main__":
    main()
