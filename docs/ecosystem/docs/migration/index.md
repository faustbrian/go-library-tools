# Migration routes

[Documentation map](../index.md) · [Ecosystem entry point](../../README.md)

- [Laravel/PHP migration](laravel-php.md): concept map, coexistence, singular
  writer/migration ownership, payload compatibility and rollback.
- [Standalone Go adoption](standalone-go.md): replace one existing Go boundary
  without adopting a framework or silently changing persisted/wire formats.

Choose either the frozen [compatibility cohort](../../compatibility-sets.md) or
a deliberate package-current version boundary. New Migrations/Postgres majors
have materially different table/DSN contracts; the Laravel route distinguishes
them explicitly. Do not replay existing migration history or assume a current
major's snippets apply to the cohort.

Next: [versioning](../versioning.md), [operations](../operations/index.md).
