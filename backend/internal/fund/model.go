package fund

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/apd/v3"
)

type Fund struct {
	FundCode string `json:"fundCode"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Provider string `json:"provider"`
}

type PriceHistory struct {
	FundCode  string      `json:"fundCode"`
	PriceDate time.Time   `json:"priceDate"`
	Nav       apd.Decimal `json:"nav"`
}

type Txn struct {
	ID                  string      `json:"id"`
	FundCode            string      `json:"fundCode"`
	TxnDate             time.Time   `json:"txnDate"`
	Unit                apd.Decimal `json:"unit"`
	UnitPrice           apd.Decimal `json:"unitPrice"`
	SalesCharge         apd.Decimal `json:"salesCharge"`
	NetInvestmentAmount apd.Decimal `json:"netInvestmentAmount"`
	TotalAmount         apd.Decimal `json:"totalAmount"`
	TxnType             string      `json:"txnType"`
	Remark              string      `json:"remark"`
}

type TxnRequest struct {
	TxnDate     time.Time   `json:"txnDate"`
	Unit        apd.Decimal `json:"unit"`
	UnitPrice   apd.Decimal `json:"unitPrice"`
	SalesCharge apd.Decimal `json:"salesCharge"`
	TxnType     string      `json:"txnType"`
	Remark      string      `json:"remark"`
}

func (t *TxnRequest) UnmarshalJSON(data []byte) error {
	type Alias TxnRequest
	aux := &struct {
		TxnDate     string      `json:"txnDate"`
		Unit        json.Number `json:"unit"`
		UnitPrice   json.Number `json:"unitPrice"`
		SalesCharge json.Number `json:"salesCharge"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	parsedTime, err := time.Parse("2006-01-02", aux.TxnDate)
	if err != nil {
		return err
	}
	t.TxnDate = parsedTime

	ctx := apd.BaseContext
	for _, f := range []struct {
		dst *apd.Decimal
		src json.Number
	}{
		{&t.Unit, aux.Unit},
		{&t.UnitPrice, aux.UnitPrice},
		{&t.SalesCharge, aux.SalesCharge},
	} {
		if _, _, err := ctx.SetString(f.dst, f.src.String()); err != nil {
			return err
		}
	}
	return nil
}

func (t *Txn) MarshalJSON() ([]byte, error) {
	type Alias Txn
	return json.Marshal(&struct {
		Unit                json.Number `json:"unit"`
		UnitPrice           json.Number `json:"unitPrice"`
		SalesCharge         json.Number `json:"salesCharge"`
		NetInvestmentAmount json.Number `json:"netInvestmentAmount"`
		TotalAmount         json.Number `json:"totalAmount"`
		*Alias
	}{
		Unit:                json.Number(t.Unit.String()),
		UnitPrice:           json.Number(t.UnitPrice.String()),
		SalesCharge:         json.Number(t.SalesCharge.String()),
		NetInvestmentAmount: json.Number(t.NetInvestmentAmount.String()),
		TotalAmount:         json.Number(t.TotalAmount.String()),
		Alias:               (*Alias)(t),
	})
}

// CalculateTxnTotals computes netInvestmentAmount = unit * unitPrice (value of units transacted), and totalAmount,
// the actual cash that moved: netInvestmentAmount + salesCharge on BUY (fee paid on top), netInvestmentAmount -
// salesCharge on SELL (a redemption fee reduces proceeds), or 0 on REINVESTED (no fresh cash changes hands).
func (t *Txn) CalculateTxnTotals(ctx *apd.Context) error {
	netInvestmentAmount := new(apd.Decimal)
	if _, err := ctx.Mul(netInvestmentAmount, &t.Unit, &t.UnitPrice); err != nil {
		return err
	}
	t.NetInvestmentAmount = *netInvestmentAmount

	totalAmount := new(apd.Decimal)
	switch t.TxnType {
	case "BUY":
		if _, err := ctx.Add(totalAmount, netInvestmentAmount, &t.SalesCharge); err != nil {
			return err
		}
	case "SELL":
		if _, err := ctx.Sub(totalAmount, netInvestmentAmount, &t.SalesCharge); err != nil {
			return err
		}
	case "REINVESTED":
		*totalAmount = *apd.New(0, 0)
	default:
		return fmt.Errorf("unknown txn type: %s", t.TxnType)
	}
	t.TotalAmount = *totalAmount
	return nil
}
