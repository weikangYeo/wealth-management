package stock

import (
	"testing"
	"time"
	"wealth-management/internal/platform/decimal"

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
			TotalAmount: *toDecimal("6245"),
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
	result, err := computeOverview(mockStock, given, givenDividend, *lastDonePrice)
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
	if result.DividendPayback.Cmp(toDecimal("0.20154430048423")) != 0 {
		t.Fatalf("want %v, got %v", toDecimal("0.20154430048423"), &result.DividendPayback)
	}
}

func TestGetDividendPayback(t *testing.T) {
	position := Position{
		unit:         *toDecimal("100"),
		cost:         *toDecimal("1000"),
		realizedGain: apd.Decimal{},
	}
	totalDiv := toDecimal("100")
	result, err := getDividendPayback(position, totalDiv)
	if err != nil {
		t.Fatal(err)
	}
	expected := toDecimal("0.1")
	if result.Cmp(expected) != 0 {
		t.Fatalf("want %v, got %v", expected, &result)
	}
}

func TestGetUnrealizedGainLoss(t *testing.T) {
	position := Position{
		unit:         *toDecimal("200"),
		cost:         *toDecimal("1794"),
		realizedGain: apd.Decimal{},
	}
	lastDonePrice := toDecimal("10.120")
	result, err := position.getUnrealizedGainLoss(*lastDonePrice)
	if err != nil {
		t.Fatal(err)
	}
	expected := toDecimal("230.00")
	if result.Cmp(expected) != 0 {
		t.Fatalf("want %v, got %v", expected, result)
	}
}

func TestGetRoi(t *testing.T) {
	position := Position{
		unit:         *toDecimal("200"),
		cost:         *toDecimal("1794"),
		realizedGain: *toDecimal("2000"),
	}
	totalDividend := toDecimal("200")
	result, err := position.getRoi(*toDecimal("230.00"), *totalDividend)
	if err != nil {
		t.Fatal(err)
	}
	expected := toDecimal("1.3545150501672")
	if result.Cmp(expected) != 0 {
		t.Fatalf("want %v, got %v", expected, result)
	}
}

func toDecimal(str string) *apd.Decimal {
	d, err := decimal.ToDecimal(str)
	if err != nil {
		panic(err)
	}
	return d
}
