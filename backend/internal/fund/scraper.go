package fund

import (
	"database/sql"
	"errors"
	"log"
	"slices"
	"time"
	"wealth-management/internal/fund/provider"

	"github.com/cockroachdb/apd/v3"
	"github.com/google/uuid"
)

// ScrapeFundNavAndIncomeDist pull nav and income dist, and compute REINVESTED transaction.
// this function assumes previous transaction was computed correctly and it will be a incremental compute.
// todo build a recompute job/scripts
func ScrapeFundNavAndIncomeDist(db *sql.DB) {
	fundRepo := newRepository(db)
	funds, err := fundRepo.getAllFunds()
	if err != nil {
		log.Printf("Error getting funds: %s. Aborting fund scrape.\n", err.Error())
		return
	}

	var scraperByProvider = map[string]provider.FundDataProvider{
		"AHAM": &provider.AhamFundProvider{},
	}

	for _, fund := range funds {
		scraper, ok := scraperByProvider[fund.Provider]
		if !ok {
			log.Printf("Scraper for %s is not built, skipping.\n", fund.Provider)
			continue
		}
		log.Printf("Scraping fund %s with provider %s\n", fund.Name, fund.Provider)
		// Get today NAV
		nav, err := scraper.FetchNavByDate(fund.FundCode, time.Now())
		if err != nil {
			log.Printf("Error while scrape nav of %s: %s\n", fund.Provider, err.Error())
			continue
		}
		priceHistory := PriceHistory{
			FundCode:  fund.FundCode,
			PriceDate: nav.NavDate,
			Nav:       *nav.Nav,
		}
		log.Printf("Insert latest nav to table: %v", priceHistory)
		if err := fundRepo.insertIgnoreFundPriceHistory(priceHistory); err != nil {
			log.Printf("Error inserting nav of %s: %s\n", fund.Provider, err.Error())
			continue
		}

		// Get Income Distribution from last pulled data up till today
		startDate, ok := getIncomeDistPullStartDate(fundRepo, fund.FundCode)
		if !ok {
			continue
		}
		incomeDistributions, err := scraper.FetchIncomeDistribution(fund.FundCode, startDate)
		if err != nil {
			log.Printf("Error fetching income distribution of %s: %s\n", fund.FundCode, err.Error())
			continue
		}
		slices.SortFunc(incomeDistributions, func(a, b provider.DistributionResult) int {
			return a.PaymentDate.Compare(b.PaymentDate)
		})
		// todo compute reinvestment txn
		ctx := apd.BaseContext.WithPrecision(14)
		for _, d := range incomeDistributions {
			// change sen to RM, e.g. 50 sen to RM 0.5
			ringgitPerUnit := new(apd.Decimal)
			_, err = ctx.Quo(ringgitPerUnit, d.SenPerUnit, apd.New(100, 0))
			if err != nil {
				log.Printf("Error calculating total dividend payout of %s: %s\n", fund.FundCode, err.Error())
				break
			}
			totalUnit, err := fundRepo.getTotalUnitToDate(fund.FundCode, d.DeclareDate)
			if err != nil {
				log.Printf("Error getting total unit of %s: %s\n", fund.FundCode, err.Error())
				break
			}
			dividendPayout := new(apd.Decimal)
			_, err = ctx.Mul(dividendPayout, ringgitPerUnit, totalUnit)
			if err != nil {
				log.Printf("Error calculating total dividend payout of %s: %s\n", fund.FundCode, err.Error())
				break
			}
			navResult, err := scraper.FetchNavByDate(fund.FundCode, d.PaymentDate)
			if err != nil {
				log.Printf("Error fetching nav of %s to calculate dividend payout: %s\n", fund.FundCode, err.Error())
				break
			}
			reinvestedUnit := new(apd.Decimal)
			_, err = ctx.Quo(reinvestedUnit, dividendPayout, navResult.Nav)
			if err != nil {
				log.Printf("Error calculating reinvested unit of %s: %s\n", fund.FundCode, err.Error())
				break
			}

			reinvestedTxn := Txn{
				ID:                  uuid.NewString(),
				FundCode:            fund.FundCode,
				TxnDate:             d.PaymentDate,
				Unit:                *reinvestedUnit,
				UnitPrice:           *navResult.Nav,
				SalesCharge:         *apd.New(0, 0),
				NetInvestmentAmount: *dividendPayout,
				TotalAmount:         *apd.New(0, 0),
				TxnType:             "REINVESTED",
				Remark:              "REINVESTED",
			}
			err = fundRepo.insertFundTxn(reinvestedTxn)
			if err != nil {
				log.Printf("Error inserting reinvested txn of %s: %s\n", fund.FundCode, err.Error())
				break
			}
		}
	}
}

func getIncomeDistPullStartDate(fundRepo *repository, fundCode string) (time.Time, bool) {
	latestReinvestedTxn, err := fundRepo.getLatestReinvestmentTxn(fundCode)
	if err == nil {
		return latestReinvestedTxn.TxnDate.Add(-time.Hour * 24), true
	}
	if !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error fetching latest reinvestment txn of %s: %s\n", fundCode, err.Error())
		return time.Time{}, false
	}
	// No Reinvested txn happened, get oldest txn date
	oldestTxn, err := fundRepo.getOldestFundTxn(fundCode)
	if err == nil {
		// Transaction take T+2 to credit to holding
		return oldestTxn.TxnDate.Add(-time.Hour * 24 * 3), true
	}
	// Error or no rows, mean no subsequence pull required, always return ok = false
	if !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error fetching oldest fund txn of %s: %s\n", fundCode, err.Error())
	}
	return time.Time{}, false
}
