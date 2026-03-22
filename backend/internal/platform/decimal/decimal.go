package decimal

import (
	"regexp"

	"github.com/cockroachdb/apd/v3"
)

func ToDecimal(str string) (*apd.Decimal, error) {
	numericStr := regexp.MustCompile("[^0-9.]").ReplaceAllString(str, "")
	numeric, _, err := apd.NewFromString(numericStr)
	if err != nil {
		return nil, err
	}
	return numeric, nil
}
