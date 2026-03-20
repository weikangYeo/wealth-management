package gold

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// todo change back to lower case after testing
func ScrapeGoldPrice() {
	//html := getHtmlPage()
	html := getTestingHtmlPage()
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
	targetURL := "https://www.maybank2u.com.my/maybank2u/malaysia/en/personal/rates/gold_and_silver.page"
	log.Printf("Scraping from %s", targetURL)

	// chromedp: navigate, wait for content, grab HTML
	var html string

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true))

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitVisible(`table`, chromedp.ByQuery),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		log.Fatal(err)
	}
	// then parse `html` with goquery (like jQuery for Go)	cookie := harvestCookie(targetURL)
	log.Printf("html scraped.")
	return html
}
