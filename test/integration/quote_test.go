package integration

import (
	"testing"

	"github.com/miles170/twstock-go/twstock"
)

func TestQuote_DownloadTwse(t *testing.T) {
	client := twstock.NewClient()
	_, err := client.Quote.DownloadTwse("2330", 2022, 12)
	if err != nil {
		t.Fatalf("DownloadTwse returned error: %v", err)
	}
	_, err = client.Quote.DownloadTwse("3374", 2022, 12)
	if err != twstock.ErrNoData {
		t.Fatalf("DownloadTwse error should be %v got %v", twstock.ErrNoData, err)
	}
}

func TestQuote_DownloadTpex(t *testing.T) {
	client := twstock.NewClient()
	_, err := client.Quote.DownloadTpex("3374", 2022, 12)
	if err != nil {
		t.Fatalf("DownloadTpex returned error: %v", err)
	}
	_, err = client.Quote.DownloadTpex("2330", 2022, 12)
	if err != twstock.ErrNoData {
		t.Fatalf("DownloadTpex error should be %v got %v", twstock.ErrNoData, err)
	}
}
