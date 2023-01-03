package integration

import (
	"testing"

	"github.com/miles170/twstock-go/twstock"
)

func TestSecurity_DownloadTwseDelisted(t *testing.T) {
	client := twstock.NewClient()
	_, err := client.Security.DownloadTwseDelisted()
	if err != nil {
		t.Fatalf("DownloadTwse returned error: %v", err)
	}
}

func TestSecurity_DownloadTpexDelisted(t *testing.T) {
	client := twstock.NewClient()
	_, err := client.Security.DownloadTpexDelisted(0)
	if err != nil {
		t.Fatalf("DownloadTpex returned error: %v", err)
	}
}
