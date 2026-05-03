package scrape

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func GetHtmlStringFromUrl(targetURL string) (string, error) {
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
		// Block ad/analytics domains before navigating so their slow responses
		// don't hold up the page's load event (e.g. DoubleClick trackers).
		network.Enable(),
		network.SetBlockedURLs([]string{
			"*doubleclick.net*",
			"*googlesyndication.com*",
			"*googletagmanager.com*",
			"*google-analytics.com*",
			"*t.sharethis.com*",
		}),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(6*time.Second), // let JS redirects/challenge scripts settle
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		return "", err
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
	return html, nil
}
