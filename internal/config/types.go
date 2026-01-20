// Package config provides configuration types and loading functionality
// for the web-to-epub scraper.
package config

// Config is the root configuration structure
type Config struct {
	Book             BookConfig             `yaml:"book"`
	Scraping         ScrapingConfig         `yaml:"scraping"`
	Navigation       NavigationConfig       `yaml:"navigation"`
	ContentDetection ContentDetectionConfig `yaml:"contentDetection"`
	ChapterExtract   ChapterExtractConfig   `yaml:"chapterExtraction"`
	Output           OutputConfig           `yaml:"output"`
}

// BookConfig contains metadata about the book being scraped
type BookConfig struct {
	Title       string `yaml:"title"`
	Author      string `yaml:"author"`
	Description string `yaml:"description"`
	Cover       string `yaml:"cover,omitempty"`
}

// ScrapingConfig controls scraping behavior
type ScrapingConfig struct {
	StartURL  string       `yaml:"startUrl"`
	Polite    PoliteConfig `yaml:"polite"`
	UserAgent string       `yaml:"userAgent,omitempty"`
	Timeout   int          `yaml:"timeout,omitempty"`
}

// PoliteConfig controls rate limiting and ethical scraping
type PoliteConfig struct {
	DelayMS          int  `yaml:"delayMs"`
	MaxConcurrent    int  `yaml:"maxConcurrent"`
	RespectRobotsTxt bool `yaml:"respectRobotsTxt"`
}

// NavigationConfig controls how chapters are discovered
type NavigationConfig struct {
	Method           string `yaml:"method"`
	URLPattern       string `yaml:"urlPattern,omitempty"`
	NumberStart      int    `yaml:"numberStart,omitempty"`
	NumberEnd        int    `yaml:"numberEnd,omitempty"`
	NextLinkSelector string `yaml:"nextLinkSelector,omitempty"`
	TOCUrl           string `yaml:"tocUrl,omitempty"`
	TOCLinkSelector  string `yaml:"tocLinkSelector,omitempty"`
	MaxChapters      int    `yaml:"maxChapters,omitempty"`
}

// ContentDetectionConfig controls how content is extracted from pages
type ContentDetectionConfig struct {
	Strategy         string            `yaml:"strategy"`
	CSSSelector      string            `yaml:"cssSelector,omitempty"`
	ExcludeSelectors []string          `yaml:"excludeSelectors,omitempty"`
	TextDensity      TextDensityConfig `yaml:"textDensity,omitempty"`
	XPathPatterns    []string          `yaml:"xpathPatterns,omitempty"`
	RegexPatterns    []string          `yaml:"regexPatterns,omitempty"`
	DOMPosition      DOMPositionConfig `yaml:"domPosition,omitempty"`
}

// TextDensityConfig controls text density analysis
type TextDensityConfig struct {
	MinDensityScore float64 `yaml:"minDensityScore"`
	MinBlockSize    int     `yaml:"minBlockSize"`
}

// DOMPositionConfig controls DOM-based content detection
type DOMPositionConfig struct {
	MaxDepth int `yaml:"maxDepth"`
	MinWidth int `yaml:"minWidth"`
}

// ChapterExtractConfig controls chapter title extraction
type ChapterExtractConfig struct {
	TitleSelector string `yaml:"titleSelector,omitempty"`
	TitleRegex    string `yaml:"titleRegex,omitempty"`
	TitleXPath    string `yaml:"titleXPath,omitempty"`
}

// OutputConfig controls output generation
type OutputConfig struct {
	Format       string             `yaml:"format"`
	OutputPath   string             `yaml:"outputPath"`
	EPUBMetadata EPUBMetadataConfig `yaml:"epubMetadata,omitempty"`
}

// EPUBMetadataConfig contains EPUB-specific metadata
type EPUBMetadataConfig struct {
	Lang       string `yaml:"lang"`
	Rights     string `yaml:"rights"`
	Publisher  string `yaml:"publisher,omitempty"`
	Identifier string `yaml:"identifier,omitempty"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Scraping: ScrapingConfig{
			Polite: PoliteConfig{
				DelayMS:          2000,
				MaxConcurrent:    1,
				RespectRobotsTxt: true,
			},
			UserAgent: "Mozilla/5.0 (compatible; WebToEPUB/1.0)",
			Timeout:   30,
		},
		Navigation: NavigationConfig{
			Method:      "url_pattern",
			NumberStart: 1,
		},
		ContentDetection: ContentDetectionConfig{
			Strategy: "hybrid",
			TextDensity: TextDensityConfig{
				MinDensityScore: 0.3,
				MinBlockSize:    100,
			},
			DOMPosition: DOMPositionConfig{
				MaxDepth: 10,
				MinWidth: 300,
			},
		},
		Output: OutputConfig{
			Format:     "epub",
			OutputPath: "./books",
			EPUBMetadata: EPUBMetadataConfig{
				Lang:   "en",
				Rights: "Personal use only",
			},
		},
	}
}
