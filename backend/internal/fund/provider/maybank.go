package provider

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
	"wealth-management/internal/platform/scrape"

	"github.com/PuerkitoBio/goquery"
	"github.com/cockroachdb/apd/v3"
)

type MaybankFundProvider struct {
}

const baseUrl = "https://www.maybank-am.com.my/list-of-funds/-/fund/view_entry/"

func (p MaybankFundProvider) FetchNavByDate(fund FundRef, date time.Time) (*NavResult, error) {
	u, err := url.Parse(baseUrl + fund.ScrapeParamValue)
	if err != nil {
		return nil, err
	}
	formatedFromDate := date.AddDate(0, 0, -3).Format("2006-01-02")
	formatedToDate := date.Format("2006-01-02")
	queryParam := u.Query()
	queryParam.Set("fromDate", formatedFromDate)
	queryParam.Set("toDate", formatedToDate)
	u.RawQuery = queryParam.Encode()
	html, err := scrape.GetHtmlStringFromUrlViaChrome(u.String())
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	return p.parseMaybankNavFromHtml(doc)
}

func (p MaybankFundProvider) FetchIncomeDistribution(fund FundRef, fromDate time.Time) ([]DistributionResult, error) {
	return nil, fmt.Errorf("so far Maybank Funds purchased dont have income distribution yet")
}

func (p MaybankFundProvider) parseMaybankNavFromHtml(doc *goquery.Document) (*NavResult, error) {
	navByDate, ok := doc.Find("input#fundHistoryData").First().Attr("value")
	if !ok {
		return nil, fmt.Errorf("navByDate not found in html")
	}
	log.Printf("navByDate: %s", navByDate)
	var result [][]string
	err := json.Unmarshal([]byte(navByDate), &result)
	if err != nil {
		return nil, err
	}
	// each string now is a "[\"yyyy-MM-dd\",\"1.2345\"]"
	latestNavArr := result[len(result)-1]
	log.Printf("latestNavArr: %s", latestNavArr)
	navDateStr := latestNavArr[0]
	navStr := latestNavArr[1]
	nav, _, err := apd.NewFromString(navStr)
	if err != nil {
		return nil, err
	}
	navDate, err := time.Parse("2006-01-02", navDateStr)
	if err != nil {
		return nil, err
	}
	return &NavResult{
		Nav:     nav,
		NavDate: navDate,
	}, nil
}
