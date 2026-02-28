package domains

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type GoldTxn struct {
	ID          string      `json:"id"`
	Bank        string      `json:"bank"`
	TxnDate     time.Time   `json:"txnDate"`
	Gram        apd.Decimal `json:"gram"` //todo consider using https://github.com/cockroachdb/apd
	UnitPrice   apd.Decimal `json:"unitPrice"`
	TotalPrice  apd.Decimal `json:"totalPrice"`
	TxnType     string      `json:"txnType"`
	EntrySource string      `json:"entrySource"`
}
