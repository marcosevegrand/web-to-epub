// Web-to-EPUB Converter
// A flexible web scraper that converts multi-chapter websites into EPUB files.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"web-to-epub-go/internal/config"
	"web-to-epub-go/internal/scraper"
)

const (
	AppName    = "web-to-epub"
	AppVersion = "1.0.0"
)

func main() {
	var (
		configFile = flag.String("config", "config.yaml", "Path to configuration file (YAML or JSON)")
		outputPath = flag.String("output", "", "Output directory (overrides config)")
		dryRun     = flag.Bool("dry-run", false, "Parse config and show plan without scraping")
		testURL    = flag.String("test", "", "Test extraction on a single URL")
		verbose    = flag.Bool("verbose", false, "Enable verbose output")
		version    = flag.Bool("version", false, "Show version information")
		help       = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s v%s - Web-to-EPUB Converter

A flexible web scraper that converts multi-chapter websites into EPUB files.

Usage:
  %s [options]

Options:
`, AppName, AppVersion, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  # Scrape with default config
  %s --config config.yaml

  # Dry run to verify configuration
  %s --config config.yaml --dry-run

  # Test extraction on a single URL
  %s --config config.yaml --test "https://example.com/chapter-1"

  # Override output directory
  %s --config config.yaml --output ./my-books

Configuration:
  See examples/ directory for sample configuration files.
  Supported formats: YAML (.yaml, .yml) and JSON (.json)

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	}

	flag.Parse()

	if *version {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		log.Fatalf("❌ Configuration file not found: %s\n\nRun '%s --help' for usage information.", *configFile, os.Args[0])
	}

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	if *outputPath != "" {
		cfg.Output.OutputPath = *outputPath
	}

	printHeader(cfg)

	if *dryRun {
		printDryRunSummary(cfg)
		return
	}

	ws, err := scraper.NewWebScraper(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create scraper: %v", err)
	}

	if *testURL != "" {
		runTestMode(ws, *testURL, *verbose)
		return
	}

	runFullScrape(ws, cfg, *verbose)
}

func printHeader(cfg *config.Config) {
	fmt.Println()
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("📚 %s v%s\n", AppName, AppVersion)
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("📖 Book: %s\n", cfg.Book.Title)
	fmt.Printf("✍️  Author: %s\n", cfg.Book.Author)
	fmt.Printf("📍 Start URL: %s\n", cfg.Scraping.StartURL)
	fmt.Printf("🔍 Detection Strategy: %s\n", cfg.ContentDetection.Strategy)
	fmt.Printf("🧭 Navigation Method: %s\n", cfg.Navigation.Method)
	fmt.Printf("📂 Output Path: %s\n", cfg.Output.OutputPath)
	fmt.Println(strings.Repeat("─", 60))
}

func printDryRunSummary(cfg *config.Config) {
	fmt.Println("\n🔍 [DRY RUN] Configuration Summary")
	fmt.Println(strings.Repeat("─", 40))

	fmt.Println("\n📋 Scraping Settings:")
	fmt.Printf("  • Delay between requests: %dms\n", cfg.Scraping.Polite.DelayMS)
	fmt.Printf("  • Respect robots.txt: %v\n", cfg.Scraping.Polite.RespectRobotsTxt)
	fmt.Printf("  • Max concurrent: %d\n", cfg.Scraping.Polite.MaxConcurrent)

	fmt.Println("\n🧭 Navigation:")
	fmt.Printf("  • Method: %s\n", cfg.Navigation.Method)
	switch cfg.Navigation.Method {
	case "url_pattern":
		fmt.Printf("  • URL Pattern: %s\n", cfg.Navigation.URLPattern)
		fmt.Printf("  • Chapter Range: %d - %d\n", cfg.Navigation.NumberStart, cfg.Navigation.NumberEnd)
		totalChapters := cfg.Navigation.NumberEnd - cfg.Navigation.NumberStart + 1
		fmt.Printf("  • Total Chapters: %d\n", totalChapters)
	case "next_link":
		fmt.Printf("  • Next Link Selector: %s\n", cfg.Navigation.NextLinkSelector)
		if cfg.Navigation.MaxChapters > 0 {
			fmt.Printf("  • Max Chapters: %d\n", cfg.Navigation.MaxChapters)
		}
	case "toc":
		fmt.Printf("  • TOC URL: %s\n", cfg.Navigation.TOCUrl)
		fmt.Printf("  • TOC Link Selector: %s\n", cfg.Navigation.TOCLinkSelector)
	}

	fmt.Println("\n🔬 Content Detection:")
	fmt.Printf("  • Strategy: %s\n", cfg.ContentDetection.Strategy)
	if cfg.ContentDetection.CSSSelector != "" {
		fmt.Printf("  • CSS Selector: %s\n", cfg.ContentDetection.CSSSelector)
	}
	if len(cfg.ContentDetection.ExcludeSelectors) > 0 {
		fmt.Printf("  • Exclude Selectors: %v\n", cfg.ContentDetection.ExcludeSelectors)
	}

	fmt.Println("\n📤 Output:")
	fmt.Printf("  • Format: %s\n", cfg.Output.Format)
	fmt.Printf("  • Path: %s\n", cfg.Output.OutputPath)
	fmt.Printf("  • Language: %s\n", cfg.Output.EPUBMetadata.Lang)

	fmt.Println("\n✅ Configuration is valid!")
	fmt.Println("Run without --dry-run to start scraping.")
	fmt.Println()
}

func runTestMode(ws *scraper.WebScraper, testURL string, verbose bool) {
	fmt.Println("\n🧪 [TEST MODE] Testing extraction on single URL")
	fmt.Printf("📍 URL: %s\n", testURL)
	fmt.Println(strings.Repeat("─", 40))

	chapter, err := ws.ScrapeTest(testURL)
	if err != nil {
		log.Fatalf("❌ Test failed: %v", err)
	}

	fmt.Printf("\n✅ Extraction successful!\n")
	fmt.Printf("📝 Title: %s\n", chapter.Title)
	fmt.Printf("📊 Content length: %d characters\n", len(chapter.Content))

	if verbose {
		fmt.Println("\n📄 Content Preview (first 500 chars):")
		fmt.Println(strings.Repeat("─", 40))
		preview := chapter.Content
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		fmt.Println(preview)
	}

	fmt.Println()
}

func runFullScrape(ws *scraper.WebScraper, cfg *config.Config, verbose bool) {
	fmt.Println("\n⏳ Starting scrape...")
	fmt.Println()

	if err := ws.ScrapeAll(); err != nil {
		log.Fatalf("❌ Scraping failed: %v", err)
	}

	ws.PrintSummary()

	fmt.Println("\n📦 Generating EPUB...")
	if err := ws.GenerateEPUB(cfg.Output.OutputPath); err != nil {
		log.Fatalf("❌ EPUB generation failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("✅ Done!")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println()
}
