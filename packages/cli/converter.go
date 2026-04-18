package main

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
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
	return mc.convert(markdown, false)
}

// ConvertForConfluence renders markdown to Confluence storage HTML. Fenced code
// blocks become the native Code Block macro (code snippet UI) instead of plain
// <pre><code>, so Confluence applies syntax highlighting and the copy control.
func (mc *MarkdownConverter) ConvertForConfluence(markdown string) (string, error) {
	return mc.convert(markdown, true)
}

func (mc *MarkdownConverter) convert(markdown string, confluenceCodeMacros bool) (string, error) {
	content := removeFirstH1(markdown)

	var buf bytes.Buffer
	if err := mc.gm.Convert([]byte(content), &buf); err != nil {
		return "", fmt.Errorf("converting markdown: %w", err)
	}

	out := buf.String()
	if confluenceCodeMacros {
		out = replaceFencedCodeWithConfluenceMacro(out)
	}
	return "<div class=\"content-wrapper\">\n" + out + "</div>", nil
}

// Goldmark fenced blocks render as <pre><code class="language-foo">...</code></pre>.
// Confluence expects the structured "code" macro for the editor's code snippets.
var goldmarkFencedCodeRE = regexp.MustCompile(`(?s)<pre>\s*<code(?:\s+class="([^"]*)")?\s*>(.*?)</code>\s*</pre>`)

func replaceFencedCodeWithConfluenceMacro(htmlDoc string) string {
	idx := goldmarkFencedCodeRE.FindAllStringSubmatchIndex(htmlDoc, -1)
	if len(idx) == 0 {
		return htmlDoc
	}

	var b strings.Builder
	last := 0
	for _, loc := range idx {
		whole0, whole1 := loc[0], loc[1]
		lang0, lang1 := loc[2], loc[3]
		code0, code1 := loc[4], loc[5]

		b.WriteString(htmlDoc[last:whole0])

		lang := ""
		if lang0 >= 0 && lang1 > lang0 {
			lang = languageFromCodeClass(htmlDoc[lang0:lang1])
		}
		rawEscaped := htmlDoc[code0:code1]
		code := html.UnescapeString(rawEscaped)
		b.WriteString(confluenceCodeMacro(lang, code))

		last = whole1
	}
	b.WriteString(htmlDoc[last:])
	return b.String()
}

func confluenceCodeMacro(language, code string) string {
	code = escapeCDATAEnd(code)

	var sb strings.Builder
	sb.WriteString(`<ac:structured-macro ac:name="code" ac:schema-version="1">`)
	if language != "" {
		sb.WriteString(`<ac:parameter ac:name="language">`)
		sb.WriteString(xmlEscapeText(language))
		sb.WriteString(`</ac:parameter>`)
	}
	sb.WriteString(`<ac:parameter ac:name="theme">Confluence</ac:parameter>`)
	sb.WriteString(`<ac:plain-text-body><![CDATA[`)
	sb.WriteString(code)
	sb.WriteString(`]]></ac:plain-text-body>`)
	sb.WriteString(`</ac:structured-macro>`)
	return sb.String()
}

func languageFromCodeClass(class string) string {
	for _, tok := range strings.Fields(class) {
		if strings.HasPrefix(tok, "language-") {
			return strings.TrimPrefix(tok, "language-")
		}
	}
	return ""
}

func escapeCDATAEnd(s string) string {
	return strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>")
}

func xmlEscapeText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
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
