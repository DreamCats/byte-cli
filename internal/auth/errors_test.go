package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchNormalTokenHTTPErrorHasActionableMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("x-tt-logid", "log-1")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":-1,"error":"invalid token"}`))
	}))
	defer server.Close()

	region := Region{Value: "us", AuthURL: server.URL, CookieName: "CAS_SESSION"}
	_, err := NewManager(region).fetchNormalToken("cookie")
	if err == nil {
		t.Fatal("expected error")
	}
	text := err.Error()
	for _, want := range []string{
		"区域 us 获取 Token 失败",
		"HTTP 401",
		"logid=log-1",
		"invalid token",
		"byte-cli auth login -r us",
		"byte-cli auth config show --show-secret",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in %q", want, text)
		}
	}
}
