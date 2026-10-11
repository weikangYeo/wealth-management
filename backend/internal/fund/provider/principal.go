package provider

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wealth-management/internal/platform/decimal"

	"github.com/PuerkitoBio/goquery"
)

type PrincipalFundProvider struct {
}

const principalBaseUrl = "https://www.principal.com.my/en/"

func (p PrincipalFundProvider) FetchNavByDate(fund FundRef, date time.Time) (*NavResult, error) {
	u, err := url.Parse(principalBaseUrl + fund.ScrapeParamValue)
	if err != nil {
		return nil, err
	}
	// make few day earlier because weekend and holiday dont have nav
	formatedFromDate := date.AddDate(0, 0, -3).Format("02-01-2006")
	formatedToDate := date.Format("02-01-2006")
	queryParam := u.Query()
	queryParam.Set("field_fund_nav_date_value[min]", formatedFromDate)
	queryParam.Set("field_fund_nav_date_value[max]", formatedToDate)
	u.RawQuery = queryParam.Encode()
	resp, err := http.Get(u.String())
	log.Printf("Fetching %s from %s\n", fund.ScrapeParamValue, u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Body is a html page
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return p.parseNavFromHtml(doc)

}

func (p PrincipalFundProvider) FetchIncomeDistribution(fund FundRef, fromDate time.Time) ([]DistributionResult, error) {
	// Principal funds, at least the one i'm holding, rarely have income dist,
	// main holding are put in bond funds and index funds
	u, err := url.Parse(principalBaseUrl + fund.ScrapeParamValue)
	if err != nil {
		return nil, err
	}
	formatedFromDate := fromDate.Format("01-02-2006")
	queryParam := u.Query()
	queryParam.Set("field_fund_nav_date_value[min]", formatedFromDate)
	u.RawQuery = queryParam.Encode()
	resp, err := http.Get(u.String())
	log.Printf("Fetching %s from %s\n", fund.ScrapeParamValue, u.String())
	defer resp.Body.Close()
	// Body is a html page
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseIncomeDistFromHtml(doc, fromDate)
}

func (p PrincipalFundProvider) parseNavFromHtml(doc *goquery.Document) (*NavResult, error) {
	// first time tag show latest nav datetime str
	timeTagSelector := doc.Find("time")
	navDateTimeStr, ok := timeTagSelector.Attr("datetime")
	log.Printf("Nav Date Time scaped: %s", navDateTimeStr)
	if !ok {
		return nil, fmt.Errorf("nav Datetime not found")
	}
	parsedTime, err := time.Parse(time.RFC3339, navDateTimeStr)
	if err != nil {
		return nil, err
	}
	// td next to its parent hold nav value
	navValueStr := timeTagSelector.
		Parent().
		NextFiltered("td.views-field-field-fund-nav").
		First().Text()
	log.Printf("navValueStr scraped: %s", navValueStr)
	nav, err := decimal.ToDecimal(navValueStr)
	if err != nil {
		return nil, err
	}

	return &NavResult{
		Nav:     nav,
		NavDate: parsedTime,
	}, nil
}

func parseIncomeDistFromHtml(doc *goquery.Document, fromDate time.Time) ([]DistributionResult, error) {
	// find td with class contain views-field-field-my-distribution-period-mon and dont have scope attr
	incomeDistList := make([]DistributionResult, 0)
	var scrapedErr error
	doc.Find("td.views-field-field-my-distribution-period-mon").
		Each(func(_ int, td *goquery.Selection) {
			tdValue := strings.TrimSpace(td.Text())
			log.Printf("Income Dist Date scraped: %s", tdValue)
			parsedTime, err := time.Parse("Jan-2006", tdValue)
			if err != nil {
				// found header row, malformat content
				log.Println("Skip malformat content")
				return
			}
			// set to last day of month, since dont have any structured public available to
			// tell us the exact exDate and paymentDate (which will make the calculation inaccurate
			// later too).
			// workaround 1: with nav comparison the whole month, found case that a income dist day
			// didnt has significant nav dropped.
			// workaround 2: a web search isnt guarantee the result, some fund from principal have its
			// own pdf/income dist (like prs)info some didnt.
			// probably an over-engineer approach is call LLM/AI agent, but I can't assert the result too,
			// it is non-determinative.
			lastDayOfMonth := time.Date(
				parsedTime.Year(),
				parsedTime.Month()+1,
				// day set to 0 so it is 1 day before
				0, 0, 0, 0, 0,
				parsedTime.Location(),
			)
			// all income dist in Principal website is show in 1 page and cannot filtered
			// after read unwanted row we will need to manually skip it.
			if lastDayOfMonth.After(fromDate) {
				dividendStr := td.NextFiltered("td.views-field-field-my-distribution-gross").First().Text()
				log.Printf("Income Dist value: %s", dividendStr)
				dividend, err := decimal.ToDecimal(dividendStr)
				if err != nil {
					log.Println("Error converting dividend to decimal")
					scrapedErr = err
				}
				incomeDist := DistributionResult{
					DeclareDate: lastDayOfMonth,
					PaymentDate: lastDayOfMonth,
					SenPerUnit:  dividend,
					Remark:      "Estimated Payment Date. Need to manually patch (unit and value, not date to avoid dup txn) from FSM/IUTA source after booked.",
				}
				incomeDistList = append(incomeDistList, incomeDist)
			}
		})
	if scrapedErr != nil {
		return nil, scrapedErr
	}
	return incomeDistList, nil
}
