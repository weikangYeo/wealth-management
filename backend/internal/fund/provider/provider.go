package provider

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type NavResult struct {
	Nav     *apd.Decimal
	NavDate time.Time
}

type DistributionResult struct {
	DeclareDate time.Time
	PaymentDate time.Time
	SenPerUnit  *apd.Decimal
}

type FundDataProvider interface {
	FetchNavByDate(scrapeParamValue string, date time.Time) (*NavResult, error)
	FetchIncomeDistribution(scrapeParamValue string, fromDate time.Time) ([]DistributionResult, error)
}
