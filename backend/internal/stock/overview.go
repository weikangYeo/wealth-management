package stock

import (
	"fmt"
	"slices"

	"github.com/cockroachdb/apd/v3"
)

// todo unit test individual computation logic
func computeOverview(stock Stock, txns []Txn, dividends []Dividend) (Overview, error) {
	totalDiv, err := getTotalDividend(dividends)
	if err != nil {
		return Overview{}, err
	}
	position, err := computePosition(txns)
	if err != nil {
		return Overview{}, err
	}
	averageCost, err := position.getAveragePrice()
	if err != nil {
		return Overview{}, err
	}

	dividendPayback := new(apd.Decimal)
	if !position.cost.IsZero() {
		decimalCtx := apd.BaseContext.WithPrecision(14)
		if _, err := decimalCtx.Quo(dividendPayback, totalDiv, &position.cost); err != nil {
			return Overview{}, err
		}
	}

	overview := Overview{
		StockName:        stock.StockName,
		DisplayName:      stock.DisplayName,
		BursaStockId:     stock.BursaStockId,
		TotalDividend:    *totalDiv,
		AveragePrice:     averageCost,
		TotalUnit:        position.unit,
		RealizedGainLoss: position.realizedGain,
		DividendPayback:  *dividendPayback,
	}
	return overview, nil
}

// computePosition replay all transaction asc by txnDate, then compute unit, cost, realized gain at given time.
// reason this is required because realized gain need the given/exact time average cost (unit and cost) to compute
func computePosition(txns []Txn) (Position, error) {
	decimalCtx := apd.BaseContext.WithPrecision(14)
	// create a copy of param to mutate (sort), do not cause side effect
	sorted := slices.Clone(txns)
	slices.SortStableFunc(sorted, func(a, b Txn) int {
		return a.TxnDate.Compare(b.TxnDate)
	})
	var position Position
	for _, txn := range sorted {
		switch txn.TxnType {
		case "BUY":
			// during BUY, accumulate unit and cost position
			if _, err := decimalCtx.Add(&position.unit, &position.unit, &txn.Unit); err != nil {
				return Position{}, err
			}
			if _, err := decimalCtx.Add(&position.cost, &position.cost, &txn.TotalAmount); err != nil {
				return Position{}, err
			}
		case "SELL":
			if txn.Unit.Cmp(&position.unit) > 0 {
				return Position{}, fmt.Errorf("SELL of %v on %s exceed %v holding", txn.Unit, txn.TxnDate, position.unit)
			}
			// average cost = position.cost / position.unit
			// cost of sold unit =  average cost * txn.unit
			// cost of sold unit =  (position.cost / position.unit) * txn.unit
			// do multiply first then only do division to avoid rounding been multiplied (avoid inaccuracy)
			// cost of sold unit =  position.cost * txn.unit / position.unit
			costOfSoldUnits := new(apd.Decimal)
			if _, err := decimalCtx.Mul(costOfSoldUnits, &position.cost, &txn.Unit); err != nil {
				return Position{}, err
			}
			if _, err := decimalCtx.Quo(costOfSoldUnits, costOfSoldUnits, &position.unit); err != nil {
				return Position{}, err
			}
			// proceed = (unit * sell price) - broker fee
			proceed := new(apd.Decimal)
			if _, err := decimalCtx.Mul(proceed, &txn.UnitPrice, &txn.Unit); err != nil {
				return Position{}, err
			}
			if _, err := decimalCtx.Sub(proceed, proceed, &txn.BrokerFee); err != nil {
				return Position{}, err
			}
			// realized gain = proceed - cost of sold units
			gain := new(apd.Decimal)
			if _, err := decimalCtx.Sub(gain, proceed, costOfSoldUnits); err != nil {
				return Position{}, err
			}
			if _, err := decimalCtx.Add(&position.realizedGain, &position.realizedGain, gain); err != nil {
				return Position{}, err
			}
			// reduce position unit and cost
			if _, err := decimalCtx.Sub(&position.unit, &position.unit, &txn.Unit); err != nil {
				return Position{}, err
			}
			if _, err := decimalCtx.Sub(&position.cost, &position.cost, costOfSoldUnits); err != nil {
				return Position{}, err
			}
		default:
			return Position{}, fmt.Errorf("unknown txn type %v", txn.TxnType)
		}
	}
	return position, nil
}

func (p Position) getAveragePrice() (apd.Decimal, error) {
	if p.unit.IsZero() {
		return apd.Decimal{}, nil
	}
	decimalCtx := apd.BaseContext.WithPrecision(14)
	result := new(apd.Decimal)
	if _, err := decimalCtx.Quo(result, &p.cost, &p.unit); err != nil {
		return *result, err
	}
	return *result, nil
}

func getTotalDividend(dividends []Dividend) (*apd.Decimal, error) {
	totalDiv := new(apd.Decimal)
	decimalCtx := apd.BaseContext.WithPrecision(14)
	// sum dividend per year
	for _, div := range dividends {
		if _, err := decimalCtx.Add(totalDiv, totalDiv, &div.NetAmount); err != nil {
			return nil, err
		}
	}
	return totalDiv, nil
}
