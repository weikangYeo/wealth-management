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

func (p PrincipalFundProvider) FetchNavByDate(scrapeParamValue string, date time.Time) (*NavResult, error) {
	u, err := url.Parse(principalBaseUrl + scrapeParamValue)
	if err != nil {
		return nil, err
	}
	// make few day earlier because weekend and holiday dont have nav
	formatedFromDate := date.AddDate(0, 0, -3).Format("01-02-2006")
	formatedToDate := date.Format("01-02-2006")
	queryParam := u.Query()
	queryParam.Set("field_fund_nav_date_value[min]", formatedFromDate)
	queryParam.Set("field_fund_nav_date_value[max]", formatedToDate)
	queryParam.Encode()
	u.RawQuery = queryParam.Encode()
	resp, err := http.Get(u.String())
	log.Printf("Fetching %s from %s\n", scrapeParamValue, u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Body is a html page
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseNavFromHtml(doc)

}

func (p PrincipalFundProvider) FetchIncomeDistribution(scrapeParamValue string, fromDate time.Time) ([]DistributionResult, error) {
	// Principal funds, at least the one i'm holding, rarely have income dist,
	// main holding are put in bond funds and index funds
	u, err := url.Parse(principalBaseUrl + scrapeParamValue)
	if err != nil {
		return nil, err
	}
	formatedFromDate := fromDate.Format("01-02-2006")
	queryParam := u.Query()
	queryParam.Set("field_fund_nav_date_value[min]", formatedFromDate)
	queryParam.Encode()
	u.RawQuery = queryParam.Encode()
	resp, err := http.Get(u.String())
	log.Printf("Fetching %s from %s\n", scrapeParamValue, u.String())
	defer resp.Body.Close()
	// Body is a html page
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseIncomeDistFromHtml(doc)
}

func parseNavFromHtml(doc *goquery.Document) (*NavResult, error) {
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

func parseIncomeDistFromHtml(doc *goquery.Document) ([]DistributionResult, error) {
	// find td with class contain views-field-field-my-distribution-period-mon and dont have scope attr
	incomeDistList := make([]DistributionResult, 0)
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
			// set to last day of month, since Principal dont have exact day date in income dist
			// day set to 0 so it is 1 day before
			lastDayOfMonth := time.Date(
				parsedTime.Year(),
				parsedTime.Month()+1,
				0, 0, 0, 0, 0,
				parsedTime.Location(),
			)
			dividendStr := td.NextFiltered("td.views-field-field-my-distribution-gross").First().Text()
			log.Printf("Income Dist value: %s", dividendStr)
			dividend, err := decimal.ToDecimal(dividendStr)
			if err != nil {
				log.Fatalln("Cannot parse Decimal")
				return
			}
			incomeDist := DistributionResult{
				DeclareDate: lastDayOfMonth,
				PaymentDate: lastDayOfMonth,
				SenPerUnit:  dividend,
			}
			incomeDistList = append(incomeDistList, incomeDist)
		})

	return incomeDistList, nil
}
