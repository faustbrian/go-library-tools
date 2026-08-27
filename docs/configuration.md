# Configuration

`.golib.yaml` contains repository-specific policy and an exact released tool
version. Its schema is published at [`schema/golib.schema.json`](../schema/golib.schema.json).
Unknown fields, future schema versions, duplicate operations, unsafe paths,
and arbitrary executables are rejected.

```yaml
schema_version: 1
tool_version: v1.0.0
tool_checksums_sha256: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
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

Consumer repositories pin both `tool_version` and the lowercase SHA-256 digest
of that release's `checksums.txt` asset. The setup action verifies this digest
before trusting the archive checksum selected for the current platform. The
tooling repository may omit the checksum only while bootstrapping and testing
the release that will produce it; released consumers must include it.

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

Typed operations may invoke bounded `go test` runs or one named target from a
repository-owned Makefile. A `test` operation runs after the module's standard
test command and preserves package-specific stress, leak, or lifecycle checks.
A `make` step is reserved for package-owned verification that cannot be
expressed as a direct Go test, such as generated-source comparison or an
external-runtime conformance suite:

```yaml
operations:
  - module: .
    gate: conformance
    steps:
      - type: make
        makefile: verification/package.mk
        target: generated
        timeout: 10m
```

The Makefile path is repository-relative, symlinks are rejected, the file is
read with a fixed size limit, and its exact validated bytes are provided to
Make through standard input. Configuration cannot select an executable, add
command-line flags, interpolate shell text, or provide secrets. Enabling fuzz,
benchmark, or conformance in a module manifest requires a matching typed
operation. Declaring an interoperability tool has the same requirement.
Missing operations fail while loading the repository rather than silently
skipping an enabled gate.
