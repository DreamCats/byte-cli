package commands

import (
	"encoding/json"
	"testing"
)

func TestDecodeJSONUseNumberPreservesLargeIntegers(t *testing.T) {
	var payload any
	if err := decodeJSONUseNumber(`{"product_id":1234567890123456789}`, &payload); err != nil {
		t.Fatal(err)
	}
	raw := payload.(map[string]any)["product_id"]
	num, ok := raw.(json.Number)
	if !ok {
		t.Fatalf("got %T", raw)
	}
	if num.String() != "1234567890123456789" {
		t.Fatalf("got %q", num.String())
	}
}
