package provider

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type NavResult struct {
	FundCode string
	Nav      apd.Decimal
	NavDate  time.Time
}

type DistributionResult struct {
	FundCode    string
	DeclareDate time.Time
	PaymentDate time.Time
	IncomeDist  apd.Decimal
}

type FundDataProvider interface {
	FetchLatestNav(fundCode string) (*NavResult, error)
	FetchIncomeDistribution(fundCode string) ([]DistributionResult, error)
}
