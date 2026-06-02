package provider

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type NavResult struct {
	FundCode string
	Nav      *apd.Decimal
	NavDate  time.Time
}

type DistributionResult struct {
	FundCode    string
	DeclareDate time.Time
	PaymentDate time.Time
	SenPerUnit  *apd.Decimal
}

type FundDataProvider interface {
	FetchLatestNav(fundCode string) (*NavResult, error)
	FetchNavByDate(fundCode string, date time.Time) (*NavResult, error)
	FetchIncomeDistribution(fundCode string, fromDate time.Time) ([]DistributionResult, error)
}
