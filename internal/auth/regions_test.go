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
