package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DreamCats/byte-cli/internal/config"
)

func TestAuthConfigShowMasksCookiesByDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.Save(config.AppConfig{
		Regions: map[string]config.RegionConfig{
			"cn": {Cookie: "abcd1234wxyz"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code, err := Auth([]string{"config", "show"}, Output{Out: &out, Err: &errOut})
	if err != nil || code != 0 {
		t.Fatalf("Auth returned code=%d err=%v stderr=%q", code, err, errOut.String())
	}
	text := out.String()
	if strings.Contains(text, "abcd1234wxyz") {
		t.Fatalf("default show leaked cookie: %q", text)
	}
	if !strings.Contains(text, "cn: abcd****wxyz") {
		t.Fatalf("default show did not mask cookie: %q", text)
	}
}

func TestAuthConfigShowSecretPrintsFullCookies(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.Save(config.AppConfig{
		Regions: map[string]config.RegionConfig{
			"cn": {Cookie: "abcd1234wxyz"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code, err := Auth([]string{"config", "show", "--show-secret"}, Output{Out: &out, Err: &errOut})
	if err != nil || code != 0 {
		t.Fatalf("Auth returned code=%d err=%v stderr=%q", code, err, errOut.String())
	}
	text := out.String()
	if !strings.Contains(text, "cn: abcd1234wxyz") {
		t.Fatalf("show --show-secret did not print full cookie: %q", text)
	}
	if strings.Contains(text, "abcd****wxyz") {
		t.Fatalf("show --show-secret printed masked cookie: %q", text)
	}
}
