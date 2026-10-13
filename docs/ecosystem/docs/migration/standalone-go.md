# Adopt Golib in an existing Go application

[Migration routes](index.md) · [Documentation map](../index.md)

Replace a single consumed contract, not the application's bootstrap all at
once. There is no required umbrella module, container or runtime.

1. Record the old observable behavior and direct callers: success, failures,
   timeout/cancellation, resource ownership, persisted/wire formats and cleanup.
2. Choose the owning module in the [catalog](../packages.md). Retain standard
   interfaces where they express the contract; optional adapters are independent
   module choices rather than mandatory imports.
3. Pin an exact public version, preserve Go's major import suffix, and follow
   its migration/compatibility guide. Do not combine an old recipe constructor
   with latest dependencies. Prove public resolution with no workspace or local
   replacement before treating an integration as installed.
4. Wire the new dependency explicitly behind the application seam. Keep a
   compatible read path/traffic switch while changing writers; do not shadow
   real side effects or hide errors to make a comparison pass.
5. Exercise the old/new contract on an equivalent bounded corpus, including
   unknown outcomes and cleanup. Deploy a limited slice, then expand only when
   the actual application operation and dependency signals agree.
6. Roll back the application boundary only while old readers can still handle
   new persisted data and payloads. Remove compatibility only after supported
   consumers and the owner's deprecation policy permit it.

Existing protocol clients, SQL schemas, cache codecs, broker consumer groups
and application idempotency keys each have their own compatibility lifetime.
A module upgrade alone does not authorize resetting any of them. Backend
recovery, replay and historical repair remain distinct operations.

Next: [architecture](../architecture.md), [versioning](../versioning.md),
[operations](../operations/index.md).
