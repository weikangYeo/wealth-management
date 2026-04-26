package stock

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"wealth-management/internal/platform/database"
	"wealth-management/internal/platform/decimal"
	"wealth-management/internal/platform/scrape"

	"github.com/PuerkitoBio/goquery"
	"github.com/cockroachdb/apd/v3"
)

type klseDividend struct {
	ExDate          time.Time
	PaymentDate     time.Time
	DividendPerUnit apd.Decimal
	Remark          string
}

func ScrapeStockLastDonePrice() {
	db, err := database.InitDbConnection(false)
	if err != nil {
		log.Fatal(err)
	}
	if db != nil {
		defer db.Close()
	}
	stockRepo := newStockRepository(db)
	stocks, err := stockRepo.getAllStocks()
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range stocks {
		scrapeStockData(stockRepo, s)
	}
}

func scrapeStockData(stockRepo *repository, s Stock) {
	klseURL := os.Getenv("KLSE_STOCK_BASE_URL") + strconv.Itoa(s.BursaStockId)
	html, err := scrape.GetHtmlStringFromUrl(klseURL)
	if err != nil {
		log.Printf("KLSE scrape failed for %s: %v, falling back to old source", s.StockName, err)
		scrapeAndSavePriceFromOldSource(stockRepo, s)
		return
	}

	price, err := scrapeKlsePrice(html)
	if err != nil {
		log.Printf("KLSE price parse failed for %s: %v, falling back to old source", s.StockName, err)
		scrapeAndSavePriceFromOldSource(stockRepo, s)
	} else {
		log.Printf("KLSE price for %s: %s", s.StockName, price.String())
		if err := stockRepo.createStockPrice(Price{
			StockName:     s.StockName,
			PriceDate:     time.Now().In(time.FixedZone("GMT+8", 8*60*60)),
			LastDonePrice: *price,
		}); err != nil {
			log.Printf("Error saving KLSE price for %s: %v", s.StockName, err)
		}
	}

	scraped, err := scrapeKlseDividends(html)
	if err != nil {
		log.Printf("KLSE dividend parse failed for %s: %v", s.StockName, err)
		return
	}
	for _, kd := range scraped {
		// skipped inserted dividend
		exists, err := stockRepo.existsDividend(s.StockName, kd.ExDate)
		if err != nil {
			log.Printf("Error checking dividend for %s: %v", s.StockName, err)
			continue
		}
		if exists {
			continue
		}
		netUnit, err := stockRepo.getNetStockUnitAtDate(s.StockName, kd.ExDate)
		if err != nil {
			log.Printf("Error getting stock unit for %s at %s: %v", s.StockName, kd.ExDate.Format("2006-01-02"), err)
			continue
		}

		// no need to proceed if as that date i didnt purchase any stock yet
		if netUnit.IsZero() {
			log.Printf("As of exDate %s, dont have any stock for %s", kd.ExDate.Format("2006-01-02"), s.StockName)
			continue
		}

		dividend := Dividend{
			StockName:       s.StockName,
			ExDate:          kd.ExDate,
			PaymentDate:     kd.PaymentDate,
			StockUnit:       netUnit,
			DividendPerUnit: kd.DividendPerUnit,
			Remark:          kd.Remark,
		}
		ctx := apd.BaseContext
		ctx.Precision = 14 // match DECIMAL(14,4): enough headroom, results rounded to 4dp by DB
		if err := dividend.CalculateDividendTotalAmount(&ctx); err != nil {
			log.Printf("Error calculating dividend for %s ex %s: %v", s.StockName, kd.ExDate.Format("2006-01-02"), err)
			continue
		}
		if err := stockRepo.createDividend(dividend); err != nil {
			log.Printf("Error saving dividend for %s ex %s: %v", s.StockName, kd.ExDate.Format("2006-01-02"), err)
		}
	}
}

func scrapeAndSavePriceFromOldSource(stockRepo *repository, s Stock) {
	targetURL := os.Getenv("STOCK_URL") + "?stock_code=" + strconv.Itoa(s.BursaStockId)
	html, err := scrape.GetHtmlStringFromUrl(targetURL)
	if err != nil {
		log.Printf("Old source scrape failed for %s: %v", s.StockName, err)
		return
	}
	price, err := decimal.ToDecimal(getLastDonePriceFromHtmlStr(html))
	if err != nil {
		log.Printf("Old source price parse failed for %s: %v", s.StockName, err)
		return
	}
	log.Printf("Old source price for %s: %s", s.StockName, price.String())
	if err := stockRepo.createStockPrice(Price{
		StockName:     s.StockName,
		PriceDate:     time.Now().In(time.FixedZone("GMT+8", 8*60*60)),
		LastDonePrice: *price,
	}); err != nil {
		log.Printf("Error saving old source price for %s: %v", s.StockName, err)
	}
}

func scrapeKlsePrice(html string) (*apd.Decimal, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	val, exists := doc.Find("#price").Attr("data-value")
	if !exists || strings.TrimSpace(val) == "" {
		return nil, fmt.Errorf("price element not found or empty")
	}
	return decimal.ToDecimal(strings.TrimSpace(val))
}

// scrapeKlseDividends extracts all dividend rows from the #dividends table.
// It skips financial-year separator rows (which have fewer than 6 cells) and logs
// any row that cannot be parsed rather than aborting the whole scrape.
func scrapeKlseDividends(html string) ([]klseDividend, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	var dividends []klseDividend
	doc.Find("#dividends table tbody tr").Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 6 {
			return // separator row (colspan="100") or malformed
		}
		// columns: [0]=Announced [1]=FinancialYear [2]=Subject [3]=ExDate [4]=PaymentDate [5]=Amount
		remark := strings.TrimSpace(cells.Eq(2).Text())
		exDateStr := strings.TrimSpace(cells.Eq(3).Find("strong").Text())
		paymentDateStr := strings.TrimSpace(cells.Eq(4).Text())
		amountStr := strings.TrimSpace(cells.Eq(5).Text())

		exDate, err := time.Parse("02 Jan 2006", exDateStr)
		if err != nil {
			log.Printf("Skipping dividend row: invalid ex date %q: %v", exDateStr, err)
			return
		}
		paymentDate, err := time.Parse("02 Jan 2006", paymentDateStr)
		if err != nil {
			log.Printf("Skipping dividend row: invalid payment date %q: %v", paymentDateStr, err)
			return
		}
		amount, err := decimal.ToDecimal(amountStr)
		if err != nil {
			log.Printf("Skipping dividend row: invalid amount %q: %v", amountStr, err)
			return
		}
		dividends = append(dividends, klseDividend{
			ExDate:          exDate,
			PaymentDate:     paymentDate,
			DividendPerUnit: *amount,
			Remark:          remark,
		})
	})
	return dividends, nil
}

func getLastDonePriceFromHtmlStr(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}
	priceText := strings.TrimSpace(
		doc.Find(".company-last-done .col-md-6").
			First().
			Find("h3.h5.bold.mb-0").
			Eq(1).
			Text(),
	)
	priceText = strings.Join(strings.Fields(priceText), " ")
	log.Println("Old source price:", priceText)
	return priceText
}

func getTestingHtmlPage() string {
	log.Println("Scraping from local testing file")
	contentBytes, err := os.ReadFile("../resources/test-material/scrape/stock_company_profile_capital_a.htm")
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}
	return string(contentBytes)
}
