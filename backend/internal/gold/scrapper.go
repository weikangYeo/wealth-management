package gold

import (
	"context"
	"log"
	"os"
	"strings"
	"time"
	"wealth-management/internal/platform/database"
	"wealth-management/internal/platform/decimal"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func ScrapeGoldPrice(isTrial bool) {
	var html string
	if isTrial {
		log.Println("Running in trial mode")
		html = getTestingHtmlPage()
	} else {
		log.Println("Running in live mode")
		html = getHtmlPage()
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
	log.Printf("Scraping from %s", targetURL)

	// chromedp: navigate, wait for content, grab HTML
	var html string

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:148.0) Gecko/20100101 Firefox/148.0"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 45*time.Second)
	defer cancelTimeout()

	// Network telemetry: request/response/failure events.
	var mainDocReqID network.RequestID
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			if e.Type == network.ResourceTypeDocument {
				mainDocReqID = e.RequestID
				log.Printf("[REQ] doc %s %s", e.Request.Method, e.Request.URL)
			}
		case *network.EventResponseReceived:
			if e.Type == network.ResourceTypeDocument || e.RequestID == mainDocReqID {
				log.Printf("[RES] doc status=%v proto=%s mime=%s server=%s url=%s",
					e.Response.Status,
					e.Response.Protocol,
					e.Response.MimeType,
					e.Response.Headers["server"],
					e.Response.URL,
				)
			}
		case *network.EventLoadingFailed:
			if e.Type == network.ResourceTypeDocument || e.RequestID == mainDocReqID {
				log.Printf("[FAIL] doc requestId=%s canceled=%v blocked=%v error=%s",
					e.RequestID, e.Canceled, e.BlockedReason, e.ErrorText)
			}
		}
	})

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.Sleep(6*time.Second), // let redirects/challenge scripts settle
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		log.Fatal(err)
	}
	// then parse `html` with goquery (like jQuery for Go)	cookie := harvestCookie(targetURL)
	log.Printf("html scraped.")

	lc := strings.ToLower(html)
	if strings.Contains(lc, "just a moment") ||
		strings.Contains(lc, "captcha") ||
		strings.Contains(lc, "attention required") ||
		strings.Contains(lc, "access denied") {
		log.Printf("Possible anti-bot/challenge page detected")
	}
	return html
}
