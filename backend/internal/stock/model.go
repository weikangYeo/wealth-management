package stock

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/apd/v3"
)

type Stock struct {
	StockName    string `json:"stockName"`
	DisplayName  string `json:"displayName"`
	BursaStockId int    `json:"bursaStockId"`
}

type Txn struct {
	ID         string      `json:"id"`
	StockName  string      `json:"stockName"`
	TxnDate    time.Time   `json:"txnDate"`
	Unit       apd.Decimal `json:"unit"`
	UnitPrice  apd.Decimal `json:"unitPrice"`
	BrokerFee  apd.Decimal `json:"brokerFee"`
	TotalPrice apd.Decimal `json:"totalPrice"`
	TxnType    string      `json:"txnType"`
	Remark     string      `json:"remark"`
}

type TxnRequest struct {
	TxnDate   time.Time   `json:"txnDate"`
	Unit      apd.Decimal `json:"unit"`
	UnitPrice apd.Decimal `json:"unitPrice"`
	BrokerFee apd.Decimal `json:"brokerFee"`
	TxnType   string      `json:"txnType"`
	Remark    string      `json:"remark"`
}

type Dividend struct {
	StockName string      `json:"stockName"`
	TxnDate   time.Time   `json:"txnDate"`
	Amount    apd.Decimal `json:"amount"`
}

type Price struct {
	StockName     string
	PriceDate     time.Time
	LastDonePrice apd.Decimal
}

// UnmarshalJSON Custom JSON unmarshaling for date parsing,
// auto called when ShouldBindJSON called
func (t *TxnRequest) UnmarshalJSON(data []byte) error {
	type Alias TxnRequest
	aux := &struct {
		TxnDate   string      `json:"txnDate"`
		Unit      json.Number `json:"unit"`      // Parse as string
		UnitPrice json.Number `json:"unitPrice"` // Parse as string
		BrokerFee json.Number `json:"brokerFee"`
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
	// Parse decimal fields from json.Number to apd.Decimal
	ctx := apd.BaseContext
	if _, _, err := ctx.SetString(&t.Unit, aux.Unit.String()); err != nil {
		return err
	}
	if _, _, err := ctx.SetString(&t.UnitPrice, aux.UnitPrice.String()); err != nil {
		return err
	}
	if _, _, err := ctx.SetString(&t.BrokerFee, aux.BrokerFee.String()); err != nil {
		return err
	}
	return nil
}

func (t *Txn) MarshalJSON() ([]byte, error) {
	type Alias Txn
	return json.Marshal(&struct {
		Unit       json.Number `json:"unit"`
		UnitPrice  json.Number `json:"unitPrice"`
		BrokerFee  json.Number `json:"brokerFee"`
		TotalPrice json.Number `json:"totalPrice"`
		*Alias
	}{
		Unit:       json.Number(t.Unit.String()),
		UnitPrice:  json.Number(t.UnitPrice.String()),
		BrokerFee:  json.Number(t.BrokerFee.String()),
		TotalPrice: json.Number(t.TotalPrice.String()),
		Alias:      (*Alias)(t),
	})
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
