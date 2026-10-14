# Historical Go input packaging

`historical-go-inputs.tar` preserves seventeen captured Go source, module and
checksum files at their original repository-relative paths, with unchanged
extracted bytes. SHA-256:
`6019f25fc5781ef830701409066a700e23ee2373380403610c955ab4a7aec13a`.

The original files also remain in control checkpoint
`d9d4c5bcaf768ef2d32af036d258c940ea3502b7`. They are historical native test
inputs, not new current Tools packages or modules. Keeping loose Go sources
and go.mod files in this repository made the source audit assign them current
production ownership and audit a copied older Go standard-library file.

Do not modify the captured bytes to satisfy current source checks. For historical
inspection or evidence reuse, extract into task-owned disposable storage and
check original identity and applicability. Packaging does not make old results
fresh, remove stale-input constraints, or qualify any current unit. Existing
native reports, assignments and their original execution identities are unchanged.

```sh
tar -tf .ai/identity-platform/evidence/historical-go-inputs.tar
# From a selected task-owned disposable directory:
tar -xf /Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform/.ai/identity-platform/evidence/historical-go-inputs.tar
```
