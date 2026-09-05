package provider

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/cockroachdb/apd/v3"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

const ahamBaseURL = "https://aham.com.my/clients/asset_C0C09289-21F6-4E4F-BA45-A8A98943FE33/api.ashx"

func decodeJSON(r io.Reader, v any) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	body = bytes.TrimPrefix(body, utf8BOM)
	return json.Unmarshal(body, v)
}

// todo test
type AhamFundProvider struct {
}

type fundNavResBody struct {
	Result []fundNavItem `json:"productnavhistoricalResult"`
}

type fundNavItem struct {
	NavDate string  `json:"NAVDate"`
	Nav     float64 `json:"NAV"`
}

type incomeDistResBody struct {
	Result []incomeDistItem `json:"incomedistributiondetailResult"`
}
type incomeDistItem struct {
	DeclareDate string  `json:"declareDate"`
	PaymentDate string  `json:"paymentDate"`
	IncomeDist  float64 `json:"IncomeDist"`
}

func (p *AhamFundProvider) FetchNavByDate(scrapeParamValue string, date time.Time) (*NavResult, error) {
	u, err := url.Parse(ahamBaseURL)
	if err != nil {
		return nil, err
	}
	// make few day earlier because weekend and holiday dont have nav
	formatedFromDate := date.AddDate(0, 0, -3).Format("2006-01-02")
	formatedToDate := date.Format("2006-01-02")
	// build url
	queryParam := u.Query()
	queryParam.Set("action", "product_nav_historical")
	queryParam.Set("from_date", formatedFromDate)
	queryParam.Set("to_date", formatedToDate)
	queryParam.Set("pf_code", scrapeParamValue)
	u.RawQuery = queryParam.Encode()
	resp, err := http.Get(u.String())
	log.Printf("Fetching %s from %s\n", scrapeParamValue, u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response fundNavResBody
	log.Printf("Decoding response %s\n", resp.Status)
	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}
	decimal := new(apd.Decimal)
	log.Printf("Getting latest nav, last index of response")
	latestNavRes := response.Result[len(response.Result)-1]
	price, err := decimal.SetFloat64(latestNavRes.Nav)
	if err != nil {
		log.Printf("Error parsing daily fund price: %v", err)
		return nil, err
	}
	navDate, err := time.Parse("2006-01-02", latestNavRes.NavDate)
	if err != nil {
		log.Printf("Error parsing daily fund date: %v", err)
		return nil, err
	}
	return &NavResult{
		NavDate: navDate,
		Nav:     price,
	}, nil
}

func (p *AhamFundProvider) FetchIncomeDistribution(scrapeParamValue string, fromDate time.Time) ([]DistributionResult, error) {
	filteredFromDate := ""
	if !fromDate.IsZero() {
		filteredFromDate = fromDate.Format("2006-01-02")
	}
	u, err := url.Parse(ahamBaseURL)
	if err != nil {
		return nil, err
	}
	queryParam := u.Query()
	queryParam.Set("action", "income_distribution")
	queryParam.Set("from_date", filteredFromDate)
	queryParam.Set("pf_code", scrapeParamValue)
	u.RawQuery = queryParam.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response incomeDistResBody
	if err := decodeJSON(resp.Body, &response); err != nil {
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
			log.Printf("Error parsing incomeDist %v, error: %v", v.IncomeDist, err)
			continue
		}

		item := DistributionResult{
			DeclareDate: declareDate,
			PaymentDate: paymentDate,
			SenPerUnit:  incomeDist,
		}
		result = append(result, item)
	}
	return result, nil
}
