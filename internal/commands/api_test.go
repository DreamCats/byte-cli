package commands

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadJSONHTTPErrorIncludesLogIDAndBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Header: http.Header{
			"X-Tt-Logid": []string{"log-1"},
		},
		Body: io.NopCloser(strings.NewReader(`{"error":"invalid jwt-token"}`)),
	}

	err := readJSON(resp, &struct{}{})
	if err == nil {
		t.Fatal("expected error")
	}
	text := err.Error()
	for _, want := range []string{"HTTP 401", "logid=log-1", "invalid jwt-token"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in %q", want, text)
		}
	}
}
