package fund

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"
	"wealth-management/internal/fund/provider"

	"github.com/cockroachdb/apd/v3"
	"github.com/google/uuid"
)

var errNoTxn = errors.New("no transaction found")

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
		"AHAM":      &provider.AhamFundProvider{},
		"PRINCIPAL": &provider.PrincipalFundProvider{},
		"TA":        &provider.TaFundProvider{},
		"MAYBANK":   &provider.MaybankFundProvider{},
	}

	for _, fund := range funds {
		scraper, ok := scraperByProvider[fund.Provider]
		if !ok {
			log.Printf("Scraper for %s is not built, skipping.\n", fund.Provider)
			continue
		}
		log.Printf("Scraping fund %s with provider %s\n", fund.Name, fund.Provider)
		fundRef := provider.FundRef{
			FundCode:         fund.FundCode,
			Name:             fund.Name,
			ScrapeParamValue: fund.ScrapeParamValue,
		}
		if err := scrapeNav(fundRepo, scraper, fundRef); err != nil {
			log.Printf("Error while scraping fund nav %s due to %s. Proceed to scrape income dist if any.\n", fund.Name, err.Error())
		}
		if err := scrapeIncomeDist(fundRepo, scraper, fundRef); err != nil {
			log.Printf("Error while scraping fund income dist %s due to %s.\n", fundRef.Name, err.Error())
		}
	}
}

func scrapeNav(fundRepo *repository, scraper provider.FundDataProvider, fundRef provider.FundRef) error {
	// Get today NAV
	nav, err := scraper.FetchNavByDate(fundRef, time.Now())
	if err != nil {
		return fmt.Errorf("error when fetch nav by date %s: %w", time.Now(), err)
	}
	priceHistory := PriceHistory{
		FundCode:  fundRef.FundCode,
		PriceDate: nav.NavDate,
		Nav:       *nav.Nav,
	}
	log.Printf("Insert latest nav to table: %v", priceHistory)
	if err := fundRepo.insertIgnoreFundPriceHistory(priceHistory); err != nil {
		return fmt.Errorf("error when insert fund price history %v: %w", priceHistory, err)
	}
	return nil
}

// todo add test
func scrapeIncomeDist(fundRepo *repository, scraper provider.FundDataProvider, fundRef provider.FundRef) error {
	// Get Income Distribution from last pulled data up till today
	startDate, err := getIncomeDistPullStartDate(fundRepo, fundRef.FundCode)
	if errors.Is(err, errNoTxn) {
		// no txn, just skip income dist computation.
		return nil
	}
	if err != nil {
		return err
	}
	log.Printf("Fetch Income Dist of %s from %s", fundRef.Name, startDate)
	incomeDistributions, err := scraper.FetchIncomeDistribution(fundRef, startDate)
	if err != nil {
		return err
	}
	slices.SortFunc(incomeDistributions, func(a, b provider.DistributionResult) int {
		return a.PaymentDate.Compare(b.PaymentDate)
	})
	ctx := apd.BaseContext.WithPrecision(14)
	for _, d := range incomeDistributions {
		// txn table use ID as PK because it could happen same fund + date + amount + txn_type happen more than once.
		// with that we can't do insert ignore in this reinvested case,
		// we need safeguard manually, skip if found reinvested txn for this fund at this date.
		exist, err := fundRepo.isReinvestedTxnExists(fundRef.FundCode, d.PaymentDate)
		if err != nil {
			return fmt.Errorf("error check existing reinvested txn: %w", err)
		}
		if exist {
			log.Printf("Found reinvested txn: %s, skip current income dist computation", d.PaymentDate)
			continue
		}
		// change sen to RM, e.g. 50 sen to RM 0.5
		ringgitPerUnit := new(apd.Decimal)
		_, err = ctx.Quo(ringgitPerUnit, d.SenPerUnit, apd.New(100, 0))
		if err != nil {
			return fmt.Errorf("error when calculating ringgit per unit: %w", err)
		}
		totalUnit, err := fundRepo.getTotalUnitToDate(fundRef.FundCode, d.DeclareDate)
		if err != nil {
			return fmt.Errorf("error when calculating total unit to date: %w", err)
		}
		dividendPayout := new(apd.Decimal)
		_, err = ctx.Mul(dividendPayout, ringgitPerUnit, totalUnit)
		if err != nil {
			return fmt.Errorf("error when calculating dividend payout: %w", err)
		}
		// during payment date, IUTA use the income dist to re-invest
		// there is a catch/iuta found during 2025 AHAM PRS record, where it use Declare Date + 2 NAV to compute dist.
		navResult, err := scraper.FetchNavByDate(fundRef, d.PaymentDate)
		if err != nil {
			return fmt.Errorf("error when fetch nav by date %s: %w", d.PaymentDate, err)
		}
		log.Printf("Income Dist: %v\n", navResult)
		reinvestedUnit := new(apd.Decimal)
		_, err = ctx.Quo(reinvestedUnit, dividendPayout, navResult.Nav)
		if err != nil {
			return fmt.Errorf("error when calculating reinvested unit payout: %w", err)
		}

		reinvestedTxn := Txn{
			ID:                  uuid.NewString(),
			FundCode:            fundRef.FundCode,
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
			return fmt.Errorf("error when inserting reinvested txn: %w", err)
		}
	}
	return nil
}

// todo add test
func getIncomeDistPullStartDate(fundRepo *repository, fundCode string) (time.Time, error) {
	// try to see which was the last pulled reinvested txn
	// then try to pull around there, because not all provider provide declare date & payment date info
	// a slight overlap make sure no data miss out.
	latestReinvestedTxn, err := fundRepo.getLatestReinvestmentTxn(fundCode)
	if err == nil {
		return latestReinvestedTxn.TxnDate.Add(time.Hour * -24), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, fmt.Errorf("error when get latest reinvestment txn: %w", err)
	}
	// No Reinvested txn happened, get oldest txn date
	oldestTxn, err := fundRepo.getOldestFundTxn(fundCode)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, errNoTxn
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("error when get oldest fund txn: %w", err)
	}
	// Transaction take T+2 to credit to holding
	return oldestTxn.TxnDate.Add(-time.Hour * 24 * 3), nil
}
