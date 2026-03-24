package stock

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type Stock struct {
	StockCode   string `json:"stock_code"`
	DisplayName string `json:"display_name"`
}

type Txn struct {
	ID         string      `json:"id"`
	StockCode  string      `json:"stock_code"`
	TxnDate    time.Time   `json:"txnDate"`
	Unit       apd.Decimal `json:"unit"`
	UnitPrice  apd.Decimal `json:"unitPrice"`
	BrokerFee  apd.Decimal `json:"brokerFee"`
	TotalPrice apd.Decimal `json:"totalPrice"`
	TxnType    string      `json:"txnType"`
	Remark     string      `json:"remark"`
}

type Dividend struct {
	StockCode string      `json:"stock_code"`
	TxnDate   time.Time   `json:"txn_date"`
	Amount    apd.Decimal `json:"amount"`
}
