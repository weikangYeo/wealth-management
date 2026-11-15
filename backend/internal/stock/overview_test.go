package stock

import (
	"log"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
)

func TestGetTotalDividend(t *testing.T) {
	given := []Dividend{
		{
			StockName:       "CIMB",
			ExDate:          time.Date(2021, time.January, 1, 1, 0, 0, 0, time.UTC),
			PaymentDate:     time.Date(2021, time.March, 1, 1, 0, 0, 0, time.UTC),
			StockUnit:       *toDecimal("200"),
			DividendPerUnit: *toDecimal("1"),
			GrossAmount:     *toDecimal("200"),
			NetAmount:       *toDecimal("200"),
		},
		{
			StockName:       "CIMB",
			ExDate:          time.Date(2023, time.January, 1, 1, 0, 0, 0, time.UTC),
			PaymentDate:     time.Date(2023, time.March, 1, 1, 0, 0, 0, time.UTC),
			StockUnit:       *toDecimal("400"),
			DividendPerUnit: *toDecimal("1"),
			GrossAmount:     *toDecimal("400"),
			NetAmount:       *toDecimal("400"),
		},
	}

	result, err := getTotalDividend(given)
	if err != nil {
		t.Fatal(err)
	}
	if result.Cmp(toDecimal("600")) != 0 {
		t.Fatalf("want %v, got %v", toDecimal("600"), result)
	}
}

func TestComputeOverview(t *testing.T) {
	given := []Txn{
		{
			ID:          "some-uuid1",
			StockName:   "CIMB",
			TxnDate:     time.Date(2021, time.January, 1, 1, 0, 0, 0, time.UTC),
			Unit:        *toDecimal("300"),
			UnitPrice:   *toDecimal("7.45"),
			BrokerFee:   *toDecimal("10"),
			TotalAmount: *toDecimal("2245"),
			TxnType:     "BUY",
		},
		{
			ID:          "some-uuid2",
			StockName:   "CIMB",
			TxnDate:     time.Date(2022, time.January, 1, 1, 0, 0, 0, time.UTC),
			Unit:        *toDecimal("800"),
			UnitPrice:   *toDecimal("7.79"),
			BrokerFee:   *toDecimal("13"),
			TotalAmount: *toDecimal("6245"),
			TxnType:     "BUY",
		},
		{
			ID:          "some-uuid3",
			StockName:   "CIMB",
			TxnDate:     time.Date(2023, time.January, 1, 1, 0, 0, 0, time.UTC),
			Unit:        *toDecimal("200"),
			UnitPrice:   *toDecimal("8"),
			BrokerFee:   *toDecimal("20"),
			TotalAmount: *toDecimal("1580"),
			TxnType:     "SELL",
		},
	}
	mockStock := Stock{
		StockName:    "CIMB",
		DisplayName:  "CIMB",
		BursaStockId: "1234",
	}
	givenDividend := []Dividend{
		{
			StockName:       "CIMB",
			PaymentDate:     time.Date(2021, time.December, 1, 1, 0, 0, 0, time.UTC),
			StockUnit:       *toDecimal("300"),
			DividendPerUnit: *toDecimal("100"),
			GrossAmount:     *toDecimal("300"),
			NetAmount:       *toDecimal("300"),
		},
		{
			StockName:       "CIMB",
			PaymentDate:     time.Date(2022, time.December, 1, 1, 0, 0, 0, time.UTC),
			StockUnit:       *toDecimal("800"),
			DividendPerUnit: *toDecimal("100"),
			GrossAmount:     *toDecimal("1100"),
			NetAmount:       *toDecimal("1100"),
		},
	}
	lastDonePrice := toDecimal("2.1")
	today := time.Date(2023, time.December, 1, 1, 0, 0, 0, time.UTC)
	result, err := computeOverview(mockStock, given, givenDividend, lastDonePrice, today)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalUnit.Cmp(toDecimal("900")) != 0 {
		t.Fatalf("want %v, got %v", toDecimal("900"), &result.TotalUnit)
	}
	// by right average price wont be changed even there is a sell txn
	if result.AveragePrice.Cmp(toDecimal("7.7181818181818")) != 0 {
		t.Fatalf("want %v, got %v", toDecimal("7.7181818181818"), &result.AveragePrice)
	}
	if result.RealizedGainLoss.Cmp(toDecimal("36.3636363636")) != 0 {
		t.Fatalf("want %v, got %v", toDecimal("36.3636363636"), &result.RealizedGainLoss)
	}
	if result.DividendPayback.Cmp(toDecimal("0.16489988221437")) != 0 {
		t.Fatalf("want %v, got %v", toDecimal("0.16489988221437"), &result.DividendPayback)
	}
}

func TestGetDividendPayback(t *testing.T) {
	tests := []struct {
		name          string
		givenPosition Position
		givenTotalDiv *apd.Decimal
		expected      *apd.Decimal
	}{
		{
			name: "holding, dividends divided by total invested",
			givenPosition: Position{
				unit:          *toDecimal("100"),
				cost:          *toDecimal("1000"),
				totalInvested: *toDecimal("1000"),
			},
			givenTotalDiv: toDecimal("100"),
			expected:      toDecimal("0.1"),
		},
		{
			// cost (500) shrank after the sell; dividing by it would give 0.2
			name: "after partial sell, divides by total invested, not remaining cost",
			givenPosition: Position{
				unit:          *toDecimal("50"),
				cost:          *toDecimal("500"),
				totalInvested: *toDecimal("1000"),
			},
			givenTotalDiv: toDecimal("100"),
			expected:      toDecimal("0.1"),
		},
		{
			name: "fully sold, still computed from total invested",
			givenPosition: Position{
				totalInvested: *toDecimal("1000"),
			},
			givenTotalDiv: toDecimal("250"),
			expected:      toDecimal("0.25"),
		},
		{
			// e.g. dividend in specie received at 0 cost; current behaviour returns 0
			name:          "nothing invested, returns nil",
			givenPosition: Position{},
			givenTotalDiv: toDecimal("100"),
			expected:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getDividendPayback(tt.givenPosition, tt.givenTotalDiv)
			if err != nil {
				t.Fatal(err)
			}
			assertDecimal(t, tt.expected, result)
		})
	}
}

func TestGetUnrealizedGainLoss(t *testing.T) {
	tests := []struct {
		name               string
		givenPosition      Position
		givenLastDonePrice *apd.Decimal
		expected           *apd.Decimal
	}{
		{
			name: "happy scenario, with last done price present, expected a result",
			givenPosition: Position{
				unit:         *toDecimal("200"),
				cost:         *toDecimal("1794"),
				realizedGain: apd.Decimal{},
			},
			givenLastDonePrice: toDecimal("10.120"),
			expected:           toDecimal("230.00"),
		},
		{
			name: "no last done price, return nil",
			givenPosition: Position{
				unit:         *toDecimal("200"),
				cost:         *toDecimal("1794"),
				realizedGain: apd.Decimal{},
			},
			givenLastDonePrice: nil,
			expected:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.givenPosition.getUnrealizedGainLoss(tt.givenLastDonePrice)
			if err != nil {
				t.Fatal(err)
			}
			expected := tt.expected
			if expected != nil {
				if result.Cmp(expected) != 0 {
					t.Fatalf("want %v, got %v", expected, result)
				}
			} else {
				if result != nil {
					t.Fatalf("want nil, got %v", result)
				}
			}

		})
	}

}

func TestGetRoi(t *testing.T) {
	tests := []struct {
		name               string
		givenPosition      Position
		givenUnrealizedGL  *apd.Decimal
		givenTotalDividend *apd.Decimal
		expected           *apd.Decimal
	}{
		{
			// (230 + 2000 + 200) / 1794
			name: "holding with realized gain and dividends",
			givenPosition: Position{
				unit:          *toDecimal("200"),
				cost:          *toDecimal("1794"),
				totalInvested: *toDecimal("1794"),
				realizedGain:  *toDecimal("2000"),
			},
			givenUnrealizedGL:  toDecimal("230.00"),
			givenTotalDividend: toDecimal("200"),
			expected:           toDecimal("1.3545150501672"),
		},
		{
			// buy 1000 @ 2.00, buy 1000 @ 3.00, sell 500 @ 4.00, price now 3.00:
			// unrealized = 1500 × 3 − 3765 = 735, realized = 735, (735 + 735) / 5000.
			// dividing by remaining cost (3765) would give 0.39043824701195
			name: "after partial sell, divides by total invested, not remaining cost",
			givenPosition: Position{
				unit:          *toDecimal("1500"),
				cost:          *toDecimal("3765"),
				totalInvested: *toDecimal("5000"),
				realizedGain:  *toDecimal("735"),
			},
			givenUnrealizedGL:  toDecimal("735"),
			givenTotalDividend: toDecimal("0"),
			expected:           toDecimal("0.294"),
		},
		{
			// (0 + 200 + 50) / 1000
			name: "fully sold, unrealized is 0 and roi is still computed",
			givenPosition: Position{
				totalInvested: *toDecimal("1000"),
				realizedGain:  *toDecimal("200"),
			},
			givenUnrealizedGL:  toDecimal("0"),
			givenTotalDividend: toDecimal("50"),
			expected:           toDecimal("0.25"),
		},
		{
			name: "no unrealized gain loss (no last done price), return nil",
			givenPosition: Position{
				unit:          *toDecimal("200"),
				cost:          *toDecimal("1794"),
				totalInvested: *toDecimal("1794"),
			},
			givenUnrealizedGL:  nil,
			givenTotalDividend: toDecimal("200"),
			expected:           nil,
		},
		{
			name:               "nothing invested, return nil",
			givenPosition:      Position{},
			givenUnrealizedGL:  toDecimal("0"),
			givenTotalDividend: toDecimal("100"),
			expected:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.givenPosition.getRoi(tt.givenUnrealizedGL, *tt.givenTotalDividend)
			if err != nil {
				t.Fatal(err)
			}
			assertDecimal(t, tt.expected, result)
		})
	}
}

// assertDecimal fails the test unless got and want are both nil, or both non-nil and numerically equal.
func assertDecimal(t *testing.T, want, got *apd.Decimal) {
	t.Helper()
	switch {
	case want == nil && got == nil:
		return
	case want == nil:
		t.Fatalf("want nil, got %v", got)
	case got == nil:
		t.Fatalf("want %v, got nil", want)
	case got.Cmp(want) != 0:
		t.Fatalf("want %v, got %v", want, got)
	}
}

func toDecimal(str string) *apd.Decimal {
	d, _, err := apd.NewFromString(str)
	if err != nil {
		log.Fatal(err)
	}
	return d
}
