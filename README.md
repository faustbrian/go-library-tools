# Go Library Tools

`golib` provides a versioned repository contract for independently released Go
libraries. It validates canonical manifests, runs proportional quality gates
in task-owned environments, and owns generic service fixtures without copying
automation into every repository.

The prepared ecosystem-milestone release is `v1.8.0`; `v1.7.2` remains the
current published release until the milestone tag is created. Consumer
repositories must use checksum-verified release binaries and immutable workflow
references.

## Quick Start

```bash
go build -o ./bin/golib ./cmd/golib
./bin/golib config validate
./bin/golib repository check
make local-ci
```

A consumer keeps policy in `.golib.yaml`, module facts in `modules.json`,
package facts in `packages.json`, and source-specific evidence under
`.verification/`. Shared scripts and tool versions remain here.

## Guarantees

- fast repository-local feedback with expensive checks selected by material
  risk;
- task-owned Go caches, temporary files, credentials, and service resources;
- no shell evaluation of repository configuration;
- strict schemas, deterministic ordering, bounded input, and fail-closed gates;
- central consumer catalogs and known-good compatibility selections without a
  fleet-wide dependency bundle; and
- immutable release and GitHub Actions consumption.

Start with the [documentation index](docs/README.md). See
[CONTRIBUTING.md](CONTRIBUTING.md) for development, [SECURITY.md](SECURITY.md)
for private reports, and [SUPPORT.md](SUPPORT.md) for support channels.

Consumers composing independently released Golib modules should start at the
[Golib ecosystem index](docs/ecosystem/README.md).

## Compatibility

The required Go version is recorded in [`.go-version`](.go-version). Released
minor versions preserve documented CLI, configuration, workflow, and evidence
contracts according to [DEPRECATION.md](DEPRECATION.md).

## License

Go Library Tools is available under the [MIT License](LICENSE).
