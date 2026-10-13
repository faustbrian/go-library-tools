# Package catalog

[Documentation map](index.md)

The [generated consumer catalog](../catalog-consumer.md) is the authoritative
human-readable module inventory for the frozen compatibility cohort. Its
[machine-readable companion](../catalog-consumer.json) records import paths,
status, responsibility, non-goals, required and optional dependencies,
companions, construction/lifecycle styles, Go/platform support, specification
or backend scope, and package-owned documentation links. Do not maintain a
second version matrix in an application guide.

## Read an entry before installing

1. Choose the problem in [package selection](choosing-packages.md).
2. Check lifecycle status and deliberate non-goals. An adapter, compatibility
   facade and planned module are not interchangeable with an active core.
3. Follow the entry's README and API links. Check package-level security,
   compatibility, performance, examples and changelog before adopting it.
4. Select an exact published module/version from the
   [compatibility set](../compatibility-sets.md). For a newer major or version,
   follow the owning repository's migration guide and test your composition;
   that is a package-current choice outside the frozen set.
5. Install the nested module separately when its README requires it. Being
   inside an owner's repository does not make a nested module part of its core.

## Catalog boundaries

The [engineering catalog](../catalog-engineering.md) additionally contains
fixtures, examples, benchmarks, interoperability harnesses and internal tools.
These are not production dependencies just because they have a `go.mod`.
Planned Identity-family repositories and `go-secret-store` have no install
instructions. Deprecated public paths remain compatibility obligations, not
preferred new integrations; use the released canonical successor where one
exists. See [status](status.md) and [limitations](limitations.md).

Catalog publication versions and Go floors are cohort facts, not a statement
of the latest repository release. A package's current README and native release
history own newer support claims. An unexplained disagreement is a gap to
resolve, not permission to average the two claims.

Next: [install and version modules](versioning.md), [selection](choosing-packages.md).
