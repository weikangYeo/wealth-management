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

// FundRef is what a provider needs to locate a fund at its data source. Not every
// provider uses every field — e.g. AHAM/Principal only use ScrapeParamValue (a
// provider-specific code or URL fragment addressing the fund directly); TA uses Name
// to match a row in its shared NAV listing and ScrapeParamValue for its per-fund
// page/PDF path.
type FundRef struct {
	Name             string
	ScrapeParamValue string
}

type FundDataProvider interface {
	FetchNavByDate(fund FundRef, date time.Time) (*NavResult, error)
	FetchIncomeDistribution(fund FundRef, fromDate time.Time) ([]DistributionResult, error)
}
