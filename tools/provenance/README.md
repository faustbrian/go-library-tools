# Provenance input checker

`check.py` inventories the exact 25 schema entries and the local release and
oracle/review files required before `cohesion-schema-provenance.json` can be
authentically assembled. It reports exact-byte SHA-256 values for files that
exist and names every missing input.

The checker is deliberately not a generator: it never invents an oracle,
review, source revision, release locator, or digest. Use `--strict` in a gate
once all required evidence is present; without it, the report is useful while
the evidence set is incomplete.

```sh
python3 tools/provenance/check.py
python3 tools/provenance/check.py --strict
```
