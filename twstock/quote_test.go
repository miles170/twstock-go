package twstock

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-sql/civil"
	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

func TestQuoteService_DownloadTwse(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stat": "OK",
			"date": "20220801",
			"title": "111年08月 2330 台積電           各日成交資訊",
			"fields": [
			  "日期",
			  "成交股數",
			  "成交金額",
			  "開盤價",
			  "最高價",
			  "最低價",
			  "收盤價",
			  "漲跌價差",
			  "成交筆數"
			],
			"data": [
			  [
				"111/08/01",
				"24,991,291",
				"12,569,771,761",
				"506.00",
				"508.00",
				"500.00",
				"504.00",
				"-5.00",
				"26,792"
			  ],
			  [
				"111/08/02",
				"42,669,591",
				"20,973,293,337",
				"494.00",
				"496.00",
				"488.50",
				"492.00",
				"-12.00",
				"63,879"
			  ],
			  [
				"111/08/03",
				"29,838,832",
				"14,823,224,632",
				"494.00",
				"--",
				"493.00",
				"501.00",
				"+9.00",
				"25,570"
			  ],
			  [
				"111/08/04",
				"26,589,086",
				"13,279,624,282",
				"499.00",
				"503.00",
				"495.00",
				"500.00",
				"-1.00",
				"27,173"
			  ],
			  [
				"111/08/05",
				"35,052,642",
				"17,966,410,242",
				"509.00",
				"516.00",
				"507.00",
				"516.00",
				"+16.00",
				"49,928"
			  ]
			],
			"notes": [
			  "符號說明:+/-/X表示漲/跌/不比價",
			  "當日統計資訊僅含一般交易，不含零股、盤後定價、鉅額、拍賣、標購。",
			  "ETF證券代號第六碼為K、M、S、C者，表示該ETF以外幣交易。"
			]
		  }`)
	})

	quotes, err := client.Quote.Download("2330", 2022, 8)
	if err != nil {
		t.Errorf("Quote.Download returned error: %v", err)
	}
	want := []Quote{
		{
			Date:   civil.Date{Year: 2022, Month: time.August, Day: 1},
			Open:   decimal.NewFromFloat(506),
			High:   decimal.NewFromFloat(508),
			Low:    decimal.NewFromFloat(500),
			Close:  decimal.NewFromFloat(504),
			Volume: 24991291,
		},
		{
			Date:   civil.Date{Year: 2022, Month: time.August, Day: 2},
			Open:   decimal.NewFromFloat(494),
			High:   decimal.NewFromFloat(496),
			Low:    decimal.NewFromFloat(488.50),
			Close:  decimal.NewFromFloat(492),
			Volume: 42669591,
		},
		{
			Date:   civil.Date{Year: 2022, Month: time.August, Day: 4},
			Open:   decimal.NewFromFloat(499),
			High:   decimal.NewFromFloat(503),
			Low:    decimal.NewFromFloat(495),
			Close:  decimal.NewFromFloat(500),
			Volume: 26589086,
		},
		{
			Date:   civil.Date{Year: 2022, Month: time.August, Day: 5},
			Open:   decimal.NewFromFloat(509),
			High:   decimal.NewFromFloat(516),
			Low:    decimal.NewFromFloat(507),
			Close:  decimal.NewFromFloat(516),
			Volume: 35052642,
		},
	}
	if !cmp.Equal(quotes, want) {
		t.Errorf("Quote.Download returned %v, want %v", quotes, want)
	}
}

func TestQuoteService_DownloadTwseError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusBadRequest)
	})

	_, err := client.Quote.DownloadTwse("2330", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTwse returned nil; expected error")
	}
	testErrorContains(t, err, ": 400")

	_, err = client.Quote.DownloadTwse("", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTwse returned nil; expected error")
	}

	_, err = client.Quote.Download("", 2022, 8)
	if err == nil {
		t.Error("Quote.Download returned nil; expected error")
	}

	_, err = client.Quote.DownloadTwse("2330", 2009, 12)
	if err == nil {
		t.Error("Quote.DownloadTwse returned nil; expected error")
	}
}

func TestQuoteService_DownloadTwseErrNoData(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stat": "很抱歉，沒有符合條件的資料!",
			"date": "20220801",
			"title": "111年08月 2330 台積電           各日成交資訊",
			"fields": [],
			"notes": [
			  "符號說明:+/-/X表示漲/跌/不比價",
			  "當日統計資訊僅含一般交易，不含零股、盤後定價、鉅額、拍賣、標購。",
			  "ETF證券代號第六碼為K、M、S、C者，表示該ETF以外幣交易。"
			]
		  }`)
	})

	_, err := client.Quote.DownloadTwse("2330", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTwse returned nil; expected error")
	}
	if !errors.Is(err, ErrNoData) {
		t.Errorf("Quote.DownloadTwse returned %v, want %v", err, ErrNoData)
	}
}

func TestQuoteService_DownloadTwseErrDateOutOffRange(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stat": "查詢日期大於今日，請重新查詢!",
			"date": "20220801",
			"title": "111年08月 2330 台積電           各日成交資訊",
			"fields": [],
			"notes": [
			  "符號說明:+/-/X表示漲/跌/不比價",
			  "當日統計資訊僅含一般交易，不含零股、盤後定價、鉅額、拍賣、標購。",
			  "ETF證券代號第六碼為K、M、S、C者，表示該ETF以外幣交易。"
			]
		  }`)
	})

	_, err := client.Quote.DownloadTwse("2330", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTwse returned nil; expected error")
	}
	if !errors.Is(err, ErrDateOutOffRange) {
		t.Errorf("Quote.DownloadTwse returned %v, want %v", err, ErrDateOutOffRange)
	}
}

func TestQuoteService_DownloadTwseBadStat(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stat": "FAIL",
			"date": "20220801",
			"title": "111年08月 2330 台積電           各日成交資訊",
			"fields": [
			  "日期",
			  "成交股數",
			  "成交金額",
			  "開盤價",
			  "最高價",
			  "最低價",
			  "收盤價",
			  "漲跌價差",
			  "成交筆數"
			],
			"data": [],
			"notes": [
			  "符號說明:+/-/X表示漲/跌/不比價",
			  "當日統計資訊僅含一般交易，不含零股、盤後定價、鉅額、拍賣、標購。",
			  "ETF證券代號第六碼為K、M、S、C者，表示該ETF以外幣交易。"
			]
		  }`)
	})

	_, err := client.Quote.DownloadTwse("2330", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTwse returned nil; expected error")
	}
}

func TestQuoteService_DownloadTwseBadFields(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stat": "OK",
			"date": "20220801",
			"title": "111年08月 2330 台積電           各日成交資訊",
			"fields": [
			  "日期",
			  "成交股數",
			  "成交金額",
			  "開盤價",
			  "最高價",
			  "最低價",
			  "=====",
			  "漲跌價差",
			  "成交筆數"
			],
			"data": [],
			"notes": [
			  "符號說明:+/-/X表示漲/跌/不比價",
			  "當日統計資訊僅含一般交易，不含零股、盤後定價、鉅額、拍賣、標購。",
			  "ETF證券代號第六碼為K、M、S、C者，表示該ETF以外幣交易。"
			]
		  }`)
	})

	_, err := client.Quote.DownloadTwse("2330", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTwse returned nil; expected error")
	}
}

func TestQuoteService_DownloadTwseBadData(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stat": "OK",
			"date": "20220801",
			"title": "111年08月 2330 台積電           各日成交資訊",
			"fields": [
			  "日期",
			  "成交股數",
			  "成交金額",
			  "開盤價",
			  "最高價",
			  "最低價",
			  "收盤價",
			  "漲跌價差",
			  "成交筆數"
			],
			"data": [
				[
				"111/0801",
				"24,991,291",
				"12,569,771,761",
				"506.00",
				"508.00",
				"500.00",
				"504.00",
				"-5.00",
				"26,792"
			  ]
			],
			"notes": [
			  "符號說明:+/-/X表示漲/跌/不比價",
			  "當日統計資訊僅含一般交易，不含零股、盤後定價、鉅額、拍賣、標購。",
			  "ETF證券代號第六碼為K、M、S、C者，表示該ETF以外幣交易。"
			]
		  }`)
	})

	_, err := client.Quote.DownloadTwse("2330", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTwse returned nil; expected error")
	}
}

func TestQuoteService_DownloadTpex(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stkNo": "3374",
			"stkName": "精材            ",
			"showListPriceNote": false,
			"showListPriceLink": false,
			"reportDate": "111/08",
			"iTotalRecords": 5,
			"aaData": [
			  [
				"111/08/01",
				"1,328",
				"168,265",
				"127.50",
				"128.00",
				"125.50",
				"127.00",
				"-2.00",
				"1,272"
			  ],
			  [
				"111/08/02",
				"1,593",
				"199,305",
				"125.00",
				"127.00",
				"123.00",
				"127.00",
				"0.00",
				"1,078"
			  ],
			  [
				"111/08/03",
				"1,603",
				"201,304",
				"124.50",
				"127.00",
				"124.00",
				"126.00",
				"-1.00",
				"1,124"
			  ],
			  [
				"111/08/04",
				"3,920",
				"500,389",
				"--",
				"130.00",
				"124.50",
				"129.50",
				"3.50",
				"2,474"
			  ],
			  [
				"111/08/05",
				"7,244",
				"940,126",
				"129.50",
				"132.00",
				"126.50",
				"129.50",
				"0.00",
				"5,073"
			  ]
			]
		  }`)
	})

	quotes, err := client.Quote.Download("3374", 2022, 8)
	if err != nil {
		t.Errorf("Quote.Download returned error: %v", err)
	}
	want := []Quote{
		{
			Date:   civil.Date{Year: 2022, Month: time.August, Day: 1},
			Open:   decimal.NewFromFloat(127.5),
			High:   decimal.NewFromFloat(128),
			Low:    decimal.NewFromFloat(125.5),
			Close:  decimal.NewFromFloat(127),
			Volume: 1328000,
		},
		{
			Date:   civil.Date{Year: 2022, Month: time.August, Day: 2},
			Open:   decimal.NewFromFloat(125),
			High:   decimal.NewFromFloat(127),
			Low:    decimal.NewFromFloat(123),
			Close:  decimal.NewFromFloat(127),
			Volume: 1593000,
		},
		{
			Date:   civil.Date{Year: 2022, Month: time.August, Day: 3},
			Open:   decimal.NewFromFloat(124.5),
			High:   decimal.NewFromFloat(127),
			Low:    decimal.NewFromFloat(124),
			Close:  decimal.NewFromFloat(126),
			Volume: 1603000,
		},
		{
			Date:   civil.Date{Year: 2022, Month: time.August, Day: 5},
			Open:   decimal.NewFromFloat(129.5),
			High:   decimal.NewFromFloat(132),
			Low:    decimal.NewFromFloat(126.5),
			Close:  decimal.NewFromFloat(129.5),
			Volume: 7244000,
		},
	}
	if !cmp.Equal(quotes, want) {
		t.Errorf("Quote.Download returned %v, want %v", quotes, want)
	}
}

func TestQuoteService_DownloadTpexError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusBadRequest)
	})

	_, err := client.Quote.DownloadTpex("3374", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTpex returned nil; expected error")
	}
	testErrorContains(t, err, ": 400")

	_, err = client.Quote.DownloadTpex("", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTpex returned nil; expected error")
	}

	_, err = client.Quote.DownloadTpex("3374", 1993, 12)
	if err == nil {
		t.Error("Quote.DownloadTpex returned nil; expected error")
	}
}

func TestQuoteService_DownloadTpexErrNoData(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stkNo": "3374",
			"stkName": "精材            ",
			"showListPriceNote": false,
			"showListPriceLink": false,
			"reportDate": "111/08",
			"iTotalRecords": 0
		}`)
	})

	_, err := client.Quote.Download("3374", 2022, 8)
	if err == nil {
		t.Errorf("Quote.Download returned nil; expected error")
	}
	if !errors.Is(err, ErrNoData) {
		t.Errorf("Quote.Download returned %v, want %v", err, ErrNoData)
	}
}

func TestQuoteService_DownloadBadCode(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stkNo": "3375",
			"stkName": "精材            ",
			"showListPriceNote": false,
			"showListPriceLink": false,
			"reportDate": "111/08",
			"iTotalRecords": 5,
			"aaData": []
		  }`)
	})

	_, err := client.Quote.DownloadTpex("3374", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTpex returned nil; expected error")
	}
}

func TestQuoteService_DownloadBadDataLength(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stkNo": "3374",
			"stkName": "精材            ",
			"showListPriceNote": false,
			"showListPriceLink": false,
			"reportDate": "111/08",
			"iTotalRecords": 5,
			"aaData": []
		  }`)
	})

	_, err := client.Quote.DownloadTpex("3374", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTpex returned nil; expected error")
	}
}

func TestQuoteService_DownloadBadFieldCount(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stkNo": "3374",
			"stkName": "精材            ",
			"showListPriceNote": false,
			"showListPriceLink": false,
			"reportDate": "111/08",
			"iTotalRecords": 1,
			"aaData": [
				[
					"111/08/01",
					"1,328",
					"168,265",
					"127.50",
					"128.00",
					"125.50",
					"127.00",
					"-2.00",
					"1,272",
					""
				]
			]
		  }`)
	})

	_, err := client.Quote.DownloadTpex("3374", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTpex returned nil; expected error")
	}
}

func TestQuoteService_DownloadBadData(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"stkNo": "3374",
			"stkName": "精材            ",
			"showListPriceNote": false,
			"showListPriceLink": false,
			"reportDate": "111/08",
			"iTotalRecords": 1,
			"aaData": [
			  [
				"111/08/01",
				"1,328",
				"168,265",
				"BADDATA",
				"128.00",
				"125.50",
				"127.00",
				"-2.00",
				"1,272"
			  ]
			]
		}`)
	})

	_, err := client.Quote.DownloadTpex("3374", 2022, 8)
	if err == nil {
		t.Error("Quote.DownloadTpex returned nil; expected error")
	}
}

func TestParse(t *testing.T) {
	_, err := parse([]string{})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}

	_, err = parse([]string{
		"111/08/01",
		"1,328",
		"168,265",
		"--",
		"--",
		"--",
		"--",
		"-2.00",
		"1,272",
	})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}

	_, err = parse([]string{
		"a/08/01",
		"1,328",
		"168,265",
		"127.50",
		"128.00",
		"125.50",
		"127.00",
		"-2.00",
		"1,272",
	})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}

	_, err = parse([]string{
		"111/a/01",
		"1,328",
		"168,265",
		"127.50",
		"128.00",
		"125.50",
		"127.00",
		"-2.00",
		"1,272",
	})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}

	_, err = parse([]string{
		"111/08/a",
		"1,328",
		"168,265",
		"127.50",
		"128.00",
		"125.50",
		"127.00",
		"-2.00",
		"1,272",
	})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}

	_, err = parse([]string{
		"111/08/01",
		"1,328",
		"168,265",
		"a",
		"128.00",
		"125.50",
		"127.00",
		"-2.00",
		"1,272",
	})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}

	_, err = parse([]string{
		"111/08/01",
		"1,328",
		"168,265",
		"127.50",
		"a",
		"125.50",
		"127.00",
		"-2.00",
		"1,272",
	})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}

	_, err = parse([]string{
		"111/08/01",
		"1,328",
		"168,265",
		"127.50",
		"128.00",
		"a",
		"127.00",
		"-2.00",
		"1,272",
	})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}

	_, err = parse([]string{
		"111/08/01",
		"1,328",
		"168,265",
		"127.50",
		"128.00",
		"125.50",
		"a",
		"-2.00",
		"1,272",
	})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}

	_, err = parse([]string{
		"111/08/01",
		"a",
		"168,265",
		"127.50",
		"128.00",
		"125.50",
		"127.00",
		"-2.00",
		"1,272",
	})
	if err == nil {
		t.Error("parse returned nil; expected error")
	}
}

func TestParseBidAsk(t *testing.T) {
	_, err := parseBidAsk("1_2_", "1_")
	if err == nil {
		t.Error("parseBidAsk returned nil; expected error")
	}

	_, err = parseBidAsk("a", "1")
	if err == nil {
		t.Error("parseBidAsk returned nil; expected error")
	}

	_, err = parseBidAsk("1", "a")
	if err == nil {
		t.Error("parseBidAsk returned nil; expected error")
	}
}

func TestTimestamp_UnmarshalJSON(t *testing.T) {
	var testCases = map[string]struct {
		data      []byte
		want      timestamp
		wantError bool
	}{
		"valid": {
			data:      []byte(`"1640567145000"`),
			want:      timestamp{Time: time.Date(2021, 12, 27, 1, 5, 45, 0, time.UTC)},
			wantError: false,
		},
		"not string": {
			data:      []byte(`1640567145000`),
			want:      timestamp{},
			wantError: true,
		},
		"empty string": {
			data:      []byte(`""`),
			want:      timestamp{},
			wantError: true,
		},
	}
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			date := timestamp{}
			err := date.UnmarshalJSON(test.data)
			if err != nil && !test.wantError {
				t.Errorf("Timestamp.UnmarshalJSON returned an error when we expected nil")
			}
			if err == nil && test.wantError {
				t.Errorf("Timestamp.UnmarshalJSON returned no error when we expected one")
			}
			if !cmp.Equal(test.want, date) {
				t.Errorf("Timestamp.UnmarshalJSON expected date %v, got %v", test.want, date)
			}
		})
	}
}

func TestParseRealtimeData(t *testing.T) {
	_, err := parseRealtimeData(realtimeData{Price: "BAD"})
	if err == nil {
		t.Error("parseRealtimeData returned nil; expected error")
	}

	_, err = parseRealtimeData(realtimeData{Price: "1.0", Open: "BAD"})
	if err == nil {
		t.Error("parseRealtimeData returned nil; expected error")
	}

	_, err = parseRealtimeData(realtimeData{Price: "1.0", Open: "1.0", High: "BAD"})
	if err == nil {
		t.Error("parseRealtimeData returned nil; expected error")
	}

	_, err = parseRealtimeData(realtimeData{Price: "1.0", Open: "1.0", High: "1.0", Low: "BAD"})
	if err == nil {
		t.Error("parseRealtimeData returned nil; expected error")
	}

	_, err = parseRealtimeData(realtimeData{Price: "1.0", Open: "1.0", High: "1.0", Low: "1.0", Volume: "BAD"})
	if err == nil {
		t.Error("parseRealtimeData returned nil; expected error")
	}

	_, err = parseRealtimeData(realtimeData{Price: "1.0", Open: "1.0", High: "1.0", Low: "1.0", Volume: "1", BidPrices: ""})
	if err == nil {
		t.Error("parseRealtimeData returned nil; expected error")
	}

	_, err = parseRealtimeData(realtimeData{Price: "1.0", Open: "1.0", High: "1.0", Low: "1.0", Volume: "1", BidPrices: "1", BidVolumes: "1"})
	if err == nil {
		t.Error("parseRealtimeData returned nil; expected error")
	}
}

func TestQuoteService_Realtime(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(realtimeQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"msgArray": [
			  {
				"tv": "-",
				"ps": "5350",
				"pz": "510.0000",
				"bp": "0",
				"fv": "36",
				"oa": "512.0000",
				"ob": "511.0000",
				"a": "511.0000_512.0000_513.0000_514.0000_515.0000_",
				"b": "510.0000_509.0000_508.0000_507.0000_506.0000_",
				"c": "2330",
				"d": "20220822",
				"ch": "2330.tw",
				"ot": "14:30:00",
				"tlong": "1661149800000",
				"f": "149_211_165_352_434_",
				"ip": "0",
				"g": "403_922_848_441_432_",
				"mt": "000000",
				"ov": "29811",
				"h": "514.0000",
				"i": "24",
				"it": "12",
				"oz": "512.0000",
				"l": "510.0000",
				"n": "台積電",
				"o": "511.0000",
				"p": "0",
				"ex": "tse",
				"s": "5401",
				"t": "13:30:00",
				"u": "570.0000",
				"v": "20813",
				"w": "467.5000",
				"nf": "台灣積體電路製造股份有限公司",
				"y": "519.0000",
				"z": "510.0000",
				"ts": "0"
			  },
			  {
				"tv": "-",
				"ps": "263",
				"pz": "128.5000",
				"bp": "0",
				"fv": "12",
				"oa": "128.5000",
				"ob": "127.5000",
				"a": "128.5000_129.0000_129.5000_130.0000_130.5000_",
				"b": "128.0000_127.5000_127.0000_126.5000_126.0000_",
				"c": "3374",
				"d": "20220822",
				"ch": "3374.tw",
				"ot": "14:30:00",
				"tlong": "1661149800000",
				"f": "19_7_25_37_46_",
				"ip": "0",
				"g": "152_87_69_48_76_",
				"mt": "000000",
				"ov": "626",
				"h": "133.5000",
				"i": "24",
				"it": "12",
				"oz": "128.5000",
				"l": "128.5000",
				"n": "精材",
				"o": "128.5000",
				"p": "0",
				"ex": "otc",
				"s": "263",
				"t": "13:30:00",
				"u": "142.0000",
				"v": "5086",
				"w": "117.0000",
				"nf": "精材科技股份有限公司",
				"y": "129.5000",
				"z": "128.5000",
				"ts": "0"
			  }
			],
			"referer": "",
			"userDelay": 5000,
			"rtcode": "0000",
			"queryTime": {
			  "sysDate": "20220822",
			  "stockInfoItem": 28,
			  "stockInfo": 10621,
			  "sessionStr": "UserSession",
			  "sysTime": "15:44:03",
			  "showChart": false,
			  "sessionFromTime": -1,
			  "sessionLatestTime": -1
			},
			"rtmessage": "OK"
		  }`)
	})

	quotes, err := client.Quote.Realtime([]string{"2330", "3374"})
	if err != nil {
		t.Errorf("Quote.Realtime returned error: %v", err)
	}
	want := map[string]RealtimeQuote{
		"2330": {
			At:       time.Date(2022, 8, 22, 6, 30, 0, 0, time.UTC),
			Code:     "2330",
			Name:     "台積電",
			FullName: "台灣積體電路製造股份有限公司",
			Price:    decimal.NewFromInt(510),
			Open:     decimal.NewFromInt(511),
			High:     decimal.NewFromInt(514),
			Low:      decimal.NewFromInt(510),
			Volume:   20813,
			Bids: []BidAsk{
				{decimal.NewFromInt(510), 403},
				{decimal.NewFromInt(509), 922},
				{decimal.NewFromInt(508), 848},
				{decimal.NewFromInt(507), 441},
				{decimal.NewFromInt(506), 432},
			},
			Asks: []BidAsk{
				{decimal.NewFromInt(511), 149},
				{decimal.NewFromInt(512), 211},
				{decimal.NewFromInt(513), 165},
				{decimal.NewFromInt(514), 352},
				{decimal.NewFromInt(515), 434},
			},
		},
		"3374": {
			At:       time.Date(2022, 8, 22, 6, 30, 0, 0, time.UTC),
			Code:     "3374",
			Name:     "精材",
			FullName: "精材科技股份有限公司",
			Price:    decimal.NewFromFloat(128.5),
			Open:     decimal.NewFromFloat(128.5),
			High:     decimal.NewFromFloat(133.5),
			Low:      decimal.NewFromFloat(128.5),
			Volume:   5086,
			Bids: []BidAsk{
				{decimal.NewFromFloat(128), 152},
				{decimal.NewFromFloat(127.5), 87},
				{decimal.NewFromFloat(127), 69},
				{decimal.NewFromFloat(126.5), 48},
				{decimal.NewFromFloat(126), 76},
			},
			Asks: []BidAsk{
				{decimal.NewFromFloat(128.5), 19},
				{decimal.NewFromFloat(129), 7},
				{decimal.NewFromFloat(129.5), 25},
				{decimal.NewFromFloat(130), 37},
				{decimal.NewFromFloat(130.5), 46},
			},
		},
	}
	if !cmp.Equal(quotes, want) {
		t.Errorf("Quote.Realtime returned %v, want %v", quotes, want)
	}
}

func TestQuoteService_RealtimeError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(realtimeQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusBadRequest)
	})

	_, err := client.Quote.Realtime([]string{"2330", "3374"})
	if err == nil {
		t.Error("Quote.Realtime returned nil; expected error")
	}
	testErrorContains(t, err, ": 400")

	_, err = client.Quote.Realtime([]string{"BAD"})
	if err == nil {
		t.Error("Quote.Realtime returned nil; expected error")
	}
}

func TestQuoteService_RealtimeBadStat(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(realtimeQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"msgArray": [],
			"referer": "",
			"userDelay": 5000,
			"rtcode": "0000",
			"queryTime": {
			  "sysDate": "20220822",
			  "stockInfoItem": 28,
			  "stockInfo": 10621,
			  "sessionStr": "UserSession",
			  "sysTime": "15:44:03",
			  "showChart": false,
			  "sessionFromTime": -1,
			  "sessionLatestTime": -1
			},
			"rtmessage": "BAD"
		  }`)
	})

	_, err := client.Quote.Realtime([]string{"2330", "3374"})
	if err == nil {
		t.Error("Quote.Realtime returned nil; expected error")
	}
}

func TestQuoteService_RealtimeBadContent(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(realtimeQuotesPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		{
			"msgArray": [
				{
					"tv": "-",
					"ps": "5350",
					"pz": "510.0000",
					"bp": "0",
					"fv": "36",
					"oa": "512.0000",
					"ob": "511.0000",
					"a": "511.0000_512.0000_513.0000_514.0000_515.0000_",
					"b": "510.0000_509.0000_508.0000_507.0000_506.0000_",
					"c": "2330",
					"d": "20220822",
					"ch": "2330.tw",
					"ot": "14:30:00",
					"tlong": "1661149800000",
					"f": "149_211_165_352_434_",
					"ip": "0",
					"g": "403_922_848_441_432_",
					"mt": "000000",
					"ov": "29811",
					"h": "514.0000",
					"i": "24",
					"it": "12",
					"oz": "512.0000",
					"l": "510.0000",
					"n": "台積電",
					"o": "511.0000",
					"p": "0",
					"ex": "tse",
					"s": "5401",
					"t": "13:30:00",
					"u": "570.0000",
					"v": "20813",
					"w": "467.5000",
					"nf": "台灣積體電路製造股份有限公司",
					"y": "519.0000",
					"z": "--",
					"ts": "0"
				}
			],
			"referer": "",
			"userDelay": 5000,
			"rtcode": "0000",
			"queryTime": {
			  "sysDate": "20220822",
			  "stockInfoItem": 28,
			  "stockInfo": 10621,
			  "sessionStr": "UserSession",
			  "sysTime": "15:44:03",
			  "showChart": false,
			  "sessionFromTime": -1,
			  "sessionLatestTime": -1
			},
			"rtmessage": "OK"
		  }`)
	})

	_, err := client.Quote.Realtime([]string{"2330", "3374"})
	if err == nil {
		t.Error("Quote.Realtime returned nil; expected error")
	}
}
