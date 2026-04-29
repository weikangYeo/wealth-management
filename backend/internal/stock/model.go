package stock

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/apd/v3"
)

type Stock struct {
	StockName    string `json:"stockName"`
	DisplayName  string `json:"displayName"`
	BursaStockId string `json:"bursaStockId"`
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
	StockName       string      `json:"stockName"`
	ExDate          time.Time   `json:"exDate"`
	PaymentDate     time.Time   `json:"paymentDate"`
	StockUnit       apd.Decimal `json:"stockUnit"`
	DividendPerUnit apd.Decimal `json:"dividendPerUnit"`
	TaxPercentage   apd.Decimal `json:"taxPercentage"`
	GrossAmount     apd.Decimal `json:"grossAmount"`
	NetAmount       apd.Decimal `json:"netAmount"`
	Remark          string      `json:"remark"`
}

type DividendRequest struct {
	ExDate          time.Time   `json:"exDate"`
	PaymentDate     time.Time   `json:"paymentDate"`
	StockUnit       apd.Decimal `json:"stockUnit"`
	DividendPerUnit apd.Decimal `json:"dividendPerUnit"`
	TaxPercentage   apd.Decimal `json:"taxPercentage"`
	Remark          string      `json:"remark"`
}

type Price struct {
	StockName     string
	PriceDate     time.Time
	LastDonePrice apd.Decimal
}

// UnmarshalJSON Custom JSON unmarshaling for date parsing,
// auto called when ShouldBindJSON called.
// Go principle: types own their behavior. The standard library itself does this — time.Time
// implements MarshalJSON/UnmarshalJSON. When a type has a non-trivial wire format, the
// conversion belongs to the type, not the caller.
// with this, handler will only concern about business logic (separate of concern)
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

func (d *DividendRequest) UnmarshalJSON(data []byte) error {
	type Alias DividendRequest
	aux := &struct {
		ExDate          string      `json:"exDate"`
		PaymentDate     string      `json:"paymentDate"`
		StockUnit       json.Number `json:"stockUnit"`
		DividendPerUnit json.Number `json:"dividendPerUnit"`
		TaxPercentage   json.Number `json:"taxPercentage"`
		GrossAmount     json.Number `json:"grossAmount"`
		NetAmount       json.Number `json:"netAmount"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	if d.ExDate, err = time.Parse("2006-01-02", aux.ExDate); err != nil {
		return err
	}
	if d.PaymentDate, err = time.Parse("2006-01-02", aux.PaymentDate); err != nil {
		return err
	}
	ctx := apd.BaseContext
	for _, f := range []struct {
		dst *apd.Decimal
		src json.Number
	}{
		{&d.StockUnit, aux.StockUnit},
		{&d.DividendPerUnit, aux.DividendPerUnit},
		{&d.TaxPercentage, aux.TaxPercentage},
	} {
		if _, _, err := ctx.SetString(f.dst, f.src.String()); err != nil {
			return err
		}
	}
	return nil
}

func (dividend *Dividend) MarshalJSON() ([]byte, error) {
	type Alias Dividend
	return json.Marshal(&struct {
		StockUnit       json.Number `json:"stockUnit"`
		DividendPerUnit json.Number `json:"dividendPerUnit"`
		TaxPercentage   json.Number `json:"TaxPercentage"`
		GrossAmount     json.Number `json:"grossAmount"`
		NetAmount       json.Number `json:"netAmount"`
		*Alias
	}{
		StockUnit:       json.Number(dividend.StockUnit.String()),
		DividendPerUnit: json.Number(dividend.DividendPerUnit.String()),
		TaxPercentage:   json.Number(dividend.TaxPercentage.String()),
		GrossAmount:     json.Number(dividend.GrossAmount.String()),
		NetAmount:       json.Number(dividend.NetAmount.String()),
		Alias:           (*Alias)(dividend),
	})
}

// CalculateStockTxnTotalPrice computes totalPrice = (unitPrice * unit) + brokerFee
func (t *Txn) CalculateStockTxnTotalPrice(ctx *apd.Context) error {
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

func (dividend *Dividend) CalculateDividendTotalAmount(ctx *apd.Context) error {
	grossAmount := new(apd.Decimal)

	if _, err := ctx.Mul(grossAmount, &dividend.DividendPerUnit, &dividend.StockUnit); err != nil {
		return err
	}

	taxAmount := new(apd.Decimal)
	if _, err := ctx.Mul(taxAmount, &dividend.TaxPercentage, &dividend.GrossAmount); err != nil {
		return err
	}

	netAmount := new(apd.Decimal)
	if _, err := ctx.Sub(netAmount, grossAmount, taxAmount); err != nil {
		return err
	}

	dividend.NetAmount = *netAmount
	dividend.GrossAmount = *grossAmount
	return nil
}
