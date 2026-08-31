# Performance Results

This directory retains the raw aggregate accepted for each tooling release
that changes the performance harness or measured orchestration behavior. Files
are named after the release, contain every sample, and identify the immutable
representative revisions. The generation workflow and interpretation rules are
documented in [Performance Rehearsals](../performance.md).

## v1.1.0

[`v1.1.0.json`](v1.1.0.json) contains the four raw reports measured by
compatibility run
[`33353902578`](https://github.com/faustbrian/go-library-tools/actions/runs/33353902578)
at tooling revision `9e4fbe90e0f08ee7f3ba5f40c64115f909600624`. Run
[`33358502106`](https://github.com/faustbrian/go-library-tools/actions/runs/33358502106)
authenticated that source run and validated its run attempt, tooling revision,
repository, representative revisions, profiles, cleanup state, and complete
sample contract. The retained aggregate has SHA-256 digest
`d93d8edb0f2e5647b28c073e17f34d92534b7d622ee29814ac65781d9c63021a`.
