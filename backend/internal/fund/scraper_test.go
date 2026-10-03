package fund

import (
	"database/sql"
	"errors"
	"testing"
	"time"
	"wealth-management/internal/fund/provider"

	"github.com/cockroachdb/apd/v3"
)

var reinvestedFoundDate = time.Date(2022, time.January, 20, 0, 0, 0, 0, time.UTC)

// mockRepo impl pullIncomeDistStartDateRepo, each field controls what one method returns.
type mockRepo struct {
	latestDate        time.Time
	latestReinvestErr error
	oldestDate        time.Time
	oldestTxnErr      error
	totalUnitToDate   apd.Decimal
	insertedFundTxn   []Txn
}

func (r *mockRepo) getLatestReinvestedTxnDate(fundCode string) (time.Time, error) {
	if r.latestReinvestErr != nil {
		return time.Time{}, r.latestReinvestErr
	}
	return r.latestDate, nil
}

func (r *mockRepo) getOldestTxnDate(fundCode string) (time.Time, error) {
	if r.oldestTxnErr != nil {
		return time.Time{}, r.oldestTxnErr
	}
	return r.oldestDate, nil
}

func (r *mockRepo) isReinvestedTxnExists(fundCode string, txnDate time.Time) (bool, error) {
	if txnDate.Equal(reinvestedFoundDate) {
		return true, nil
	}
	return false, nil
}

func (r *mockRepo) getTotalUnitToDate(fundCode string, txnDate time.Time) (*apd.Decimal, error) {
	return &r.totalUnitToDate, nil
}
func (r *mockRepo) insertFundTxn(fundTxn Txn) error {
	r.insertedFundTxn = append(r.insertedFundTxn, fundTxn)
	return nil
}

func TestGetIncomeDistPullStartDate(t *testing.T) {
	// latest and oldest are deliberately different so a case proves which source the result came from.
	latest := time.Date(2021, time.January, 20, 0, 0, 0, 0, time.UTC)
	oldest := time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		repo    mockRepo
		want    time.Time // only checked when wantErr is nil
		wantErr error     // matched with errors.Is, nil means expect success
	}{
		{
			// start date = latest reinvested txn date -1 day,
			// to cater variance of declaration/payment date issue from scrape source.
			name: "use latest reinvested txn date minus 1 day",
			repo: mockRepo{latestDate: latest, oldestDate: oldest},
			want: time.Date(2021, time.January, 19, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "no reinvested txn yet, fall back to oldest txn date",
			repo: mockRepo{latestDate: latest, latestReinvestErr: sql.ErrNoRows, oldestDate: oldest},
			want: oldest,
		},
		{
			name:    "no txn at all",
			repo:    mockRepo{latestReinvestErr: sql.ErrNoRows, oldestTxnErr: sql.ErrNoRows},
			wantErr: errNoTxn,
		},
		{
			name:    "db failure on latest reinvested lookup",
			repo:    mockRepo{latestReinvestErr: sql.ErrConnDone},
			wantErr: sql.ErrConnDone,
		},
		{
			name:    "db failure on oldest txn lookup",
			repo:    mockRepo{latestReinvestErr: sql.ErrNoRows, oldestTxnErr: sql.ErrConnDone},
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIncomeDistPullStartDate(&tt.repo, "any")

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v, want no error", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}

type mockFundProvider struct {
	IncomeDistFromSource []provider.DistributionResult
	navResult            provider.NavResult
}

func (p *mockFundProvider) FetchNavByDate(fund provider.FundRef, date time.Time) (*provider.NavResult, error) {
	return &p.navResult, nil
}

func (p *mockFundProvider) FetchIncomeDistribution(fund provider.FundRef, fromDate time.Time) ([]provider.DistributionResult, error) {
	return p.IncomeDistFromSource, nil
}

// to test dividend payout calculation logic
func TestScrapeIncomeDistHappyPathCalculationLogic(t *testing.T) {
	latest := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	oldest := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)
	totalUnit, _, err := apd.NewFromString("2000")
	if err != nil {
		t.Fatal(err)
	}
	nav, _, err := apd.NewFromString("7.5")
	if err != nil {
		t.Fatal(err)
	}
	repo := mockRepo{
		latestDate:      latest,
		oldestDate:      oldest,
		totalUnitToDate: *totalUnit,
	}
	senPerUnit, _, err := apd.NewFromString("13")
	if err != nil {
		t.Fatal(err)
	}
	paymentDate := time.Date(2023, time.January, 5, 0, 0, 0, 0, time.UTC)
	mockProvider := mockFundProvider{
		IncomeDistFromSource: []provider.DistributionResult{
			{
				DeclareDate: time.Date(2023, time.January, 2, 0, 0, 0, 0, time.UTC),
				PaymentDate: paymentDate,
				SenPerUnit:  senPerUnit,
			},
		},
		navResult: provider.NavResult{
			NavDate: reinvestedFoundDate,
			Nav:     nav,
		},
	}
	err = scrapeIncomeDist(&repo, &mockProvider, provider.FundRef{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	insertedTxns := repo.insertedFundTxn
	if len(insertedTxns) != 1 {
		t.Fatalf("Expected 1 inserted txn, got %d", len(insertedTxns))
	}
	txn := insertedTxns[0]
	if !txn.TxnDate.Equal(paymentDate) {
		t.Fatalf("Expected txn date %v, got %v", paymentDate, txn.TxnDate)
	}
	if txn.TxnType != "REINVESTED" {
		t.Fatalf("Expected txn type REINVESTED, got %s", txn.TxnType)
	}
	if !txn.TotalAmount.IsZero() {
		t.Fatalf("Expected total amount to be zero during REINVESTED, got %v", txn.TotalAmount)
	}
	if txn.UnitPrice.Cmp(nav) != 0 {
		t.Fatalf("Expected unit price to be %v, got %v", nav, txn.UnitPrice)
	}
	if !txn.SalesCharge.IsZero() {
		t.Fatalf("Expected sales charge to be zero, got %v", txn.SalesCharge)
	}
	wantUnit, _, err := apd.NewFromString("34.666666666667")
	if err != nil {
		t.Fatal(err)
	}
	if txn.Unit.Cmp(wantUnit) != 0 {
		t.Fatalf("Expected unit to be %v, got %v", wantUnit, txn.Unit)
	}
	wantNetInvestmentAmount, _, err := apd.NewFromString("260")
	if txn.NetInvestmentAmount.Cmp(wantNetInvestmentAmount) != 0 {
		t.Fatalf("Expected net investment amount to be %v, got %v", wantNetInvestmentAmount, txn.NetInvestmentAmount)
	}

}
