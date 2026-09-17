package gold

import (
	"os"
	"testing"
)

func TestGetGoldPriceStrFromHtml(t *testing.T) {
	contentBytes, err := os.ReadFile("testdata/gold_price_maybank.html")
	if err != nil {
		t.Fatalf("Failed to read gold price file: %v", err)
	}
	content := string(contentBytes)
	result := getGoldPriceStrFromHtml(content)
	if result != "557.06" {
		t.Fatalf("Failed to get gold price: %v", result)
	}
}
