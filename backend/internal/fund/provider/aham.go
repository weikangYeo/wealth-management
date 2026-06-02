package provider

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cockroachdb/apd/v3"
)

// todo test
type AhamFundProvider struct {
}

type fundNavResBody struct {
	Result []fundNavItem `json:"dailyfundperformanceResult"`
}

type fundNavItem struct {
	AsOfDate string  `json:"AsOfDate"`
	Fund     float64 `json:"Fund"`
}

type incomeDistResBody struct {
	Result []incomeDistItem `json:"incomedistributiondetailResult"`
}
type incomeDistItem struct {
	DeclareDate string  `json:"declareDate"`
	PaymentDate string  `json:"paymentDate"`
	IncomeDist  float64 `json:"IncomeDist"`
}

func (p *AhamFundProvider) FetchLatestNav(fundCode string) (*NavResult, error) {
	todayDate := time.Now().Format("2006-01-02")
	previousMonthDate := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	url := fmt.Sprintf("https://aham.com.my/clients/asset_C0C09289-21F6-4E4F-BA45-A8A98943FE33/api.ashx?action=daily_fund_performance&from_date=%s&to_date=%s&pf_code=%s",
		previousMonthDate, todayDate, fundCode)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response fundNavResBody
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	decimal := new(apd.Decimal)
	latestNavRes := response.Result[len(response.Result)-1]
	price, err := decimal.SetFloat64(latestNavRes.Fund)
	if err != nil {
		log.Printf("Error parsing daily fund price: %v", err)
		return nil, err
	}
	navDate, err := time.Parse("2026-01-02", latestNavRes.AsOfDate)
	if err != nil {
		log.Printf("Error parsing daily fund date: %v", err)
		return nil, err
	}
	return &NavResult{
		FundCode: fundCode,
		NavDate:  navDate,
		Nav:      price,
	}, nil
}

func (p *AhamFundProvider) FetchNavByDate(fundCode string, date time.Time) (*NavResult, error) {
	dateStr := date.Format("2006-01-02")
	url := fmt.Sprintf("https://aham.com.my/clients/asset_C0C09289-21F6-4E4F-BA45-A8A98943FE33/api.ashx?action=daily_fund_performance&from_date=%s&to_date=%s&pf_code=%s",
		dateStr, dateStr, fundCode)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response fundNavResBody
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if len(response.Result) == 0 {
		return nil, fmt.Errorf("empty response")
	}
	decimal := new(apd.Decimal)
	latestNavRes := response.Result[0]
	price, err := decimal.SetFloat64(latestNavRes.Fund)
	if err != nil {
		log.Printf("Error parsing daily fund price: %v", err)
		return nil, err
	}
	navDate, err := time.Parse("2026-01-02", latestNavRes.AsOfDate)
	if err != nil {
		log.Printf("Error parsing daily fund date: %v", err)
		return nil, err
	}
	return &NavResult{
		FundCode: fundCode,
		NavDate:  navDate,
		Nav:      price,
	}, nil
}

func (p *AhamFundProvider) FetchIncomeDistribution(fundCode string, fromDate time.Time) ([]DistributionResult, error) {
	filteredFromDate := ""
	if !fromDate.IsZero() {
		filteredFromDate = fromDate.Format("2006-01-02")
	}

	url := fmt.Sprintf("https://aham.com.my/clients/asset_C0C09289-21F6-4E4F-BA45-A8A98943FE33/api.ashx?action=income_distribution&pf_code=%s&from=%s",
		fundCode, filteredFromDate)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response incomeDistResBody
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	decimal := new(apd.Decimal)
	var result []DistributionResult
	for _, v := range response.Result {
		declareDate, err := time.Parse("2006-01-02", v.DeclareDate)
		if err != nil {
			log.Printf("Error parsing declareDate %s, error: %v", v.DeclareDate, err)
			continue
		}
		paymentDate, err := time.Parse("2006-01-02", v.PaymentDate)
		if err != nil {
			log.Printf("Error parsing paymentDate %s, error: %v", v.PaymentDate, err)
			continue
		}
		incomeDist, err := decimal.SetFloat64(v.IncomeDist)
		if err != nil {
			log.Printf("Error parsing incomeDist %s, error: %v", v.IncomeDist, err)
			continue
		}

		item := DistributionResult{
			FundCode:    fundCode,
			DeclareDate: declareDate,
			PaymentDate: paymentDate,
			SenPerUnit:  incomeDist,
		}
		result = append(result, item)
	}
	return result, nil
}
