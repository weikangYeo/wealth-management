package provider

import (
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cockroachdb/apd/v3"
)

func TestParseNavFromHtml(t *testing.T) {
	file, err := os.Open("testdata/ta_nav_page.html")
	if err != nil {
		t.Errorf("could not open testdata/ta_nav_page.html, %v", err)
	}
	defer file.Close()
	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		t.Errorf("could not read testdata/ta_nav_page.html, %v", err)
	}
	navResult, err := TaFundProvider{}.parseNavFromHtml(doc, "TA GLOBAL TECHNOLOGY FUND MYR CLASS")
	if err != nil {
		t.Fatal(err)
	}
	if navResult.NavDate.Format("2006-01-02") != "2026-08-05" {
		t.Errorf("NavDate = %v, want 2026-08-05", navResult.NavDate)
	}

	expected, _, _ := apd.NewFromString("0.5813")
	if navResult.Nav.Cmp(expected) != 0 {
		t.Errorf("Nav = %v, want 0.5813", navResult.Nav)
	}
}

func TestParseFactSheetPdfUrlFromHtml(t *testing.T) {
	file, err := os.Open("testdata/ta_fund_info.html")
	if err != nil {
		t.Errorf("could not open testdata/ta_fund_info.html, %v", err)
	}
	defer file.Close()
	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		t.Errorf("could not read testdata/ta_fund_info.html, %v", err)
	}
	href, err := TaFundProvider{}.parseFactSheetPdfUrlFromHtml(doc, "tagtf")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://www.tainvest.com.my/wp-content/pdf/fund-info/tagtf-ffs.pdf"
	if href != want {
		t.Errorf("href = %s, want %s", href, want)
	}
}

func TestFetchIncomeDistributionFullFlow(t *testing.T) {
	fundRef := FundRef{
		Name:             "TA GLOBAL TECHNOLOGY FUND MYR CLASS",
		ScrapeParamValue: "tagtf",
	}
	dist, err := TaFundProvider{}.FetchIncomeDistribution(fundRef, time.Now())
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
		{wantDate(2023, time.May, 29), "5.0"},
		{wantDate(2024, time.February, 6), "7.0"},
		{wantDate(2025, time.July, 17), "6.0"},
		{wantDate(2026, time.April, 30), "7.0"},
	}
	if len(dist) != len(want) {
		t.Fatalf("got %d distributions, want %d: %+v", len(dist), len(want), dist)
	}

	for i, w := range want {
		got := dist[i]
		if !got.DeclareDate.Equal(w.date) {
			t.Errorf("dist[%d].DeclareDate = %v, want %v", i, got.DeclareDate, w.date)
		}
		if !got.PaymentDate.Equal(w.date) {
			t.Errorf("dist[%d].PaymentDate = %v, want %v", i, got.PaymentDate, w.date)
		}
		wantGross, _, _ := apd.NewFromString(w.gross)
		if got.SenPerUnit.Cmp(wantGross) != 0 {
			t.Errorf("dist[%d].SenPerUnit = %v, want %v", i, got.SenPerUnit, wantGross)
		}
	}

}

func TestParseIncomeDistFromTextFile(t *testing.T) {
	file, err := os.Open("testdata/tagtf-ffs-layout.txt")
	result, err := TaFundProvider{}.parseIncomeDistFromTextFile(file)
	if err != nil {
		t.Errorf("could not parse income distributions: %v", err)
	}
	wantDate := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	want := []struct {
		date  time.Time
		gross string
	}{
		{wantDate(2023, time.May, 29), "5.0"},
		{wantDate(2024, time.February, 6), "7.0"},
		{wantDate(2025, time.July, 17), "6.0"},
		{wantDate(2026, time.April, 30), "7.0"},
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
