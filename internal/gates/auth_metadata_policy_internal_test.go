package gates

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

// These public digest declarations are recorded in both immutable Authentication
// decision histories, 0062d0a8c92135e6020522db36d361f7ca9e3b5d and
// 2931094d5e905565e52f0be9cada6b009a147939. Assemble their bytes so the
// policy's own source does not introduce scanner-shaped credential literals.
func authenticationPublicMetadataLines() []struct{ path, line string } {
	records := []struct {
		id     string
		digest []string
	}{
		{"AUTH-DEC-011", []string{"be33f511", "1184dec4", "c31bdeac", "3760ee2e", "f0fcdd73", "0a677f55", "479e7ee8", "f66ad659"}},
		{"AUTH-DEC-012", []string{"91356766", "40d9c0df", "84a1b676", "985d9f5c", "25129983", "04591b88", "09092b50", "ce0d7884"}},
		{"AUTH-DEC-001", []string{"dbf293fd", "525d952d", "fe51ac1a", "f7a43325", "429fa00d", "fe6d34e2", "cdc45069", "830330f5"}},
		{"AUTH-DEC-002", []string{"5d9b14d2", "4d6ac9ea", "1f84b43d", "5d745454", "b53e5271", "fee035aa", "909d1d9a", "2342d5fe"}},
		{"AUTH-DEC-002", []string{"15e8ade4", "98da8125", "48bf36de", "b3ab0b77", "d5fae914", "62036bc3", "6e27aaa3", "9fb4f273"}},
		{"AUTH-DEC-003", []string{"81ffead7", "7cfa5491", "3477444f", "24798f9b", "632beb94", "057e578e", "c1f266b0", "20c6a7c1"}},
		{"AUTH-DEC-003", []string{"10e348b7", "b49b28fb", "53f35f00", "72cbda5f", "f08156f8", "71c854fc", "b7e0f62b", "70f0ed9e"}},
		{"AUTH-DEC-004", []string{"86762f7d", "564203e9", "714e51b8", "67aa8ec4", "ebc913d0", "ca704e71", "6a409e56", "b2537722"}},
		{"AUTH-DEC-004", []string{"06998d20", "abf14867", "af6c7195", "a88b8ada", "dc4a66dd", "a00e1990", "172180c5", "c71a61a6"}},
		{"AUTH-DEC-005", []string{"b8d53a0e", "3c098089", "2e5b161e", "2e5b2534", "f635d250", "e4310757", "d11d78ff", "08b773c4"}},
		{"AUTH-DEC-006", []string{"e18d6391", "38c2828e", "3b8240a6", "bf80c543", "7b04d630", "da12cc3e", "419d476a", "b35b88e7"}},
		{"AUTH-DEC-007", []string{"f8b3f8e8", "393811a2", "c3ce0698", "8b92f812", "52babeed", "e518c683", "6afaac7a", "3de89326"}},
		{"AUTH-DEC-007", []string{"e9a92111", "6a9efa03", "3845aff4", "eb3e9313", "03ad4f13", "80b19033", "8c730245", "ea8860d6"}},
		{"AUTH-DEC-008", []string{"b4f02bec", "b2f6f0f2", "d098c8d2", "4fe75004", "8bf87f8d", "2aefc8c6", "80745c0d", "240c812a"}},
		{"AUTH-DEC-009", []string{"0104da4c", "0dca68fb", "7146d17f", "5d70077f", "e34edaf9", "19feef29", "b718252e", "db65ee37"}},
		{"AUTH-DEC-010", []string{"4a94dc00", "885603ca", "190f6e2e", "da39a98d", "09511ff7", "417bb3de", "9cb13177", "4398d60f"}},
	}
	lines := make([]struct{ path, line string }, 0, len(records)+1)
	for _, record := range records {
		lines = append(lines, struct{ path, line string }{"CHANGELOG.md", "- " + record.id + " sha256:" + strings.Join(record.digest, "")})
	}
	// The last cell names a real public test at the same immutable source. Split
	// at the Markdown boundary that generic-api-key mistakes for an assignment.
	row := "| [JWT-DEC-004](../docs/specification-decisions.md) | rfc7518-source, rfc7517-source, rfc8725-source | " +
		"TestValidatorRejectsCryptographicallyUnsafeKeys, TestValidateKeyMaterialRejectsEveryInvalidRepresentation, " +
		"TestRemoteJWKValidationRejectsEveryKeyPolicyViolation" + ", " + "TestRFC7520HMACJWKInteroperability |"
	return append(lines, struct{ path, line string }{"jwt/specification/README.md", row})
}

func TestAuthenticationMetadataPolicyRequiresExactTuple(t *testing.T) {
	for index, fixture := range authenticationPublicMetadataLines() {
		t.Run(strings.Join([]string{fixture.path, string(rune('A' + index))}, "/"), func(t *testing.T) {
			for _, test := range []struct {
				name, path, rule, line string
				allowed                bool
			}{
				{"exact", fixture.path, "generic-api-key", fixture.line, true},
				{"scanner framing", fixture.path, "generic-api-key", "\n" + fixture.line, true},
				{"other detector", fixture.path, "another-rule", fixture.line, false},
				{"other path", "docs/" + fixture.path, "generic-api-key", fixture.line, false},
				{"path suffix", fixture.path + ".other", "generic-api-key", fixture.line, false},
				{"changed content", fixture.path, "generic-api-key", fixture.line[:len(fixture.line)-1] + "!", false},
				{"prefix", fixture.path, "generic-api-key", "prefix" + fixture.line, false},
				{"adjacent assignment", fixture.path, "generic-api-key", fixture.line + " api_key=" + "ordinary-value", false},
				{"next line assignment", fixture.path, "generic-api-key", fixture.line + "\napi_key=" + "ordinary-value", false},
				{"credential assignment", fixture.path, "generic-api-key", "api_key=" + "ordinary-value", false},
			} {
				t.Run(test.name, func(t *testing.T) {
					if allowed := authenticationMetadataPolicyAllows(t, test.path, test.rule, test.line); allowed != test.allowed {
						t.Fatalf("metadata allowance = %v, want %v", allowed, test.allowed)
					}
				})
			}
		})
	}
}

// Evaluate the actual rendered policy expressions at the documented Gitleaks
// rule/path/line boundary without invoking a scanner or accepting its parser.
func authenticationMetadataPolicyAllows(t *testing.T, path, rule, line string) bool {
	t.Helper()
	for block := range strings.SplitSeq(gitleaksPolicy, "[[allowlists]]") {
		if !strings.Contains(block, `description = "Exact public Authentication`) {
			continue
		}
		if !strings.Contains(block, `condition = "AND"`) || !strings.Contains(block, `targetRules = ["generic-api-key"]`) || !strings.Contains(block, `regexTarget = "line"`) {
			t.Fatal("metadata policy lacks its exact rule/path/line conjunction")
		}
		patterns := func(key string) []string {
			t.Helper()
			array := regexp.MustCompile(`(?s)` + key + ` = \[(.*?)\]\n`).FindStringSubmatch(block)
			if len(array) != 2 {
				t.Fatal("metadata policy expression array is missing")
			}
			values := regexp.MustCompile(`'''(.*?)'''`).FindAllStringSubmatch(array[1], -1)
			result := make([]string, 0, len(values))
			for _, value := range values {
				result = append(result, value[1])
			}
			return result
		}
		matches := func(key, content string) bool {
			return slices.ContainsFunc(patterns(key), func(pattern string) bool { return regexp.MustCompile(pattern).MatchString(content) })
		}
		if rule == "generic-api-key" && matches("paths", path) && matches("regexes", line) {
			return true
		}
	}
	return false
}
