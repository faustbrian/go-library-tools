# Validate documentation claims

[Documentation map](index.md)

Run checks from the owning standalone repository. Native
[CLI documentation](../../cli.md) and
[gate policy](../../verification.md) define what each check actually proves.
No root workspace or sibling `replace` is part of public-consumer evidence.

## Routine affected changes

The current native tooling provides:

```sh
golib docs check --local
```

This checks local Markdown structure/link destinations without network access
or runtime commands. It does not compile fenced Go snippets, prove prose truth,
validate every heading fragment or establish public release compatibility.
Inspect changed anchors, navigation, actual module/status claims and executable
examples separately; a static registration/string check is not behavior proof.

```sh
golib docs check
```

The full native documentation route adds lockfile-pinned spelling, checksum-
pinned external-link checking and owner-declared typed documentation operations.
Read `.golib.yaml` before running it: nested modules and typed operations can
execute examples/tests. Every Go run uses bounded task-owned `GOCACHE`,
`GOMODCACHE` and `GOTMPDIR`; shared caches are not appropriate agent resources.
Repository cspell allowlists and deterministic link exceptions are documented
inputs, not permission to suppress genuine broken public targets.

Changed-content secret scanning uses `golib secrets check --base-sha` with the
actual full ancestor SHA. Do not infer absence of private URLs, machine paths
or production identifiers from a secret scanner alone. Review only safe field
names/counts if a finding contains sensitive values.

## Explicit navigation validation

From the owning Tools source checkout, `make docs-navigation` builds a
task-owned verifier, runs the optional navigation check and removes its caches
and binary. The updated source-built CLI additionally supports:

```sh
golib docs check --navigation
```

This additive local-only route validates actual Markdown heading fragments,
reference links, directory entry links and root-README reachability across
manifest-declared module documentation trees. It rejects personal machine
paths, symlinks/escaping paths and aggregate input above 32 MiB, without
executing examples or accessing the network. Do not combine it with `--local`
or `--module`; existing local/full commands retain their prior behavior.
Use the current source contract until a release actually publishes this new
option; older pinned tooling does not acquire it through a documentation link.

## Portal milestone

The complete milestone additionally needs:

- Required pages, heading fragments, parent/next navigation, entry-point
  reachability and package-to-ecosystem backlinks.
- Module/path/status/release claims against their actual owner and immutable
  tag, keeping package-current inventory separate from frozen cohort catalogs.
- Native catalog generation/drift validation against the declared immutable
  projection/source inputs; current-main changes do not rewrite the old set.
- Every supported [recipe](recipes/index.md) once at its stated execution
  boundary, with clean public resolution and no workspace/local replacements.
  Shared source executions can cover multiple IDs only when they assert both
  contracts. External or recovery work stays in its controlled/hosted boundary.
- Changed executable Go snippets compiled/tested through their actual owner.
  Clearly labelled excerpts/pseudocode must not masquerade as runnable programs.
- Public-document leakage review, terminology/style, contradictory or duplicate
  recommendations, evidence-backed caveats and fair comparisons.
- Adoption walkthroughs for every audience in the [documentation map](index.md),
  recording actual selection/installation/operations obstacles rather than
  treating counts or badges as comprehension evidence.

These requirements are not all covered by the baseline native local command.
A final report must name actual checker/runtime outcomes and remaining gaps;
neither this checklist nor source presence is a passing milestone result.
Unchanged immutable evidence may be reused only with its original identity and
when it still exercises the exact claim. Do not invent proof hashes for prose.

Next: [contribution](contributing.md), [recipes](recipes/index.md),
[versioning](versioning.md).
