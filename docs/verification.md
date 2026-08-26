# Verification

The standard contract validates formatting, module tidiness, unsafe imports,
vet, tests, races, exact package coverage, mutation evidence, linting,
vulnerabilities, secrets, licenses, fuzzing, documentation, API compatibility,
conformance, interoperability, and benchmarks when enabled by the manifests.

The documentation gate requires a regular root README, bounds document count
and size, rejects trailing whitespace and symlinks, and verifies local Markdown
targets without making network requests. It then runs CSpell `10.0.0` from an
embedded lockfile in the task-owned workspace. Consumer repositories retain
only their word policy in `cspell.json`, not shared npm manifests or lockfiles.

Coverage is evaluated per production package and must be exactly 100%.
Mutation reports must account for every viable mutant and kill 100%; equivalent
or unreachable cases require narrow reviewed records.

Evidence is keyed by complete behavior-affecting content and verifier identity,
not Git history. It is persisted atomically when available. Code, tests,
fixtures, relevant configuration, tools, or service identity invalidate only
the affected evidence. Missing or ambiguous identity fails closed.
