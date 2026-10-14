# Native independent identity outbox persistence review

**APPROVE — local-development batch only. No findings.**

Principal: `/root/outbox_identity_persistence_review`; default GPT6.1 Sol medium,
fork_turns=none, single structured reviewer, read-only, no further agents or
mutations.

Boundary: `feature/identity-platform` at base/HEAD
`6ac8115d013cc38ba4b6437721d653cf97a2fd94`, including the complete dirty patch
in `docs/README.md`, both manifests, and all `postgres/identity` files.
Consumed identity PostgreSQL source verified at
`677584cab9c08e7aa15b8e0fcc3abd06035ebb1a`.

Source-backed closure:

- Public contribution requires complete audited `Work.StagedLifecycle()`;
  execution remains through that same Work, with no transaction constructor,
  closure, callback, delivery or cascade authority.
- Root `Writer.PrepareInsert()` owns insert validation/SQL. Safe projection
  preserves uint64 versions as decimal strings and canonical Work effect IDs.
- Unique retained effect and tenant/command/index bindings serialize competing
  inserts. Exact replay compares immutable fields, tolerates relay-owned
  scheduling/state changes, and cannot recreate pruned payloads.
- Binding mutation triggers reject ordinary UPDATE/DELETE/TRUNCATE without
  leaking values. The additive migration does not modify populated root tables;
  downgrade fails explicitly. Documentation correctly excludes hostile
  administrators from trigger guarantees.
- Failure prefixes remain inside the owner's transaction; documented abort
  responsibility prevents partial binding/message commits.

Verification: independently ran `git diff --check` successfully. Reviewed
supplied compiler-red evidence and reported passing
`go test -tags integration -race -count=2 -timeout=120s ./...`,
`go vet ./...`, public tests and external-package registry example.
Runtime results were supplied by the parent, not independently rerun.

**Not release-ready or complete-unit evidence:** full inventory remains pending
because the CLI is absent; unpublished dependency clean-consumer proof, exact
coverage/mutation and release gates remain pending. The private SQL seam proves
owned persistence, not a public Work/Coordinator command journey, complete
lifecycle/cascade composition or all-67-unit acceptance.

Final unchanged reviewed patch committed locally at
`3e1ab9a8c7c327538b6daee014683c45d25f606c`, not pushed.
