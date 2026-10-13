# Limitations and intentional boundaries

[Documentation map](index.md)

The frozen [residual register](../../../release/cohesion-residuals.json) belongs
to `golib-compat-v1-20260912.2`. It is not silently rewritten when newer
documentation or package releases close a gap. This page explains every
material disposition and the adoption workaround. Repository-specific current
security/compatibility findings remain with their owning release records.

| Scope and versions | Classification and impact | Alternative, workaround and revisit owner |
| --- | --- | --- |
| Planned Identity, OAuth Server, Organization, Passkey, SCIM, Secret Store, SSO, WebAuthn roots | Planned product boundaries; no installable public implementation | Use an existing reviewed identity provider/credential source. Identity-platform owns implementation/review/publication before promotion. |
| Verkle Tree, all published v1 versions and current source | Experimental/non-production security boundary: do not use it to protect production integrity or confidentiality. A stable Go API or frozen catalog's active/stable label is not cryptographic qualification | Restrict it to isolated research without security-sensitive assets. The [owner's security warning](https://github.com/faustbrian/go-verkle-tree/blob/75ac4456e84dfbdc057a197de5abd9e4053e83ed/SECURITY.md#experimental-status-and-known-limitations) records unaudited backend/transcript, cancellation, mutable-global and side-channel limitations. Maintainers must resolve those boundaries and obtain independent cryptographic and affected-contract qualification before any production-readiness claim. |
| Released legacy adapter/facade paths listed in the residual register | Compatibility constraint, not a defect merely because naming differs | New adoption selects the released target-oriented successor. Existing imports retain support; removal requires owned/clean-consumer migration, the longer of 180 days and two stable minor releases, and an authorized major. |
| Golib math package identifier `gomath` | Intentional naming exception avoiding standard math collision | Use its documented package identifier; reconsider only at an authorized major with demonstrated migration benefit. |
| External consumers outside the owned fleet | Missing evidence, not proof of no consumers | Preserve compatibility obligations. Maintainers own authenticated usage discovery and reconciliation. |
| Tooling internal assets/cohesion/JCS packages at the frozen scope | Permanent assurance exception; exact statement coverage is not claimed | Rely on stated focused contracts and applicable checks, not an invented 100% coverage badge. Revisit only if boundaries materially change. |
| Repository-owned authority/checksum findings listed in the frozen register | Affected next-release boundaries remain separately owned; old cohort membership does not waive them | Inspect each owner's current upstream authority, checksum and release record. Resolve before the affected publication; do not characterize every listed package as an unsafe runtime. |
| Missing migration/operator/contributor routes in the historical residual | Downstream discoverability gaps, not API defects | This portal supplies those routes when published and linked. Publication does not retroactively alter the frozen register or certify a combined deployment. |
| Fourteen frozen composition recipes | Bounded composition support, not production deployment certification | Read each recipe's external/backend boundary; qualify actual infrastructure, business behavior and deployment separately. |
| OpenAPI representative handler | No generated-contract proof from a handwritten handler | Test runtime/description/client alignment before making that claim. |
| Kafka/CloudEvents and Track/Location projection recipes | Application-owned Kafka/schema seams and bounded HTTP search seam; no broker-runtime or live OpenSearch proof | Run the owning external interoperability checks for your backend/version before deployment. |
| Track, Postal and Location service-role fixtures | Public construction/lifecycle evidence, not application business behavior | Test each real application's behavior, persisted formats and deployment separately. |
| Ambiguous commit, webhook timeout or broker settlement | Unknown external outcome, not known rollback/failure | Reconcile by a stable application operation identity; do not blindly retry side effects. |

The application deliberately owns configuration, policy order, transactions,
provider setup and reverse-order cleanup. Absence of a global service container
or default resilience stack is an intentional non-goal, not a missing feature.
Conversely, an unsupported backend/specification feature must remain an explicit
limitation with its owner, not be relabelled as design philosophy.

Newer releases may close a recorded release boundary or add a capability; use
their native changelog/security evidence rather than editing history. See
[status](status.md) and [versioning](versioning.md).

Next: [package selection](choosing-packages.md), [operations](operations/index.md),
[security](../security/README.md).
