package twstock

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-sql/civil"
	"github.com/shopspring/decimal"
)

func TestMarketDataService_DownloadMarginBalance(t *testing.T) {
	for _, tc := range []struct {
		market     Market
		fixture    string
		date       civil.Date
		days       int
		wantAmount string
		wantLots   int
	}{
		{TWSE, "margin_balance_twse.json", civil.Date{Year: 2026, Month: time.October, Day: 1}, 5, "629867193000", 233322},
		{TPEx, "margin_balance_tpex.json", civil.Date{Year: 2026, Month: time.October, Day: 1}, 5, "218872000000", 36462},
		{TWSE, "margin_balance_twse_first.json", civil.Date{Year: 2026, Month: time.August, Day: 3}, 1, "514755393000", 185322},
		{TPEx, "margin_balance_tpex_first.json", civil.Date{Year: 2026, Month: time.August, Day: 3}, 1, "167008000000", 31855},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			client, mux, teardown := setup()
			t.Cleanup(teardown)
			body, err := os.ReadFile("testdata/" + tc.fixture)
			if err != nil {
				t.Fatal(err)
			}
			path, date := "/rwd/zh/marginTrading/MI_MARGN_TREND", fmt.Sprintf("%04d%02d%02d", tc.date.Year, tc.date.Month, tc.date.Day)
			if tc.market == TPEx {
				path, date = "/www/zh-tw/dashboardOtc/marginTrend", fmt.Sprintf("%03d/%02d/%02d", tc.date.Year-1911, tc.date.Month, tc.date.Day)
			}
			mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if q.Get("date") != date || q.Get("days") != fmt.Sprint(tc.days) {
					t.Errorf("unexpected query: %s", r.URL.RawQuery)
				}
				if tc.market == TWSE && q.Get("response") != "json" {
					t.Error("missing JSON response parameter")
				}
				if tc.market == TPEx && q.Get("lang") != "zh-tw" {
					t.Error("missing TPEx language parameter")
				}
				_, _ = w.Write(body)
			})
			got, err := client.MarketData.DownloadMarginBalance(tc.market, tc.date, tc.days)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tc.days {
				t.Fatalf("got %d records, want %d", len(got), tc.days)
			}
			last := got[len(got)-1]
			if last.Date != tc.date || !last.FinancingAmount.Equal(decimal.RequireFromString(tc.wantAmount)) || last.ShortSellingLots != tc.wantLots {
				t.Errorf("last record = %+v, want %s, %s yuan, %d lots", last, tc.date, tc.wantAmount, tc.wantLots)
			}
			if tc.days == 5 && got[0].Date != (civil.Date{Year: 2026, Month: time.September, Day: 23}) {
				t.Errorf("unexpected first date: %s", got[0].Date)
			}
		})
	}
}

func TestMarketDataService_DownloadMarginBalanceInvalidInput(t *testing.T) {
	for _, market := range []Market{TWSE, TPEx, Market("invalid")} {
		for _, tc := range []struct {
			date civil.Date
			days int
		}{
			{civil.Date{Year: 2026, Month: time.August, Day: 2}, 1},
			{civil.Date{Year: 2026, Month: time.August, Day: 32}, 1},
			{civil.Date{Year: 2026, Month: time.August, Day: 3}, 0},
			{civil.Date{Year: 2026, Month: time.August, Day: 3}, -1},
		} {
			client, mux, teardown := setup()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { t.Error("invalid input issued HTTP request") })
			got, err := client.MarketData.DownloadMarginBalance(market, tc.date, tc.days)
			teardown()
			if err == nil || got != nil {
				t.Fatalf("got %v, %v; want nil result and error", got, err)
			}
		}
	}
	client, _, teardown := setup()
	t.Cleanup(teardown)
	if _, err := client.MarketData.DownloadMarginBalance(Market("invalid"), civil.Date{Year: 2026, Month: time.August, Day: 3}, 1); err == nil {
		t.Fatal("invalid market accepted")
	}
}

func TestMarketDataService_DownloadMarginBalanceFailures(t *testing.T) {
	empty, err := os.ReadFile("testdata/margin_balance_tpex_empty.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, market := range []Market{TWSE, TPEx} {
		path, stat, key, amount, lots := "/rwd/zh/marginTrading/MI_MARGN_TREND", "OK", "data", "marginAmt", "shortShr"
		if market == TPEx {
			path, stat, key, amount, lots = "/www/zh-tw/dashboardOtc/marginTrend", "ok", "trend", "marginBalance", "shortBalance"
		}
		validRow := fmt.Sprintf(`{"date":"20260803","%s":1,"%s":2}`, amount, lots)
		for _, tc := range []struct {
			name    string
			status  int
			body    string
			wantErr error
		}{
			{"HTTP", 503, "", nil},
			{"JSON", 200, `{`, nil},
			{"state", 200, `{"stat":"BAD"}`, nil},
			{"no data", 200, `{"stat":"很抱歉，沒有符合條件的資料!"}`, ErrNoData},
			{"official no data", 200, string(empty), ErrNoData},
			{"date out of range", 200, `{"stat":"查詢日期大於今日，請重新查詢!"}`, ErrDateOutOffRange},
			{"empty", 200, fmt.Sprintf(`{"stat":%q,"%s":[]}`, stat, key), ErrNoData},
			{"date", 200, fmt.Sprintf(`{"stat":%q,"%s":[{"date":"20260832","%s":1,"%s":2}]}`, stat, key, amount, lots), nil},
			{"missing amount", 200, fmt.Sprintf(`{"stat":%q,"%s":[{"date":"20260803","%s":2}]}`, stat, key, lots), nil},
			{"null lots", 200, fmt.Sprintf(`{"stat":%q,"%s":[{"date":"20260803","%s":1,"%s":null}]}`, stat, key, amount, lots), nil},
			{"bad amount", 200, fmt.Sprintf(`{"stat":%q,"%s":[{"date":"20260803","%s":"BAD","%s":2}]}`, stat, key, amount, lots), nil},
			{"fractional lots", 200, fmt.Sprintf(`{"stat":%q,"%s":[{"date":"20260803","%s":1,"%s":2.5}]}`, stat, key, amount, lots), nil},
			{"later bad row", 200, fmt.Sprintf(`{"stat":%q,"%s":[%s,{}]}`, stat, key, validRow), nil},
		} {
			t.Run(string(market)+"/"+tc.name, func(t *testing.T) {
				client, mux, teardown := setup()
				t.Cleanup(teardown)
				mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); fmt.Fprint(w, tc.body) })
				got, err := client.MarketData.DownloadMarginBalance(market, civil.Date{Year: 2026, Month: time.October, Day: 1}, 5)
				if err == nil || got != nil {
					t.Fatalf("got %v, %v; want nil result and error", got, err)
				}
				if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
					t.Errorf("got %v, want %v", err, tc.wantErr)
				}
			})
		}
	}
}

func TestMarketDataService_DownloadMarginBalanceZeroBalances(t *testing.T) {
	for _, market := range []Market{TWSE, TPEx} {
		t.Run(string(market), func(t *testing.T) {
			client, mux, teardown := setup()
			t.Cleanup(teardown)
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if market == TWSE {
					fmt.Fprint(w, `{"stat":"OK","data":[{"date":"20260803","marginAmt":0,"shortShr":0}]}`)
				} else {
					fmt.Fprint(w, `{"stat":"ok","trend":[{"date":"20260803","marginBalance":0,"shortBalance":0}]}`)
				}
			})
			got, err := client.MarketData.DownloadMarginBalance(market, civil.Date{Year: 2026, Month: time.August, Day: 3}, 5)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || !got[0].FinancingAmount.IsZero() || got[0].ShortSellingLots != 0 {
				t.Errorf("got %+v, want one record with zero balances", got)
			}
		})
	}
}
