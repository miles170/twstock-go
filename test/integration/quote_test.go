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
}

func TestQuote_DownloadTpex(t *testing.T) {
	client := twstock.NewClient()
	_, err := client.Quote.DownloadTpex("3374", 2022, 12)
	if err != nil {
		t.Fatalf("DownloadTwse returned error: %v", err)
	}
}
