# Go library ecosystem threat model

Contract version: 1

## Assets and trust boundaries

Protected assets are source integrity, published module contents, tags,
workflow credentials, vulnerability and scanner evidence, module-proxy
archives, authentication and encryption keys, tokens, personal data,
operational metadata, and consumers' runtime data. Attacker-controlled input
includes pull-request content and Git history; URLs, headers, bodies, schemas,
regular expressions, archives, and filesystem paths; database values, queue
records, cache entries, webhook deliveries, and plugin results; environment and
process output; third-party actions and Go tools; and generated or uploaded
evidence.

Trust changes at every network, DNS, proxy, redirect, filesystem, database,
queue, cache, parser, resolver, plugin, CI, artifact, and release boundary.
Authentication proves an identity only for the explicitly configured issuer,
audience, key, algorithm, and use. Authorization is a separate fail-closed
decision and is never inferred from successful parsing, authentication, tenancy
metadata, or possession of an administrative API object.

## Threats and controls

| Threat | Boundary | Required control |
| --- | --- | --- |
| Vulnerable dependency or reachable standard-library flaw | module graph | pinned `govulncheck`, dependency review, and a reviewed release verdict |
| Unsafe implementation or sensitive data flow | Go source | standalone pinned `gosec` plus centrally owned `go-analysis` security policy |
| A broad or unexplained scanner suppression hides a new finding | Go source and analyzer policy | exact-rule native gosec suppressions with a rationale, and same-comment-line explanations for `nolint:gosec`; centrally generated policies ignore repository-owned scanner configuration |
| Credential committed then deleted | Git history | gitleaks scans Git history, with a tooling-owned policy that repositories cannot widen |
| Workflow privilege or supply-chain substitution | GitHub Actions | immutable action SHAs, least privilege, Actionlint, and the owned workflow security analysis |
| Archive traversal, links, devices, mode abuse, or expansion exhaustion | bootstrap download | HTTPS and digest verification followed by bounded entry/type/path/size validation before extraction |
| SSRF, unsafe proxying, redirect escalation, or DNS rebinding | outbound network | caller-owned destinations and transports, explicit redirect/proxy policy, resolved-address checks where applicable, and bounded requests |
| Path traversal, symlink escape, or check/use replacement | filesystem | canonical containment, link rejection or descriptor-relative operations, explicit roots, and checks at the operation boundary |
| SQL, command, template, header, or protocol injection and request smuggling | database, process, renderer, and protocol parsers | parameters and argument arrays, strict grammar and framing, explicit transaction ownership, and differential tests |
| Decompression, reference, parser, regex, or fan-out exhaustion | archives, schemas, formats, resolvers, and worker orchestration | pre-allocation size, depth, count, concurrency, timeout, retry, and memory limits |
| Replay, duplicate, poison-message, reordering, or partial success | queues, schedulers, webhooks, outboxes, and idempotency stores | authenticated resource-scoped replay identity, durable atomic claims, bounded retries, and explicit dead-letter and recovery ownership |
| Timing or policy bypass | authentication and authorization | constant-time secret comparison, issuer/key/use binding, explicit capabilities, and fail-closed middleware and tenant policy |
| Secret or personal-data disclosure | logs, traces, metrics, errors, panics, examples, fixtures, and CI artifacts | typed redaction, bounded diagnostics, safe panic containment, and secret scanning of history and the current tree |
| Race, deadlock, leak, stale cache, or stuck shutdown | concurrent callbacks, caches, background work, and lifecycle hooks | caller cancellation, bounded ownership, no locks across callbacks or I/O, race/leak tests, and explicit close semantics |
| Hidden network, filesystem, process, environment, goroutine, or global registry behavior | constructors and package initialization | inert defaults, explicit injected dependencies, caller-owned lifecycle, and no process-global mutable registry |
| Scanner output leaks source or credentials | retained evidence | suppressed scanner streams with independent size limits, process-tree termination on overflow, payload-free diagnostic classes, evidence pointers, and source-equivalent access control |
| A mutable or incomplete result pointer is presented as exact evidence | scanner and release records | immutable source revision, tool version, command, completion time, and SHA-256 result digest |
| A scanner result is mistaken for a release decision | release boundary | an owner records a per-revision verdict and residual risks; automated scans never replace review |
| Dependency, action, release-key, or maintainer compromise | module, CI, and publication control planes | minimum permissions, immutable pins and checksums, independent review, provenance, scoped releases, and revocation/advisory procedures |

NilAway remains warning-only because it is an advisory correctness signal, not
a promoted security control. Race, fuzz, mutation, performance, conformance,
and aggregate fleet checks remain risk-selected under proportional assurance.

Security-sensitive repositories refine this shared model with their concrete
inputs, assets, trust decisions, resource budgets, failure states, tests, and
residual risks. Scanner success is not evidence that these design obligations
were examined.

## Review conditions

Review this model when a trust boundary, release mechanism, scanner policy,
archive format, credential scope, supported platform, or public security CLI
contract changes, and after any confirmed ecosystem security incident.
