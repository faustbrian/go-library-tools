package gates

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The oracle is immutable source, not the policy. The four aggregate native
// identifiers have source-proven non-credential roles; this test does not
// reconstruct unavailable historical manifests or assert old execution.
func identityEvidenceScannerFixtures(t *testing.T) map[string]string {
	t.Helper()
	source := strings.Join([]string{"362f94d2", "0c20d25b", "f69148a8", "93719bae", "81174121"}, "")
	inputs := []struct {
		path   string
		bytes  int64
		digest string
	}{
		{".ai/identity-platform/EXECUTION_LEDGER.md", 1789932, "679a03dfe3128ff496f8f12ccf2f8503a7ed277495d699e1b8491ac052f2f4c5"},
		{".ai/identity-platform/evidence/assignments/authentication-g0.md", 8699, "941e825917e9a76dfe96afd24cfbef4c187166e1ba8e27c7a23a084a7f0024ca"},
		{".ai/identity-platform/evidence/assignments/authorization-g0.md", 7649, "cfeee7a9c13bdfb6049508331f58207102f131c6e0fa2d1f971d17816ccdfb24"},
		{".ai/identity-platform/evidence/authentication-authorization-consumers-g0-attempt1/results/authorization-contracts/inputs.json", 1200, "3699e39a566dee62995e647f974ab02f032cf73ae74404b45c06c54a1ecef823"},
		{".ai/identity-platform/evidence/authentication-authorization-consumers-g0-attempt1/results/middleware-siblings/inputs.json", 1124, "0d69e877e713b9804bc18ae9d9fff62c686b0c7763708dc55b601163352f5c17"},
		{".ai/identity-platform/evidence/authentication-consumers-primary/results/queue-authz-server/inputs.json", 1117, "80c0458c589be8b2015cfe96ccebfe29f754d68cef24b4bf9fc5da44b2272807"},
		{".ai/identity-platform/evidence/authentication-consumers-primary/results/service-compatibility/inputs.json", 1184, "2b174a609f7ec65816e48be6268d0a454f2a3bb36243757d044ff3d872b4719f"},
		{".ai/identity-platform/evidence/authentication-consumers-primary/results/service-reference-http/inputs.json", 1198, "267c64c83b7d3d09b2466c8ecf90f06b6ae1f468f6ccf5f276f0873cd7fff0f9"},
		{".ai/identity-platform/evidence/authentication-g0-coordinator-native/artifacts/mutation.json", 139045, "ddad9c41cf2f64c5304fab1c39be98826034b97fa9b26cb02f12693f9dde1b94"},
		{".ai/identity-platform/evidence/authentication-g0-native/.artifacts/mutation.json", 139045, "96953c07f109125e6b73bbd612f1077639da850141b325f4b6135b98a23a2915"},
		{".ai/identity-platform/evidence/authorization-main-adoption.md", 2567, "aceb84a2f792eb3b8809d3bb8d7cecb42cdcb76b1ea29605d7ecec61ea765378"},
		{".ai/identity-platform/evidence/capability-g0-coordinator-native/README.md", 5131, "8db66181e81e81dbfc86082de45ec29526d32e97621c3d89253f2e2445fc132e"},
		{".ai/identity-platform/evidence/capability-g0-coordinator-native/artifacts/mutation.json", 113987, "87307a9342279ac92a9156352e9ce6e6bf2afc610a2cb1874273d2f03d242ce6"},
		{".ai/identity-platform/evidence/capability-g0-coordinator-native/focused/vulnerability.log", 392079, "46a6248e0026e66d86badc5e72708894e7d68107ab21cea7c907d0a126257f33"},
		{".ai/identity-platform/evidence/capability-g0-native/.artifacts/mutation.json", 113987, "ebac2483c9fb5e626de7d2e19d3de5a91bb915d4e2e2fdb9b557ec0f2716035c"},
		{".ai/identity-platform/evidence/delivery-lease-runtime.json", 12636, "0ddb0223913153a12619444c00251b93313c21b29bfb3f2781b8a8aae5234bc4"},
		{".ai/identity-platform/evidence/otp-development.md", 504216, "497c5d84f8420d8f4babf5d6e2983fae8933683f2c983ece80fbd9fea72d5cf5"},
		{".ai/identity-platform/evidence/password-g0-narrow/README.md", 5030, "b928542593cae082f2e06cda4c3119d934d019ecb99b63093b12a443135fff4d"},
		{".ai/identity-platform/evidence/password-g0-narrow/vulnerability.json-stream", 464888, "119103bde334c1bb1752eab1bc6771c00ad58e8bdaca92b4b17850c860a89c73"},
		{".ai/identity-platform/evidence/risk-captcha-issuance-runtime.json", 93564, "5b50f5363af804c63b4d268ce18ef769b6465f5ecffb254b5a0c2516c97c55ae"},
		{".ai/identity-platform/evidence/risk-captcha-native-runtime.json", 94866, "2b5a8e17fe09e3810f02f718facb84ebdd4654f2889099e7720933dff4e5e69c"},
		{".ai/identity-platform/evidence/risk-captcha-replay-runtime.json", 92863, "4d7ff2ba8b066448ace794c5965b0ac428a71f78855d40b798bc84e232c698ea"},
		{".ai/identity-platform/evidence/risk-decision-audit-buffer-runtime.json", 39687, "607b06bbeb6dde1945767613d77bd48a444cf7d465ac302f886e9a1bad3ed2bf"},
		{".ai/identity-platform/evidence/risk-decision-audit-runtime.json", 93380, "444c16cd8e4d8691726931521e542a984f6cc354b0e0b519b0381dc8bd891a1d"},
		{".ai/identity-platform/evidence/risk-decision-builder-runtime.json", 8867, "7d69bbf1180ff866b5b3237be8ec0de3c2e4b2fac3d9cf5ca8fe2d11044f8b66"},
		{".ai/identity-platform/evidence/risk-lockout-admission-runtime.json", 47644, "6badea41a09f0af342355fad1ebaad54048120c8f7c578d6dd14ee7b0b7f6dae"},
		{".ai/identity-platform/evidence/risk-lockout-native-race.json", 330386, "0d9fcece2d78d925068208a8bb414ffbb006d4a432945bd4b32494264ed9c009"},
		{".ai/identity-platform/evidence/risk-lockout-native-review.md", 10096, "076d568b75e337dfb252cf4625933a95060a0796fb62cd57a520f4bf3e1f5366"},
		{".ai/identity-platform/evidence/risk-lockout-native-vet.json", 330288, "fe1268b434af1a4e0044ecb3892b9171f1f0bda2f0e1b4a383b3a892963a5da0"},
		{".ai/identity-platform/evidence/risk-lockout-policy-race.json", 11494, "ee128cc92f7e0f686518a5969fafac269646e243d68ba9797f738c1deb54d641"},
		{".ai/identity-platform/evidence/risk-lockout-policy-recovery-green.json", 11560, "34a5d164811b4883785f818af4ad6140afe8c0d1964619118cfc0824bc765c19"},
		{".ai/identity-platform/evidence/risk-lockout-policy-vet.json", 11410, "fa9ba9db5c800671b1a379ff57e7b5e601655fa6f9c8d79f48c299bb520266eb"},
		{".ai/identity-platform/evidence/risk-lockout-state-final-race.json", 41393, "e28a0d23b254a9a51e48cc06e558570ca4561fa968bf4b3a35cf1a4b8f0c6d8f"},
		{".ai/identity-platform/evidence/risk-lockout-state-final-vet.json", 41154, "4d7ac462c1b99d95212733a320b9f4df10cadec95caf650c34e9bce217416faa"},
		{".ai/identity-platform/evidence/risk-lockout-state-race.json", 41520, "ccc0f562028c5c25188fbbb9d451c6e1da2be7b7468b75e5c6e23eefbbb88890"},
		{".ai/identity-platform/evidence/risk-lockout-state-vet.json", 41281, "a7f69f02ad8a8b6f075a86e6518c599100178b1f4e00b7fa4126073946e5baf9"},
	}
	files := make(map[string]string, len(inputs))
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		t.Fatal("immutable scanner oracle requires the default HTTP transport")
	}
	transport := defaultTransport.Clone()
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Timeout: 45 * time.Second, Transport: transport}
	for _, input := range inputs {
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://raw.githubusercontent.com/faustbrian/go-library-tools/"+source+"/"+input.path, nil)
		if err != nil {
			t.Fatal("immutable scanner oracle request failed")
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatal("immutable scanner oracle unavailable")
		}
		data, readErr := io.ReadAll(io.LimitReader(response.Body, input.bytes+1))
		closeErr := response.Body.Close()
		if response.StatusCode != http.StatusOK || readErr != nil || closeErr != nil || int64(len(data)) != input.bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != input.digest {
			t.Fatal("immutable scanner oracle integrity differs")
		}
		files[input.path] = string(data)
	}
	return files
}

func TestIdentityEvidencePolicyPinnedScanner(t *testing.T) {
	if os.Getenv("GOLIB_GITLEAKS_INTEGRATION") != "1" {
		t.Skip("opt-in pinned scanner contract")
	}
	files := identityEvidenceScannerFixtures(t)
	credential := strings.Join([]string{"A1b2C3d4", "E5f6G7h8", "I9j0K1l2", "M3n4O5p6", "Q7r8S9t0", "U1v2W3x4", "Y5z6A7b8", "C9d0E1f2"}, "")
	token := strings.Join([]string{"sg", "p_", "b42f631d", "80ec943a", "9bd22715", "f044ad68", "3e759c12"}, "")
	check := func(t *testing.T, corpus map[string]string, rejectEveryPath bool) {
		t.Helper()
		root, policy := gitleaksRepository(t, corpus)
		for _, mode := range []string{"git", "dir"} {
			findings, _, err := runGitleaksIntegration(t, root, policy, mode)
			if !rejectEveryPath {
				if err != nil || len(findings) != 0 {
					t.Fatal("authenticated literal records are not scanner-clean")
				}
				continue
			}
			if err == nil {
				t.Fatal("different-value credentials were ignored")
			}
			for path := range corpus {
				generic, sourcegraph := false, false
				for _, finding := range findings {
					if finding.File == path {
						generic = generic || finding.RuleID == "generic-api-key"
						sourcegraph = sourcegraph || finding.RuleID == "sourcegraph-access-token"
					}
				}
				if !generic || !sourcegraph {
					t.Logf("safe detector boundary: mode=%s expectedPath=%s generic=%t sourcegraph=%t findings=%d", mode, path, generic, sourcegraph, len(findings))
					for index, finding := range findings {
						if index >= 16 {
							break
						}
						t.Logf("safe finding: file=%s rule=%s", finding.File, finding.RuleID)
					}
					t.Fatal("same-path credential detectors were weakened")
				}
			}
		}
	}
	t.Run("authenticated source records", func(t *testing.T) { check(t, files, false) })
	t.Run("adjacent different credentials in every allowed path", func(t *testing.T) {
		corpus := make(map[string]string, len(files))
		for path, records := range files {
			corpus[path] = records + "api_key=" + credential + "\n" + token + "\n"
		}
		check(t, corpus, true)
	})
	t.Run("new credential field and same-line extension in every allowed path", func(t *testing.T) {
		corpus := make(map[string]string, len(files))
		for path, records := range files {
			first, _, _ := strings.Cut(records, "\n")
			corpus[path] = first + " api_key=" + credential + "\n" + token + "\n"
		}
		check(t, corpus, true)
	})
	t.Run("identical records at different paths remain findings", func(t *testing.T) {
		corpus := make(map[string]string, len(files))
		for path, records := range files {
			corpus["other/"+path] = records
		}
		root, policy := gitleaksRepository(t, corpus)
		baseline := filepath.Join(t.TempDir(), "default-policy.toml")
		if err := os.WriteFile(baseline, []byte("[extend]\nuseDefault = true\n"), 0600); err != nil {
			t.Fatal("baseline policy unavailable")
		}
		for _, mode := range []string{"git", "dir"} {
			candidates, _, candidateErr := runGitleaksIntegration(t, root, policy, mode)
			defaults, _, defaultErr := runGitleaksIntegration(t, root, baseline, mode)
			if candidateErr == nil || defaultErr == nil || len(candidates) != len(defaults) {
				t.Fatal("different-path source records were improperly allowed")
			}
			for path := range corpus {
				found := false
				for _, f := range candidates {
					found = found || f.File == path
				}
				if !found {
					t.Logf("safe path boundary: mode=%s expectedPath=%s findings=%d", mode, path, len(candidates))
					for index, finding := range candidates {
						if index >= 16 {
							break
						}
						t.Logf("safe finding: file=%s rule=%s", finding.File, finding.RuleID)
					}
					t.Fatal("different-path detector partition disappeared")
				}
			}
		}
	})
}
