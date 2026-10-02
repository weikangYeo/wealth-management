package fund

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

// create mock that impl testable interface
type mockRepo struct {
	startDate         time.Time
	latestReinvestErr error
	oldestTxnErr      error
}

func (r mockRepo) getLatestReinvestedTxnDate(fundCode string) (time.Time, error) {
	if r.latestReinvestErr != nil {
		return time.Time{}, r.latestReinvestErr
	}
	return r.startDate, nil
}
func (r mockRepo) getOldestTxnDate(fundCode string) (time.Time, error) {
	if r.oldestTxnErr != nil {
		return time.Time{}, r.oldestTxnErr
	}
	return r.startDate, nil
}

func TestGetIncomeDistPullStartDateBaseOnLastReinvestedTxnDate(t *testing.T) {
	mock := mockRepo{
		startDate:         time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC),
		latestReinvestErr: nil,
		oldestTxnErr:      nil,
	}
	startDate, err := getIncomeDistPullStartDate(mock, "any")
	if err != nil {
		t.Errorf("getIncomeDistPullStartDate = %v, want no error", err)
	}
	// startDate should be latest reinvested txnDate -1 day
	// to cater variance of declaration/payment date issue from scrape source.
	if startDate != time.Date(2021, time.January, 9, 0, 0, 0, 0, time.UTC) {
		t.Errorf("getIncomeDistPullStartDate = %v, want 2021, time", startDate)
	}
}

func TestGetIncomeDistPullStartDateBaseOnOldestTxnDate(t *testing.T) {
	mock := mockRepo{
		startDate:         time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC),
		latestReinvestErr: sql.ErrNoRows,
		oldestTxnErr:      nil,
	}
	startDate, err := getIncomeDistPullStartDate(mock, "any")
	if err != nil {
		t.Errorf("getIncomeDistPullStartDate = %v, want no error", err)
	}
	if startDate != time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC) {
		t.Errorf("getIncomeDistPullStartDate = %v, want 2021, time", startDate)
	}
}

func TestGetIncomeDistPullStartDateNoTxnFound(t *testing.T) {
	mock := mockRepo{
		startDate:         time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC),
		latestReinvestErr: sql.ErrNoRows,
		oldestTxnErr:      errNoTxn,
	}
	_, err := getIncomeDistPullStartDate(mock, "any")
	if err != nil && !errors.Is(err, errNoTxn) {
		t.Errorf("err = %v, want ErrNoRows error", err)
	}
}
