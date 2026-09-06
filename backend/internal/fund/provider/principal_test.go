package provider

import (
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cockroachdb/apd/v3"
)

// test goquery logic that scape data from testdata/principal_fund_page.html
func TestParseNavResult(t *testing.T) {
	f, err := os.Open("testdata/principal_fund_page.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatal(err)
	}

	navResult, err := parseNavFromHtml(doc)
	if err != nil {
		t.Fatal(err)
	}
	if navResult.NavDate.Format("2006-01-02") != "2026-09-03" {
		t.Errorf("NavDate = %v, want 2026-09-03", navResult.NavDate)
	}

	expected, _, _ := apd.NewFromString("0.9180")
	if navResult.Nav.Cmp(expected) != 0 {
		t.Errorf("Nav = %v, want 2026-09-03", navResult.Nav)
	}
}

func TestParseIncomeDistFromHtml(t *testing.T) {
	f, err := os.Open("testdata/principal_fund_page.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatal(err)
	}

	result, err := parseIncomeDistFromHtml(doc)
	if err != nil {
		t.Fatal(err)
	}

	wantDate := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	want := []struct {
		date  time.Time
		gross string
	}{
		{wantDate(2022, time.August, 31), "1.86"},
		{wantDate(2021, time.October, 31), "2.98"},
		{wantDate(2020, time.December, 31), "3.84"},
	}

	if len(result) != len(want) {
		t.Fatalf("got %d distributions, want %d: %+v", len(result), len(want), result)
	}

	for i, w := range want {
		got := result[i]
		if !got.DeclareDate.Equal(w.date) {
			t.Errorf("result[%d].DeclareDate = %v, want %v", i, got.DeclareDate, w.date)
		}
		if !got.PaymentDate.Equal(w.date) {
			t.Errorf("result[%d].PaymentDate = %v, want %v", i, got.PaymentDate, w.date)
		}
		wantGross, _, _ := apd.NewFromString(w.gross)
		if got.SenPerUnit.Cmp(wantGross) != 0 {
			t.Errorf("result[%d].SenPerUnit = %v, want %v", i, got.SenPerUnit, wantGross)
		}
	}
}
