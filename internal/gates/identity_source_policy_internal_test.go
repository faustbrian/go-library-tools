package gates

import (
	"os"
	"strings"
	"testing"
)

// These source literals come from the independently classified Identity history:
// public cursor IDs, deterministic retry inputs, prose, event names and generated
// API digests. Split detector-shaped literals so the test source stays scanned.
func TestIdentitySourcePolicyPinnedScanner(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("opt-in pinned scanner contract")
	}
	join := func(parts ...string) string { return strings.Join(parts, "") }
	fixtures := map[string]string{
		"apikey/postgres/metadata_test.go":  join("canonical-key\", \"", "01ARZ3ND", "EKTSV4RR", "FFQ69G5FAV", "\"\n"),
		"http/selection.go":                 "// ID/OpenAPI ID 128, risk/" + "idempotency 32, event semantics 4096 bytes. Canonical\n",
		"risk/valkey/README.md":             "`rediss` endpoint, verified TLS trust, namespace/replay key, " + "policy/digest keys,\n",
		"postgres/otp_execution.go":         "// during exact-owner finalization. Secret-handling, " + "missing/foreign/returned\n",
		"session/preauth.go":                "// uint32 key version, " + "uint16-length-prefixed ID/tenant/action/origin/callback/\n",
		"session/README.md":                 "  credential, " + "factor/passkey applicability and exact derived scope keys.\n",
		"postgres/credential_compromise.go": "\t\top, event = \"identity.passkey.mark-compromised\", \"" + "passkey.compromised.v1\"\n",
	}
	for _, path := range []string{"delivery/receipt_test.go", "delivery/postgres/receipt_test.go", "delivery/postgres/receipt_integration_test.go"} {
		fixtures[path] = "IdempotencyKey: \"" + join("synthetic-", "receipt-0001") + "\"\n"
	}
	for _, value := range []string{"dedupe-0001", "receipt-0001", "reconcile-0001"} {
		fixtures["delivery/service_test.go"] += "IdempotencyKey: \"" + "synthetic-" + value + "\"\n"
	}
	// The historical generator uses Digest::SHA256.file(api).hexdigest. These
	// six distinct values account for nine historical findings, not credentials;
	// this test does not claim to regenerate the unavailable API preimages.
	for _, digest := range []string{
		join("af6f4235", "feabeccf", "1bca7a7d", "8ef57dcf", "660b5153", "a40d6ecd", "454e7183", "dedb8245"),
		join("80701423", "b438a6e9", "f4c51c1b", "5f18ff95", "20641b4f", "a85f2c7a", "44e0ddd3", "8cc08b1b"),
		join("e6c5d002", "a9e9b28a", "fddc3616", "71754678", "29050473", "e936bd3a", "71d02750", "5fca1b2a"),
		join("8942f6cc", "8c105a6c", "ddc440a1", "77f0160d", "589dea5e", "a4bc54fe", "b5972768", "9dd7248d"),
		join("18fd1dde", "5da68f8c", "08c9d7de", "142178e7", "e95c7de8", "efa54699", "abb7906b", "f5507b78"),
		join("0ee4dfe4", "b28e0d70", "f7c96ed1", "f2749ec2", "d240444d", "19b5dbeb", "3c06b1ee", "b8601634"),
	} {
		fixtures["http/catalog.json"] += "{\"api_" + "sha256\":\"" + digest + "\"}\n"
	}
	credential := join("A1b2C3d4", "E5f6G7h8", "I9j0K1l2", "M3n4O5p6", "Q7r8S9t0", "U1v2W3x4", "Y5z6A7b8", "C9d0E1f2")
	for _, scenario := range []string{"classified literals", "same-path changed values", "same-line adjacent credentials", "different paths"} {
		t.Run(scenario, func(t *testing.T) {
			files := make(map[string]string, len(fixtures))
			for path, literal := range fixtures {
				switch scenario {
				case "same-path changed values":
					switch {
					case path == "http/catalog.json":
						literal = "{\"api_" + "sha256\":\"" + credential + "\"}\n"
					case strings.HasPrefix(path, "delivery/"):
						literal = "IdempotencyKey: \"" + credential + "\"\n"
					case path == "apikey/postgres/metadata_test.go":
						literal = "canonical-key\", \"" + credential + "\"\n"
					case path == "postgres/credential_compromise.go":
						literal = "\t\top, event = \"identity.passkey.mark-compromised\", \"" + credential + "\"\n"
					default:
						literal = strings.TrimSuffix(literal, "\n") + " api_key=" + credential + "\n"
					}
				case "same-line adjacent credentials":
					literal = strings.ReplaceAll(literal, "\n", " api_key="+credential+"\n")
				case "different paths":
					path = "other/" + path
				}
				files[path] = literal
			}
			root, policy := gitleaksRepository(t, files)
			for _, mode := range []string{"git", "dir"} {
				findings, _, err := runGitleaksIntegration(t, root, policy, mode)
				if scenario == "classified literals" {
					if err != nil || len(findings) != 0 {
						for _, finding := range findings {
							t.Logf("safe finding: mode=%s path=%s rule=%s", mode, finding.File, finding.RuleID)
						}
						t.Fatalf("classified literals rejected: mode=%s findings=%d", mode, len(findings))
					}
					continue
				}
				if err == nil {
					t.Fatal("scanner did not reject negative controls")
				}
				for path := range files {
					found := false
					for _, finding := range findings {
						found = found || finding.File == path && finding.RuleID == "generic-api-key"
					}
					if !found {
						t.Fatalf("negative control lost: mode=%s path=%s", mode, path)
					}
				}
			}
		})
	}
}
