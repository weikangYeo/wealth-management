package stock

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/cockroachdb/apd/v3"
)

func computeOverview(stock Stock, txns []Txn, dividends []Dividend, lastDonePrice *apd.Decimal, today time.Time) (Overview, error) {
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
	unrealizedGainLoss, err := position.getUnrealizedGainLoss(lastDonePrice)
	if err != nil {
		return Overview{}, err
	}
	roi, err := position.getRoi(unrealizedGainLoss, *totalDiv)
	if err != nil {
		return Overview{}, err
	}
	dividendPayback, err := getDividendPayback(position, totalDiv)
	if err != nil {
		return Overview{}, err
	}
	var annualizedReturn *float64
	// XIRR needs today's holding value (units × price), so skip it when no price has been scraped.
	if lastDonePrice != nil {
		holding := 0.0
		lastDonePriceFloat, err := lastDonePrice.Float64()
		if err != nil {
			return Overview{}, err
		}

		if !position.unit.IsZero() {
			totalUnit, err := position.unit.Float64()
			if err != nil {
				return Overview{}, err
			}
			holding = totalUnit * lastDonePriceFloat
		}
		annualizedReturn, err = getAnnualizedReturn(txns, dividends, holding, today)
		if err != nil {
			return Overview{}, err
		}
	}

	overview := Overview{
		StockName:          stock.StockName,
		DisplayName:        stock.DisplayName,
		BursaStockId:       stock.BursaStockId,
		TotalDividend:      *totalDiv,
		AveragePrice:       averageCost,
		TotalUnit:          position.unit,
		DividendPayback:    dividendPayback,
		RealizedGainLoss:   position.realizedGain,
		UnrealizedGainLoss: unrealizedGainLoss,
		Roi:                roi,
		AnnualizedReturn:   annualizedReturn,
	}
	return overview, nil
}

// getAnnualizedReturn calculate base on XIRR formula, referred "npv-and-xirr.md" for more details. using *float because we need nil to represent N/A annualize return case
func getAnnualizedReturn(txns []Txn, dividends []Dividend, holding float64, today time.Time) (*float64, error) {
	if len(txns) == 0 {
		return nil, nil
	}
	firstInvestmentDate := slices.MinFunc(txns, func(a, b Txn) int { return a.TxnDate.Compare(b.TxnDate) }).TxnDate
	if today.Sub(firstInvestmentDate) < 365*24*time.Hour {
		// the whole holding < 1 year which could cause wrong annualized representation
		return nil, nil
	}
	cashFlows, err := getCashFlows(txns, dividends, holding, today, firstInvestmentDate)
	if err != nil {
		return nil, err
	}
	if len(cashFlows) == 0 {
		// no cash flow
		return nil, nil
	}
	// below are trying to find the rate that can make npv = 0 (XIRR)
	// using root finding (binary search), so it would be O(logN) time complexity
	// first set a boundary for starting ceiling and floor, FYI 1 = 100 %,
	low, high := -0.99, 10.0
	// edge case check, by right npv(extreme low) should return positive value, and npv(extreme high) should return negative
	// so if they are in both side, mean it is a zero cost stock or stock without value case (referred npv-and-xirr.md)
	npvLow, npvHigh := npv(cashFlows, low), npv(cashFlows, high)
	if (npvLow > 0) == (npvHigh > 0) {
		// no rate case
		return nil, nil
	}
	// as long as gap close and smaller than toleranceNum, we consider it is close enough and we take the final rate
	const toleranceNum = 1e-7
	// binary search rate
	for i := 0; i < 100 && high-low > toleranceNum; i++ {
		mid := (low + high) / 2
		midNpv := npv(cashFlows, mid)
		if midNpv > 0 {
			// npv beat rate, mean rate is higher
			low = mid
		} else {
			high = mid
		}
	}
	r := (low + high) / 2
	return &r, nil
}

func getCashFlows(txns []Txn, dividends []Dividend, holding float64, today time.Time, firstInvestmentDate time.Time) ([]CashFlow, error) {
	// compute cash flow
	var cashFlows []CashFlow
	for _, t := range txns {
		amount, err := t.TotalAmount.Float64()
		if err != nil {
			return nil, err
		}
		c := CashFlow{
			value:                   amount,
			yearFromFirstInvestment: t.TxnDate.Sub(firstInvestmentDate).Hours() / 24 / 365,
		}
		if t.TxnType == "BUY" {
			c.value = c.value * -1
		}
		cashFlows = append(cashFlows, c)
	}
	for _, d := range dividends {
		// skip future date dividend since we not receiving the cash yet
		if d.PaymentDate.After(today) {
			continue
		}
		amount, err := d.NetAmount.Float64()
		if err != nil {
			return nil, err
		}
		cashFlows = append(cashFlows,
			CashFlow{
				value:                   amount,
				yearFromFirstInvestment: d.PaymentDate.Sub(firstInvestmentDate).Hours() / 24 / 365,
			})
	}
	cashFlows = append(cashFlows, CashFlow{
		value:                   holding,
		yearFromFirstInvestment: today.Sub(firstInvestmentDate).Hours() / 24 / 365,
	})
	return cashFlows, nil
}

// referred npv-and-xirr.md for formula
func npv(cashFlows []CashFlow, rate float64) float64 {
	var result float64
	for _, c := range cashFlows {
		result += c.value / math.Pow(1+rate, c.yearFromFirstInvestment)
	}
	return result
}

func getDividendPayback(position Position, totalDiv *apd.Decimal) (*apd.Decimal, error) {
	if position.totalInvested.IsZero() {
		// return nil since all sold out, show N/A in FE
		return nil, nil
	}
	dividendPayback := new(apd.Decimal)
	decimalCtx := apd.BaseContext.WithPrecision(14)
	if _, err := decimalCtx.Quo(dividendPayback, totalDiv, &position.totalInvested); err != nil {
		return nil, err
	}
	return dividendPayback, nil
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
			if _, err := decimalCtx.Add(&position.totalInvested, &position.totalInvested, &txn.TotalAmount); err != nil {
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

func (p Position) getUnrealizedGainLoss(lastDonePrice *apd.Decimal) (*apd.Decimal, error) {
	if p.unit.IsZero() {
		// no unit, unrealized gain/loss = 0
		return apd.New(0, 0), nil
	}
	// cant calculate without last done price
	if lastDonePrice == nil {
		return nil, nil
	}
	gainLossDiff := new(apd.Decimal)
	ctx := apd.BaseContext.WithPrecision(14)
	// unrealized gain loss = (lastDonePrice − avgPrice) × unit
	// but avgPrice was a result from a division, divide-then-multi might increase the inaccurate rounding result
	// so expand it (lastDonePrice − avgPrice) × unit
	// = lastDonePrice * unit - avgPrice * unit
	// = lastDonePrice * unit - (cost / unit) * unit
	// = lastDonePrice * unit - cost
	if _, err := ctx.Mul(gainLossDiff, lastDonePrice, &p.unit); err != nil {
		return nil, err
	}
	if _, err := ctx.Sub(gainLossDiff, gainLossDiff, &p.cost); err != nil {
		return nil, err
	}
	return gainLossDiff, nil
}

// ROI = (unrealized P/L + realized P/L + dividends received) / total cost invested
func (p Position) getRoi(unrealizedRealGain *apd.Decimal, totalDividend apd.Decimal) (*apd.Decimal, error) {
	if p.totalInvested.IsZero() {
		return nil, nil
	}
	if unrealizedRealGain == nil {
		// not accurate roi without unrealizedRealGain
		return nil, nil
	}
	roi := new(apd.Decimal)
	ctx := apd.BaseContext.WithPrecision(14)
	if _, err := ctx.Add(roi, unrealizedRealGain, &p.realizedGain); err != nil {
		return nil, err
	}
	if _, err := ctx.Add(roi, roi, &totalDividend); err != nil {
		return nil, err
	}
	if _, err := ctx.Quo(roi, roi, &p.totalInvested); err != nil {
		return nil, err
	}
	return roi, nil
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
