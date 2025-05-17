package analyzer

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestExtractTitle(t *testing.T) {
	htmlContent := `<html><head><title>Test Page</title></head><body></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	title := extractTitle(doc)
	if title != "Test Page" {
		t.Errorf("expected 'Test Page', got '%s'", title)
	}
}

func TestHasLoginForm(t *testing.T) {
	htmlContent := `<html><body><form><input type="text" name="user"/><input type="password"/></form></body></html>`
	doc, _ := html.Parse(strings.NewReader(htmlContent))
	if !hasLoginForm(doc) {
		t.Errorf("expected to detect login form, but did not")
	}
}

func TestDetermineHTMLVersion(t *testing.T) {
	doctypeHTML5 := `<!DOCTYPE html><html></html>`
	doc5, _ := html.Parse(strings.NewReader(doctypeHTML5))
	if determineHTMLVersion(doc5) != "HTML5" {
		t.Errorf("expected HTML5")
	}
}

func TestCheckLinkAccessibility(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := "https://example.com"
	baseURL := "https://example.com"

	accessible, err := checkLinkAccessibility(ctx, url, baseURL, client)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !accessible {
		t.Errorf("expected link to be accessible, but got false")
	}
}
