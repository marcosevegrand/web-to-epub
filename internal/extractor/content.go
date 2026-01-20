package extractor

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractTitle extracts the chapter title from a document
func ExtractTitle(doc *goquery.Document, selector, regexPattern string) string {
	if selector != "" {
		title := doc.Find(selector).First()
		if title.Length() > 0 {
			text := strings.TrimSpace(title.Text())
			if text != "" {
				return text
			}
		}
	}

	if regexPattern != "" {
		html, _ := doc.Html()
		re := regexp.MustCompile("(?i)" + regexPattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			text := strings.TrimSpace(matches[1])
			text = stripHTMLTags(text)
			if text != "" {
				return text
			}
		}
	}

	fallbackSelectors := []string{
		"h1.title",
		"h1.chapter-title",
		"h1.entry-title",
		"h1.post-title",
		"h1",
		"h2.title",
		"h2.chapter-title",
		".chapter-title",
		".story-title",
		"title",
	}

	for _, sel := range fallbackSelectors {
		title := doc.Find(sel).First()
		if title.Length() > 0 {
			text := strings.TrimSpace(title.Text())
			if text != "" && len(text) < 200 {
				return text
			}
		}
	}

	return ""
}

func stripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// CleanContent cleans and normalizes extracted HTML content
func CleanContent(html string) string {
	scriptRe := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	html = scriptRe.ReplaceAllString(html, "")

	styleRe := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	html = styleRe.ReplaceAllString(html, "")

	noscriptRe := regexp.MustCompile(`(?is)<noscript[^>]*>.*?</noscript>`)
	html = noscriptRe.ReplaceAllString(html, "")

	eventRe := regexp.MustCompile(`\s+on\w+="[^"]*"`)
	html = eventRe.ReplaceAllString(html, "")

	styleAttrRe := regexp.MustCompile(`\s+style="[^"]*"`)
	html = styleAttrRe.ReplaceAllString(html, "")

	classRe := regexp.MustCompile(`\s+class="[^"]*"`)
	html = classRe.ReplaceAllString(html, "")

	idRe := regexp.MustCompile(`\s+id="[^"]*"`)
	html = idRe.ReplaceAllString(html, "")

	dataRe := regexp.MustCompile(`\s+data-[a-z-]+="[^"]*"`)
	html = dataRe.ReplaceAllString(html, "")

	html = regexp.MustCompile(`<(\w+)\s+>`).ReplaceAllString(html, "<$1>")

	html = regexp.MustCompile(`<p>\s*</p>`).ReplaceAllString(html, "")

	html = regexp.MustCompile(`(<br\s*/?>){3,}`).ReplaceAllString(html, "<br/><br/>")

	html = strings.ReplaceAll(html, "\r\n", "\n")
	html = strings.ReplaceAll(html, "\r", "\n")

	html = regexp.MustCompile(`\n{3,}`).ReplaceAllString(html, "\n\n")

	return strings.TrimSpace(html)
}

// NormalizeForEPUB converts HTML to EPUB-compatible XHTML
func NormalizeForEPUB(html string) string {
	selfClosingTags := []string{"br", "hr", "img", "input", "meta", "link"}
	for _, tag := range selfClosingTags {
		re := regexp.MustCompile(`<` + tag + `([^/>]*)>`)
		html = re.ReplaceAllString(html, "<"+tag+"$1/>")
		re = regexp.MustCompile(`<` + tag + `([^/>]*)\s*/>`)
		html = re.ReplaceAllString(html, "<"+tag+"$1/>")
	}

	return html
}

// ExtractImages finds all image URLs in content
func ExtractImages(html string) []string {
	var images []string
	re := regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
	matches := re.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			images = append(images, match[1])
		}
	}
	return images
}

// WordCount returns an approximate word count for content
func WordCount(html string) int {
	text := stripHTMLTags(html)
	words := strings.Fields(text)
	return len(words)
}

// EstimateReadingTime estimates reading time in minutes (assuming 200 wpm)
func EstimateReadingTime(html string) int {
	words := WordCount(html)
	minutes := words / 200
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}
