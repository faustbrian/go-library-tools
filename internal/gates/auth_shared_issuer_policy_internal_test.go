package gates

import (
	"regexp"
	"strings"
	"testing"
)

// Public synthetic OpenSSL verification vectors from Authentication commit
// d9f5132c7f682a4c84d3cba3c0b07243dd841b3e. The generator discards its
// ephemeral signing key; fixture SHA-256 is 0d99421c2b3359327ef778fa5ca329a45ce9f971298c3fd99c7049b2f5982cde.
// Split token bytes so policy tests do not themselves resemble credentials.
func authenticationSharedIssuerPublicLines() []string {
	return []string{
		strings.Join([]string{"    \"valid\": \"eyJhbGciOi", "JFZERTQSIsInR5cCI6ImF0K2", "p3dCIsImtpZCI6ImtleSJ9.e", "yJpc3MiOiJodHRwczovL2lzc", "3Vlci5leGFtcGxlIiwiYXVkI", "joicmVzb3VyY2UiLCJzdWIiO", "iJvcmdhbml6YXRpb24iLCJjb", "GllbnRfaWQiOiJ3b3JrbG9hZ", "CIsImlhdCI6MTc5MDg1NjAwM", "CwiZXhwIjoxNzkwODU2MDYwL", "CJqdGkiOiJhYmNkZWZnaGlqa", "2xtbm9wcXJzdHV2IiwicHVyc", "G9zZSI6ImFjY2Vzcy10b2tlb", "iIsImVudmlyb25tZW50Ijoid", "GVzdCIsIm1ldGhvZCI6InBhc", "3N3b3JkIiwiYXNzdXJhbmNlI", "joyLCJvd25lciI6eyJraW5kI", "joib3JnYW5pemF0aW9uIiwia", "WQiOiJvcmdhbml6YXRpb24iL", "CJ2ZXJzaW9uIjozfSwiYWN0b", "3IiOnsia2luZCI6InVzZXIiL", "CJpZCI6Imh1bWFuIiwidmVyc", "2lvbiI6NH0sInRlbmFudCI6I", "nRlbmFudCIsInRlbmFudF92Z", "XJzaW9uIjo1LCJjbGllbnRfd", "mVyc2lvbiI6Niwic2lnbmluZ", "192ZXJzaW9uIjo3LCJtZW1iZ", "XJzaGlwX2lkIjoibWVtYmVyc", "2hpcCIsIm1lbWJlcnNoaXBfd", "mVyc2lvbiI6OCwicm9vdF9jc", "mVkZW50aWFsIjp7ImlkIjoiY", "3JlZGVudGlhbCIsInZlcnNpb", "24iOjEsImV4cGlyZXNfYXQiO", "iIwMDAxLTAxLTAxVDAwOjAwO", "jAwWiJ9LCJyb290X2dyYW50I", "jp7ImlkIjoicm9vdCIsInZlc", "nNpb24iOjIsImV4cGlyZXNfY", "XQiOiIyMDI2LTEwLTAxVDEyO", "jA1OjAwWiJ9LCJwYXJlbnRzI", "jpbeyJpZCI6InBhcmVudCIsI", "nZlcnNpb24iOjksImV4cGlyZ", "XNfYXQiOiIyMDI2LTEwLTAxV", "DEyOjAzOjAwWiJ9XSwic2Nvc", "GVzIjpbInJlYWQiXX0.Qkp_f", "aZF7mpspJ4dkieyR2tjTOpsi", "vOFwGYlsT4tg9pKSVxtVhxdo", "ZckPBFF9GOje9tbfWrxdnAkM", "t7gd9KtAg\","}, ""),
		strings.Join([]string{"    \"wrong_environment\":", " \"eyJhbGciOiJFZERTQSIsIn", "R5cCI6ImF0K2p3dCIsImtpZC", "I6ImtleSJ9.eyJpc3MiOiJod", "HRwczovL2lzc3Vlci5leGFtc", "GxlIiwiYXVkIjoicmVzb3VyY", "2UiLCJzdWIiOiJvcmdhbml6Y", "XRpb24iLCJjbGllbnRfaWQiO", "iJ3b3JrbG9hZCIsImlhdCI6M", "Tc5MDg1NjAwMCwiZXhwIjoxN", "zkwODU2MDYwLCJqdGkiOiJhY", "mNkZWZnaGlqa2xtbm9wcXJzd", "HV2IiwicHVycG9zZSI6ImFjY", "2Vzcy10b2tlbiIsImVudmlyb", "25tZW50Ijoib3RoZXIiLCJtZ", "XRob2QiOiJwYXNzd29yZCIsI", "mFzc3VyYW5jZSI6Miwib3duZ", "XIiOnsia2luZCI6Im9yZ2Fua", "XphdGlvbiIsImlkIjoib3JnY", "W5pemF0aW9uIiwidmVyc2lvb", "iI6M30sImFjdG9yIjp7Imtpb", "mQiOiJ1c2VyIiwiaWQiOiJod", "W1hbiIsInZlcnNpb24iOjR9L", "CJ0ZW5hbnQiOiJ0ZW5hbnQiL", "CJ0ZW5hbnRfdmVyc2lvbiI6N", "SwiY2xpZW50X3ZlcnNpb24iO", "jYsInNpZ25pbmdfdmVyc2lvb", "iI6NywibWVtYmVyc2hpcF9pZ", "CI6Im1lbWJlcnNoaXAiLCJtZ", "W1iZXJzaGlwX3ZlcnNpb24iO", "jgsInJvb3RfY3JlZGVudGlhb", "CI6eyJpZCI6ImNyZWRlbnRpY", "WwiLCJ2ZXJzaW9uIjoxLCJle", "HBpcmVzX2F0IjoiMDAwMS0wM", "S0wMVQwMDowMDowMFoifSwic", "m9vdF9ncmFudCI6eyJpZCI6I", "nJvb3QiLCJ2ZXJzaW9uIjoyL", "CJleHBpcmVzX2F0IjoiMjAyN", "i0xMC0wMVQxMjowNTowMFoif", "SwicGFyZW50cyI6W3siaWQiO", "iJwYXJlbnQiLCJ2ZXJzaW9uI", "jo5LCJleHBpcmVzX2F0IjoiM", "jAyNi0xMC0wMVQxMjowMzowM", "FoifV0sInNjb3BlcyI6WyJyZ", "WFkIl19.Eb8IXgmdfxZn7hFx", "hjUfvs2N7LDhvyyX0tVdi_B1", "b5PFoMw8kazD3MVD6gf2PWC5", "eKgZbWvfyIyDzFKzX4xhBw\","}, ""),
		strings.Join([]string{"    \"id_token\": \"eyJhbGc", "iOiJFZERTQSIsInR5cCI6ImF", "0K2p3dCIsImtpZCI6ImtleSJ", "9.eyJpc3MiOiJodHRwczovL2", "lzc3Vlci5leGFtcGxlIiwiYX", "VkIjoicmVzb3VyY2UiLCJzdW", "IiOiJvcmdhbml6YXRpb24iLC", "JjbGllbnRfaWQiOiJ3b3JrbG", "9hZCIsImlhdCI6MTc5MDg1Nj", "AwMCwiZXhwIjoxNzkwODU2MD", "YwLCJqdGkiOiJhYmNkZWZnaG", "lqa2xtbm9wcXJzdHV2IiwicH", "VycG9zZSI6ImlkLXRva2VuIi", "wiZW52aXJvbm1lbnQiOiJ0ZX", "N0IiwibWV0aG9kIjoicGFzc3", "dvcmQiLCJhc3N1cmFuY2UiOj", "IsIm93bmVyIjp7ImtpbmQiOi", "Jvcmdhbml6YXRpb24iLCJpZC", "I6Im9yZ2FuaXphdGlvbiIsIn", "ZlcnNpb24iOjN9LCJhY3Rvci", "I6eyJraW5kIjoidXNlciIsIm", "lkIjoiaHVtYW4iLCJ2ZXJzaW", "9uIjo0fSwidGVuYW50IjoidG", "VuYW50IiwidGVuYW50X3Zlcn", "Npb24iOjUsImNsaWVudF92ZX", "JzaW9uIjo2LCJzaWduaW5nX3", "ZlcnNpb24iOjcsIm1lbWJlcn", "NoaXBfaWQiOiJtZW1iZXJzaG", "lwIiwibWVtYmVyc2hpcF92ZX", "JzaW9uIjo4LCJyb290X2NyZW", "RlbnRpYWwiOnsiaWQiOiJjcm", "VkZW50aWFsIiwidmVyc2lvbi", "I6MSwiZXhwaXJlc19hdCI6Ij", "AwMDEtMDEtMDFUMDA6MDA6MD", "BaIn0sInJvb3RfZ3JhbnQiOn", "siaWQiOiJyb290IiwidmVyc2", "lvbiI6MiwiZXhwaXJlc19hdC", "I6IjIwMjYtMTAtMDFUMTI6MD", "U6MDBaIn0sInBhcmVudHMiOl", "t7ImlkIjoicGFyZW50Iiwidm", "Vyc2lvbiI6OSwiZXhwaXJlc1", "9hdCI6IjIwMjYtMTAtMDFUMT", "I6MDM6MDBaIn1dLCJzY29wZX", "MiOlsicmVhZCJdfQ.-LQHaPv", "Be_-VCtGLPV0YUFtlHHR1qOu", "bJNODdyjyW8LRGnJdY7Rrk0U", "X0Vlp6TFZp5r7TtwGfzB7qQM", "fusCrAg\","}, ""),
		strings.Join([]string{"    \"missing_root\": \"eyJ", "hbGciOiJFZERTQSIsInR5cCI", "6ImF0K2p3dCIsImtpZCI6Imt", "leSJ9.eyJpc3MiOiJodHRwcz", "ovL2lzc3Vlci5leGFtcGxlIi", "wiYXVkIjoicmVzb3VyY2UiLC", "JzdWIiOiJvcmdhbml6YXRpb2", "4iLCJjbGllbnRfaWQiOiJ3b3", "JrbG9hZCIsImlhdCI6MTc5MD", "g1NjAwMCwiZXhwIjoxNzkwOD", "U2MDYwLCJqdGkiOiJhYmNkZW", "ZnaGlqa2xtbm9wcXJzdHV2Ii", "wicHVycG9zZSI6ImFjY2Vzcy", "10b2tlbiIsImVudmlyb25tZW", "50IjoidGVzdCIsIm1ldGhvZC", "I6InBhc3N3b3JkIiwiYXNzdX", "JhbmNlIjoyLCJvd25lciI6ey", "JraW5kIjoib3JnYW5pemF0aW", "9uIiwiaWQiOiJvcmdhbml6YX", "Rpb24iLCJ2ZXJzaW9uIjozfS", "wiYWN0b3IiOnsia2luZCI6In", "VzZXIiLCJpZCI6Imh1bWFuIi", "widmVyc2lvbiI6NH0sInRlbm", "FudCI6InRlbmFudCIsInRlbm", "FudF92ZXJzaW9uIjo1LCJjbG", "llbnRfdmVyc2lvbiI6Niwic2", "lnbmluZ192ZXJzaW9uIjo3LC", "JtZW1iZXJzaGlwX2lkIjoibW", "VtYmVyc2hpcCIsIm1lbWJlcn", "NoaXBfdmVyc2lvbiI6OCwicm", "9vdF9jcmVkZW50aWFsIjp7Im", "lkIjoiY3JlZGVudGlhbCIsIn", "ZlcnNpb24iOjEsImV4cGlyZX", "NfYXQiOiIwMDAxLTAxLTAxVD", "AwOjAwOjAwWiJ9LCJyb290X2", "dyYW50Ijp7ImlkIjoiIiwidm", "Vyc2lvbiI6MiwiZXhwaXJlc1", "9hdCI6IjIwMjYtMTAtMDFUMT", "I6MDU6MDBaIn0sInBhcmVudH", "MiOlt7ImlkIjoicGFyZW50Ii", "widmVyc2lvbiI6OSwiZXhwaX", "Jlc19hdCI6IjIwMjYtMTAtMD", "FUMTI6MDM6MDBaIn1dLCJzY2", "9wZXMiOlsicmVhZCJdfQ.n57", "N4Ru_sLmTbVsrs8ynKswV3CO", "o05ByPImjeNasqfKYVeU5_0C", "ujvmuOAtmHUxZ5ZufwtyGV4z", "dTt2yEn-tDw\","}, ""),
		strings.Join([]string{"    \"cycle\": \"eyJhbGciOi", "JFZERTQSIsInR5cCI6ImF0K2", "p3dCIsImtpZCI6ImtleSJ9.e", "yJpc3MiOiJodHRwczovL2lzc", "3Vlci5leGFtcGxlIiwiYXVkI", "joicmVzb3VyY2UiLCJzdWIiO", "iJvcmdhbml6YXRpb24iLCJjb", "GllbnRfaWQiOiJ3b3JrbG9hZ", "CIsImlhdCI6MTc5MDg1NjAwM", "CwiZXhwIjoxNzkwODU2MDYwL", "CJqdGkiOiJhYmNkZWZnaGlqa", "2xtbm9wcXJzdHV2IiwicHVyc", "G9zZSI6ImFjY2Vzcy10b2tlb", "iIsImVudmlyb25tZW50Ijoid", "GVzdCIsIm1ldGhvZCI6InBhc", "3N3b3JkIiwiYXNzdXJhbmNlI", "joyLCJvd25lciI6eyJraW5kI", "joib3JnYW5pemF0aW9uIiwia", "WQiOiJvcmdhbml6YXRpb24iL", "CJ2ZXJzaW9uIjozfSwiYWN0b", "3IiOnsia2luZCI6InVzZXIiL", "CJpZCI6Imh1bWFuIiwidmVyc", "2lvbiI6NH0sInRlbmFudCI6I", "nRlbmFudCIsInRlbmFudF92Z", "XJzaW9uIjo1LCJjbGllbnRfd", "mVyc2lvbiI6Niwic2lnbmluZ", "192ZXJzaW9uIjo3LCJtZW1iZ", "XJzaGlwX2lkIjoibWVtYmVyc", "2hpcCIsIm1lbWJlcnNoaXBfd", "mVyc2lvbiI6OCwicm9vdF9jc", "mVkZW50aWFsIjp7ImlkIjoiY", "3JlZGVudGlhbCIsInZlcnNpb", "24iOjEsImV4cGlyZXNfYXQiO", "iIwMDAxLTAxLTAxVDAwOjAwO", "jAwWiJ9LCJyb290X2dyYW50I", "jp7ImlkIjoicm9vdCIsInZlc", "nNpb24iOjIsImV4cGlyZXNfY", "XQiOiIyMDI2LTEwLTAxVDEyO", "jA1OjAwWiJ9LCJwYXJlbnRzI", "jpbeyJpZCI6ImFiY2RlZmdoa", "WprbG1ub3BxcnN0dXYiLCJ2Z", "XJzaW9uIjo5LCJleHBpcmVzX", "2F0IjoiMjAyNi0xMC0wMVQxM", "jowMzowMFoifV0sInNjb3Blc", "yI6WyJyZWFkIl19.Vyna9ETf", "nk7fn_-pr9g9GEqY_Gu6t6De", "Qu3cfmH8iqacJEfAhBwNys7y", "lEudXxZ5C2UCL_Mgbmi4ykpF", "6AmQCA\","}, ""),
		strings.Join([]string{"    \"excessive_depth\": \"", "eyJhbGciOiJFZERTQSIsInR5", "cCI6ImF0K2p3dCIsImtpZCI6", "ImtleSJ9.eyJpc3MiOiJodHR", "wczovL2lzc3Vlci5leGFtcGx", "lIiwiYXVkIjoicmVzb3VyY2U", "iLCJzdWIiOiJvcmdhbml6YXR", "pb24iLCJjbGllbnRfaWQiOiJ", "3b3JrbG9hZCIsImlhdCI6MTc", "5MDg1NjAwMCwiZXhwIjoxNzk", "wODU2MDYwLCJqdGkiOiJhYmN", "kZWZnaGlqa2xtbm9wcXJzdHV", "2IiwicHVycG9zZSI6ImFjY2V", "zcy10b2tlbiIsImVudmlyb25", "tZW50IjoidGVzdCIsIm1ldGh", "vZCI6InBhc3N3b3JkIiwiYXN", "zdXJhbmNlIjoyLCJvd25lciI", "6eyJraW5kIjoib3JnYW5pemF", "0aW9uIiwiaWQiOiJvcmdhbml", "6YXRpb24iLCJ2ZXJzaW9uIjo", "zfSwiYWN0b3IiOnsia2luZCI", "6InVzZXIiLCJpZCI6Imh1bWF", "uIiwidmVyc2lvbiI6NH0sInR", "lbmFudCI6InRlbmFudCIsInR", "lbmFudF92ZXJzaW9uIjo1LCJ", "jbGllbnRfdmVyc2lvbiI6Niw", "ic2lnbmluZ192ZXJzaW9uIjo", "3LCJtZW1iZXJzaGlwX2lkIjo", "ibWVtYmVyc2hpcCIsIm1lbWJ", "lcnNoaXBfdmVyc2lvbiI6OCw", "icm9vdF9jcmVkZW50aWFsIjp", "7ImlkIjoiY3JlZGVudGlhbCI", "sInZlcnNpb24iOjEsImV4cGl", "yZXNfYXQiOiIwMDAxLTAxLTA", "xVDAwOjAwOjAwWiJ9LCJyb29", "0X2dyYW50Ijp7ImlkIjoicm9", "vdCIsInZlcnNpb24iOjIsImV", "4cGlyZXNfYXQiOiIyMDI2LTE", "wLTAxVDEyOjA1OjAwWiJ9LCJ", "wYXJlbnRzIjpbeyJpZCI6InB", "hcmVudC0wIiwidmVyc2lvbiI", "6MSwiZXhwaXJlc19hdCI6IjI", "wMjYtMTAtMDFUMTI6MDM6MDB", "aIn0seyJpZCI6InBhcmVudC0", "xIiwidmVyc2lvbiI6MSwiZXh", "waXJlc19hdCI6IjIwMjYtMTA", "tMDFUMTI6MDM6MDBaIn0seyJ", "pZCI6InBhcmVudC0yIiwidmV", "yc2lvbiI6MSwiZXhwaXJlc19", "hdCI6IjIwMjYtMTAtMDFUMTI", "6MDM6MDBaIn0seyJpZCI6InB", "hcmVudC0zIiwidmVyc2lvbiI", "6MSwiZXhwaXJlc19hdCI6IjI", "wMjYtMTAtMDFUMTI6MDM6MDB", "aIn0seyJpZCI6InBhcmVudC0", "0IiwidmVyc2lvbiI6MSwiZXh", "waXJlc19hdCI6IjIwMjYtMTA", "tMDFUMTI6MDM6MDBaIn1dLCJ", "zY29wZXMiOlsicmVhZCJdfQ.", "y9TrBrjDrQCbKQpbrY2_Otfe", "dsf1wvIwoHnbKwr4oVK82SlP", "jPzeCBz6HqBPOBw2PMDQVc_0", "uM3DKuziyDinDg\""}, ""),
	}
}

func TestAuthenticationSharedIssuerPolicyRequiresExactTuple(t *testing.T) {
	const path = "testdata/shared-issuer/openssl-eddsa.json"
	for _, line := range authenticationSharedIssuerPublicLines() {
		for _, test := range []struct {
			name, path, rule, line string
			allowed                bool
		}{
			{"exact", path, "jwt", line, true},
			{"scanner framing", path, "jwt", "\n" + line, true},
			{"different path", "other/" + path, "jwt", line, false},
			{"different detector", path, "generic-api-key", line, false},
			{"changed token", path, "jwt", strings.Replace(line, ".", ".A", 1), false},
			{"additional assignment", path, "jwt", line + " api_key=unrelated", false},
			{"additional line", path, "jwt", line + "\napi_key=unrelated", false},
		} {
			t.Run(test.name, func(t *testing.T) {
				allowed := false
				for _, block := range strings.Split(gitleaksPolicy, "[[allowlists]]") {
					if !strings.Contains(block, `description = "Exact synthetic shared-issuer`) {
						continue
					}
					if !strings.Contains(block, `condition = "AND"`) || !strings.Contains(block, `targetRules = ["jwt"]`) || !strings.Contains(block, `regexTarget = "line"`) {
						t.Fatal("fixture policy lost rule/path/line conjunction")
					}
					match := func(key, value string) bool {
						a := regexp.MustCompile(`(?s)` + key + ` = \[(.*?)\]\n`).FindStringSubmatch(block)
						if len(a) != 2 {
							t.Fatal("missing expression array")
						}
						for _, p := range regexp.MustCompile(`'''(.*?)'''`).FindAllStringSubmatch(a[1], -1) {
							if regexp.MustCompile(p[1]).MatchString(value) {
								return true
							}
						}
						return false
					}
					allowed = allowed || test.rule == "jwt" && match("paths", test.path) && match("regexes", test.line)
				}
				if allowed != test.allowed {
					t.Fatalf("fixture allowance = %v, want %v", allowed, test.allowed)
				}
			})
		}
	}
}
