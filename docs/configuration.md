# Configuration

`.golib.yaml` contains repository-specific policy and an exact released tool
version. Its schema is published at [`schema/golib.schema.json`](../schema/golib.schema.json).
Unknown fields, future schema versions, duplicate operations, unsafe paths,
and arbitrary executables are rejected.

```yaml
schema_version: 1
tool_version: v1.0.0
manifest:
  modules: modules.json
  packages: packages.json
evidence:
  root: .verification
mutation:
  root: .verification/mutation
runtimes:
  deno: 2.9.4
  zsh: "5.9"
```

Defaults match the example paths. `modules.json` owns module identity, Go
version, lifecycle, gates, tags, services, and release metadata.
`packages.json` owns package classification. Configuration references those
facts rather than duplicating them.

`runtimes` declares exact non-Go executables used by repository tests. Deno is
installed at the declared semantic version. The current zsh fixture supports
exactly `5.9`; unsupported versions fail configuration validation rather than
silently selecting a different shell. Node is owned by the documentation tool
chain and therefore is not repeated as repository policy.

Typed operations may invoke bounded `go test` runs for docs, fuzz,
conformance, interoperability, API, or benchmarks. Shell commands and secrets
are not valid configuration.
