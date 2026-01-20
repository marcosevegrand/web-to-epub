package navigator

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"web-to-epub-go/internal/config"
)

// ChapterInfo contains information about a discovered chapter
type ChapterInfo struct {
	URL    string
	Title  string
	Index  int
	Status string
}

// ChapterNavigator handles chapter discovery and navigation
type ChapterNavigator struct {
	config   config.NavigationConfig
	chapters []ChapterInfo
	visited  map[string]bool
}

// NewChapterNavigator creates a new chapter navigator
func NewChapterNavigator(cfg config.NavigationConfig) *ChapterNavigator {
	return &ChapterNavigator{
		config:   cfg,
		chapters: make([]ChapterInfo, 0),
		visited:  make(map[string]bool),
	}
}

// DiscoverChapters discovers all chapter URLs based on configuration
func (cn *ChapterNavigator) DiscoverChapters(startURL string) ([]ChapterInfo, error) {
	switch cn.config.Method {
	case "url_pattern":
		return cn.discoverByPattern()
	case "next_link":
		return cn.discoverByNextLink(startURL)
	case "toc":
		return cn.discoverByTOC()
	default:
		return nil, fmt.Errorf("unknown navigation method: %s", cn.config.Method)
	}
}

func (cn *ChapterNavigator) discoverByPattern() ([]ChapterInfo, error) {
	if cn.config.URLPattern == "" {
		return nil, fmt.Errorf("URL pattern not configured")
	}

	start := cn.config.NumberStart
	if start < 1 {
		start = 1
	}
	end := cn.config.NumberEnd
	if end < start {
		return nil, fmt.Errorf("numberEnd (%d) must be >= numberStart (%d)", end, start)
	}

	urls := GenerateURLsFromPattern(cn.config.URLPattern, start, end)

	chapters := make([]ChapterInfo, len(urls))
	for i, url := range urls {
		chapters[i] = ChapterInfo{
			URL:    url,
			Index:  i + 1,
			Title:  fmt.Sprintf("Chapter %d", start+i),
			Status: "pending",
		}
	}

	cn.chapters = chapters
	return chapters, nil
}

func (cn *ChapterNavigator) discoverByNextLink(startURL string) ([]ChapterInfo, error) {
	chapters := []ChapterInfo{
		{
			URL:    startURL,
			Index:  1,
			Title:  "Chapter 1",
			Status: "pending",
		},
	}
	cn.chapters = chapters
	return chapters, nil
}

func (cn *ChapterNavigator) discoverByTOC() ([]ChapterInfo, error) {
	return nil, fmt.Errorf("TOC discovery requires scraper to fetch TOC page first")
}

// ParseTOCPage parses a table of contents page and extracts chapter links
func (cn *ChapterNavigator) ParseTOCPage(doc *goquery.Document, baseURL string) ([]ChapterInfo, error) {
	if cn.config.TOCLinkSelector == "" {
		return nil, fmt.Errorf("TOC link selector not configured")
	}

	chapters := make([]ChapterInfo, 0)
	index := 1

	doc.Find(cn.config.TOCLinkSelector).Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		url := ResolveRelativeURL(baseURL, href)

		if cn.visited[url] {
			return
		}
		cn.visited[url] = true

		title := strings.TrimSpace(s.Text())
		if title == "" {
			title = fmt.Sprintf("Chapter %d", index)
		}

		if cn.config.MaxChapters > 0 && index > cn.config.MaxChapters {
			return
		}

		chapters = append(chapters, ChapterInfo{
			URL:    url,
			Title:  title,
			Index:  index,
			Status: "pending",
		})
		index++
	})

	cn.chapters = chapters
	return chapters, nil
}

// FindNextChapterLink finds the next chapter link in a document
func (cn *ChapterNavigator) FindNextChapterLink(doc *goquery.Document, currentURL string) (string, bool) {
	if cn.config.NextLinkSelector == "" {
		return "", false
	}

	link := doc.Find(cn.config.NextLinkSelector).First()
	if link.Length() == 0 {
		return "", false
	}

	href, exists := link.Attr("href")
	if !exists || href == "" {
		return "", false
	}

	nextURL := ResolveRelativeURL(currentURL, href)

	if nextURL == currentURL {
		return "", false
	}

	if cn.visited[nextURL] {
		return "", false
	}

	return nextURL, true
}

// MarkVisited marks a URL as visited
func (cn *ChapterNavigator) MarkVisited(url string) {
	cn.visited[url] = true
}

// IsVisited checks if a URL has been visited
func (cn *ChapterNavigator) IsVisited(url string) bool {
	return cn.visited[url]
}

// AddChapter adds a discovered chapter
func (cn *ChapterNavigator) AddChapter(chapter ChapterInfo) {
	cn.chapters = append(cn.chapters, chapter)
}

// GetChapters returns all discovered chapters
func (cn *ChapterNavigator) GetChapters() []ChapterInfo {
	return cn.chapters
}

// UpdateChapterStatus updates the status of a chapter
func (cn *ChapterNavigator) UpdateChapterStatus(index int, status string) {
	if index > 0 && index <= len(cn.chapters) {
		cn.chapters[index-1].Status = status
	}
}

// UpdateChapterTitle updates the title of a chapter
func (cn *ChapterNavigator) UpdateChapterTitle(index int, title string) {
	if index > 0 && index <= len(cn.chapters) && title != "" {
		cn.chapters[index-1].Title = title
	}
}

// GetProgress returns scraping progress statistics
func (cn *ChapterNavigator) GetProgress() (total, scraped, errors int) {
	total = len(cn.chapters)
	for _, ch := range cn.chapters {
		switch ch.Status {
		case "scraped":
			scraped++
		case "error":
			errors++
		}
	}
	return
}
