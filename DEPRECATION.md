# Deprecation Policy

Deprecations MUST identify the replacement, reason, migration steps, and
earliest removal version. At `v1` and later, a supported replacement SHOULD
exist for at least one minor release before removal.

Security or correctness defects MAY require faster removal when continued
support is unsafe. Silent behavior changes, undocumented aliases, and
indefinite compatibility paths are prohibited.
