# Troubleshooting

`configuration invalid` indicates an unknown field, unsupported schema, unsafe
path, duplicate operation, or invalid typed step. Validate against the
published schema and run `golib config validate`.

`evidence ... mismatched` means behavior-affecting inputs differ from the stored
record. Inspect with `golib evidence inspect --json`; rerun only the affected
package and gate.

Service failures report the first lifecycle boundary and still attempt exact
cleanup. Docker Compose startup failures include a bounded diagnostic excerpt;
generated fixture credentials are redacted. Do not run broad Docker cleanup.
Confirm the pinned image is available, the runtime supports dynamic loopback
ports, and no task-owned resource remains.

Missing tools and required fixtures are failures, not skips. NilAway is the
only advisory analyzer.
