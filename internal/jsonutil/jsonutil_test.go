package jsonutil

import "testing"

func TestDecodeNormalizedHandlesAcronyms(t *testing.T) {
	var dst struct {
		ID      int `json:"id"`
		IDLInfo struct {
			RepoName string `json:"repo_name"`
		} `json:"idl_info"`
	}
	if err := DecodeNormalized([]byte(`{"ID":7,"IDLInfo":{"RepoName":"g/p"}}`), &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 7 || dst.IDLInfo.RepoName != "g/p" {
		t.Fatalf("unexpected decoded value: %#v", dst)
	}
}
