# Go Library Tools

`golib` provides one strict, versioned repository contract for independently
released Go libraries. It validates canonical manifests, runs quality gates in
task-owned environments, preserves content-addressed verification evidence,
and owns generic service fixtures without copying automation into every
repository.

The current stable release is `v1.1.0`. Consumer repositories must use
checksum-verified release binaries and immutable workflow references.

## Quick Start

```bash
go build -o ./bin/golib ./cmd/golib
./bin/golib config validate
./bin/golib repository check
./bin/golib check --all
```

A consumer keeps policy in `.golib.yaml`, module facts in `modules.json`,
package facts in `packages.json`, and source-specific evidence under
`.verification/`. Shared scripts and tool versions remain here.

## Guarantees

- exact production-package coverage and mutation requirements;
- task-owned Go caches, temporary files, credentials, and service resources;
- no shell evaluation of repository configuration;
- evidence keyed by behavior-affecting content rather than Git history;
- strict schemas, deterministic ordering, bounded input, and fail-closed gates;
- auditable specification decisions with executable evidence and monitored
  authoritative errata or release feeds;
- immutable release and GitHub Actions consumption.

Start with the [documentation index](docs/README.md). See
[CONTRIBUTING.md](CONTRIBUTING.md) for development, [SECURITY.md](SECURITY.md)
for private reports, and [SUPPORT.md](SUPPORT.md) for support channels.

## Compatibility

The required Go version is recorded in [`.go-version`](.go-version). Released
minor versions preserve documented CLI, configuration, workflow, and evidence
contracts according to [DEPRECATION.md](DEPRECATION.md).

## License

Go Library Tools is available under the [MIT License](LICENSE).
