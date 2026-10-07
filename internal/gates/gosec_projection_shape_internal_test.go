package gates

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestGosecProjectionExactReportSize(t *testing.T) {
	root := t.TempDir()
	data := ordinaryGosecReport(root, false, true)
	// Trailing JSON whitespace preserves the same valid report at the limit.
	data = append(data, []byte(strings.Repeat(" ", (4<<20)-len(data)))...)
	if got := gosecFailureMetadata(data, root); got != "gosec-loading-failure findings=0 loading_packages=1 loading_errors=1" {
		t.Fatalf("exact-size projection = %q", got)
	}
	if got := gosecFailureMetadata(append(data, ' '), root); got != "gosec-tool-or-report-failure" {
		t.Fatalf("oversize projection = %q", got)
	}
}

func TestGosecProjectionUnambiguousShape(t *testing.T) {
	root := t.TempDir()
	valid := string(ordinaryGosecReport(root, false, true))
	for name, data := range map[string]string{
		"identical duplicate": strings.Replace(valid, `"found":0`, `"found":0,"found":0`, 1),
		"case duplicate":      strings.Replace(valid, `"found":0`, `"found":0,"Found":0`, 1),
		"statistics type":     strings.Replace(valid, `"lines":20`, `"lines":"private"`, 1),
		"issues type":         strings.Replace(valid, `"Issues":[]`, `"Issues":{}`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if data == valid {
				t.Fatal("fixture did not change the intended field")
			}
			if got := gosecFailureMetadata([]byte(data), root); got != "gosec-tool-or-report-failure" {
				t.Fatalf("ambiguous or invalid shape projection = %q", got)
			}
		})
	}
}

func TestGosecProjectionIssueCountBoundaries(t *testing.T) {
	for _, count := range []int{4096, 4097} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			root := t.TempDir()
			report, _ := ordinaryGosecFindings(root, count)
			data, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			want := "gosec-tool-or-report-failure"
			if count == 4096 {
				var projection strings.Builder
				fmt.Fprintf(&projection, "gosec-findings findings=4096 loading_packages=0 loading_errors=0")
				for index := range 64 {
					fmt.Fprintf(&projection, "\ngosec-location %x G115 %d 4", sha256.Sum256([]byte(fmt.Sprintf("resource-%02d.go", index))), index+1)
				}
				want = projection.String()
			}
			if got := gosecFailureMetadata(data, root); got != want {
				t.Fatal("issue-count projection must preserve complete counts, bounded locations and private text exclusion")
			}
		})
	}
}

func TestGosecProjectionIssueRequiredState(t *testing.T) {
	for _, field := range []string{"nosec", "file"} {
		t.Run(field, func(t *testing.T) {
			root := t.TempDir()
			report, issues := ordinaryGosecFindings(root, 1)
			if field == "nosec" {
				delete(issues[0], field)
			} else {
				issues[0][field] = "."
			}
			if got := projectReport(t, report, root); got != "gosec-tool-or-report-failure" {
				t.Fatalf("incomplete or module-root issue projection = %q", got)
			}
		})
	}
}
