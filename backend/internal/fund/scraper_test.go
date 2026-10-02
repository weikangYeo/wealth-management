package fund

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

// mockRepo impl pullIncomeDistStartDateRepo, each field controls what one method returns.
type mockRepo struct {
	latestDate        time.Time
	latestReinvestErr error
	oldestDate        time.Time
	oldestTxnErr      error
}

func (r mockRepo) getLatestReinvestedTxnDate(fundCode string) (time.Time, error) {
	if r.latestReinvestErr != nil {
		return time.Time{}, r.latestReinvestErr
	}
	return r.latestDate, nil
}

func (r mockRepo) getOldestTxnDate(fundCode string) (time.Time, error) {
	if r.oldestTxnErr != nil {
		return time.Time{}, r.oldestTxnErr
	}
	return r.oldestDate, nil
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
			got, err := getIncomeDistPullStartDate(tt.repo, "any")

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
