package scrape

import (
	"regexp"
	"strings"
	"testing"
)

// TestGetHtmlStringFromUrlViaChrome is a live network test against a real, currently
// bot-protected target (Imperva/Incapsula) — used to manually verify that chromedp can
// still get through. Swap targetURL to check a different fund house's page later.
func TestGetHtmlStringFromUrlViaChrome(t *testing.T) {
	targetURL := "https://www.maybank-am.com.my/list-of-funds/-/fund/view_entry/408?fromDate=2026-01-18&toDate=2026-09-14"

	html, err := GetHtmlStringFromUrlViaChrome(targetURL)
	if err != nil {
		t.Fatalf("GetHtmlStringFromUrlViaChrome error: %v", err)
	}
	t.Logf("HTML length: %d", len(html))

	titleRe := regexp.MustCompile(`(?is)<title>(.*?)</title>`)
	m := titleRe.FindStringSubmatch(html)
	if m == nil {
		t.Fatal("no <title> found in scraped html")
	}
	title := strings.TrimSpace(m[1])
	t.Logf("Title: %s", title)

	lc := strings.ToLower(html)
	blockMarkers := []string{"just a moment", "captcha", "attention required", "access denied"}
	for _, marker := range blockMarkers {
		if strings.Contains(lc, marker) {
			t.Errorf("possible bot-detection page: html contains %q", marker)
		}
	}

	if len(html) < 1000 {
		t.Errorf("html length = %d, too short to be a real content page", len(html))
	}
}
