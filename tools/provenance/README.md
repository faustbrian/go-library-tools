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

## Proportional V2 envelope

`check_v2.py` inventories the same frozen 25 schema entries using a compact
v2 contract: one multiplexed oracle asset and one aggregate review asset cover
the complete set. It is an inventory gate, not a provenance generator; absent
assets remain explicit and v1 files and semantics are unchanged.

```sh
python3 tools/provenance/check_v2.py
python3 tools/provenance/check_v2.py --strict
```

`generate_v2.py` can emit a deterministic, non-claiming inventory for review;
`verify_v2.py` checks its exact 25-schema binding and rejects semantic outcome
claims. These tools do not create the multiplexed oracle or aggregate review.

```sh
python3 tools/provenance/generate_v2.py --output /tmp/provenance-v2.json
python3 tools/provenance/verify_v2.py /tmp/provenance-v2.json
```
