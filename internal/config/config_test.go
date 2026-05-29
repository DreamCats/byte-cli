package config

import "testing"

func TestMaskCookie(t *testing.T) {
	if got := MaskCookie("abcd1234wxyz"); got != "abcd****wxyz" {
		t.Fatalf("got %q", got)
	}
	if got := MaskCookie("short"); got != "*****" {
		t.Fatalf("got %q", got)
	}
}

func TestLoadSaveConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := AppConfig{Regions: map[string]RegionConfig{"cn": {Cookie: "cookie"}}, Proxy: ProxyConfig{HTTPS: "http://proxy"}}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Regions["cn"].Cookie != "cookie" || got.Proxy.HTTPS != "http://proxy" {
		t.Fatalf("unexpected config: %#v", got)
	}
}
