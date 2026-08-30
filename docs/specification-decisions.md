# Specification Decision Contract

A module is specification-backed when its `modules.json` entry has a nonempty
`specifications` array. Each such module owns these repository-relative files:

- `docs/specification-decisions.md`, the human decision register;
- `specification/README.md`, the human conformance matrix;
- `specification/decisions.json`, the authoritative decision data;
- `specification/conformance.json`, the authoritative evidence bindings;
- `specification/decision-history.json`, the append-only content-digest ledger; and
- `specification/monitoring.json`, the source and change-authority policy.

The module must also declare its module-owned source manifests in
`provenance`. Nested modules cannot borrow a parent or sibling module's source
manifest, fixture, fuzz target, interoperability artifact, documentation, or
decision identifier.

## Decision data

Markdown headings use `## FAMILY-DEC-NNN: Title`; family prefixes may contain
hyphens. The headings, titles, and identifier set must exactly match the two
JSON manifests and the human conformance matrix. Identifiers are unique across
the repository and are never reused.

`specification/decisions.json` uses this strict shape:

```json
{
  "schema_version": 1,
  "change_control": {
    "readme": "README.md",
    "conformance": "specification/README.md",
    "compatibility": "COMPATIBILITY.md",
    "contribution": "CONTRIBUTING.md",
    "changelog": "CHANGELOG.md",
    "pull_request_template": ".github/pull_request_template.md"
  },
  "decisions": [
    {
      "id": "HTTP-DEC-001",
      "title": "Request target normalization",
      "status": "resolved",
      "owner": "HTTP maintainers",
      "classification": "omission",
      "decision_scope": "defensive",
      "specification": "RFC 9110 HTTP Semantics",
      "version": "RFC 9110",
      "source_authority": "rfc9110-source",
      "section": "7.1",
      "requirement_strength": "not specified",
      "issue": "The specification does not select an application normalization policy.",
      "interpretations": ["Preserve bytes", "Normalize equivalent paths"],
      "peer_behavior": "Maintained peers implement both policies.",
      "selected_behavior": "Preserve exact bytes.",
      "rationale": "Normalization can invalidate signatures.",
      "security_consequences": "Signed input is not rewritten.",
      "resource_consequences": "Input remains bounded by the parser limit.",
      "compatibility_consequences": "Existing byte-exact behavior remains stable.",
      "wire_consequences": "Request-target bytes remain unchanged.",
      "executable_evidence": ["TestRequestTargetContract"],
      "fixture_evidence": ["testdata/request-target.json"],
      "fuzz_evidence": ["FuzzRequestTargetContract"],
      "interoperability_evidence": ["testdata/peer.tsv"],
      "public_apis": ["Parse"],
      "documentation": ["docs/specification-decisions.md"],
      "upstream_status": "No upstream issue exists.",
      "reconsider_when": "RFC 9110 is superseded."
    }
  ]
}
```

Allowed statuses are `resolved`, `unresolved`, and `superseded`. Allowed
classifications are `ambiguity`, `contradiction`, `omission`, `erratum`,
`implementation-defined behavior`, `optional behavior`, and
`interoperability policy`. Allowed decision scopes are `normative`,
`recommended`, `defensive`, `extension-specific`, `transport-specific`, and
`application-policy`. Requirement strength is the exact applicable BCP 14
keyword, `not specified`, or `informative`.

Resolved decisions must name at least one current Go `Test` function. Fixture,
fuzz, and interoperability arrays may be empty when those evidence lanes are
not applicable; every supplied value must still resolve to current module-owned
evidence. Unresolved decisions follow the same validation for supplied evidence
but may omit executable evidence and remain visible while blocking repository
and release checks. Superseded decisions retain their historical evidence and
exact cross-register bindings without requiring retired symbols or data files
to remain current. Their documentation remains current, they name a known
replacement in `replacement`, and their own Markdown section links to the
replacement decision's exact generated heading anchor.

## Conformance bindings

`specification/conformance.json` binds each decision to evidence without
duplicating prose:

```json
{
  "schema_version": 1,
  "decisions": [
    {
      "id": "HTTP-DEC-001",
      "authoritative_sources": ["rfc9110-source"],
      "executable_evidence": ["TestRequestTargetContract"],
      "fixtures": ["testdata/request-target.json"],
      "fuzz": ["FuzzRequestTargetContract"],
      "differential_evidence": ["testdata/peer.tsv"],
      "differential_classification": "deliberate policy difference",
      "public_behavior": ["Parse preserves request-target bytes."]
    }
  ]
}
```

Evidence arrays must exactly match the decision data. `authoritative_sources`
must include the decision's primary source authority. Additional specification,
registry, extension, or recommendation authorities may cover other
module-declared specifications; the Markdown entry must expose each additional
authority's exact ID, version, URL, and covered specification names as one
inline-code JSON record. For example:

```markdown
Additional authoritative source: `{"id":"rfc9110-extension","version":"Extension 1","url":"https://example.com/extension","specifications":["Example Extension"]}`
```

The record must remain inside that decision's level-two section.
Differential evidence uses `local defect`, `peer defect`, `fixture defect`,
`harness defect`, `specification ambiguity`, or `deliberate policy difference`.
An empty differential evidence array must use `not assessed`, and `not assessed`
is invalid when differential evidence is present.

## Source and authority pins

TSV source manifests require `id`, `version`, `url`, `sha256`, and `status`
columns. Every row is an HTTPS pin with status `pinned`. JSON source manifests
must be valid; every `*sha256` value must be in the same object as a nonempty
source identity such as an ID, name, version, URL, path, source, commit, or tag.

`specification/monitoring.json` uses this strict shape:

```json
{
  "schema_version": 1,
  "reviewed_at": "2026-08-30",
  "review_interval_days": 90,
  "authorities": [
    {
      "id": "rfc9110-source",
      "kind": "specification",
      "version": "RFC 9110",
      "url": "https://www.rfc-editor.org/rfc/rfc9110.txt",
      "sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "specifications": ["RFC 9110 HTTP Semantics"]
    },
    {
      "id": "rfc9110-errata",
      "kind": "errata",
      "url": "https://www.rfc-editor.org/errata/rfc9110",
      "sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "specifications": ["RFC 9110 HTTP Semantics"]
    }
  ]
}
```

The review interval is 1 through 90 days. Source authorities require an exact
version. Authority kinds are `specification`, `registry`, `extension`,
`recommendation`, `errata`, and `releases`. Every exact `modules.json`
specification label requires at least one source authority and one errata or
release authority. A module may monitor at most 64 authorities.

## Change control and online checks

The declared README, conformance, compatibility, contribution, and changelog
documents must link to the module's decision register. Module-specific
documents remain inside that module. The pull-request template must contain a
`Specification Decisions` review section covering the decision identifier,
compatibility impact, changelog entry, and any replacement or supersession.

Every Markdown decision entry must contain every exact value from its typed
decision record plus the authoritative source URL. This keeps the human
register and machine record in agreement instead of accepting field-name
keywords. Every current decision identifier and canonical decision SHA-256
must also have a durable changelog record. Historical changelog identifiers
must remain in the current register as superseded entries, preventing silent
erasure; a changed record requires a new digest record. The strict
`decision-history.json` ledger contains every retained identifier and every
canonical digest that identifier has held. The checker validates that current
content is internally complete without depending on Git topology. Because no
self-contained tree can prove that a previous tree did not contain an entry,
append-only preservation is a protected review obligation: the required pull
request section makes every removed or superseded identifier explicit, and
reviewers must require it to remain in the register and ledger.

`golib specification check` performs the complete offline structural check.
`golib specification check --online` additionally downloads bounded authority
content, verifies each reviewed digest, preserves the original HTTPS authority
across redirects, rejects any authority resolution containing a non-public
network address, revalidates redirect resolution, and bounds the complete online
run to two minutes.
Repositories without declared specifications pass without network access.
