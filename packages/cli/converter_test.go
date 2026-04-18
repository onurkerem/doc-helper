package main

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func TestReplaceFencedCodeWithConfluenceMacro(t *testing.T) {
	gm := goldmark.New(goldmark.WithExtensions(extension.GFM))
	input := "```json\n{\"success\": true}\n```\n"
	var buf strings.Builder
	if err := gm.Convert([]byte(input), &buf); err != nil {
		t.Fatal(err)
	}
	htmlDoc := buf.String()
	out := replaceFencedCodeWithConfluenceMacro(htmlDoc)

	if strings.Contains(out, "<pre>") {
		t.Fatalf("expected <pre> removed, got: %s", out)
	}
	if !strings.Contains(out, `ac:name="code"`) {
		t.Fatalf("expected code macro: %s", out)
	}
	if !strings.Contains(out, `<ac:parameter ac:name="language">json</ac:parameter>`) {
		t.Fatalf("expected json language param: %s", out)
	}
	if !strings.Contains(out, `{"success": true}`) {
		t.Fatalf("expected code body preserved: %s", out)
	}
	if !strings.Contains(out, `<ac:plain-text-body><![CDATA[`) {
		t.Fatalf("expected CDATA body: %s", out)
	}
}

func TestConfluenceCodeMacroCDATAEnd(t *testing.T) {
	s := confluenceCodeMacro("text", "a]]>b")
	if !strings.Contains(s, "]]]]><![CDATA[>") {
		t.Fatalf("expected split CDATA for ]]> in body: %s", s)
	}
}

func TestLanguageFromCodeClass(t *testing.T) {
	if languageFromCodeClass(`language-json hljs`) != "json" {
		t.Fatalf("got %q", languageFromCodeClass(`language-json hljs`))
	}
	if languageFromCodeClass("") != "" {
		t.Fatal()
	}
}
