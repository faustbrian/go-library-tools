# Native independent review: public creation/recovery composition

Principal: `/root/identity_public_composition_final`, default agent,
GPT 6.1 Sol medium, isolated `fork_turns=none`, read-only. No runtime access.
Root reviewed dirty bytes committed as `ff5367713d5081021d765bd679519bbc8e1337fb`;
optional reviewed dirty bytes committed as `5335ad7f195e55a1ca895d71578fbf5b00e2d71a`.

## Initial native disposition

**Root-only development commit: APPROVE. No findings.**

Reviewed base/HEAD `b956e68880ab4fcee4a0e14947db0f38dc2e2b4a` plus dirty
`postgres/writer.go`, `postgres/prepared_test.go`,
`postgres/mutation_boundaries_internal_test.go`, and root `CHANGELOG.md`.

The metadata cast preserves direct pgx semantics and fixes explicitly typed
`Work` execution without changing validation, argument ownership, batch
atomicity, or dependency boundaries.

**Optional composition batch: REQUEST_CHANGES.**

- **Blocker:** `composition_integration_test.go:80` unconditionally deletes
  fixed audit roles even when preflight fails because those roles already
  exist. Cleanup is registered before role reservation; failed preflight rolls
  back newly created roles but preserves pre-existing ones, which cleanup can
  then delete. Guard deletion with ownership established by successful
  installation. Falsification: the loopback/database guard does not establish
  cluster-role ownership, and transactional preflight does not protect
  pre-existing roles from the later administrative connection.
- **Consumer readiness gap:** `go.mod:14` still pins root `6ac8115d013c`, whose
  insert lacks the fix. Workspace-backed passes do not establish behavior
  using that declared dependency. Advance the pin after the root commit and
  reassess the complete optional diff.

The remaining optional dirty paths and new composition test were reviewed
completely. Real public components are composed; independent SQL observations
check receipts, changed-intent conflict, denied requests, success/failure audit
contents, lifecycle contents, and abandoned-reservation recovery.

**Evidence supplied:** composition race/integration suite, root race tests,
both vet commands passed; independently inspected `git diff --check` passed.
No tests or runtime access performed by this reviewer.

Coverage/mutation, clean-consumer, provider, and release gates remain pending;
this approves only the root local development commit, not completion or
publication.

## Final native disposition

**APPROVE for the optional local development commit. No findings.**

Reviewed exact base/HEAD `ff5367713d5081021d765bd679519bbc8e1337fb` plus
complete dirty `postgres/identity/{go.mod,README.md,CHANGELOG.md}` and new
`composition_integration_test.go`.

Both prior findings are resolved:
- Role deletion now requires successful atomic installation; failed preflight
  preserves pre-existing roles.
- The root dependency pins the committed JSONB fix, with matching revision
  and timestamp.

Public composition, replay/conflict, denial, terminal observation,
real-participant recovery, and audit/lifecycle oracles remain intact.
Dependencies remain optional and acyclic.

Supplied evidence: collision experiment preserved the existing role;
`go test -race -tags=integration -count=2 ./postgres/identity` and integration
vet passed; fixture resources `0|0`. Independently checked
`git diff --check`: passed.

No reviewer runtime access or test execution. Published clean-consumer,
coverage/mutation, provider, and release evidence remain pending; approval
does not certify release or platform completion.
