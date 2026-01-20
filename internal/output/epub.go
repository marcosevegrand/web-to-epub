package output

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-shiori/go-epub"
)

// EPUBOptions contains options for EPUB generation
type EPUBOptions struct {
	IncludeTableOfContents bool
	CustomCSS              string
	EmbedFonts             bool
}

// DefaultEPUBOptions returns default EPUB generation options
func DefaultEPUBOptions() *EPUBOptions {
	return &EPUBOptions{
		IncludeTableOfContents: true,
		EmbedFonts:             false,
	}
}

// GenerateEPUB generates an EPUB file from a book
func GenerateEPUB(book *Book, outputPath string) error {
	return GenerateEPUBWithOptions(book, outputPath, DefaultEPUBOptions())
}

// GenerateEPUBWithOptions generates an EPUB file with custom options
func GenerateEPUBWithOptions(book *Book, outputPath string, opts *EPUBOptions) error {
	if len(book.Chapters) == 0 {
		return fmt.Errorf("no chapters to include in EPUB")
	}

	e, err := epub.NewEpub(book.Title)
	if err != nil {
		return fmt.Errorf("failed to create EPUB: %w", err)
	}

	e.SetAuthor(book.Author)
	if book.Description != "" {
		e.SetDescription(book.Description)
	} else {
		e.SetDescription(fmt.Sprintf("Web novel - %d chapters", len(book.Chapters)))
	}
	if book.Lang != "" {
		e.SetLang(book.Lang)
	}
	if book.Identifier != "" {
		e.SetIdentifier(book.Identifier)
	}

	// Add CSS - AddCSS expects CSS content as string, not file path
	css := getDefaultCSS()
	if opts.CustomCSS != "" {
		css += "\n" + opts.CustomCSS
	}
	
	var cssPath string
	if css != "" {
		// The AddCSS function signature: AddCSS(cssContent string, destFilename string) (string, error)
		cssPath, err = e.AddCSS(css, "styles.css")
		if err != nil {
			fmt.Printf("⚠ Warning: Failed to add CSS: %v\n", err)
			cssPath = ""
		}
	}

	for i, ch := range book.Chapters {
		cleanContent := sanitizeHTML(ch.Content)
		cleanContent = normalizeForEPUB(cleanContent)

		sectionBody := formatChapterHTML(ch.Title, cleanContent)

		_, err := e.AddSection(sectionBody, ch.Title, "", cssPath)
		if err != nil {
			fmt.Printf("⚠ Warning: Error adding chapter %d (%s): %v\n", i+1, ch.Title, err)
		}
	}

	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := SanitizeFilename(book.Title) + ".epub"
	fullPath := filepath.Join(outputPath, filename)

	if err := e.Write(fullPath); err != nil {
		return fmt.Errorf("failed to write EPUB: %w", err)
	}

	info, err := os.Stat(fullPath)
	if err == nil {
		fmt.Printf("✓ EPUB generated: %s (%s)\n", fullPath, FormatFileSize(info.Size()))
	} else {
		fmt.Printf("✓ EPUB generated: %s\n", fullPath)
	}

	return nil
}

func getDefaultCSS() string {
	return `body {
    font-family: Georgia, "Times New Roman", serif;
    line-height: 1.6;
    margin: 1em;
    padding: 0;
}

h1 {
    font-size: 1.5em;
    margin-top: 1em;
    margin-bottom: 0.5em;
    page-break-after: avoid;
}

h2 {
    font-size: 1.3em;
    margin-top: 0.8em;
    margin-bottom: 0.4em;
}

h3 {
    font-size: 1.1em;
    margin-top: 0.6em;
    margin-bottom: 0.3em;
}

p {
    margin: 0.5em 0;
    text-align: justify;
    text-indent: 1.5em;
}

p:first-of-type {
    text-indent: 0;
}

blockquote {
    margin: 1em 2em;
    padding: 0.5em 1em;
    border-left: 3px solid #ccc;
    font-style: italic;
}

pre, code {
    font-family: "Courier New", Courier, monospace;
    font-size: 0.9em;
}

pre {
    padding: 1em;
    background-color: #f5f5f5;
    overflow-x: auto;
    white-space: pre-wrap;
}

hr {
    border: none;
    border-top: 1px solid #ccc;
    margin: 2em 0;
}

a {
    color: #0066cc;
    text-decoration: none;
}

img {
    max-width: 100%;
    height: auto;
}

.chapter-title {
    text-align: center;
    margin-bottom: 2em;
}

.separator {
    text-align: center;
    margin: 2em 0;
}
`
}

func formatChapterHTML(title, content string) string {
	escapedTitle := html.EscapeString(title)

	return fmt.Sprintf(`<h1 class="chapter-title">%s</h1>
%s`, escapedTitle, content)
}

func sanitizeHTML(htmlContent string) string {
	htmlContent = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`(?is)<noscript[^>]*>.*?</noscript>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`<!--[\s\S]*?-->`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`\s+on\w+="[^"]*"`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`\s+style="[^"]*"`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`\s+class="[^"]*"`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`\s+id="[^"]*"`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`\s+data-[a-z-]+="[^"]*"`).ReplaceAllString(htmlContent, "")

	return strings.TrimSpace(htmlContent)
}

func normalizeForEPUB(htmlContent string) string {
	selfClosing := []string{"br", "hr", "img", "input", "meta", "link"}
	for _, tag := range selfClosing {
		re := regexp.MustCompile(fmt.Sprintf(`(?i)<%s([^/>]*)>`, tag))
		htmlContent = re.ReplaceAllString(htmlContent, fmt.Sprintf("<%s$1/>", tag))
	}

	if !regexp.MustCompile(`^\s*<(p|div|h[1-6]|ul|ol|table|blockquote)`).MatchString(htmlContent) {
		paragraphs := regexp.MustCompile(`\n\s*\n`).Split(htmlContent, -1)
		var wrapped []string
		for _, p := range paragraphs {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if !regexp.MustCompile(`^<(p|div|h[1-6]|ul|ol|table|blockquote)`).MatchString(p) {
				p = "<p>" + p + "</p>"
			}
			wrapped = append(wrapped, p)
		}
		htmlContent = strings.Join(wrapped, "\n")
	}

	return htmlContent
}

// CreateTableOfContents generates a table of contents chapter
func CreateTableOfContents(chapters []Chapter) string {
	var toc strings.Builder
	toc.WriteString("<h1>Table of Contents</h1>\n<ul>\n")

	for _, ch := range chapters {
		escapedTitle := html.EscapeString(ch.Title)
		toc.WriteString(fmt.Sprintf("  <li>%s</li>\n", escapedTitle))
	}

	toc.WriteString("</ul>\n")
	return toc.String()
}
