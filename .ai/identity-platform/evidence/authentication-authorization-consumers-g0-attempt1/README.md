# Candidate-bound consumer attempt 1

The coordinator copied immutable committed repository snapshots and wired
disposable consumer manifests to the integrated authentication, authorization
and capability candidates. No user checkout, package implementation or tracked
package manifest was changed. `results/source-inputs.json` captures the exact
commits, Git trees and complete snapshot file digests before temporary wiring.
Each successful test has separate pre-execution inputs and native output hashes.

Passed with race detection:

- Authorization `integration/contracts`.
- HTTP middleware `integration/siblings`.

Six consumers did not reach behavioral tests:

- Authentication `authotel`, `jwt` and `oidc`: their committed clock v1.0.0
  checksums differ from the fetched module and from the root checksum. The
  checksum verifier rejected acquisition; it was not bypassed.
- Service `compatibility` and `integration/reference-http`: the committed
  consumers import authentication `adapters/http`, absent from the selected
  integrated candidate. An independent compatibility advisory is pending.
- Queue `authz`/`server`: acquisition failed. This attempt retained only the
  first 1,500 stderr bytes, dominated by downloads, so its first failing
  boundary is not established by this artifact. It is not a test failure or
  a pass; a separately recorded focused acquisition is required.

`batch-result.json` is failed, not an aggregate pass. Failed acquisitions have
their own records and no invented test result. The runner retains the exact
attempt behavior, including its error-capture limitation. All disposable build,
module, temporary and source-snapshot directories were removed after execution.
This attempt does not change inventory status or certify the shared-issuer
runtime milestone, publication, hosted CI or deployment.

The default Gitleaks directory scan returned exit 1 for two generic-api-key
matches at inputs.json line 7 in the successful consumers. Both values are the
public authentication candidate Git commit in `candidate_revisions`, not
credentials. The redacted native report is retained; no scanner pass or blanket
exclusion is claimed.
