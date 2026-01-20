package formatter

import (
	"regexp"
	"strings"
	"unicode"
)

// TextProcessor provides text processing utilities
type TextProcessor struct {
	SmartQuotes bool
	SmartDashes bool
	TrimLines   bool
}

// NewTextProcessor creates a processor with default settings
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{
		SmartQuotes: true,
		SmartDashes: true,
		TrimLines:   true,
	}
}

// Process applies text processing transformations
func (p *TextProcessor) Process(text string) string {
	if p.TrimLines {
		text = trimLines(text)
	}
	if p.SmartQuotes {
		text = convertSmartQuotes(text)
	}
	if p.SmartDashes {
		text = convertSmartDashes(text)
	}
	return text
}

func trimLines(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

func convertSmartQuotes(text string) string {
	re := regexp.MustCompile(`(^|[\s\n\r\(])\"`)
	text = re.ReplaceAllString(text, "$1"")

	re = regexp.MustCompile(`\"([\s\n\r\.,!?\)\:]|$)`)
	text = re.ReplaceAllString(text, ""$1")

	re = regexp.MustCompile(`(^|[\s\n\r\(])\'`)
	text = re.ReplaceAllString(text, "$1'")

	re = regexp.MustCompile(`\'([\s\n\r\.,!?\)\:]|$)`)
	text = re.ReplaceAllString(text, "'$1")

	re = regexp.MustCompile(`([a-zA-Z])\'([a-zA-Z])`)
	text = re.ReplaceAllString(text, "$1'$2")

	return text
}

func convertSmartDashes(text string) string {
	text = strings.ReplaceAll(text, "---", "—")
	text = strings.ReplaceAll(text, "--", "–")
	return text
}

// NormalizeWhitespace normalizes various whitespace characters
func NormalizeWhitespace(text string) string {
	text = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) && r != '\n' && r != '\r' {
			return ' '
		}
		return r
	}, text)

	re := regexp.MustCompile(` +`)
	text = re.ReplaceAllString(text, " ")

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	return text
}

// RemoveControlCharacters removes non-printable control characters
func RemoveControlCharacters(text string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r >= 32 {
			return r
		}
		return -1
	}, text)
}

// TruncateText truncates text to a maximum length with ellipsis
func TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	truncated := text[:maxLen-3]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLen/2 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// CleanTitle cleans a chapter title
func CleanTitle(title string) string {
	title = ExtractTextContent(title)
	title = NormalizeWhitespace(title)
	title = strings.TrimSpace(title)
	title = RemoveControlCharacters(title)

	if len(title) > 200 {
		title = TruncateText(title, 200)
	}

	return title
}

// SplitIntoParagraphs splits text into paragraphs
func SplitIntoParagraphs(text string) []string {
	re := regexp.MustCompile(`\n\s*\n`)
	parts := re.Split(text, -1)

	paragraphs := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			paragraphs = append(paragraphs, p)
		}
	}

	return paragraphs
}

// WordCount returns word count for text
func WordCount(text string) int {
	text = ExtractTextContent(text)
	words := strings.Fields(text)
	return len(words)
}

// SentenceCount returns approximate sentence count
func SentenceCount(text string) int {
	text = ExtractTextContent(text)
	re := regexp.MustCompile(`[.!?]+`)
	matches := re.FindAllString(text, -1)

	if len(matches) == 0 {
		return 1
	}
	return len(matches)
}

// CharacterCount returns character count excluding HTML
func CharacterCount(text string) int {
	text = ExtractTextContent(text)
	return len(text)
}

// DetectLanguage attempts to detect the primary language (simplified)
func DetectLanguage(text string) string {
	text = strings.ToLower(text)

	englishWords := []string{"the", "and", "is", "it", "to", "of", "in", "that", "was", "for"}
	spanishWords := []string{"el", "la", "de", "que", "y", "en", "un", "es", "por", "con"}
	frenchWords := []string{"le", "la", "de", "et", "est", "un", "une", "que", "dans", "pour"}

	englishCount := 0
	spanishCount := 0
	frenchCount := 0

	words := strings.Fields(text)
	for _, word := range words {
		for _, en := range englishWords {
			if word == en {
				englishCount++
			}
		}
		for _, es := range spanishWords {
			if word == es {
				spanishCount++
			}
		}
		for _, fr := range frenchWords {
			if word == fr {
				frenchCount++
			}
		}
	}

	if spanishCount > englishCount && spanishCount > frenchCount {
		return "es"
	}
	if frenchCount > englishCount && frenchCount > spanishCount {
		return "fr"
	}
	return "en"
}
