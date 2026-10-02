package integration

import (
	"testing"
	"time"

	"github.com/golang-sql/civil"
	"github.com/miles170/twstock-go/twstock"
)

func TestMarketData_DownloadMarginMaintenanceRatio(t *testing.T) {
	client := twstock.NewClient()
	data, err := client.MarketData.DownloadMarginMaintenanceRatio(civil.Date{Year: 2026, Month: time.October, Day: 1}, 60)
	if err != nil {
		t.Fatalf("DownloadMarginMaintenanceRatio returned error: %v", err)
	}
	if len(data) != 42 {
		t.Fatalf("DownloadMarginMaintenanceRatio returned %d records, want 42", len(data))
	}
	first := data[0]
	if first.Date != (civil.Date{Year: 2026, Month: time.August, Day: 3}) || first.Ratio.String() != "178.76" {
		t.Fatalf("first record = %+v, want 2026-08-03 with ratio 178.76", first)
	}
	last := data[len(data)-1]
	if last.Date != (civil.Date{Year: 2026, Month: time.October, Day: 1}) || last.Ratio.String() != "195.28" {
		t.Fatalf("last record = %+v, want 2026-10-01 with ratio 195.28", last)
	}
}

func TestMarketData_DownloadMarginBalance(t *testing.T) {
	for _, tc := range []struct {
		market      twstock.Market
		count       int
		firstDate   civil.Date
		firstAmount string
		firstLots   int
		lastAmount  string
		lastLots    int
	}{
		{twstock.TWSE, 60, civil.Date{Year: 2026, Month: time.July, Day: 7}, "610945256000", 213844, "629867193000", 233322},
		{twstock.TPEx, 42, civil.Date{Year: 2026, Month: time.August, Day: 3}, "167008000000", 31855, "218872000000", 36462},
	} {
		t.Run(string(tc.market), func(t *testing.T) {
			date := civil.Date{Year: 2026, Month: time.October, Day: 1}
			data, err := twstock.NewClient().MarketData.DownloadMarginBalance(tc.market, date, 60)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) != tc.count {
				t.Fatalf("got %d records, want %d", len(data), tc.count)
			}
			first, last := data[0], data[len(data)-1]
			if first.Date != tc.firstDate || first.FinancingAmount.String() != tc.firstAmount || first.ShortSellingLots != tc.firstLots {
				t.Errorf("unexpected first record: %+v", first)
			}
			if last.Date != date || last.FinancingAmount.String() != tc.lastAmount || last.ShortSellingLots != tc.lastLots {
				t.Errorf("unexpected last record: %+v", last)
			}
		})
	}
}

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
