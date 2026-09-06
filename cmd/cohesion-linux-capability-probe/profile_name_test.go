package main

import (
	"strings"
	"testing"
)

func TestValidAppArmorProfileName(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		value string
		want  bool
	}{
		{name: "amd64", value: "golib-schema-v3-1-1-amd64", want: true},
		{name: "arm64", value: "golib-schema-v3-1-1-arm64", want: true},
		{name: "missing suffix", value: "golib-schema-v3-"},
		{name: "wrong prefix", value: "schema-v3-1-1-amd64"},
		{name: "path separator", value: "golib-schema-v3-1/1-amd64"},
		{name: "uppercase", value: "golib-schema-v3-1-1-AMD64"},
		{name: "too long", value: "golib-schema-v3-" + strings.Repeat("a", 112)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := validAppArmorProfileName(test.value); got != test.want {
				t.Fatalf("validAppArmorProfileName(%q)=%t want=%t", test.value, got, test.want)
			}
		})
	}
}

func validAppArmorProfileName(name string) bool {
	const prefix = "golib-schema-v3-"
	if !strings.HasPrefix(name, prefix) || len(name) == len(prefix) || len(name) >= 128 {
		return false
	}
	for _, character := range name {
		if character != '-' && (character < '0' || character > '9') && (character < 'a' || character > 'z') {
			return false
		}
	}
	return true
}
