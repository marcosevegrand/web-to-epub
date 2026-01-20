package extractor

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// CSSSelectorStrategy extracts content using CSS selectors
type CSSSelectorStrategy struct {
	Selector         string
	ExcludeSelectors []string
}

func (s *CSSSelectorStrategy) Name() string {
	return "css_selector"
}

func (s *CSSSelectorStrategy) Extract(doc *goquery.Document, opts *ExtractionOptions) (string, error) {
	if s.Selector == "" {
		return "", fmt.Errorf("no CSS selector configured")
	}

	content := doc.Find(s.Selector)
	if content.Length() == 0 {
		return "", fmt.Errorf("selector '%s' matched no elements", s.Selector)
	}

	content = content.Clone()

	for _, excludeSelector := range s.ExcludeSelectors {
		content.Find(excludeSelector).Remove()
	}

	if opts != nil && opts.CleanScripts {
		content.Find("script, style, noscript").Remove()
	}

	html, err := content.Html()
	if err != nil {
		return "", fmt.Errorf("failed to get HTML: %w", err)
	}

	return strings.TrimSpace(html), nil
}

// TextDensityStrategy extracts content based on text density analysis
type TextDensityStrategy struct {
	MinDensityScore float64
	MinBlockSize    int
}

func (s *TextDensityStrategy) Name() string {
	return "text_density"
}

func (s *TextDensityStrategy) Extract(doc *goquery.Document, opts *ExtractionOptions) (string, error) {
	body := doc.Find("body")
	if body.Length() == 0 {
		return "", fmt.Errorf("no body element found")
	}

	minDensity := s.MinDensityScore
	if minDensity <= 0 {
		minDensity = 0.3
	}
	minSize := s.MinBlockSize
	if minSize <= 0 {
		minSize = 100
	}

	bestBlock := findBestBlockByDensity(body, minDensity, minSize)
	if bestBlock == nil {
		return "", fmt.Errorf("no suitable content block found (min density: %.2f, min size: %d)", minDensity, minSize)
	}

	newDoc := goquery.NewDocumentFromNode(bestBlock)

	if opts != nil && opts.CleanScripts {
		newDoc.Find("script, style, noscript").Remove()
	}

	html, err := newDoc.Html()
	if err != nil {
		return "", fmt.Errorf("failed to get HTML: %w", err)
	}

	return strings.TrimSpace(html), nil
}

func findBestBlockByDensity(selection *goquery.Selection, minDensity float64, minSize int) *html.Node {
	var bestNode *html.Node
	var bestScore float64

	contentTags := map[string]bool{
		"div": true, "article": true, "section": true, "main": true,
		"p": true, "td": true, "blockquote": true,
	}

	selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tagName := goquery.NodeName(s)
		if !contentTags[tagName] {
			return
		}

		class, _ := s.Attr("class")
		id, _ := s.Attr("id")
		combined := strings.ToLower(class + " " + id)
		skipPatterns := []string{"nav", "menu", "sidebar", "footer", "header", "ad", "comment", "share", "social"}
		for _, pattern := range skipPatterns {
			if strings.Contains(combined, pattern) {
				return
			}
		}

		text := strings.TrimSpace(s.Text())
		if len(text) < minSize {
			return
		}

		htmlContent, _ := s.Html()
		if len(htmlContent) == 0 {
			return
		}

		density := float64(len(text)) / float64(len(htmlContent))

		score := density
		if len(text) > 500 {
			score *= 1.2
		}
		if len(text) > 1000 {
			score *= 1.3
		}

		if s.Find("p").Length() > 2 {
			score *= 1.5
		}

		if score > bestScore && density >= minDensity {
			bestScore = score
			bestNode = s.Nodes[0]
		}
	})

	return bestNode
}

// XPathRegexStrategy extracts content using regex patterns on raw HTML
type XPathRegexStrategy struct {
	Patterns []string
}

func (s *XPathRegexStrategy) Name() string {
	return "xpath_regex"
}

func (s *XPathRegexStrategy) Extract(doc *goquery.Document, opts *ExtractionOptions) (string, error) {
	if len(s.Patterns) == 0 {
		return "", fmt.Errorf("no regex patterns configured")
	}

	fullHTML, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("failed to get HTML: %w", err)
	}

	for _, pattern := range s.Patterns {
		re, err := regexp.Compile("(?is)" + pattern)
		if err != nil {
			continue
		}

		matches := re.FindStringSubmatch(fullHTML)
		if len(matches) > 1 {
			content := strings.TrimSpace(matches[1])
			if len(content) > 0 {
				return content, nil
			}
		}
	}

	return "", fmt.Errorf("no regex patterns matched")
}

// DOMPositionStrategy finds content by DOM structure heuristics
type DOMPositionStrategy struct {
	MaxDepth int
	MinWidth int
}

func (s *DOMPositionStrategy) Name() string {
	return "dom_position"
}

func (s *DOMPositionStrategy) Extract(doc *goquery.Document, opts *ExtractionOptions) (string, error) {
	patterns := []string{
		"article.content",
		"article.post-content",
		"article.entry-content",
		"article.story-content",
		"article",
		"main.content",
		"main",
		".chapter-content",
		".story-content",
		".post-content",
		".entry-content",
		".article-content",
		".content-body",
		".post-body",
		".text-content",
		".content",
		".post",
		".story",
		".article",
		"#content",
		"#main-content",
		"#article-content",
		"#story-content",
		"#post-content",
		"[role='main']",
		"[role='article']",
		"div.content",
		"div.main",
	}

	removePatterns := []string{
		"nav", "aside", "header", "footer",
		".ads", ".ad", ".advertisement",
		".sidebar", ".side-bar",
		".comments", ".comment-section",
		".share", ".social", ".sharing",
		".related", ".recommended",
		".navigation", ".breadcrumb",
		".author-bio", ".author-info",
		"script", "style", "noscript",
	}

	for _, pattern := range patterns {
		content := doc.Find(pattern)
		if content.Length() == 0 {
			continue
		}

		content = content.Clone()

		for _, removePattern := range removePatterns {
			content.Find(removePattern).Remove()
		}

		html, err := content.Html()
		if err != nil {
			continue
		}

		html = strings.TrimSpace(html)
		if len(html) > 100 {
			return html, nil
		}
	}

	return "", fmt.Errorf("no content found using DOM position heuristics")
}

// HybridStrategy tries multiple strategies in sequence until one succeeds
type HybridStrategy struct {
	Strategies []DetectionStrategy
}

func (s *HybridStrategy) Name() string {
	return "hybrid"
}

func (s *HybridStrategy) Extract(doc *goquery.Document, opts *ExtractionOptions) (string, error) {
	if len(s.Strategies) == 0 {
		return "", fmt.Errorf("no strategies configured for hybrid extraction")
	}

	var errors []string
	for _, strategy := range s.Strategies {
		content, err := strategy.Extract(doc, opts)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", strategy.Name(), err))
			continue
		}
		if len(content) > 0 {
			return content, nil
		}
	}

	return "", fmt.Errorf("all hybrid strategies failed: %s", strings.Join(errors, "; "))
}

// NewHybridStrategy creates a hybrid strategy with default sub-strategies
func NewHybridStrategy(cssSelector string, excludeSelectors []string, textDensity TextDensityStrategy) *HybridStrategy {
	strategies := []DetectionStrategy{}

	if cssSelector != "" {
		strategies = append(strategies, &CSSSelectorStrategy{
			Selector:         cssSelector,
			ExcludeSelectors: excludeSelectors,
		})
	}

	strategies = append(strategies, &textDensity)
	strategies = append(strategies, &DOMPositionStrategy{})

	return &HybridStrategy{Strategies: strategies}
}
