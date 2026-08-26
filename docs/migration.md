# Migration From Copied Tooling

1. Inventory modules, packages, gates, external runtimes, service fixtures,
   API baselines, and mutation checkpoints.
2. Add strict `.golib.yaml` without duplicating manifest facts.
3. Move source-specific evidence to `.verification/` and preserve exact content
   identity. Keep package fixtures in durable descriptive locations.
4. Replace the Makefile with thin `golib` commands and CI with the immutable
   reusable workflow.
5. Remove `.golib` scripts, shared tool pins, and obsolete workflow code.
6. Run repository validation and content-identical parity before committing.

Migrate in bounded cohorts. Do not rerun expensive evidence because Git history
changed, and do not migrate a checkpoint whose semantic identity cannot be
proven.
