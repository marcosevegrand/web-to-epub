// Package navigator provides chapter discovery and navigation functionality.
package navigator

import (
	"regexp"
	"strconv"
	"strings"
)

// GenerateURLsFromPattern generates chapter URLs from a pattern
func GenerateURLsFromPattern(pattern string, start, end int) []string {
	if end < start {
		return nil
	}

	urls := make([]string, 0, end-start+1)
	for i := start; i <= end; i++ {
		url := strings.ReplaceAll(pattern, "{number}", strconv.Itoa(i))
		url = strings.ReplaceAll(url, "{num}", strconv.Itoa(i))
		url = strings.ReplaceAll(url, "{n}", strconv.Itoa(i))
		urls = append(urls, url)
	}
	return urls
}

// ExtractChapterNumber attempts to extract a chapter number from a URL
func ExtractChapterNumber(url string) (int, bool) {
	patterns := []string{
		`chapter[_-]?(\d+)`,
		`ch[_-]?(\d+)`,
		`c(\d+)`,
		`/(\d+)/?$`,
		`page[_-]?(\d+)`,
		`p(\d+)`,
		`episode[_-]?(\d+)`,
		`ep[_-]?(\d+)`,
	}

	urlLower := strings.ToLower(url)
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(urlLower)
		if len(matches) > 1 {
			num, err := strconv.Atoi(matches[1])
			if err == nil {
				return num, true
			}
		}
	}
	return 0, false
}

// NormalizeURL ensures URL has proper format
func NormalizeURL(url string) string {
	url = strings.TrimSpace(url)

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	url = strings.TrimSuffix(url, "/")

	return url
}

// ResolveRelativeURL resolves a relative URL against a base URL
func ResolveRelativeURL(baseURL, relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL
	}

	if strings.HasPrefix(relativeURL, "//") {
		if strings.HasPrefix(baseURL, "https://") {
			return "https:" + relativeURL
		}
		return "http:" + relativeURL
	}

	baseURL = NormalizeURL(baseURL)

	if strings.HasPrefix(relativeURL, "/") {
		re := regexp.MustCompile(`^(https?://[^/]+)`)
		matches := re.FindStringSubmatch(baseURL)
		if len(matches) > 1 {
			return matches[1] + relativeURL
		}
	}

	lastSlash := strings.LastIndex(baseURL, "/")
	if lastSlash > 8 {
		return baseURL[:lastSlash+1] + relativeURL
	}

	return baseURL + "/" + relativeURL
}

// IsValidURL checks if a string looks like a valid URL
func IsValidURL(url string) bool {
	url = strings.TrimSpace(url)
	if url == "" {
		return false
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	re := regexp.MustCompile(`^https?://[a-zA-Z0-9][a-zA-Z0-9-]*\.[a-zA-Z]{2,}`)
	return re.MatchString(url)
}

// ExtractDomain extracts the domain from a URL
func ExtractDomain(url string) string {
	re := regexp.MustCompile(`^https?://([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
