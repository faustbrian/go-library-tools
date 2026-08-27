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
api:
  baselines:
    - module: .
      mode: apidiff
      path: api/v1.export
runtimes:
  deno: 2.9.4
  zsh: "5.9"
```

Defaults match the example paths. `modules.json` owns module identity, Go
version, lifecycle, gates, tags, services, and release metadata.
`packages.json` owns package classification. Configuration references those
facts rather than duplicating them.

`api.baselines` assigns at most one API policy to each module. `apidiff` mode
tracks incompatible exported API changes against a repository-relative
baseline. `go-doc` mode compares an exact normalized documentation snapshot
for an explicit package list and is intended for repositories whose existing
contract includes exported documentation. The API update command writes either
format atomically. Checksum-only and Git-history scripts migrate to explicit
baseline files; they do not remain executable hooks.

`runtimes` declares exact non-Go executables used by repository tests. Deno is
installed at the declared semantic version. The current zsh fixture supports
exactly `5.9`; unsupported versions fail configuration validation rather than
silently selecting a different shell. Node is owned by the documentation tool
chain and therefore is not repeated as repository policy.

Typed operations may invoke bounded `go test` runs for tests, docs, fuzz,
conformance, interoperability, API, or benchmarks. A `test` operation runs
after the module's standard test command and preserves package-specific stress,
leak, or lifecycle checks without admitting arbitrary shell hooks. Shell
commands and secrets are not valid configuration. Enabling fuzz, benchmark, or
conformance in a module manifest requires a matching typed operation. Declaring
an interoperability tool has the same requirement. Missing operations fail
while loading the repository rather than silently skipping an enabled gate.
