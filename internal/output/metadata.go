// Package output provides EPUB generation and metadata handling.
package output

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

// Book represents a book with metadata and chapters
type Book struct {
	Title       string
	Author      string
	Description string
	Lang        string
	Rights      string
	Publisher   string
	Identifier  string
	Cover       []byte
	CoverType   string
	Chapters    []Chapter
	CreatedAt   time.Time
}

// Chapter represents a book chapter
type Chapter struct {
	Title   string
	Content string
	Index   int
}

// Metadata represents EPUB metadata
type Metadata struct {
	Title       string
	Author      string
	Description string
	Language    string
	Rights      string
	Publisher   string
	Identifier  string
	Date        time.Time
	Subject     []string
}

// NewBook creates a new book with defaults
func NewBook(title, author string) *Book {
	return &Book{
		Title:      title,
		Author:     author,
		Lang:       "en",
		Rights:     "Personal use only",
		Chapters:   make([]Chapter, 0),
		CreatedAt:  time.Now(),
		Identifier: GenerateUUID(),
	}
}

// AddChapter adds a chapter to the book
func (b *Book) AddChapter(title, content string) {
	index := len(b.Chapters) + 1
	b.Chapters = append(b.Chapters, Chapter{
		Title:   title,
		Content: content,
		Index:   index,
	})
}

// GetMetadata returns the book's metadata
func (b *Book) GetMetadata() *Metadata {
	return &Metadata{
		Title:       b.Title,
		Author:      b.Author,
		Description: b.Description,
		Language:    b.Lang,
		Rights:      b.Rights,
		Publisher:   b.Publisher,
		Identifier:  b.Identifier,
		Date:        b.CreatedAt,
	}
}

// TotalWordCount returns the total word count of all chapters
func (b *Book) TotalWordCount() int {
	total := 0
	for _, ch := range b.Chapters {
		total += wordCount(ch.Content)
	}
	return total
}

// EstimatedReadingTime returns estimated reading time in minutes
func (b *Book) EstimatedReadingTime() int {
	words := b.TotalWordCount()
	minutes := words / 200
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

func wordCount(content string) int {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(content, "")
	return len(strings.Fields(text))
}

// GenerateUUID generates a unique identifier for the book
func GenerateUUID() string {
	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Sprintf("urn:uuid:generated-%d", time.Now().UnixNano())
	}
	return "urn:uuid:" + id.String()
}

// SanitizeFilename creates a safe filename from a string
func SanitizeFilename(name string) string {
	invalid := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "")
	}

	result = strings.ReplaceAll(result, " ", "_")

	if len(result) > 100 {
		result = result[:100]
	}

	if result == "" {
		result = "book"
	}

	return strings.TrimSpace(result)
}

// FormatFileSize formats a file size in bytes to human-readable format
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// FormatDuration formats a duration to human-readable format
func FormatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	}

	hours := minutes / 60
	mins := minutes % 60

	if mins == 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	return fmt.Sprintf("%d hours %d minutes", hours, mins)
}
