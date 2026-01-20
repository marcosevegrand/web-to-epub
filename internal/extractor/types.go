// Package extractor provides content extraction strategies for web scraping.
package extractor

import (
	"github.com/PuerkitoBio/goquery"
)

// ExtractionResult contains the result of content extraction
type ExtractionResult struct {
	Content      string
	Title        string
	StrategyUsed string
	Warnings     []string
}

// ExtractionOptions provides options for extraction strategies
type ExtractionOptions struct {
	TitleSelector    string
	TitleRegex       string
	MinContentLength int
	PreserveImages   bool
	CleanScripts     bool
}

// DefaultExtractionOptions returns sensible defaults
func DefaultExtractionOptions() *ExtractionOptions {
	return &ExtractionOptions{
		MinContentLength: 100,
		PreserveImages:   true,
		CleanScripts:     true,
	}
}

// DetectionStrategy interface for pluggable detection methods
type DetectionStrategy interface {
	Extract(doc *goquery.Document, opts *ExtractionOptions) (string, error)
	Name() string
}

// ContentExtractor manages multiple extraction strategies
type ContentExtractor struct {
	strategies []DetectionStrategy
	options    *ExtractionOptions
}

// NewContentExtractor creates a new extractor with the given strategies
func NewContentExtractor(strategies []DetectionStrategy, opts *ExtractionOptions) *ContentExtractor {
	if opts == nil {
		opts = DefaultExtractionOptions()
	}
	return &ContentExtractor{
		strategies: strategies,
		options:    opts,
	}
}

// Extract attempts content extraction using registered strategies in order
func (ce *ContentExtractor) Extract(doc *goquery.Document) (*ExtractionResult, error) {
	result := &ExtractionResult{
		Warnings: make([]string, 0),
	}

	var lastErr error
	for _, strategy := range ce.strategies {
		content, err := strategy.Extract(doc, ce.options)
		if err != nil {
			result.Warnings = append(result.Warnings,
				"Strategy '"+strategy.Name()+"' failed: "+err.Error())
			lastErr = err
			continue
		}

		if len(content) < ce.options.MinContentLength {
			result.Warnings = append(result.Warnings,
				"Strategy '"+strategy.Name()+"' returned content too short")
			continue
		}

		result.Content = content
		result.StrategyUsed = strategy.Name()
		return result, nil
	}

	if lastErr != nil {
		return result, lastErr
	}
	return result, nil
}

// AddStrategy adds a new strategy to the extractor
func (ce *ContentExtractor) AddStrategy(s DetectionStrategy) {
	ce.strategies = append(ce.strategies, s)
}

// SetOptions updates extraction options
func (ce *ContentExtractor) SetOptions(opts *ExtractionOptions) {
	ce.options = opts
}
