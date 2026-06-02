package gold

import (
	"log"
	"os"
	"strings"
	"time"
	"wealth-management/internal/platform/database"
	"wealth-management/internal/platform/decimal"
	"wealth-management/internal/platform/scrape"

	"github.com/PuerkitoBio/goquery"
)

func ScrapeGoldPrice() {
	targetURL := os.Getenv("GOLD_URL")
	html, err := scrape.GetHtmlStringFromUrl(targetURL)
	if err != nil {
		log.Fatal(err)
	}
	price, err := decimal.ToDecimal(getGoldPriceStrFromHtml(html))
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.InitDbConnection(false)
	if err != nil {
		log.Fatal(err)
	}
	if db != nil {
		defer db.Close()
	}

	goldRepo := newGoldRepository(db)

	goldPrice := PriceHistory{
		BuyPrice: *price,
		Date:     time.Now().In(time.FixedZone("GMT+8", 8*60*60)),
	}

	err = goldRepo.insertOrUpdatePriceHistory(goldPrice)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully scraped gold price")

}

func getGoldPriceStrFromHtml(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}

	// what price would bank buy from me
	var buyPrice string
	doc.Find("p").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// downloaded html might seperated by nextline in unknown position
		// remove all whitespace to compare
		text := strings.Join(strings.Fields(s.Text()), "")
		if text == "MaybankGoldInvestmentAccount" {
			log.Println("found MaybankGoldInvestmentAccount")
			// first table appear after Maybank Gold Investment Account title
			// use nextAll to skip any unwanted tag (like div) in between
			table := s.NextAll().Find("table").First()
			if table.Length() == 0 {
				log.Println("Warning: Found header but no table follows")
				log.Println("Continue searching")
				return true // Continue searching just in case
			}

			table.Find("tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
				// skip tr that contain table header
				if s.Find("th").Length() > 0 {
					log.Println("Skipping header row:", s.Text())
					return true
				}
				// get third td node after header node
				td := s.Find("td").Eq(2)
				if td.Length() > 0 {
					log.Printf("Get buyPrice: %s\n", td.Text())
					buyPrice = td.Text()
					return false
				}
				return true
			})
			// break
			return false
		}
		return true
	})
	log.Println("buyPrice:", buyPrice)
	return buyPrice
}

func getTestingHtmlPage() string {
	log.Println("Scraping from local testing file")
	filename := "../resources/test-material/scrape/gold_and_silver_price.html"
	contentBytes, err := os.ReadFile(filename)

	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}

	// Convert the byte slice to a string
	return string(contentBytes)
}

func getHtmlPage() string {
	targetURL := os.Getenv("GOLD_URL")
	html, err := scrape.GetHtmlStringFromUrl(targetURL)
	if err != nil {
		log.Fatal(err)
	}
	return html
}
