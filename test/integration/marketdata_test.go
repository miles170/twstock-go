package integration

import (
	"testing"

	"github.com/miles170/twstock-go/twstock"
)

func TestMarketData_DownloadTwse(t *testing.T) {
	client := twstock.NewClient()
	_, err := client.MarketData.DownloadTwse(2022, 12)
	if err != nil {
		t.Fatalf("DownloadTwse returned error: %v", err)
	}
}

func TestMarketData_DownloadTpex(t *testing.T) {
	client := twstock.NewClient()
	_, err := client.MarketData.DownloadTpex(2022, 12)
	if err != nil {
		t.Fatalf("DownloadTpex returned error: %v", err)
	}
}
