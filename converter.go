package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type MarkdownConverter struct {
	gm goldmark.Markdown
}

func NewMarkdownConverter() *MarkdownConverter {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)
	return &MarkdownConverter{gm: gm}
}

func (mc *MarkdownConverter) Convert(markdown string) (string, error) {
	content := removeFirstH1(markdown)

	var buf bytes.Buffer
	if err := mc.gm.Convert([]byte(content), &buf); err != nil {
		return "", fmt.Errorf("converting markdown: %w", err)
	}

	html := buf.String()
	return "<div class=\"content-wrapper\">\n" + html + "</div>", nil
}

func (mc *MarkdownConverter) ExtractTitle(markdown string) string {
	for line := range strings.SplitSeq(markdown, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}

func removeFirstH1(content string) string {
	found := false
	var lines []string
	for line := range strings.SplitSeq(content, "\n") {
		if !found && strings.HasPrefix(line, "# ") {
			found = true
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
