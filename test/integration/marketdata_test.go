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

func TestMarketData_DownloadTAIEX(t *testing.T) {
	client := twstock.NewClient()
	_, err := client.MarketData.DownloadTAIEX(1999, 1)
	if err != nil {
		t.Fatalf("DownloadTAIEX returned error: %v", err)
	}
}

func TestMarketData_DownloadTPExIndex(t *testing.T) {
	client := twstock.NewClient()
	_, err := client.MarketData.DownloadTPExIndex(1999, 9)
	if err != nil {
		t.Fatalf("DownloadTPExIndex returned error: %v", err)
	}
}
