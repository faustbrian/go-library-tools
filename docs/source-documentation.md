# Source documentation checks

Run the source-only commands from a Git worktree or one of its subdirectories.
They inspect tracked module boundaries without requiring `.golib.yaml`, a root
module, or generated framework manifests. Policy-enforcement commands retain
their separate configuration and version checks.

`golib source-docs inventory --json` inventories current contents of Git-tracked
Go files and module boundaries without compiling packages or selecting a build.
It includes nested modules, platform/tag variants, tests and examples. Tracked
`testdata`, `fixture` and `fixtures` paths are identified as fixtures; vendor Go
files are counted but excluded. Negative Go fixtures remain visible as manual
review findings instead of masquerading as audited production packages.
Generated source is identified with Go's canonical generated header and is not
subject to hand-maintained exported-comment requirements.

The deterministic schema-version-1 report contains `modules`, `files`,
`excluded_vendor_files` and sorted `findings`. Each finding has a repository
path, line, code, severity and fixed diagnostic. Source text, link destinations,
generator commands and credential matches are never copied into diagnostics;
credential-shaped file paths are redacted. Paths, individual reads, total source
bytes and finding counts are bounded. Unsafe/symlinked inputs, missing tracked
files, invalid production Go syntax, budget exhaustion and acquisition failures
fail both modes instead of publishing a partial success. Untracked files are
not covered. The caller must provide a stable repository filesystem.

`golib source-docs check --json` uses the same inventory and additionally exits
1 for objective error findings. Inventory mode exits 0 with findings, 1 for an
acquisition failure and 2 for invalid command syntax. Either command can omit
`--json` for bounded human-readable diagnostics. `make source-docs-inventory`
and `make source-docs-check` provide source-built local routes with owned caches.

Objective checks require a canonical `Package name` comment (or `Command` for a
main package), exported declaration comments beginning with their name, and
documented consumer-visible struct fields/interface methods. Public type
definitions, signatures and value types/initializers retain reachable anonymous,
aliased and promoted members, including concrete private types exposed to callers.
Private wire layouts, implementation receivers and function-local types are not
forced to repeat contracts owned by public interfaces. This is bounded lexical
reachability, not Go type checking; imported or dynamically inferred surfaces
still require the repository's API and manual review. Shared grouped-value
comments and grouped field names are supported; embedded members inherit type
documentation. Test and fixture declarations are not mistaken for public APIs,
nor are external `_test` namespaces required to have package documentation.
`Deprecated:` must begin its own paragraph and contain guidance. Existing vet,
Staticcheck and strict lint retain their own applicable checks, including
sentence punctuation; this audit does not score prose or require internal
declaration comments.

TODO/FIXME/HACK debt must have an issue reference or enough owner and trigger
context under repository policy. The audit recognizes issue references and
explicit `owner=`/`trigger=` tokens, but accepts free-form context as a visible
manual-review item rather than imposing a new comment syntax. Bare markers
without explanations fail strict checks. NOTE/WARNING and actionable markers remain manual-review findings
for their semantic truth, hazards, removal conditions and safe modification
procedure. The check does not infer these properties from word counts.

Generated files need a same-directory `//go:generate` declaration or an explicit
`by <local-path>` producer in their canonical header. Input paths mentioned
elsewhere do not establish generator ownership; ambiguous or external identities
remain manual-review findings. Recognized local targets must
be Git-tracked, regular and inside the repository without symlinks. A header
may instead name a contained Go command-package directory with tracked,
hand-maintained source declaring a main function; test, fixture and generated
files alone cannot establish that producer. Build constraints do not hide its
source from inventory. This does not certify compilation or execution. Local targets
of supported `go run` forms are checked without executing them.
All consecutive local `.go` source arguments are checked, and Go double-quoted
arguments are recognized. Lexical admission is bounded to 64 KiB and 256
arguments; aliases, environment expansion, unsupported flags, malformed quotes
and external tools remain explicit review items without claimed availability.
This establishes an available declaration only:
it does not prove tool installation, freshness or reproducibility. Run the
repository's applicable regeneration check separately; the audit never runs a
generator. Markdown-style inline and link-definition destinations in comments
are checked for supported HTTP(S)/mailto structure and safe, existing local
targets. Remote reachability, fragment anchors, specification versions and
normative interpretation require separate applicable link/specification proof.

Narrow credential-shaped patterns in comments and runnable-example literals
are rejected without disclosure. This supplements rather than replaces
`golib secrets check`; the pinned scanner remains the comprehensive secret
authority. Existing affected example tests and vet prove example execution and
compilation; AST inventory alone cannot establish them.

Tools CI runs the source-built nonblocking inventory. The reusable library CI
input `source_documentation` defaults to blank for backward compatibility;
explicit `inventory` or `check` opts in. Unknown nonblank values fail. Adoption
requires a published tooling release supporting the command, or its reviewed
immutable `source_bootstrap` path. A workflow pin change alone does not upgrade
an older released binary. Strict fleet enforcement remains a per-repository
rollout dependency, not something an inventory report certifies.
