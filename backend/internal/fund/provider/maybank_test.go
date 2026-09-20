package provider

import (
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cockroachdb/apd/v3"
)

func TestParseMaybankNavFromHtml(t *testing.T) {
	f, err := os.Open("testdata/maybank_fund_nav.html")
	if err != nil {
		t.Fatalf("could not open maybank_fund_nav.html: %v", err)
	}
	defer f.Close()
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatalf("could not parse maybank_fund_nav.html: %v", err)
	}
	result, err := MaybankFundProvider{}.parseMaybankNavFromHtml(doc)
	if err != nil {
		t.Fatalf("could not parse maybank_fund_nav.html: %v", err)
	}
	expectedNav, _, _ := apd.NewFromString("1.6985")
	if result.Nav.Cmp(expectedNav) != 0 {
		t.Errorf("Nav = %v, want %v", result.Nav, expectedNav)
	}
	expectedNavDate := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	if !result.NavDate.Equal(expectedNavDate) {
		t.Errorf("NavDate = %v, want %v", result.NavDate, expectedNavDate)
	}

}
