package provider

import (
	"os"
	"testing"
	"time"
	"wealth-management/internal/platform/decimal"

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

	navResult, err := PrincipalFundProvider{}.parseNavFromHtml(doc)
	if err != nil {
		t.Fatal(err)
	}
	if navResult.NavDate.Format("2006-01-02") != "2026-09-03" {
		t.Errorf("NavDate = %v, want 2026-09-03", navResult.NavDate)
	}

	expected, _, _ := apd.NewFromString("0.9180")
	if navResult.Nav.Cmp(expected) != 0 {
		t.Errorf("Nav = %v, want 0.9180", navResult.Nav)
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

	toDate := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	toDecimal := func(decimalStr string) *apd.Decimal {
		d, err := decimal.ToDecimal(decimalStr)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	tests := []struct {
		name     string
		fromDate time.Time
		want     []DistributionResult
	}{
		{
			name:     "Pull all income dist from source",
			fromDate: toDate(2000, time.January, 1),
			want: []DistributionResult{
				{
					PaymentDate: toDate(2022, time.August, 31),
					DeclareDate: toDate(2022, time.August, 31),
					SenPerUnit:  toDecimal("1.86"),
				},
				{
					PaymentDate: toDate(2021, time.October, 31),
					DeclareDate: toDate(2021, time.October, 31),
					SenPerUnit:  toDecimal("2.98"),
				}, {
					PaymentDate: toDate(2020, time.December, 31),
					DeclareDate: toDate(2020, time.December, 31),
					SenPerUnit:  toDecimal("3.84"),
				},
			},
		},
		{
			name:     "Pull only matched income dist",
			fromDate: toDate(2022, time.January, 1),
			want: []DistributionResult{
				{
					PaymentDate: toDate(2022, time.August, 31),
					DeclareDate: toDate(2022, time.August, 31),
					SenPerUnit:  toDecimal("1.86"),
				},
			},
		},
		{
			name:     "fromDate is not inclusive",
			fromDate: toDate(2021, time.October, 31),
			want: []DistributionResult{
				{
					PaymentDate: toDate(2022, time.August, 31),
					DeclareDate: toDate(2022, time.August, 31),
					SenPerUnit:  toDecimal("1.86"),
				},
			},
		},
		{
			name:     "no result case",
			fromDate: toDate(2028, time.October, 31),
			want:     []DistributionResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseIncomeDistFromHtml(doc, tt.fromDate)
			if err != nil {
				t.Fatal(err)
			}
			if len(result) != len(tt.want) {
				t.Fatalf("got %d distributions, want %d: %+v", len(result), len(tt.want), result)
			}
			for i, w := range tt.want {
				got := result[i]
				if !got.DeclareDate.Equal(w.DeclareDate) {
					t.Errorf("result[%d].DeclareDate = %v, want %v", i, got.DeclareDate, w.DeclareDate)
				}
				if !got.PaymentDate.Equal(w.PaymentDate) {
					t.Errorf("result[%d].PaymentDate = %v, want %v", i, got.PaymentDate, w.PaymentDate)
				}
				if got.SenPerUnit.Cmp(w.SenPerUnit) != 0 {
					t.Errorf("result[%d].SenPerUnit = %v, want %v", i, got.SenPerUnit, w.SenPerUnit)
				}
			}
		})

	}
}
