package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseRepoName(t *testing.T) {
	cases := map[string]string{
		"git@code.byted.org:group/project.git":             "group/project",
		"ssh://git@code.byted.org:29418/group/project.git": "group/project",
		"https://code.byted.org/group/project.git":         "group/project",
	}
	for input, want := range cases {
		if got := ParseRepoName(input); got != want {
			t.Fatalf("ParseRepoName(%q)=%q want %q", input, got, want)
		}
	}
}

func TestExtractMarkdownField(t *testing.T) {
	body := "**问题描述**: something broke\n**优先级**: P1"
	if got := extractMarkdownField(body, "**问题描述**:"); got != "something broke" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintAimeComment(t *testing.T) {
	var buf bytes.Buffer
	out := Output{Out: &buf}
	printAimeComment(Comment{Content: "**问题描述**: bug\n**优先级**: P2\n**问题分类**: logic"}, "  [aime] now", out)
	text := buf.String()
	for _, want := range []string{"问题: bug", "优先级: P2", "分类: logic"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in %q", want, text)
		}
	}
}
