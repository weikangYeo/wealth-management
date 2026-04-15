package stock

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"wealth-management/internal/platform/database"
	"wealth-management/internal/platform/decimal"
	"wealth-management/internal/platform/scrape"

	"github.com/PuerkitoBio/goquery"
)

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
	for _, stock := range stocks {
		targetUrl := os.Getenv("STOCK_URL") + "?stock_code=" + strconv.Itoa(stock.BursaStockId)
		html, err := scrape.GetHtmlStringFromUrl(targetUrl)
		if err != nil {
			log.Printf("Error found when download html from url %s\n", targetUrl)
			return
		}
		price, err := decimal.ToDecimal(getLastDonePriceFromHtmlStr(html))
		if err != nil {
			log.Printf("Error found when scrape or convert to decimal for stock %s\n", stock.StockName)
			return
		}
		log.Printf("Price for stock %s is %s\n", stock.StockName, price)
		// todo save price to db and test
		stockPrice := Price{
			StockName:     stock.StockName,
			PriceDate:     time.Now().In(time.FixedZone("GMT+8", 8*60*60)),
			LastDonePrice: *price,
		}

		if err := stockRepo.createStockPrice(stockPrice); err != nil {
			log.Printf("Error when save stock price of %s, skipping this.\n", stock.StockName)
		}
	}

}

// todo consider stream read (not sure if this is a thing in golang) to reduce memory pressure
func getLastDonePriceFromHtmlStr(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}

	priceText := strings.TrimSpace(
		doc.Find(".company-last-done .col-md-6").
			First().
			// h3 tag, class = "h5 bold mb-0"
			Find("h3.h5.bold.mb-0").
			Eq(1).
			Text(),
	)
	// remove white space
	priceText = strings.Join(strings.Fields(priceText), " ")
	log.Println("buyPrice:", priceText)
	return priceText
}

func getTestingHtmlPage() string {
	log.Println("Scraping from local testing file")
	filename := "../resources/test-material/scrape/stock_company_profile_capital_a.htm"
	contentBytes, err := os.ReadFile(filename)

	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}

	// Convert the byte slice to a string
	return string(contentBytes)
}
