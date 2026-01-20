// Package formatter provides HTML and text processing utilities.
package formatter

import (
	"regexp"
	"strings"
)

// HTMLNormalizer provides HTML normalization for EPUB compatibility
type HTMLNormalizer struct {
	PreserveImages bool
	PreserveLinks  bool
	AllowedTags    map[string]bool
}

// NewHTMLNormalizer creates a normalizer with default settings
func NewHTMLNormalizer() *HTMLNormalizer {
	return &HTMLNormalizer{
		PreserveImages: true,
		PreserveLinks:  true,
		AllowedTags: map[string]bool{
			"p": true, "br": true, "hr": true,
			"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
			"div": true, "span": true,
			"strong": true, "b": true, "em": true, "i": true, "u": true, "s": true,
			"blockquote": true, "pre": true, "code": true,
			"ul": true, "ol": true, "li": true,
			"table": true, "thead": true, "tbody": true, "tr": true, "th": true, "td": true,
			"a": true, "img": true,
			"figure": true, "figcaption": true,
			"sup": true, "sub": true,
		},
	}
}

// Normalize normalizes HTML for EPUB compatibility
func (n *HTMLNormalizer) Normalize(html string) string {
	html = removeTag(html, "script")
	html = removeTag(html, "style")
	html = removeTag(html, "noscript")
	html = removeHTMLComments(html)
	html = removeEventHandlers(html)
	html = removeAttribute(html, "style")
	html = removeAttribute(html, "class")
	html = removeAttribute(html, "id")
	html = removeDataAttributes(html)
	html = fixSelfClosingTags(html)
	html = removeEmptyParagraphs(html)
	html = collapseWhitespace(html)

	if !n.PreserveImages {
		html = removeTag(html, "img")
	}

	if !n.PreserveLinks {
		html = stripTagKeepContent(html, "a")
	}

	return strings.TrimSpace(html)
}

func removeTag(html, tag string) string {
	re := regexp.MustCompile(`(?is)<` + tag + `[^>]*>.*?</` + tag + `>`)
	html = re.ReplaceAllString(html, "")
	re = regexp.MustCompile(`(?i)<` + tag + `[^>]*/?>`)
	return re.ReplaceAllString(html, "")
}

func stripTagKeepContent(html, tag string) string {
	re := regexp.MustCompile(`(?i)<` + tag + `[^>]*>`)
	html = re.ReplaceAllString(html, "")
	re = regexp.MustCompile(`(?i)</` + tag + `>`)
	return re.ReplaceAllString(html, "")
}

func removeHTMLComments(html string) string {
	re := regexp.MustCompile(`<!--[\s\S]*?-->`)
	return re.ReplaceAllString(html, "")
}

func removeEventHandlers(html string) string {
	re := regexp.MustCompile(`\s+on\w+="[^"]*"`)
	html = re.ReplaceAllString(html, "")
	re = regexp.MustCompile(`\s+on\w+='[^']*'`)
	return re.ReplaceAllString(html, "")
}

func removeAttribute(html, attr string) string {
	re := regexp.MustCompile(`\s+` + attr + `="[^"]*"`)
	html = re.ReplaceAllString(html, "")
	re = regexp.MustCompile(`\s+` + attr + `='[^']*'`)
	return re.ReplaceAllString(html, "")
}

func removeDataAttributes(html string) string {
	re := regexp.MustCompile(`\s+data-[a-zA-Z0-9-]+="[^"]*"`)
	html = re.ReplaceAllString(html, "")
	re = regexp.MustCompile(`\s+data-[a-zA-Z0-9-]+'[^']*'`)
	return re.ReplaceAllString(html, "")
}

func fixSelfClosingTags(html string) string {
	selfClosingTags := []string{"br", "hr", "img", "input", "meta", "link", "area", "base", "col", "embed", "param", "source", "track", "wbr"}

	for _, tag := range selfClosingTags {
		re := regexp.MustCompile(`(?i)<(` + tag + `)([^/>]*)>`)
		html = re.ReplaceAllString(html, "<$1$2/>")

		re = regexp.MustCompile(`(?i)<(` + tag + `)([^/>]*)\s+/>`)
		html = re.ReplaceAllString(html, "<$1$2/>")
	}

	return html
}

func removeEmptyParagraphs(html string) string {
	re := regexp.MustCompile(`<p[^>]*>\s*</p>`)
	return re.ReplaceAllString(html, "")
}

func collapseWhitespace(html string) string {
	re := regexp.MustCompile(`[ \t]+`)
	html = re.ReplaceAllString(html, " ")

	re = regexp.MustCompile(`\n{3,}`)
	html = re.ReplaceAllString(html, "\n\n")

	return html
}

// WrapInParagraphs wraps text content in <p> tags
func WrapInParagraphs(content string) string {
	paragraphs := regexp.MustCompile(`\n\n+`).Split(content, -1)

	var result strings.Builder
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if regexp.MustCompile(`^<(p|div|h[1-6]|blockquote|ul|ol|table|pre)`).MatchString(p) {
			result.WriteString(p)
		} else {
			result.WriteString("<p>")
			result.WriteString(p)
			result.WriteString("</p>\n")
		}
	}

	return result.String()
}

// ExtractTextContent extracts plain text from HTML
func ExtractTextContent(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")

	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&apos;", "'")

	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// ConvertBRToParagraphs converts <br><br> sequences to paragraph breaks
func ConvertBRToParagraphs(html string) string {
	re := regexp.MustCompile(`(<br\s*/?>){2,}`)
	return re.ReplaceAllString(html, "</p><p>")
}
