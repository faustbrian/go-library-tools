# Integrated authorization root checkpoint

The coordinator executed the root-module contract against package commit
`a74c24b3846323e0c1627c6fccbb79ce25248392`. `native-result.json`
records exit 0 and binds the exact captured merged output. `input-manifest.json`
and `environment-inputs.json` were captured before execution; `run.rb` retains
the launch and cleanup procedure. The official Golib v1.4.0 archive was
checksum-verified against the retained release manifest before execution.

The native output reports exact statement coverage for all twenty production
packages, including 717/717 root statements. Mutation verification freshly
killed 431 root mutants and reused content-identical evidence for 608 mutants
in the remaining packages: 1,039 total, with no surviving, uncovered, timed-out,
invalid or skipped mutants. The initial mutation-import diagnostic rejected
stale historical checkpoints; those rejections are preserved, not relabeled.

`native-artifacts/` retains the native by-input records and mutation reports,
including historical records. `native-artifact-index.json` records their exact
bytes, sizes and hashes. Historical artifacts are not automatically current
passes. The forty coordinator-created package-side files were removed only
after byte-identical custody transfer. Tracked package inputs were not edited.

This checkpoint does not set the inventory unit to verified. Candidate-bound
reverse consumers, current external construction proof, remaining selected
environment closure and independent delivery disposition are separate work.
NilAway output remains advisory. A live vulnerability-feed result does not
identify immutable feed contents, and root execution alone does not establish
real database/cache integration, full shared-issuer runtime acceptance, hosted
CI, publication or deployment. Task-owned execution caches and tools were
removed; raw output bytes remain unchanged.
