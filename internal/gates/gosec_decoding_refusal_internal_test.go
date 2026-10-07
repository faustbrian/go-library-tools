package gates

import "testing"

func TestGosecStatisticTypeErrorsAlwaysRefuseProjection(t *testing.T) {
	for _, field := range []string{"files", "lines", "nosec", "found"} {
		for _, value := range []struct {
			name string
			data any
		}{
			{"string", "ordinary text"},
			{"boolean", true},
			{"object", map[string]int{"ordinary": 1}},
			{"array", []int{1}},
		} {
			t.Run(field+"/"+value.name, func(t *testing.T) {
				root := t.TempDir()
				report := projectionReport(t, root, false)
				stats, ok := report["Stats"].(map[string]any)
				if !ok {
					t.Fatal("fixture statistics must be an object")
				}
				stats[field] = value.data
				if got := projectReport(t, report, root); got != "gosec-tool-or-report-failure" {
					t.Fatalf("wrong statistic type projected as %q", got)
				}
			})
		}
	}
}
