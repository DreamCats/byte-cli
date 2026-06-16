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
	if got := codebase.LoginURL(); got != "" {
		t.Fatalf("codebase LoginURL() = %q, want empty", got)
	}
}
