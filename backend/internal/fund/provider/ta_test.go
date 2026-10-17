package provider

import (
	"os"
	"reflect"
	"testing"
	"time"
	"wealth-management/internal/platform/decimal"

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
	provider := TaFundProvider{}
	navResult, err := provider.parseNavFromHtml(doc, "TA GLOBAL TECHNOLOGY FUND MYR CLASS")
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
	provider := TaFundProvider{}
	href, err := provider.parseFactSheetPdfUrlFromHtml(doc, "tagtf")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://www.tainvest.com.my/wp-content/pdf/fund-info/tagtf-ffs.pdf"
	if href != want {
		t.Errorf("href = %s, want %s", href, want)
	}
}

func TestParseIncomeDistFromTextFile(t *testing.T) {
	file, err := os.Open("testdata/tagtf-ffs-layout.txt")
	provider := TaFundProvider{}
	result, err := provider.parseIncomeDistFromTextFile(file)
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

type taNavLookupMock struct {
	navResult []NavResult
	resultIdx int
}

func (t *taNavLookupMock) FetchNavByDate(fund FundRef, date time.Time) (*NavResult, error) {
	if t.resultIdx >= len(t.navResult) {
		return &t.navResult[len(t.navResult)], nil
	}
	i := t.resultIdx
	t.resultIdx++
	return &t.navResult[i], nil
}

func TestGetValidIncomeDist(t *testing.T) {

	mockedDigit1, err := decimal.ToDecimal("1.0")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name             string
		fromDate         time.Time
		dueDate          time.Time
		mockedFullDists  []DistributionResult
		mockedNavResults []NavResult
		expected         []DistributionResult
	}{
		{
			name:     "All income distribution after fromDate, paymentDate is updated base on updated nav date",
			fromDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
			dueDate:  time.Date(2023, time.June, 1, 0, 0, 0, 0, time.UTC),
			mockedFullDists: []DistributionResult{
				{
					DeclareDate: time.Date(2021, time.April, 2, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2021, time.April, 2, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
				{
					DeclareDate: time.Date(2022, time.April, 1, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2022, time.April, 1, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
			},
			mockedNavResults: []NavResult{
				{
					NavDate: time.Date(2021, time.April, 3, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
				{
					NavDate: time.Date(2022, time.April, 2, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
			},
			expected: []DistributionResult{
				{
					DeclareDate: time.Date(2021, time.April, 2, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2021, time.April, 3, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
				{
					DeclareDate: time.Date(2022, time.April, 1, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2022, time.April, 2, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
			},
		},
		{
			name:     "income distribution before fromDate is not considered",
			fromDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
			dueDate:  time.Date(2023, time.June, 1, 0, 0, 0, 0, time.UTC),
			mockedFullDists: []DistributionResult{
				{
					DeclareDate: time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
				{
					DeclareDate: time.Date(2022, time.April, 1, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2022, time.April, 1, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
			},
			mockedNavResults: []NavResult{
				{
					NavDate: time.Date(2022, time.April, 2, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
			},
			expected: []DistributionResult{
				{
					DeclareDate: time.Date(2022, time.April, 1, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2022, time.April, 2, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
			},
		},
		{
			name:     "Nav refresh on 3 days later",
			fromDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
			dueDate:  time.Date(2023, time.June, 1, 0, 0, 0, 0, time.UTC),
			mockedFullDists: []DistributionResult{
				{
					DeclareDate: time.Date(2021, time.April, 2, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2021, time.April, 2, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
			},
			mockedNavResults: []NavResult{
				{
					NavDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
				{
					NavDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
				{
					NavDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
				{
					NavDate: time.Date(2021, time.April, 4, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
			},
			expected: []DistributionResult{
				{
					DeclareDate: time.Date(2021, time.April, 2, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2021, time.April, 4, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
			},
		},
		{
			name:     "income distribution where its nav date after navDueDate is not considered",
			fromDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
			dueDate:  time.Date(2021, time.April, 3, 0, 0, 0, 0, time.UTC),
			mockedFullDists: []DistributionResult{
				{
					DeclareDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
					PaymentDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
					SenPerUnit:  mockedDigit1,
				},
			},
			mockedNavResults: []NavResult{
				{
					NavDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
				{
					NavDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
				{
					NavDate: time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
				{
					NavDate: time.Date(2021, time.April, 4, 0, 0, 0, 0, time.UTC),
					Nav:     mockedDigit1,
				},
			},
			expected: []DistributionResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookupMock := taNavLookupMock{
				navResult: tt.mockedNavResults,
			}
			got, err := getValidIncomeDist(FundRef{}, tt.fromDate, tt.dueDate, tt.mockedFullDists, &lookupMock)
			if err != nil {
				t.Fatalf("getValidIncomeDist() error = %v, wantErr %v", err, tt.expected)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("getValidIncomeDist() got = %v, want %v", got, tt.expected)
			}
		})
	}

}
