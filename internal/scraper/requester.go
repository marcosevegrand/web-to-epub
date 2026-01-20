package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"web-to-epub-go/internal/config"
)

// Requester handles HTTP requests with rate limiting and politeness
type Requester struct {
	client      *http.Client
	config      config.PoliteConfig
	userAgent   string
	lastRequest time.Time
	mutex       sync.Mutex
	robotsCache map[string]*RobotsRules
	robotsMutex sync.RWMutex
}

// RobotsRules represents parsed robots.txt rules
type RobotsRules struct {
	Disallowed []string
	CrawlDelay time.Duration
	Fetched    time.Time
}

// NewRequester creates a new HTTP requester with rate limiting
func NewRequester(politeConfig config.PoliteConfig, userAgent string, timeout int) *Requester {
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (compatible; WebToEPUB/1.0)"
	}
	if timeout <= 0 {
		timeout = 30
	}

	return &Requester{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		config:      politeConfig,
		userAgent:   userAgent,
		robotsCache: make(map[string]*RobotsRules),
	}
}

// Fetch fetches a URL with rate limiting and politeness
func (r *Requester) Fetch(ctx context.Context, targetURL string) ([]byte, error) {
	r.waitForRateLimit()

	if r.config.RespectRobotsTxt {
		allowed, err := r.isAllowedByRobots(targetURL)
		if err != nil {
			fmt.Printf("⚠ Warning: Failed to check robots.txt: %v\n", err)
		} else if !allowed {
			return nil, fmt.Errorf("URL disallowed by robots.txt: %s", targetURL)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", r.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func (r *Requester) waitForRateLimit() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delay := time.Duration(r.config.DelayMS) * time.Millisecond
	elapsed := time.Since(r.lastRequest)

	if elapsed < delay {
		time.Sleep(delay - elapsed)
	}

	r.lastRequest = time.Now()
}

func (r *Requester) isAllowedByRobots(targetURL string) (bool, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return true, err
	}

	domain := parsedURL.Scheme + "://" + parsedURL.Host
	rules, err := r.getRobotsRules(domain)
	if err != nil {
		return true, err
	}

	path := parsedURL.Path
	if path == "" {
		path = "/"
	}

	for _, disallowed := range rules.Disallowed {
		if strings.HasPrefix(path, disallowed) {
			return false, nil
		}
	}

	return true, nil
}

func (r *Requester) getRobotsRules(domain string) (*RobotsRules, error) {
	r.robotsMutex.RLock()
	rules, exists := r.robotsCache[domain]
	r.robotsMutex.RUnlock()

	if exists && time.Since(rules.Fetched) < time.Hour {
		return rules, nil
	}

	robotsURL := domain + "/robots.txt"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
	if err != nil {
		return &RobotsRules{}, err
	}
	req.Header.Set("User-Agent", r.userAgent)

	resp, err := r.client.Do(req)
	if err != nil {
		return &RobotsRules{Fetched: time.Now()}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &RobotsRules{Fetched: time.Now()}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RobotsRules{Fetched: time.Now()}, nil
	}

	rules = parseRobotsTxt(string(body))
	rules.Fetched = time.Now()

	r.robotsMutex.Lock()
	r.robotsCache[domain] = rules
	r.robotsMutex.Unlock()

	return rules, nil
}

func parseRobotsTxt(content string) *RobotsRules {
	rules := &RobotsRules{
		Disallowed: make([]string, 0),
	}

	lines := strings.Split(content, "\n")
	inUserAgentBlock := false
	isRelevantAgent := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		directive := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch directive {
		case "user-agent":
			inUserAgentBlock = true
			isRelevantAgent = value == "*" || strings.Contains(strings.ToLower(value), "webtoepub")
		case "disallow":
			if inUserAgentBlock && isRelevantAgent && value != "" {
				rules.Disallowed = append(rules.Disallowed, value)
			}
		case "crawl-delay":
			if inUserAgentBlock && isRelevantAgent {
				var delay float64
				fmt.Sscanf(value, "%f", &delay)
				rules.CrawlDelay = time.Duration(delay * float64(time.Second))
			}
		}
	}

	return rules
}

// SetTimeout sets the HTTP client timeout
func (r *Requester) SetTimeout(seconds int) {
	r.client.Timeout = time.Duration(seconds) * time.Second
}

// SetUserAgent sets the user agent string
func (r *Requester) SetUserAgent(ua string) {
	r.userAgent = ua
}
