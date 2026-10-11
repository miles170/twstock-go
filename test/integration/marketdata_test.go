package integration

import (
	"testing"
	"time"

	"github.com/golang-sql/civil"
	"github.com/miles170/twstock-go/twstock"
)

// Query recent completed days so rolling upstream retention does not age out
// a fixed historical test window. The APIs handle weekends and holidays.
func dashboardQueryDate() civil.Date {
	return civil.DateOf(time.Now().In(time.FixedZone("Asia/Taipei", 8*60*60)).AddDate(0, 0, -1))
}

func TestMarketData_DownloadMarginMaintenanceRatio(t *testing.T) {
	date := dashboardQueryDate()
	const days = 60
	client := twstock.NewClient()
	data, err := client.MarketData.DownloadMarginMaintenanceRatio(date, days)
	if err != nil {
		t.Fatalf("DownloadMarginMaintenanceRatio returned error: %v", err)
	}
	// Exact dates and values are covered by fixture-based unit tests. Live
	// dashboards may return fewer records as their available history changes.
	if len(data) == 0 || len(data) > days {
		t.Fatalf("got %d records, want between 1 and %d", len(data), days)
	}
	for i, row := range data {
		if !row.Date.IsValid() || row.Date.After(date) {
			t.Errorf("record %d has invalid date %s for query ending %s", i, row.Date, date)
		}
		if i > 0 && !data[i-1].Date.Before(row.Date) {
			t.Errorf("dates are not strictly increasing: %s, %s", data[i-1].Date, row.Date)
		}
		if !row.Ratio.IsPositive() {
			t.Errorf("record %d has nonpositive ratio: %s", i, row.Ratio)
		}
	}
}

func TestMarketData_DownloadMarginBalance(t *testing.T) {
	date := dashboardQueryDate()
	const days = 60
	for _, market := range []twstock.Market{twstock.TWSE, twstock.TPEx} {
		t.Run(string(market), func(t *testing.T) {
			data, err := twstock.NewClient().MarketData.DownloadMarginBalance(market, date, days)
			if err != nil {
				t.Fatal(err)
			}
			// Availability and historical values can change upstream; exact
			// parsing and unit conversions are covered by fixture-based tests.
			if len(data) == 0 || len(data) > days {
				t.Fatalf("got %d records, want between 1 and %d", len(data), days)
			}
			for i, row := range data {
				if !row.Date.IsValid() || row.Date.After(date) {
					t.Errorf("record %d has invalid date %s for query ending %s", i, row.Date, date)
				}
				if i > 0 && !data[i-1].Date.Before(row.Date) {
					t.Errorf("dates are not strictly increasing: %s, %s", data[i-1].Date, row.Date)
				}
				if row.FinancingAmount.IsNegative() || row.ShortSellingLots < 0 {
					t.Errorf("record %d has negative balances: %+v", i, row)
				}
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
