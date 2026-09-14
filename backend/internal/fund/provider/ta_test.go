package provider

import (
	"os"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
)

func TestParseIncomeDistFromTextFile(t *testing.T) {
	file, err := os.Open("testdata/tagtf-ffs-layout.txt")
	if err != nil {
		t.Errorf("could not open test file: %v", err)
	}
	defer file.Close()

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
