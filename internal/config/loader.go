package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from a YAML or JSON file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		if err := yaml.Unmarshal(data, cfg); err != nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config (tried YAML and JSON): %w", err)
			}
		}
	}

	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// ValidateConfig validates the configuration for required fields and consistency
func ValidateConfig(cfg *Config) error {
	if cfg.Book.Title == "" {
		return fmt.Errorf("book.title is required")
	}
	if cfg.Book.Author == "" {
		return fmt.Errorf("book.author is required")
	}

	if cfg.Scraping.StartURL == "" && cfg.Navigation.Method != "toc" {
		return fmt.Errorf("scraping.startUrl is required (unless using toc navigation)")
	}

	switch cfg.Navigation.Method {
	case "url_pattern":
		if cfg.Navigation.URLPattern == "" {
			return fmt.Errorf("navigation.urlPattern is required for url_pattern method")
		}
		if cfg.Navigation.NumberEnd < cfg.Navigation.NumberStart {
			return fmt.Errorf("navigation.numberEnd must be >= navigation.numberStart")
		}
	case "next_link":
		if cfg.Navigation.NextLinkSelector == "" {
			return fmt.Errorf("navigation.nextLinkSelector is required for next_link method")
		}
	case "toc":
		if cfg.Navigation.TOCUrl == "" {
			return fmt.Errorf("navigation.tocUrl is required for toc method")
		}
		if cfg.Navigation.TOCLinkSelector == "" {
			return fmt.Errorf("navigation.tocLinkSelector is required for toc method")
		}
	default:
		return fmt.Errorf("unknown navigation method: %s (valid: url_pattern, next_link, toc)", cfg.Navigation.Method)
	}

	validStrategies := map[string]bool{
		"css_selector": true,
		"text_density": true,
		"xpath_regex":  true,
		"dom_position": true,
		"hybrid":       true,
	}
	if !validStrategies[cfg.ContentDetection.Strategy] {
		return fmt.Errorf("unknown content detection strategy: %s", cfg.ContentDetection.Strategy)
	}

	if cfg.ContentDetection.Strategy == "css_selector" && cfg.ContentDetection.CSSSelector == "" {
		return fmt.Errorf("contentDetection.cssSelector is required for css_selector strategy")
	}

	if cfg.ContentDetection.Strategy == "xpath_regex" && len(cfg.ContentDetection.RegexPatterns) == 0 {
		return fmt.Errorf("contentDetection.regexPatterns is required for xpath_regex strategy")
	}

	if cfg.Output.Format != "epub" && cfg.Output.Format != "pdf" {
		return fmt.Errorf("output.format must be 'epub' or 'pdf'")
	}

	if cfg.Scraping.Polite.DelayMS < 0 {
		return fmt.Errorf("scraping.polite.delayMs must be >= 0")
	}
	if cfg.Scraping.Polite.MaxConcurrent < 1 {
		cfg.Scraping.Polite.MaxConcurrent = 1
	}

	return nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(cfg *Config, path string) error {
	ext := strings.ToLower(filepath.Ext(path))

	var data []byte
	var err error

	switch ext {
	case ".json":
		data, err = json.MarshalIndent(cfg, "", "  ")
	default:
		data, err = yaml.Marshal(cfg)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
