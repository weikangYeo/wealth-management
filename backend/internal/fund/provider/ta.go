package provider

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
	"wealth-management/internal/platform/decimal"

	"github.com/PuerkitoBio/goquery"
)

type TaFundProvider struct{}

const navBaseUrl = "https://aims.tainvest.com.my/mod/openFPrice/openFPrice_act_list_new.cfm"
const fundInfoBaseUrl = "https://www.tainvest.com.my"
const tempPdfFileName = "ta_factsheet.pdf"
const tempTextFileName = "ta_factsheet.txt"

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
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return p.parseNavFromHtml(doc, fund.Name)
}

// FetchIncomeDistribution TA income dist logic would be tricky, it require 2 steps.
// First go to Fund Info Page and scraped "Fund Fact Sheet" anchor link, then download the pdf
// Next use pdfToText to convert pdf to text file, read it (use best guess) and get income dist
func (p TaFundProvider) FetchIncomeDistribution(fund FundRef, fromDate time.Time) ([]DistributionResult, error) {
	// Get Fact Sheet PDF url from fund info html page
	u, err := url.Parse(fundInfoBaseUrl + "/" + fund.ScrapeParamValue + "/")
	if err != nil {
		return nil, err
	}
	log.Printf("Scraping %s from %s, by visiting %s", fund.Name, fund.ScrapeParamValue, u.String())
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	pdfUrl, err := p.parseFactSheetPdfUrlFromHtml(doc, fund.ScrapeParamValue)
	if err != nil {
		return nil, err
	}

	// Download Fund Fact Sheet PDF
	pdfUrlParsed, err := url.Parse(pdfUrl)
	if err != nil {
		return nil, err
	}
	out, err := os.Create(tempPdfFileName)
	defer func() {
		err := os.Remove(tempPdfFileName)
		if err != nil {
			log.Printf("Error removing temp pdf file %s", tempPdfFileName)
		}
	}()
	if err != nil {
		return nil, err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Printf("Error closing temp pdf file %s", tempPdfFileName)
		}
	}(out)
	pdfResp, err := http.Get(pdfUrlParsed.String())
	if err != nil {
		return nil, err
	}
	defer pdfResp.Body.Close()
	if pdfResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pdf responded with %s", pdfResp.Status)
	}
	if _, err := io.Copy(out, pdfResp.Body); err != nil {
		return nil, err
	}

	// Convert PDF to TXT
	// use os installed pdfToTxt library.
	// have tried https://github.com/ledongthuc/pdf but it has little to no geometry-aware reconstruction
	// i.e. the output is difficult to parse, too much effort to tinker it.
	pdfToTextCmd := exec.Command("pdftotext", "-layout", tempPdfFileName, tempTextFileName)
	defer func() {
		err := os.Remove(tempTextFileName)
		if err != nil {
			log.Printf("Error removing temp text file %s", tempTextFileName)
		}
	}()
	// binding stderr to cmd, so os can write to stderr (we assigned)
	// and can print stderr value if exit code != 0
	var stderr bytes.Buffer
	pdfToTextCmd.Stderr = &stderr
	if err := pdfToTextCmd.Run(); err != nil {
		return nil, fmt.Errorf("error running pdftotext: %s", stderr.String())
	}
	file, err := os.Open(tempTextFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return p.parseIncomeDistFromTextFile(file)
}

func (p TaFundProvider) parseFactSheetPdfUrlFromHtml(doc *goquery.Document, scrapeParamValue string) (string, error) {
	var factSheetPdf string
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, ok := s.Attr("href")
		if !ok {
			// continue to next tag
			return true
		}
		if strings.Contains(href, scrapeParamValue) &&
			strings.Contains(href, "pdf") &&
			strings.Contains(href, "ffs") &&
			strings.Contains(href, "fund-info") {
			// found the value we want
			factSheetPdf = href
			return false
		}
		return true
	})
	if factSheetPdf == "" {
		return "", fmt.Errorf("no fact sheet pdf found")
	}
	return factSheetPdf, nil
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
	if result.Nav == nil {
		if parseErr != nil {
			return nil, parseErr
		}
		return nil, fmt.Errorf("no nav found for %s", fundName)
	}
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
			log.Printf("Found keywords\n")
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
					log.Printf("Found MYR column, assigning index: %d", i)
					myrCol = i
					break
				}
			}
			continue
		}
		if isIncomeDistTable {
			log.Printf("Reading line in Income Dist Table %s", line)
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
