# Contributing

Read [AGENTS.md](AGENTS.md) and the affected contract documentation before
editing. Keep changes focused, preserve unrelated work, and update the
changelog for user-visible behavior.

Use the pinned Go toolchain and task-owned disposable caches. Start with
focused tests and vet in the affected packages, for example:

```bash
make check-packages PACKAGES=./internal/inventory
make local-check
```

Do not run unrelated mutation, coverage, race or ecosystem campaigns for an
ordinary fix. Select additional verification for the changed contract's actual
risk. Tools' process-control, descendant-cleanup and Docker-impacting tests run
on designated hosted jobs, not an agent's local machine. `make local-ci` includes
the complete test set; it is not the locally safe substitute for focused tests.
`make milestone-check` and `make ci` are explicit aggregate entry points, not
submission prerequisites for every contribution.

See [Development](docs/development.md) for standalone library test, focused fuzz,
optional mutation and release commands. A library fix does not require sibling
checkouts, ecosystem ledgers, or an unrelated Tools release.

CLI, schema, workflow, evidence, or fixture changes require compatibility
tests and corresponding documentation. New external dependencies require a
maintenance, license, vulnerability, and supply-chain review.

Specification-backed modules must follow the
[specification decision contract](docs/specification-decisions.md), including
its decision-register, conformance, compatibility, changelog, and pull-request
review requirements.

Do not add copied automation, arbitrary execution hooks, mutable action pins,
permanent module replacements, machine-specific paths, bypass flags, or
threshold exceptions.
