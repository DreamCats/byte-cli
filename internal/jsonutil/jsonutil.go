package jsonutil

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode"
)

func DecodeNormalized(data []byte, dst any) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	normalized := normalize(raw)
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(encoded))
	return dec.Decode(dst)
}

func normalize(value any) any {
	switch v := value.(type) {
	case []any:
		out := make([]any, len(v))
		for i := range v {
			out[i] = normalize(v[i])
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[toSnake(key)] = normalize(item)
		}
		return out
	default:
		return value
	}
}

func toSnake(name string) string {
	if strings.Contains(name, "_") {
		return name
	}
	runes := []rune(name)
	var b strings.Builder
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
