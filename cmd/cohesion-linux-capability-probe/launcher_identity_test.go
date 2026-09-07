package main

import (
	"fmt"
	"strconv"
	"testing"
)

const overflowID = 65534

func TestParseLauncherIDAllowsRootGroupForNonRootUser(t *testing.T) {
	t.Parallel()

	got, err := parseLauncherID("gid", "0")
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Fatalf("launcher gid=%d want=0", got)
	}
}

func TestValidateLauncherIDBoundaries(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		kind    string
		value   int
		wantErr bool
	}{
		{name: "ordinary user", kind: "uid", value: 1000},
		{name: "ordinary group", kind: "gid", value: 1000},
		{name: "root user", kind: "uid", value: 0, wantErr: true},
		{name: "root group", kind: "gid", value: 0},
		{name: "overflow user", kind: "uid", value: overflowID, wantErr: true},
		{name: "overflow group", kind: "gid", value: overflowID, wantErr: true},
		{name: "negative user", kind: "uid", value: -1, wantErr: true},
		{name: "negative group", kind: "gid", value: -1, wantErr: true},
		{name: "unknown identity kind", kind: "group", value: 1000, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateLauncherID(test.kind, test.value)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateLauncherID(%q, %d) error=%v wantErr=%t", test.kind, test.value, err, test.wantErr)
			}
		})
	}
}

func TestParseLauncherIDRejectsNonCanonicalValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "00", "+1", "-1", "65534", "not-an-id"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			if _, err := parseLauncherID("gid", value); err == nil {
				t.Fatalf("parseLauncherID(%q) unexpectedly succeeded", value)
			}
		})
	}
}

func parseLauncherID(name, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || strconv.Itoa(parsed) != value {
		return 0, fmt.Errorf("invalid launcher %s %q", name, value)
	}
	if err := validateLauncherID(name, parsed); err != nil {
		return 0, err
	}
	return parsed, nil
}

func validateLauncherID(name string, value int) error {
	if (name != "uid" && name != "gid") || value < 0 || value == overflowID || (name == "uid" && value == 0) {
		return fmt.Errorf("invalid launcher %s %q", name, strconv.Itoa(value))
	}
	return nil
}
