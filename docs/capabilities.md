# Legacy Capability Map

The shared CLI replaces the copied `.golib` implementation. Repository policy,
fixtures, API baselines, and content-addressed evidence remain with the library
that owns them.

| Copied capability | Shared replacement | Repository-owned input |
| --- | --- | --- |
| `repository-check.sh` | `golib repository check` | manifests, modules, workspace |
| `run-modules.sh`, `check-module.sh`, `check-gates.txt` | `golib check` | enabled module gates |
| `with-disposable-go-cache.sh`, `isolated-go.sh` | process executor | none |
| `build-local-proxy.sh` | owned-module input and task workspace handling | module graph |
| `check-go-safety.go`, `check-go-safety.sh` | native safety gate | Go source |
| `check-coverage.sh` | native exact coverage gate | package classifications |
| `check-sbom.sh` | native bounded CycloneDX 1.6 generation and validation | module graph |
| mutation scripts and Gremlins patches | native mutation campaign | reports, reviews, approvals |
| `restore-ci-mutation-evidence.sh` | `golib mutation import` | approved checkpoint archive and ledger |
| gate evidence and snapshot scripts | content-addressed evidence store | `.verification` records |
| `check-api-baseline.sh`, `update-api-baseline.sh` | `golib api check`, `golib api update` | API baseline |
| package Make targets for fuzzing | typed `fuzz` operations | target and budget |
| package Make targets for conformance | typed `conformance` operations | conformance fixtures |
| package Make targets for interoperability | typed `interoperability` operations | reference fixtures and runtimes |
| package Make targets for benchmarks | typed `benchmark` operations | benchmark target and budget |
| `check-documentation.sh` and npm tool manifests | native docs gate | Markdown and `cspell.json` |
| tool installation in `versions.env` | centrally pinned Go tools | runtime versions only when package-specific |
| service start and stop scripts | gate-owned typed service leases | topology and compatibility payloads |
| `codeql-build.sh` | reusable workflow CodeQL matrix | build tags and modules |
| `release.sh`, `filter-releasable-modules.sh` | `golib release check`, `golib release dry-run` | release metadata |
| `stage-ci-evidence.sh` | reusable workflow artifact upload | `.verification` |
| `package.mk` | canonical gates and typed operations | no copied implementation |

`audit-goals.sh` and migration-only evidence adapters have no runtime
replacement. They describe the completed development process rather than a
consumer repository capability. Package-specific scripts remain only when they
are fixtures or test subjects and are invoked through a typed operation.
