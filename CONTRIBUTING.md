# Contributing

Read [AGENTS.md](AGENTS.md) and the affected contract documentation before
editing. Keep changes focused, preserve unrelated work, and update the
changelog for user-visible behavior.

Use task-owned Go caches for every command. During development, run the
narrowest affected tests, race checks, coverage gate, and static analysis.
Before submission, run:

```bash
make ci
```

CLI, schema, workflow, evidence, or fixture changes require compatibility
tests and corresponding documentation. New external dependencies require a
maintenance, license, vulnerability, and supply-chain review.

Do not add copied automation, arbitrary execution hooks, mutable action pins,
permanent module replacements, machine-specific paths, bypass flags, or
threshold exceptions.
