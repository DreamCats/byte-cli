package commands

import "testing"

func TestMessageSanitizer(t *testing.T) {
	got := newMessageSanitizer().sanitize(`hello   "LogID":"abc" {{rip=1}} world`)
	if got != `hello world` {
		t.Fatalf("got %q", got)
	}
}

func TestExtractEntriesFiltersAndTruncates(t *testing.T) {
	item := LogItem{
		Group: LogGroup{PSM: "a.b", VRegion: "US"},
		Value: []LogValue{{
			Level:  "ERROR",
			KVList: []KVPair{{Key: "_msg", Value: "prefix keyword suffix"}, {Key: "_location", Value: "file.go:1"}},
		}},
	}
	entries := extractEntries(item, newMessageSanitizer(), []string{"keyword"}, []string{"error"}, 10)
	if len(entries) != 1 {
		t.Fatalf("entries=%#v", entries)
	}
	if entries[0].Msg != "prefix key..." || entries[0].Location != "file.go:1" {
		t.Fatalf("unexpected entry: %#v", entries[0])
	}
}
