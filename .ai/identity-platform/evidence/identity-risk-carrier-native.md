# Native independent carrier-subset review

## First report

## REQUEST_CHANGES — local carrier batch

Principal: `/root/risk_carrier_final`, independent read-only reviewer.  
Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
Base/HEAD: `5639c6c23754e53146bfaa366a60c0f2fd6e656b`  
Reviewed staged tree: `9d63369ec3ce9b72e0c2187c1ffe526f3e6610d8`

### Blocker: canonical public-contract names have incompatible definitions

[risk/carrier.go:51](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/carrier.go:51), [risk/carrier.go:85](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/carrier.go:85), and [risk/carrier.go:29](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/carrier.go:29) introduce `PhoneRecoveryBinding` and `CarrierFact` with different public shapes from their exact named contracts:

- `PhoneRecoveryBinding` replaces the canonical typed `TenantID`, `SubjectID`, `AttemptID`, `RecoveryOperationID`, and `PreAuthTransactionID` fields with renamed strings and changes `PolicyVersion` from bounded string to `uint64`.
- `CarrierFact` omits required `Provider`, substitutes `State CarrierState` for `Status SignalStatus`, and changes the closed kind `carrier-risk` to `carrier`.

The authoritative definitions are `PUBLIC_CONTRACTS.json:37120–37173` and `37990–38056`. A canonical provider cannot populate its documented fields, canonical string policy versions cannot be represented, and a configured `carrier-risk` source is rejected by `carrierIndex()`. Future canonical service/adapters cannot coexist with a second declaration of these same Go names.

**Remediation:** give the preparatory carrier subset distinct public type names, preserving the canonical names for their specified contracts, or implement the canonical layouts now. Use the canonical `carrier-risk` tag and add focused compatibility assertions.

**Falsification:** checked the selected schemas, base declarations, direct-consumer searches, and development-first amendment. The base has neither named type; no existing implementation establishes these alternative layouts. The amendment explicitly changes sequencing, not product semantics. Pending full-service/catalog work does not authorize conflicting definitions of existing canonical names.

### Remaining reviewed boundary

All five changed files and relevant unchanged policy/source orchestration were inspected. No additional concrete defects found in constructor ownership/copying, disabled defaults, exact phase/parent binding, private evaluator provenance, deterministic three-reader evaluation, bounded contexts, denial normalization, original freshness/expiry, or redacted projections. No existing runtime consumers use these new types.

Parent-supplied focused tests, package race run, and vet passed; these are supplied execution evidence, not independently rerun results. Fuzz execution, exact coverage/mutation, canonical service, durable issuance/consumption, provider interoperability, and remaining release gates stay pending and do not constitute this local batch’s completion claim. No files or external state changed.

## Final report

## APPROVE — local development batch

**No findings.**

Principal: `/root/risk_carrier_final`, independent read-only reviewer.  
Repository: `/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform`  
Base/HEAD: `5639c6c23754e53146bfaa366a60c0f2fd6e656b`  
Reviewed staged tree: `a2ae105a7e9ee16401eac7c95cbc7861522f1280`

The complete five-file diff, relevant unchanged helpers/source orchestration, authoritative phone-recovery requirements, and direct-consumer fan-out were reviewed again.

**Prior finding resolved:** [carrier.go:54](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/carrier.go:54) and [carrier.go:89](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/carrier.go:89) now declare distinct preparatory types, `PhoneRecoveryScope` and `CarrierObservation`, without occupying the canonical `PhoneRecoveryBinding`/`CarrierFact` names. [carrier.go:29](/Users/brian/Developer/.worktrees/golib-go-identity-identity-platform/risk/carrier.go:29) uses canonical `carrier-risk`; the new regression exercises literal-tag startup configuration through an actual allowed evaluation, independently of the constant’s value.

The remaining paths preserve disabled startup defaults, exact phase/parent-operation and complete binding checks before reader access, private evaluator provenance, copied configuration, three deterministic bounded reads, adverse/error/unknown denial, original freshness ceilings, exclusive expiry, and redacted diagnostics. No existing runtime consumer installs this preparatory subset as recovery authority.

**Verification:** parent-supplied focused regression RED/GREEN, focused carrier tests, complete risk race/count-two run, and vet passed. Package coverage is **90.8%, not release-compliant**. New-tree fuzz was pending when supplied; old-tree fuzz is not credited to this tree. I performed read-only source/structural checks, not an independent test rerun.

**Claim boundary:** approval covers this local carrier-subset development checkpoint only. Canonical service/contracts and catalogs, full risk/counters, durable one-use issuance/shared-Work consumption, live provider interoperability, exact coverage/mutation, and remaining release gates remain unfinished. This is not risk-unit completion, release readiness, or permission to enable phone recovery in an application.
