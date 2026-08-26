package repository

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCheckLegacyReportsInspectionFailure(t *testing.T) {
	failure := errors.New("injected failure")
	err := checkLegacy("/repo", func(string) (os.FileInfo, error) { return nil, failure })
	if err == nil || !strings.Contains(err.Error(), "inspect legacy") {
		t.Fatalf("checkLegacy() error = %v", err)
	}
}
