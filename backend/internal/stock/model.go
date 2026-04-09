package stock

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/apd/v3"
)

type Stock struct {
	StockCode   string `json:"stockCode"`
	DisplayName string `json:"displayName"`
}

type Txn struct {
	ID         string      `json:"id"`
	StockCode  string      `json:"stockCode"`
	TxnDate    time.Time   `json:"txnDate"`
	Unit       apd.Decimal `json:"unit"`
	UnitPrice  apd.Decimal `json:"unitPrice"`
	BrokerFee  apd.Decimal `json:"brokerFee"`
	TotalPrice apd.Decimal `json:"totalPrice"`
	TxnType    string      `json:"txnType"`
	Remark     string      `json:"remark"`
}

type Dividend struct {
	StockCode string      `json:"stockCode"`
	TxnDate   time.Time   `json:"txnDate"`
	Amount    apd.Decimal `json:"amount"`
}

// UnmarshalJSON Custom JSON unmarshaling for date parsing,
// auto called when ShouldBindJSON called
func (t *Txn) UnmarshalJSON(data []byte) error {
	type Alias Txn
	aux := &struct {
		TxnDate string `json:"txnDate"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse date in YYYY-MM-DD format (funny go format, probably has to memorize it)
	parsedTime, err := time.Parse("2006-01-02", aux.TxnDate)
	if err != nil {
		return err
	}

	t.TxnDate = parsedTime
	return nil
}

// CalculateTotalPrice computes totalPrice = (unitPrice * unit) + brokerFee
func (t *Txn) CalculateTotalPrice(ctx *apd.Context) error {
	totalPrice := new(apd.Decimal)

	// Multiply unitPrice * unit
	if _, err := ctx.Mul(totalPrice, &t.UnitPrice, &t.Unit); err != nil {
		return err
	}

	// Add brokerFee
	if _, err := ctx.Add(totalPrice, totalPrice, &t.BrokerFee); err != nil {
		return err
	}

	t.TotalPrice = *totalPrice
	return nil
}
