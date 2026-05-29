package commands

import (
	"strings"
	"testing"
)

func TestNormalizeFlagsAllowsOptionsAfterPositionals(t *testing.T) {
	got := normalizeFlags(
		[]string{"1", "-R", "ttec/industry_solution", "--unresolved"},
		stringSet("R", "repo"),
		stringSet("unresolved"),
	)
	want := []string{"-R", "ttec/industry_solution", "--unresolved", "1"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestNormalizeFlagsSupportsEquals(t *testing.T) {
	got := normalizeFlags([]string{"server", "tool", "--region=us", "-a", "x=1"}, stringSet("region", "a"), nil)
	want := []string{"--region=us", "-a", "x=1", "server", "tool"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("got %#v want %#v", got, want)
	}
}
