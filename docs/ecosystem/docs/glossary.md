# Glossary

[Documentation map](index.md)

| Term | Meaning |
| --- | --- |
| Family | One of nine semantic selection/ownership groups; not an import layer. |
| Owner | Package responsible for the meaning and guarantees of a concern. |
| Adapter | Explicit translation to a target; not hidden infrastructure ownership. |
| Composition root | Application code that chooses dependencies, policies, acquisition and cleanup. |
| Compatibility set | Immutable evidence for exact public versions and bounded scenarios, not a release train. |
| Cohort | The exact version roster tested in a published compatibility set. |
| Package-current | Independently released newer package selection, outside the old cohort unless proven separately. |
| Facade | Retained released compatibility path; not permission to remove or retag it. |
| Acquisition | Context-bearing external I/O that makes resource ownership visible. |
| Drain | Stop new intake and wait for accepted work to reach the stated boundary. |
| Shutdown | Bounded complete ordered shutdown under the owner's contract. |
| Close | Immediate synchronous release under an `io.Closer`-compatible contract. |
| Wait | Join an already-owned operation, without acquiring hidden ownership. |
| Acknowledgement | Transport settlement after the stated processing boundary, not decode success. |
| Unknown outcome | External effect may have happened; not equivalent to known failure or safe retry. |
| Idempotency | A defined operation identity returns/preserves a prior outcome without duplicating its effect. |
| Outbox | Business change and durable delivery intent committed in one transaction. |
| Reference recipe | Executable public-API composition with stated fixture/backend scope; not production certification. |

See [design language](../design-language.md) for normative construction,
lifecycle, error, security and adapter meanings. Preserve reviewed secondary
capability names from the catalog instead of inventing synonyms.

Next: [architecture](architecture.md), [ownership map](integration-map.md).
