package analyzer

import (
	"errors"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"site-analyzer/internal/models"
)

func AnalyzePage(url string) (*models.AnalysisResult, error) {
	resp, error := http.Get(url)
	if error != nil {
		return nil, errors.New("Failed to fetch url :" + error.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("HTTP error: " + resp.Status)
	}
	doc, error := html.Parse(resp.Body)
	if error != nil {
		return nil, errors.New("Error parsing HTML document: " + error.Error())
	}

	result := &models.AnalysisResult{
		Title:       extractTitle(doc),
		HTMLVersion: determineHTMLVersion(doc),
		Headings:    countHeadings(doc),
		LoginForm:   hasLoginForm(doc),
	}

	links := extractLinks(doc, url)

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

func extractLinks(n *html.Node, baseURL string) []models.Link {
	var links []models.Link
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					internal := strings.HasPrefix(href, "/") || strings.Contains(href, baseURL)
					accessible := checkLinkAccessibility(href, baseURL)
					links = append(links, models.Link{Href: href, Internal: internal, Accessible: accessible})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return links
}

func checkLinkAccessibility(href, base string) bool {
	if strings.HasPrefix(href, "/") {
		href = base + href
	}
	resp, err := http.Head(href)
	if err != nil || resp.StatusCode >= 400 {
		return false
	}
	return true
}
