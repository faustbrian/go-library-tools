# Architecture

`cmd/golib` is a thin process boundary. Internal packages own strict
configuration, canonical inventory, gate orchestration, exact coverage,
mutation execution, API compatibility, evidence, repository validation, and
generic service lifecycle.

Consumer repositories own only facts whose meaning is repository-specific:
module and package manifests, `.golib.yaml`, API baselines, conformance data,
interoperability fixtures, and content-addressed evidence. They do not own a
copy of the implementation.

All external commands run through an executor with a task-owned workspace and
isolated Go caches. Service fixtures receive unique resources and export their
environment only to the gate they serve. Repository content is untrusted and
is never evaluated as shell input.
