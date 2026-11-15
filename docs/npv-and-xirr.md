# NPV and XIRR: annualized return for a stock holding

How the stock detail page's **Annualized Return** is meant to be calculated, and why.

## The question

"Including dividends, what yearly rate did my money in this stock earn?"

Simple total return ÷ years (or CAGR) gets this wrong when money went in at different times. CAGR assumes
everything was invested on day one. XIRR handles any number of buys, sells and dividends on any dates.

## Step 1: cash flows

Every event that moves money between you and the stock becomes a dated cash flow:

| Event               | Amount                             | Date             |
|---------------------|------------------------------------|------------------|
| BUY                 | **−** `TotalAmount` (fee included) | txn date         |
| SELL                | **+** `TotalAmount` (after fee)    | txn date         |
| Dividend            | **+** `NetAmount`                  | **payment** date |
| What you still hold | **+** units × latest price         | today            |

Realized G/L, unrealized G/L and dividends are all **inside** these flows already. Don't add any of them on top,
or they get counted twice.

## Step 2: NPV, a function of a guessed rate

NPV (net present value) takes one input, a rate `r`, and returns one number:

```
NPV(r) = Σ cashflowᵢ / (1 + r)^(yearsᵢ)        yearsᵢ = (dateᵢ − first date) in days / 365
```

`i` numbers the cash flows (1, 2, 3, …), and `Σ` means "add the term for every i". In code it's a loop.

- **Dividing by `(1 + r)^years`** moves each cash flow back to the first investment date. Money that arrives later
  is divided more, because the same amount on day one could have been growing at `r` all that time (the time
  value of money).
- **The exponent is in years, not days,** because `r` is a yearly rate.
- **The result is a surplus or shortfall, in first-date money,** compared with putting the same money in an
  account earning `r`:
    - `NPV > 0`: the investment beat `r`.
    - `NPV < 0`: the investment fell short of `r`.

Example: two BUYs of RM 100 each, on 2024-01-01 and 2025-01-01. On 2026-01-01 the holding is worth RM 400 (market value,
not gain). Comparing with an FD at 10%:

| Cash flow           | Years from first | Discounted at 10%    |
|---------------------|------------------|----------------------|
| −100 (BUY)          | 0                | −100.00              |
| −100 (BUY)          | 1                | −100 / 1.1 = −90.91  |
| +400 (market value) | 2                | 400 / 1.21 = +330.58 |
| NPV                 |                  | +139.67              |

▎
NPV > 0, so the investment beat a 10% FD. The XIRR (the rate where NPV = 0) is about 56.2%: solving −100 − 100/(1+r) +
400/(1+r)² = 0 gives 1 + r ≈ 1.5615.

## Step 3: XIRR, the rate where NPV = 0

XIRR is **not** NPV. It's the specific `r` at which `NPV(r) = 0`.

**Why zero:** at that rate, the investment and an account paying `r` are exactly equal, neither ahead nor behind.
That's what it means for the investment to have "returned `r` a year".

**The savings-account picture:** open an account paying `r`, compounding. Deposit each BUY, withdraw each dividend
and SELL on the same dates, then withdraw today's market value. At `r = XIRR`, the account ends at exactly RM 0.

- **With a lower `r`:** money is left over.
- **With a higher `r`:** you come up short.

## Worked example: MAYBANK

- **Buys:** RM 983 (2018-01-01) and RM 811 (2020-04-11).
- **Dividends:** 18 payments totalling RM 919.
- **Holding:** 200 units, at an *assumed* price of RM 10.00.

| r (guess) | NPV        | Meaning                                                       |
|-----------|------------|---------------------------------------------------------------|
| −50%      | +1,014,285 |                                                               |
| 0%        | +1,125     | no discounting: 919 + 2,000 − 1,794 = plain profit            |
| 5%        | +308       | beat a 5% account                                             |
| **7.75%** | **≈ 0**    | **XIRR, the annualized return**                               |
| 10%       | −200       | fell short of a 10% account                                   |
| 1000%     | −972       | later inflows shrink to ~0; only the day-0 buy (−983) remains |

For comparison, at the same price:

| Price    | Total return | CAGR (wrong here) | XIRR     |
|----------|--------------|-------------------|----------|
| RM 9.00  | 51.6%        | 4.8%              | **6.7%** |
| RM 10.00 | 62.7%        | 5.7%              | **7.8%** |
| RM 11.00 | 73.9%        | 6.4%              | **8.8%** |

CAGR understates the return because it treats the 2020 buy as if it had been invested since 2018.

## Step 4: finding r (bisection)

There's no way to rearrange the equation into `r = …`, because `r` is buried inside the powers. So you **search**:
call `NPV` with a guessed `r`, look at the sign, and adjust. This is *root finding*. Bisection is a smarter
search than brute force, because each guess throws away half the remaining range.

```go
lo, hi := -0.99, 10.0
if sign(npv(flows, lo)) == sign(npv(flows, hi)) {
return nil // no rate in range makes NPV zero (e.g. zero-cost holding)
}
const tolerance = 1e-7
for i := 0; i < 200 && hi-lo > tolerance; i++ {
mid := (lo + hi) / 2
if npv(flows, mid) > 0 {
lo = mid // still beating mid: true rate is higher
} else {
hi = mid // falling short of mid: true rate is lower
}
}
r := (lo + hi) / 2
```

### Why NPV is positive at `lo` and negative at `hi`

Money goes **out first** (the first BUY at year 0 isn't discounted) and comes **in later** (dividends, today's
value). Raising `r` shrinks the later inflows towards zero while the day-0 outflow stays the same. NPV goes from
positive to negative, crossing zero once, and that crossing is the answer.

### Why `lo = −0.99` and `hi = 10`

These aren't derived values. They're safe bounds: one surely too low, one surely too high.

- **`r` must be greater than −1:** at −1 the formula divides by zero, and below −1 a negative base raised to a
  fractional power is `NaN`. −0.99 means losing 99% a year.
- **10 means 1000% a year,** beyond any long-held stock. 1 (100%) would also work. A wider range costs only one
  extra step per doubling.

The sign check before the loop catches the case where the answer isn't inside the range.

### Why ~27 iterations, not 100

The range starts 10.99 wide and halves each step: `width = 10.99 / 2ⁿ`.

- **About 24 steps** give ±0.0001%, far finer than the 2-decimal display.
- **After about 55–60 steps,** `float64` can't shrink the range any further.

Stopping on `hi - lo > tolerance` states the intent ("precise enough"). The iteration cap is only a safety net.

## Implementation notes (Go)

- **Keep it a pure function** in `internal/stock/overview.go`, alongside `computePosition`. The handler loads the
  data and passes it in; the function never touches the repository:
  ```go
  func annualizedReturn(txns []Txn, dividends []Dividend, lastPrice *apd.Decimal, now time.Time) (*float64, error)
  ```
- **Time difference:** `t.Sub(first)` returns a `time.Duration` in **nanoseconds**, not days. Go has no
  `Duration.Days()`, so use `d.Hours() / 24 / 365` to get years.
- **`float64` is fine here.** Fractional powers are awkward in `apd`, and the output is an estimated rate, not
  money. Convert inputs with `d.Float64()`. Money values everywhere else stay `apd`.

## When to return `null` (UI shows "—")

| Case                                     | Why                                                                                                                    |
|------------------------------------------|------------------------------------------------------------------------------------------------------------------------|
| No latest price                          | Today's value (the final cash flow) is unknown                                                                         |
| No outflows (e.g. AAX in-specie, cost 0) | Nothing invested, so no rate can make NPV zero                                                                         |
| Holding younger than 1 year              | Annualizing a short period inflates it (+3% in a month ≈ +42%/yr). Show plain total return % instead (GIPS convention) |

A **fully sold** stock still works: today's value is 0, and XIRR shows how the closed position performed.

## Related metrics (same page)

| Metric           | Formula                                    | Answers                                           |
|------------------|--------------------------------------------|---------------------------------------------------|
| Dividend payback | total dividends ÷ total buy cost           | How much of my money have dividends returned?     |
| Total return %   | (unrealized + realized + dividends) ÷ cost | Overall, how far up or down am I?                 |
| **XIRR**         | rate where NPV = 0                         | What yearly rate did my money earn?               |
| Yield on cost    | TTM DPS ÷ average cost                     | What income rate am I earning now on what I paid? |
