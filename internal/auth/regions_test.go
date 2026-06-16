package auth

import "testing"

func TestParseRegionCodebase(t *testing.T) {
	region, err := ParseRegion("CODEBASE")
	if err != nil {
		t.Fatal(err)
	}
	if !region.IsCodebase || region.CookieName != "CAS_SESSION_API" {
		t.Fatalf("unexpected region: %#v", region)
	}
}

func TestParseRegionRejectsUnknown(t *testing.T) {
	if _, err := ParseRegion("sg"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRegionLoginURL(t *testing.T) {
	region, err := ParseRegion("cn")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := region.LoginURL(), "https://cloud.bytedance.net/auth/api/v1/login"; got != want {
		t.Fatalf("LoginURL() = %q, want %q", got, want)
	}
	codebase, err := ParseRegion("codebase")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := codebase.LoginURL(), "https://bits.bytedance.net/api/v1/identity/login?next=https%3A%2F%2Fbits.bytedance.net%2Fworkbench"; got != want {
		t.Fatalf("codebase LoginURL() = %q, want %q", got, want)
	}
}
