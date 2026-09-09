# Clean-room forward-oracle tooling

This directory contains independent tooling for authoring and checking
schema-v3 forward-oracle assets. It MUST NOT import `internal/cohesion` or
reuse the production parser, validator, canonicalizer, or digest helpers.

The generator writes explicit fixture and splice bytes. The checker recomputes
canonical JSON and SHA-256 values from those bytes and verifies the frozen case
roster. A production verifier test is retained separately as a compatibility
check; it is not a substitute for the clean-room evidence.

The diagnostic and contract-review assets are completed lanes. Their checkers
are `verify.py` and `verify_contract_review.py`; each recomputes fixture
digests, canonical accepted values, splice bounds, and the frozen case roster
without importing production code. Additional schema lanes must be added only
with their decision-matrix cases, independent expected outcomes, and reviewer
evidence.
