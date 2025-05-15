package analyzer

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"site-analyzer/internal/models"
)

const maxConcurrentRequests = 10

type linkResult struct {
	Link models.Link
	Err  error
}

func AnalyzePage(ctx context.Context, url string) (*models.AnalysisResult, error) {

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	//fetch the web page
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New("failed to create request: " + err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to fetch URL: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("HTTP error: " + resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, errors.New("error parsing HTML document: " + err.Error())
	}

	result := &models.AnalysisResult{
		Title:       extractTitle(doc),
		HTMLVersion: determineHTMLVersion(doc),
		Headings:    countHeadings(doc),
		LoginForm:   hasLoginForm(doc),
	}
	//
	links := extractLinks(ctx, doc, url, client)

	// Count link types
	for _, link := range links {
		if link.Internal {
			result.Internal++
		} else {
			result.External++
		}
		if !link.Accessible {
			result.Inaccessible++
		}
	}

	return result, nil
}

func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
		return n.FirstChild.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title := extractTitle(c)
		if title != "" {
			return title
		}
	}
	return "N/A"
}

func countHeadings(n *html.Node) map[string]int {
	counts := make(map[string]int)
	validHeadings := map[string]bool{
		"h1": true, "h2": true, "h3": true,
		"h4": true, "h5": true, "h6": true,
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && validHeadings[n.Data] {
			counts[n.Data]++
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	return counts
}

func hasLoginForm(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "input" {
		for _, attr := range n.Attr {
			if attr.Key == "type" && attr.Val == "password" {
				return true
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasLoginForm(c) {
			return true
		}
	}
	return false
}

func determineHTMLVersion(n *html.Node) string {
	for c := n; c != nil; c = c.NextSibling {
		if c.Type == html.DoctypeNode {
			if strings.Contains(strings.ToLower(c.Data), "html") {
				return "HTML5"
			}
			return "Older HTML version"
		}
	}
	return "Unknown"
}

func extractLinks(ctx context.Context, n *html.Node, baseURL string, client *http.Client) []models.Link {
	var links []models.Link
	var mu sync.Mutex // Protects the links slice

	// Collect links
	var collectLinks func(*html.Node)
	collectLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					internal := strings.HasPrefix(href, "/") || strings.Contains(href, baseURL)
					// Append link without accessibility info (will be set later)
					mu.Lock()
					links = append(links, models.Link{Href: href, Internal: internal})
					mu.Unlock()
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collectLinks(c)
		}
	}
	collectLinks(n)

	// Check accessibility concurrently
	results := checkLinksAccessibility(ctx, links, baseURL, client)

	// Update links with accessibility results
	for i, result := range results {
		links[i].Accessible = result.Err == nil && result.Link.Accessible
	}

	return links
}

func checkLinksAccessibility(ctx context.Context, links []models.Link, baseURL string, client *http.Client) []linkResult {
	var wg sync.WaitGroup
	results := make([]linkResult, len(links))
	channel := make(chan struct{}, maxConcurrentRequests) // Limits concurrency

	for i, link := range links {
		wg.Add(1)
		go func(i int, link models.Link) {
			defer wg.Done()

			// Acquire channel
			select {
			case channel <- struct{}{}:
				defer func() { <-channel }()
			case <-ctx.Done():
				results[i] = linkResult{Link: link, Err: ctx.Err()}
				return
			}

			// Check accessibility
			accessible, err := checkLinkAccessibility(ctx, link.Href, baseURL, client)
			results[i] = linkResult{
				Link: models.Link{Href: link.Href, Internal: link.Internal, Accessible: accessible},
				Err:  err,
			}
		}(i, link)
	}

	wg.Wait()
	return results
}

// checkLinkAccessibility checks if a single link is accessible using an HTTP HEAD request.
func checkLinkAccessibility(ctx context.Context, href, baseURL string, client *http.Client) (bool, error) {
	// Resolve relative URLs
	if strings.HasPrefix(href, "/") {
		href = baseURL + href
	}

	// Skip invalid URLs
	if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
		return false, errors.New("invalid URL scheme")
	}

	// Create HEAD request
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, href, nil)
	if err != nil {
		return false, err
	}

	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode < 400, nil
}
