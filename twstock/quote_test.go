package twstock

import (
	"fmt"
	"net/http"
	"testing"
	"time"

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
			Date:   time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC),
			Open:   decimal.NewFromFloat(506),
			High:   decimal.NewFromFloat(508),
			Low:    decimal.NewFromFloat(500),
			Close:  decimal.NewFromFloat(504),
			Volume: 24991291,
		},
		{
			Date:   time.Date(2022, 8, 2, 0, 0, 0, 0, time.UTC),
			Open:   decimal.NewFromFloat(494),
			High:   decimal.NewFromFloat(496),
			Low:    decimal.NewFromFloat(488.50),
			Close:  decimal.NewFromFloat(492),
			Volume: 42669591,
		},
		{
			Date:   time.Date(2022, 8, 4, 0, 0, 0, 0, time.UTC),
			Open:   decimal.NewFromFloat(499),
			High:   decimal.NewFromFloat(503),
			Low:    decimal.NewFromFloat(495),
			Close:  decimal.NewFromFloat(500),
			Volume: 26589086,
		},
		{
			Date:   time.Date(2022, 8, 5, 0, 0, 0, 0, time.UTC),
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
			Date:   time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC),
			Open:   decimal.NewFromFloat(127.5),
			High:   decimal.NewFromFloat(128),
			Low:    decimal.NewFromFloat(125.5),
			Close:  decimal.NewFromFloat(127),
			Volume: 1328000,
		},
		{
			Date:   time.Date(2022, 8, 2, 0, 0, 0, 0, time.UTC),
			Open:   decimal.NewFromFloat(125),
			High:   decimal.NewFromFloat(127),
			Low:    decimal.NewFromFloat(123),
			Close:  decimal.NewFromFloat(127),
			Volume: 1593000,
		},
		{
			Date:   time.Date(2022, 8, 3, 0, 0, 0, 0, time.UTC),
			Open:   decimal.NewFromFloat(124.5),
			High:   decimal.NewFromFloat(127),
			Low:    decimal.NewFromFloat(124),
			Close:  decimal.NewFromFloat(126),
			Volume: 1603000,
		},
		{
			Date:   time.Date(2022, 8, 5, 0, 0, 0, 0, time.UTC),
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
