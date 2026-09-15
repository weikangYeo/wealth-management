package provider

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wealth-management/internal/platform/decimal"

	"github.com/PuerkitoBio/goquery"
)

type TaFundProvider struct{}

const navBaseUrl = "https://aims.tainvest.com.my/mod/openFPrice/openFPrice_act_list_new.cfm"
const fundInfoBaseUrl = "https://www.tainvest.com.my"

func (p TaFundProvider) FetchNavByDate(fund FundRef, date time.Time) (*NavResult, error) {
	u, err := url.Parse(navBaseUrl)
	if err != nil {
		return nil, err
	}
	formatedDate := date.Format("02-01-2006")
	query := u.Query()
	query.Set("date", formatedDate)
	query.Set("tab", "1")
	query.Set("data", "01")
	u.RawQuery = query.Encode()
	resp, err := http.Get(u.String())
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return p.parseNavFromHtml(doc, fund.Name)
}

func (p TaFundProvider) FetchIncomeDistribution(fund FundRef, fromDate time.Time) ([]DistributionResult, error) {
	return nil, fmt.Errorf("FetchIncomeDistribution not yet implemented")
}

// make it as function receiver, so it will be private to TaFundProvider,
// else it would be package visible
func (p TaFundProvider) parseNavFromHtml(doc *goquery.Document, fundName string) (*NavResult, error) {
	var result NavResult
	var parseErr error
	doc.Find("td").EachWithBreak(func(i int, s *goquery.Selection) bool {
		scrapedFundName := strings.ToUpper(strings.TrimSpace(s.Text()))
		if scrapedFundName == strings.ToUpper(fundName) {
			navDate, err := time.Parse("02/01/2006", s.Next().Text())
			if err != nil {
				parseErr = err
				return false
			}
			nav, err := decimal.ToDecimal(s.Next().Next().Text())
			if err != nil {
				parseErr = err
				return false
			}
			result = NavResult{
				Nav:     nav,
				NavDate: navDate,
			}
		}
		return true
	})
	return &result, parseErr
}

func (p TaFundProvider) parseIncomeDistFromTextFile(r io.Reader) ([]DistributionResult, error) {
	isIncomeDistTable := false
	hasIncomeDistRead := false
	myrCol := 0
	scanner := bufio.NewScanner(r)
	result := make([]DistributionResult, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Fund Distribution") && !strings.Contains(line, "Unit Split History") {
			// found the table header, set flag to indicate next iteration to start scrape logic
			isIncomeDistTable = true
			// try to find which column are the myr column
			line = strings.ReplaceAll(line, "Fund Distribution", "FundDistribution")
			// consolidate "MYR Hedged" to "MYRHedged" before call `.Field()`
			// so we can tell it is MYR or MYRHedged instead of MYR, MYR, Hedged
			line = strings.ReplaceAll(line, " Hedged", "Hedged")
			columns := strings.Fields(line)
			for i, column := range columns {
				if (strings.TrimSpace(column)) == "MYR" {
					myrCol = i
					break
				}
			}
			continue
		}
		if isIncomeDistTable {
			log.Printf("Reading line %s", line)
			columns := strings.Fields(line)
			if len(columns) == 0 {
				if hasIncomeDistRead {
					log.Printf("Income Dist table finish read")
					break
				}
				log.Printf("skipping empty line")
				continue
			}
			declareDate, err := time.Parse("02/01/2006", columns[0])
			if err != nil {
				log.Printf("%s is not a valid date, skipping", columns[0])
				continue
			}
			dividend, err := decimal.ToDecimal(columns[myrCol])
			if err != nil {
				// ta MYR class history contain invalid/0 value, just skip if cannot parse
				log.Printf("%s is not a valid digit, skipping", columns[myrCol])
				continue
			}
			incomeDist := DistributionResult{
				PaymentDate: declareDate,
				DeclareDate: declareDate,
				SenPerUnit:  dividend,
			}
			result = append(result, incomeDist)
			hasIncomeDistRead = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
