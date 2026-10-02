package twstock

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-sql/civil"
	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

func TestMarketDataService_DownloadMarginMaintenanceRatio(t *testing.T) {
	for _, tc := range []struct {
		name    string
		date    civil.Date
		days    int
		fixture string
		want    []MarginMaintenanceRatio
	}{
		{
			name: "trend", date: civil.Date{Year: 2026, Month: time.October, Day: 1}, days: 5,
			fixture: "margin_maintenance_ratio.json",
			want: []MarginMaintenanceRatio{
				{civil.Date{Year: 2026, Month: time.September, Day: 23}, decimal.RequireFromString("193.85")},
				{civil.Date{Year: 2026, Month: time.September, Day: 24}, decimal.RequireFromString("193.94")},
				{civil.Date{Year: 2026, Month: time.September, Day: 29}, decimal.RequireFromString("192.45")},
				{civil.Date{Year: 2026, Month: time.September, Day: 30}, decimal.RequireFromString("194.50")},
				{civil.Date{Year: 2026, Month: time.October, Day: 1}, decimal.RequireFromString("195.28")},
			},
		},
		{
			name: "first available day", date: civil.Date{Year: 2026, Month: time.August, Day: 3}, days: 1,
			fixture: "margin_maintenance_ratio_first.json",
			want: []MarginMaintenanceRatio{
				{civil.Date{Year: 2026, Month: time.August, Day: 3}, decimal.RequireFromString("178.76")},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, mux, teardown := setup()
			t.Cleanup(teardown)
			body, err := os.ReadFile("testdata/" + tc.fixture)
			if err != nil {
				t.Fatal(err)
			}
			mux.HandleFunc("GET /rwd/zh/marginTrading/BFIJ3U_TREND", func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if q.Get("response") != "json" || q.Get("date") != fmt.Sprintf("%04d%02d%02d", tc.date.Year, tc.date.Month, tc.date.Day) || q.Get("days") != fmt.Sprint(tc.days) {
					t.Errorf("unexpected query: %s", r.URL.RawQuery)
				}
				_, _ = w.Write(body)
			})
			got, err := client.MarketData.DownloadMarginMaintenanceRatio(tc.date, tc.days)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarketDataService_DownloadMarginMaintenanceRatioInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		date civil.Date
		days int
	}{
		{"before availability", civil.Date{Year: 2026, Month: time.August, Day: 2}, 1},
		{"invalid day", civil.Date{Year: 2026, Month: time.August, Day: 32}, 1},
		{"invalid month", civil.Date{Year: 2026, Month: 13, Day: 1}, 1},
		{"zero date", civil.Date{}, 1},
		{"zero days", civil.Date{Year: 2026, Month: time.August, Day: 3}, 0},
		{"negative days", civil.Date{Year: 2026, Month: time.August, Day: 3}, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, mux, teardown := setup()
			t.Cleanup(teardown)
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { t.Error("invalid input issued HTTP request") })
			got, err := client.MarketData.DownloadMarginMaintenanceRatio(tc.date, tc.days)
			if err == nil || got != nil {
				t.Fatalf("got %v, %v; want nil result and error", got, err)
			}
		})
	}
}

func TestMarketDataService_DownloadMarginMaintenanceRatioFailures(t *testing.T) {
	empty, err := os.ReadFile("testdata/margin_maintenance_ratio_empty.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{"HTTP failure", http.StatusServiceUnavailable, "", nil},
		{"malformed JSON", http.StatusOK, `{`, nil},
		{"API failure", http.StatusOK, `{"stat":"BAD"}`, nil},
		{"date out of range", http.StatusOK, `{"stat":"查詢日期大於今日，請重新查詢!"}`, ErrDateOutOffRange},
		{"official no data", http.StatusOK, string(empty), ErrNoData},
		{"empty data", http.StatusOK, `{"stat":"OK","data":[]}`, ErrNoData},
		{"invalid date", http.StatusOK, `{"stat":"OK","data":[{"date":"20260832","keepRate":178.76}]}`, nil},
		{"missing date", http.StatusOK, `{"stat":"OK","data":[{"keepRate":178.76}]}`, nil},
		{"missing ratio", http.StatusOK, `{"stat":"OK","data":[{"date":"20260803"}]}`, nil},
		{"null ratio", http.StatusOK, `{"stat":"OK","data":[{"date":"20260803","keepRate":null}]}`, nil},
		{"invalid ratio", http.StatusOK, `{"stat":"OK","data":[{"date":"20260803","keepRate":"BAD"}]}`, nil},
		{"invalid later row", http.StatusOK, `{"stat":"OK","data":[{"date":"20260803","keepRate":178.76},{"date":"20260804","keepRate":null}]}`, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, mux, teardown := setup()
			t.Cleanup(teardown)
			mux.HandleFunc("GET /rwd/zh/marginTrading/BFIJ3U_TREND", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			})
			got, err := client.MarketData.DownloadMarginMaintenanceRatio(civil.Date{Year: 2026, Month: time.October, Day: 1}, 5)
			if err == nil || got != nil {
				t.Fatalf("got %v, %v; want nil result and error", got, err)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
			if tc.status != http.StatusOK {
				if httpErr, ok := errors.AsType[*ErrorResponse](err); !ok || httpErr.Response.StatusCode != tc.status {
					t.Errorf("got %v, want HTTP status %d", err, tc.status)
				}
			}
		})
	}
}

func TestMarketDataService_DownloadMarginMaintenanceRatioTransportFailure(t *testing.T) {
	client, _, teardown := setup()
	teardown()
	got, err := client.MarketData.DownloadMarginMaintenanceRatio(civil.Date{Year: 2026, Month: time.August, Day: 3}, 1)
	if err == nil || got != nil {
		t.Fatalf("got %v, %v; want nil result and transport error", got, err)
	}
}
