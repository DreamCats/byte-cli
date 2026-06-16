package auth

import (
	"runtime"
	"testing"
	"time"
)

func TestCDPCookieMatches(t *testing.T) {
	now := time.Unix(100, 0)
	cookie := cdpCookie{Name: "CAS_SESSION", Value: "token", Domain: ".bytedance.net", Expires: 200}
	if !cdpCookieMatches(cookie, "cloud.bytedance.net", "CAS_SESSION", now) {
		t.Fatal("expected cookie to match parent domain")
	}
}

func TestCDPCookieMatchesRejectsExpired(t *testing.T) {
	now := time.Unix(100, 0)
	cookie := cdpCookie{Name: "CAS_SESSION", Value: "token", Domain: "cloud.bytedance.net", Expires: 99}
	if cdpCookieMatches(cookie, "cloud.bytedance.net", "CAS_SESSION", now) {
		t.Fatal("expected expired cookie to be rejected")
	}
}

func TestAuthHost(t *testing.T) {
	region, err := ParseRegion("us")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := authHost(region), "cloud-ttp-us.bytedance.net"; got != want {
		t.Fatalf("authHost() = %q, want %q", got, want)
	}
}

func TestBrowserAppPath(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS app path parsing")
	}
	got := browserAppPath("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	if want := "/Applications/Google Chrome.app"; got != want {
		t.Fatalf("browserAppPath() = %q, want %q", got, want)
	}
}
