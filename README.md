# Web-to-EPUB Converter in Go

A flexible web scraper that converts multi-chapter websites into EPUB files with configurable content detection.

## Features

- **Configurable content detection** (doesn't rely solely on semantic HTML)
- **Multi-strategy content extraction**:
  - CSS selector-based extraction
  - Text density analysis (heuristic-based)
  - XPath/regex patterns
  - DOM position analysis
  - Hybrid (tries all methods in sequence)
- **Chapter navigation** (URL patterns, next-chapter links, table of contents)
- **Polite scraping** (rate limiting, robots.txt respect, user-agent)
- **EPUB generation** with proper metadata
- **Error handling & graceful degradation**
- **Configuration file support** (YAML/JSON)

## Installation

### Prerequisites

- Go 1.21 or later

### Build

```bash
# Clone or download the project
cd web-to-epub-go

# Download dependencies
go mod download

# Build the binary
go build -o scraper ./cmd/scraper

# Or install to $GOPATH/bin
go install ./cmd/scraper
```

## Quick Start

1. **Create a configuration file** (see `examples/` directory):

```yaml
book:
  title: "My Web Novel"
  author: "Author Name"

scraping:
  startUrl: "https://example.com/chapter-1"
  polite:
    delayMs: 2000
    respectRobotsTxt: true

navigation:
  method: "url_pattern"
  urlPattern: "https://example.com/chapter-{number}"
  numberStart: 1
  numberEnd: 100

contentDetection:
  strategy: "hybrid"
  cssSelector: "article.content"

output:
  format: "epub"
  outputPath: "./books"
```

2. **Run the scraper**:

```bash
# Dry run to verify configuration
./scraper --config config.yaml --dry-run

# Full scrape
./scraper --config config.yaml

# Test extraction on a single URL
./scraper --config config.yaml --test "https://example.com/chapter-1"
```

## Usage

```
Usage:
  ./scraper [options]

Options:
  -config string
        Path to configuration file (default "config.yaml")
  -output string
        Output directory (overrides config)
  -dry-run
        Parse config and show plan without scraping
  -test string
        Test extraction on a single URL
  -verbose
        Enable verbose output
  -version
        Show version information
  -help
        Show help message
```

## Project Structure

```
web-to-epub-go/
├── cmd/
│   └── scraper/
│       └── main.go              # CLI entry point
├── internal/
│   ├── config/
│   │   ├── loader.go            # Config loading & validation
│   │   └── types.go             # Config structs
│   ├── extractor/
│   │   ├── content.go           # Content utilities
│   │   ├── strategies.go        # Detection algorithms
│   │   └── types.go             # Extractor interfaces
│   ├── navigator/
│   │   ├── chapter.go           # Chapter discovery
│   │   └── patterns.go          # URL patterns
│   ├── scraper/
│   │   ├── scraper.go           # Main orchestration
│   │   └── requester.go         # HTTP with rate limiting
│   ├── formatter/
│   │   ├── html.go              # HTML normalization
│   │   └── text.go              # Text processing
│   └── output/
│       ├── epub.go              # EPUB generation
│       └── metadata.go          # Book metadata
├── examples/
│   ├── config_textdensity.yaml
│   ├── config_selector.yaml
│   ├── config_hybrid.yaml
│   ├── config_nextlink.yaml
│   └── config_toc.yaml
├── go.mod
└── README.md
```

## Configuration Reference

### Navigation Methods

#### URL Pattern
```yaml
navigation:
  method: "url_pattern"
  urlPattern: "https://example.com/chapter-{number}"
  numberStart: 1
  numberEnd: 100
```

#### Next Link
```yaml
navigation:
  method: "next_link"
  nextLinkSelector: "a.next-chapter"
  maxChapters: 500
```

#### Table of Contents
```yaml
navigation:
  method: "toc"
  tocUrl: "https://example.com/table-of-contents"
  tocLinkSelector: ".chapter-list a"
```

### Content Detection Strategies

- **css_selector**: Best for well-structured sites
- **text_density**: Best for messy HTML
- **xpath_regex**: For complex/unusual structures
- **dom_position**: Fallback using common patterns
- **hybrid**: Tries all methods in sequence (recommended)

## Legal & Ethical Notes

- Only use on content you have permission to download
- Always check the site's `robots.txt` and Terms of Service
- Add delays between requests to avoid overloading servers
- For personal use only

## License

MIT License
