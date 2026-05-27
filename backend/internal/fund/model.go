package fund

import (
	"encoding/json"
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
	FundCode      string      `json:"fundCode"`
	PriceDate     time.Time   `json:"priceDate"`
	LastDonePrice apd.Decimal `json:"lastDonePrice"`
}

type Txn struct {
	ID              string      `json:"id"`
	FundCode        string      `json:"fundCode"`
	TxnDate         time.Time   `json:"txnDate"`
	Unit            apd.Decimal `json:"unit"`
	UnitPrice       apd.Decimal `json:"unitPrice"`
	SalesCharge     apd.Decimal `json:"salesCharge"`
	GrossTotalPrice apd.Decimal `json:"grossTotalPrice"`
	NetTotalPrice   apd.Decimal `json:"netTotalPrice"`
	TxnType         string      `json:"txnType"`
	Remark          string      `json:"remark"`
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
		Unit            json.Number `json:"unit"`
		UnitPrice       json.Number `json:"unitPrice"`
		SalesCharge     json.Number `json:"salesCharge"`
		GrossTotalPrice json.Number `json:"grossTotalPrice"`
		NetTotalPrice   json.Number `json:"netTotalPrice"`
		*Alias
	}{
		Unit:            json.Number(t.Unit.String()),
		UnitPrice:       json.Number(t.UnitPrice.String()),
		SalesCharge:     json.Number(t.SalesCharge.String()),
		GrossTotalPrice: json.Number(t.GrossTotalPrice.String()),
		NetTotalPrice:   json.Number(t.NetTotalPrice.String()),
		Alias:           (*Alias)(t),
	})
}

// CalculateTxnTotals computes grossTotalPrice = unit * unitPrice, netTotalPrice = grossTotalPrice + salesCharge
func (t *Txn) CalculateTxnTotals(ctx *apd.Context) error {
	gross := new(apd.Decimal)
	if _, err := ctx.Mul(gross, &t.Unit, &t.UnitPrice); err != nil {
		return err
	}

	net := new(apd.Decimal)
	if _, err := ctx.Sub(net, gross, &t.SalesCharge); err != nil {
		return err
	}

	t.GrossTotalPrice = *gross
	t.NetTotalPrice = *net
	return nil
}
