# Ecosystem security contract

This versioned directory is the shared security authority for standalone
`go-*` repositories. It defines the [threat model](threat-model.md), the
[vulnerability-management process](vulnerability-management.md), and the
machine-readable risk and release records.

`risk-register.json` records durable risks. Every risk has one owner, a
rationale, a mitigation, and an objective review condition. Risks are reviewed
when their condition occurs and before a release verdict references them.

`security-matrix.json` records the exact tool version, command, completion time,
status, immutable result digest, evidence pointer, and release verdict for an
immutable module revision. Passing rows require a successful result from every
governed scanner. Every residual-risk identifier must be an
accepted, time-bounded risk owned by the same module; critical and high risks
cannot be accepted into a passing verdict.

Validate both records with:

```sh
golib security validate --directory docs/ecosystem/security
```

The schemas are published as
[`security-risk-register.schema.json`](../../../schema/security-risk-register.schema.json)
and [`security-matrix.schema.json`](../../../schema/security-matrix.schema.json).
The validator applies these published structural schemas before release-safety
rules, rejects duplicate JSON object keys, and accepts numerically equivalent
representations of schema version `1` (for example, `1.0` and `1e0`).
Maintainers can run `make security-differential` (with `uv` available) to
compare CLI decisions against a separately implemented Draft 2020-12 validator.
Repositories adopt this contract incrementally; missing fleet rows do not
silently become passing verdicts and do not block unrelated Tier A/B work.
