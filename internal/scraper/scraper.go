package scraper

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"web-to-epub-go/internal/config"
	"web-to-epub-go/internal/extractor"
	"web-to-epub-go/internal/navigator"
	"web-to-epub-go/internal/output"
)

// Chapter represents a scraped chapter
type Chapter struct {
	Title   string
	Content string
	URL     string
	Index   int
}

// WebScraper orchestrates the scraping process
type WebScraper struct {
	config    *config.Config
	requester *Requester
	navigator *navigator.ChapterNavigator
	strategy  extractor.DetectionStrategy
	chapters  []Chapter
	options   *extractor.ExtractionOptions
}

// NewWebScraper creates a new web scraper instance
func NewWebScraper(cfg *config.Config) (*WebScraper, error) {
	requester := NewRequester(
		cfg.Scraping.Polite,
		cfg.Scraping.UserAgent,
		cfg.Scraping.Timeout,
	)

	nav := navigator.NewChapterNavigator(cfg.Navigation)

	strategy := buildStrategy(cfg.ContentDetection)

	options := &extractor.ExtractionOptions{
		TitleSelector:    cfg.ChapterExtract.TitleSelector,
		TitleRegex:       cfg.ChapterExtract.TitleRegex,
		MinContentLength: 100,
		PreserveImages:   true,
		CleanScripts:     true,
	}

	return &WebScraper{
		config:    cfg,
		requester: requester,
		navigator: nav,
		strategy:  strategy,
		chapters:  make([]Chapter, 0),
		options:   options,
	}, nil
}

// ScrapeAll scrapes all chapters based on configuration
func (ws *WebScraper) ScrapeAll() error {
	ctx := context.Background()

	switch ws.config.Navigation.Method {
	case "url_pattern":
		return ws.scrapeByPattern(ctx)
	case "next_link":
		return ws.scrapeByNextLink(ctx)
	case "toc":
		return ws.scrapeByTOC(ctx)
	default:
		return fmt.Errorf("unknown navigation method: %s", ws.config.Navigation.Method)
	}
}

func (ws *WebScraper) scrapeByPattern(ctx context.Context) error {
	chapters, err := ws.navigator.DiscoverChapters(ws.config.Scraping.StartURL)
	if err != nil {
		return fmt.Errorf("failed to discover chapters: %w", err)
	}

	fmt.Printf("📖 Found %d chapters to scrape\n\n", len(chapters))

	for i, chapterInfo := range chapters {
		fmt.Printf("⏳ [%d/%d] Scraping: %s\n", i+1, len(chapters), chapterInfo.URL)

		if err := ws.scrapeChapter(ctx, chapterInfo.URL, chapterInfo.Index); err != nil {
			fmt.Printf("  ✗ Error: %v\n", err)
			ws.navigator.UpdateChapterStatus(chapterInfo.Index, "error")
		} else {
			ws.navigator.UpdateChapterStatus(chapterInfo.Index, "scraped")
		}

		total, scraped, errors := ws.navigator.GetProgress()
		fmt.Printf("  📊 Progress: %d/%d scraped, %d errors\n\n", scraped, total, errors)
	}

	return nil
}

func (ws *WebScraper) scrapeByNextLink(ctx context.Context) error {
	currentURL := ws.config.Scraping.StartURL
	index := 1
	maxChapters := ws.config.Navigation.MaxChapters
	if maxChapters <= 0 {
		maxChapters = 10000
	}

	fmt.Printf("📖 Starting next-link navigation from: %s\n\n", currentURL)

	for index <= maxChapters {
		fmt.Printf("⏳ [%d] Scraping: %s\n", index, currentURL)

		ws.navigator.MarkVisited(currentURL)

		body, err := ws.requester.Fetch(ctx, currentURL)
		if err != nil {
			fmt.Printf("  ✗ Error fetching: %v\n", err)
			break
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			fmt.Printf("  ✗ Error parsing: %v\n", err)
			break
		}

		content, err := ws.strategy.Extract(doc, ws.options)
		if err != nil {
			fmt.Printf("  ✗ Error extracting content: %v\n", err)
		} else {
			title := extractor.ExtractTitle(doc, ws.options.TitleSelector, ws.options.TitleRegex)
			if title == "" {
				title = fmt.Sprintf("Chapter %d", index)
			}

			content = extractor.CleanContent(content)

			ws.chapters = append(ws.chapters, Chapter{
				Title:   title,
				Content: content,
				URL:     currentURL,
				Index:   index,
			})

			fmt.Printf("  ✓ Extracted: %s (%d chars)\n", title, len(content))
		}

		nextURL, found := ws.navigator.FindNextChapterLink(doc, currentURL)
		if !found {
			fmt.Println("\n📍 No more chapters found (next link not found)")
			break
		}

		currentURL = nextURL
		index++
	}

	fmt.Printf("\n📊 Total chapters scraped: %d\n", len(ws.chapters))
	return nil
}

func (ws *WebScraper) scrapeByTOC(ctx context.Context) error {
	tocURL := ws.config.Navigation.TOCUrl
	if tocURL == "" {
		return fmt.Errorf("TOC URL not configured")
	}

	fmt.Printf("📖 Fetching table of contents: %s\n\n", tocURL)

	body, err := ws.requester.Fetch(ctx, tocURL)
	if err != nil {
		return fmt.Errorf("failed to fetch TOC: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to parse TOC: %w", err)
	}

	chapters, err := ws.navigator.ParseTOCPage(doc, tocURL)
	if err != nil {
		return fmt.Errorf("failed to parse TOC links: %w", err)
	}

	fmt.Printf("📖 Found %d chapters in TOC\n\n", len(chapters))

	for i, chapterInfo := range chapters {
		fmt.Printf("⏳ [%d/%d] Scraping: %s\n", i+1, len(chapters), chapterInfo.Title)

		if err := ws.scrapeChapter(ctx, chapterInfo.URL, chapterInfo.Index); err != nil {
			fmt.Printf("  ✗ Error: %v\n", err)
			ws.navigator.UpdateChapterStatus(chapterInfo.Index, "error")
		} else {
			if len(ws.chapters) > 0 {
				lastChapter := ws.chapters[len(ws.chapters)-1]
				if lastChapter.Title != "" && lastChapter.Title != chapterInfo.Title {
					ws.navigator.UpdateChapterTitle(chapterInfo.Index, lastChapter.Title)
				}
			}
			ws.navigator.UpdateChapterStatus(chapterInfo.Index, "scraped")
		}

		total, scraped, errors := ws.navigator.GetProgress()
		fmt.Printf("  📊 Progress: %d/%d scraped, %d errors\n\n", scraped, total, errors)
	}

	return nil
}

func (ws *WebScraper) scrapeChapter(ctx context.Context, url string, index int) error {
	body, err := ws.requester.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	content, err := ws.strategy.Extract(doc, ws.options)
	if err != nil {
		return fmt.Errorf("failed to extract content: %w", err)
	}

	title := extractor.ExtractTitle(doc, ws.options.TitleSelector, ws.options.TitleRegex)
	if title == "" {
		title = fmt.Sprintf("Chapter %d", index)
	}

	content = extractor.CleanContent(content)

	ws.chapters = append(ws.chapters, Chapter{
		Title:   title,
		Content: content,
		URL:     url,
		Index:   index,
	})

	fmt.Printf("  ✓ Extracted: %s (%d chars)\n", title, len(content))
	return nil
}

// GenerateEPUB generates an EPUB from scraped chapters
func (ws *WebScraper) GenerateEPUB(outputPath string) error {
	if len(ws.chapters) == 0 {
		return fmt.Errorf("no chapters to generate EPUB from")
	}

	outputChapters := make([]output.Chapter, len(ws.chapters))
	for i, ch := range ws.chapters {
		outputChapters[i] = output.Chapter{
			Title:   ch.Title,
			Content: ch.Content,
			Index:   ch.Index,
		}
	}

	book := &output.Book{
		Title:       ws.config.Book.Title,
		Author:      ws.config.Book.Author,
		Description: ws.config.Book.Description,
		Chapters:    outputChapters,
		Lang:        ws.config.Output.EPUBMetadata.Lang,
		Rights:      ws.config.Output.EPUBMetadata.Rights,
		Publisher:   ws.config.Output.EPUBMetadata.Publisher,
	}

	return output.GenerateEPUB(book, outputPath)
}

// GetChapters returns the scraped chapters
func (ws *WebScraper) GetChapters() []Chapter {
	return ws.chapters
}

// GetProgress returns scraping progress
func (ws *WebScraper) GetProgress() (total, scraped, errors int) {
	return ws.navigator.GetProgress()
}

func buildStrategy(cfg config.ContentDetectionConfig) extractor.DetectionStrategy {
	switch cfg.Strategy {
	case "css_selector":
		return &extractor.CSSSelectorStrategy{
			Selector:         cfg.CSSSelector,
			ExcludeSelectors: cfg.ExcludeSelectors,
		}
	case "text_density":
		return &extractor.TextDensityStrategy{
			MinDensityScore: cfg.TextDensity.MinDensityScore,
			MinBlockSize:    cfg.TextDensity.MinBlockSize,
		}
	case "xpath_regex":
		return &extractor.XPathRegexStrategy{
			Patterns: cfg.RegexPatterns,
		}
	case "dom_position":
		return &extractor.DOMPositionStrategy{
			MaxDepth: cfg.DOMPosition.MaxDepth,
			MinWidth: cfg.DOMPosition.MinWidth,
		}
	case "hybrid":
		strategies := []extractor.DetectionStrategy{}

		if cfg.CSSSelector != "" {
			strategies = append(strategies, &extractor.CSSSelectorStrategy{
				Selector:         cfg.CSSSelector,
				ExcludeSelectors: cfg.ExcludeSelectors,
			})
		}

		strategies = append(strategies, &extractor.TextDensityStrategy{
			MinDensityScore: cfg.TextDensity.MinDensityScore,
			MinBlockSize:    cfg.TextDensity.MinBlockSize,
		})

		strategies = append(strategies, &extractor.DOMPositionStrategy{
			MaxDepth: cfg.DOMPosition.MaxDepth,
			MinWidth: cfg.DOMPosition.MinWidth,
		})

		return &extractor.HybridStrategy{Strategies: strategies}
	default:
		return &extractor.DOMPositionStrategy{}
	}
}

// ScrapeTest scrapes a single URL for testing configuration
func (ws *WebScraper) ScrapeTest(url string) (*Chapter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	body, err := ws.requester.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	content, err := ws.strategy.Extract(doc, ws.options)
	if err != nil {
		return nil, fmt.Errorf("failed to extract content: %w", err)
	}

	title := extractor.ExtractTitle(doc, ws.options.TitleSelector, ws.options.TitleRegex)
	if title == "" {
		title = "Test Chapter"
	}

	content = extractor.CleanContent(content)

	return &Chapter{
		Title:   title,
		Content: content,
		URL:     url,
		Index:   1,
	}, nil
}

// PrintSummary prints a summary of the scraping results
func (ws *WebScraper) PrintSummary() {
	totalWords := 0
	totalChars := 0

	for _, ch := range ws.chapters {
		totalWords += extractor.WordCount(ch.Content)
		totalChars += len(ch.Content)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("📚 Scraping Summary\n")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Book: %s by %s\n", ws.config.Book.Title, ws.config.Book.Author)
	fmt.Printf("Chapters scraped: %d\n", len(ws.chapters))
	fmt.Printf("Total words: ~%d\n", totalWords)
	fmt.Printf("Total characters: %d\n", totalChars)
	fmt.Printf("Estimated reading time: ~%d minutes\n", totalWords/200)
	fmt.Println(strings.Repeat("=", 50))
}
