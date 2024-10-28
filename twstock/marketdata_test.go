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

func TestMarketDataService_DownloadTwse(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"stat": "OK",
			"date": "20220823",
			"title": "111年08月市場成交資訊",
			"fields": [
				"日期",
				"成交股數",
				"成交金額",
				"成交筆數",
				"發行量加權股價指數",
				"漲跌點數"
			],
			"data": [
				[
					"111/08/01",
					"5,028,214,637",
					"181,641,243,076",
					"1,645,643",
					"14,981.69",
					"-18.38"
				],
				[
					"111/08/02",
					"6,048,011,360",
					"219,100,619,458",
					"2,009,677",
					"14,747.23",
					"-234.46"
				],
				[
					"111/08/03",
					"4,784,239,317",
					"177,701,242,578",
					"1,593,708",
					"14,777.02",
					"29.79"
				],
				[
					"111/08/04",
					"5,235,083,128",
					"197,658,576,530",
					"1,723,077",
					"14,702.20",
					"-74.82"
				],
				[
					"111/08/05",
					"5,642,144,158",
					"206,698,922,021",
					"1,664,515",
					"15,036.04",
					"333.84"
				]
			],
			"notes": [
				"當日統計資訊含一般、零股、盤後定價，不含鉅額、拍賣、標購。",
				"不加計外幣交易證券交易金額。"
			]
		}`)
	})

	data, err := client.MarketData.DownloadTwse(2022, 8)
	if err != nil {
		t.Errorf("MarketData.DownloadTwse returned error: %v", err)
	}
	want := []MarketData{
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 1},
			TradeVolume: 5028214637,
			TradeValue:  decimal.NewFromInt(181641243076),
			Transaction: 1645643,
			Index:       decimal.NewFromFloat(14981.69),
			Change:      decimal.NewFromFloat(-18.38),
		},
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 2},
			TradeVolume: 6048011360,
			TradeValue:  decimal.NewFromInt(219100619458),
			Transaction: 2009677,
			Index:       decimal.NewFromFloat(14747.23),
			Change:      decimal.NewFromFloat(-234.46),
		},
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 3},
			TradeVolume: 4784239317,
			TradeValue:  decimal.NewFromInt(177701242578),
			Transaction: 1593708,
			Index:       decimal.NewFromFloat(14777.02),
			Change:      decimal.NewFromFloat(29.79),
		},
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 4},
			TradeVolume: 5235083128,
			TradeValue:  decimal.NewFromInt(197658576530),
			Transaction: 1723077,
			Index:       decimal.NewFromFloat(14702.20),
			Change:      decimal.NewFromFloat(-74.82),
		},
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 5},
			TradeVolume: 5642144158,
			TradeValue:  decimal.NewFromInt(206698922021),
			Transaction: 1664515,
			Index:       decimal.NewFromFloat(15036.04),
			Change:      decimal.NewFromFloat(333.84),
		},
	}
	if !cmp.Equal(data, want) {
		t.Errorf("MarketData.DownloadTwse returned %v, want %v", data, want)
	}
}

func TestMarketDataService_DownloadTwseError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusBadRequest)
	})

	_, err := client.MarketData.DownloadTwse(2022, 8)
	if err == nil {
		t.Error("MarketData.DownloadTwse returned nil; expected error")
	}
	testErrorContains(t, err, ": 400")

	_, err = client.MarketData.DownloadTwse(1989, 12)
	if err == nil {
		t.Error("MarketData.DownloadTwse returned nil; expected error")
	}
}

func TestMarketDataService_DownloadTwseBadStat(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"stat": "BAD",
			"date": "20220823",
			"title": "111年08月市場成交資訊",
			"fields": [
				"日期",
				"成交股數",
				"成交金額",
				"成交筆數",
				"發行量加權股價指數",
				"漲跌點數"
			],
			"data": [],
			"notes": [
				"當日統計資訊含一般、零股、盤後定價，不含鉅額、拍賣、標購。",
				"不加計外幣交易證券交易金額。"
			]
		}`)
	})

	_, err := client.MarketData.DownloadTwse(2022, 8)
	if err == nil {
		t.Error("MarketData.DownloadTwse returned nil; expected error")
	}
}

func TestMarketDataService_DownloadTwseErrDateOutOffRange(t *testing.T) {
	var testCases = []string{
		`{
		"stat": "查詢日期大於今日，請重新查詢!",
		"date": "20220823",
		"title": "111年08月市場成交資訊",
		"fields": [
			"日期",
			"成交股數",
			"成交金額",
			"成交筆數",
			"發行量加權股價指數",
			"漲跌點數"
		],
		"data": [],
		"notes": [
			"當日統計資訊含一般、零股、盤後定價，不含鉅額、拍賣、標購。",
			"不加計外幣交易證券交易金額。"
		]}`}

	for _, test := range testCases {
		t.Run("ErrDateOutOffRange", func(t *testing.T) {
			client, mux, teardown := setup()
			defer teardown()

			mux.HandleFunc(twseMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, "GET")
				fmt.Fprint(w, test)
			})

			_, err := client.MarketData.DownloadTwse(2022, 8)
			if err == nil {
				t.Error("MarketData.DownloadTwse returned nil; expected error")
			}
			if !errors.Is(err, ErrDateOutOffRange) {
				t.Errorf("MarketData.DownloadTwse returned %v, want %v", err, ErrDateOutOffRange)
			}
		})
	}
}

func TestMarketDataService_DownloadTwseBadFields(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"stat": "OK",
			"date": "20220823",
			"title": "111年08月市場成交資訊",
			"fields": [
				"日期",
				"成交股數",
				"成交金額",
				"成交筆===數",
				"發行量加權股價指數",
				"漲跌點數"
			],
			"data": [],
			"notes": [
				"當日統計資訊含一般、零股、盤後定價，不含鉅額、拍賣、標購。",
				"不加計外幣交易證券交易金額。"
			]
		}`)
	})

	_, err := client.MarketData.DownloadTwse(2022, 8)
	if err == nil {
		t.Error("MarketData.DownloadTwse returned nil; expected error")
	}
}

func TestMarketDataService_DownloadTwseBadContent(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(twseMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"stat": "OK",
			"date": "20220823",
			"title": "111年08月市場成交資訊",
			"fields": [
				"日期",
				"成交股數",
				"成交金額",
				"成交筆數",
				"發行量加權股價指數",
				"漲跌點數"
			],
			"data": [
				[
					"111/08/80",
					"5,028,214,637",
					"181,641,243,076",
					"1,645,643",
					"14,981.69",
					"-18.38"
				]
			],
			"notes": [
				"當日統計資訊含一般、零股、盤後定價，不含鉅額、拍賣、標購。",
				"不加計外幣交易證券交易金額。"
			]
		}`)
	})

	_, err := client.MarketData.DownloadTwse(2022, 8)
	if err == nil {
		t.Error("MarketData.DownloadTwse returned nil; expected error")
	}
}

func TestMarketDataService_DownloadTpex(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"tables": [
				{
					"title": "日成交量值指數",
					"date": "20220801",
					"data": [
						[
							"111/08/01",
							"630,223",
							"46,240,795",
							"436,953",
							182.75,
							-0.83
						],
						[
							"111/08/02",
							"694,615",
							"51,249,693",
							"484,905",
							179.30,
							-3.45
						],
						[
							"111/08/03",
							"683,637",
							"50,799,048",
							"473,344",
							178.17,
							-1.13
						],
						[
							"111/08/04",
							"677,880",
							"51,578,056",
							"458,468",
							178.18,
							0.01
						],
						[
							"111/08/05",
							"651,963",
							"57,144,930",
							"435,858",
							182.37,
							4.19
						]
					],
					"fields": [
						"日期",
						"成交股數（仟股）",
						"金額（仟元）",
						"筆數",
						"櫃買指數",
						"漲/跌"
					],
					"notes": [
						"上表為於等價、零股、盤後定價等交易系統交易之上櫃股票成交資訊。",
						"每日下午6:00另行產製於等價、零股、盤後定價、鉅額等交易系統交易之上櫃股票、權證、TDR、ETF、ETN、受益證券等上櫃有價證券之成交資訊，但不含轉(交)換公司債之成交統計報表，如<a href=\"daily-indices-rpk.html\">連結</a>"
					],
					"totalCount": 5,
					"summary": []
				}
			],
			"date": "20220801",
			"stat": "ok"
		}`)
	})

	data, err := client.MarketData.DownloadTpex(2022, 8)
	if err != nil {
		t.Errorf("MarketData.DownloadTpex returned error: %v", err)
	}
	want := []MarketData{
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 1},
			TradeVolume: 630223000,
			TradeValue:  decimal.NewFromInt(46240795000),
			Transaction: 436953,
			Index:       decimal.NewFromFloat(182.75),
			Change:      decimal.NewFromFloat(-0.83),
		},
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 2},
			TradeVolume: 694615000,
			TradeValue:  decimal.NewFromInt(51249693000),
			Transaction: 484905,
			Index:       decimal.NewFromFloat(179.30),
			Change:      decimal.NewFromFloat(-3.45),
		},
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 3},
			TradeVolume: 683637000,
			TradeValue:  decimal.NewFromInt(50799048000),
			Transaction: 473344,
			Index:       decimal.NewFromFloat(178.17),
			Change:      decimal.NewFromFloat(-1.13),
		},
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 4},
			TradeVolume: 677880000,
			TradeValue:  decimal.NewFromInt(51578056000),
			Transaction: 458468,
			Index:       decimal.NewFromFloat(178.18),
			Change:      decimal.NewFromFloat(0.01),
		},
		{
			Date:        civil.Date{Year: 2022, Month: time.August, Day: 5},
			TradeVolume: 651963000,
			TradeValue:  decimal.NewFromInt(57144930000),
			Transaction: 435858,
			Index:       decimal.NewFromFloat(182.37),
			Change:      decimal.NewFromFloat(4.19),
		},
	}
	if !cmp.Equal(data, want) {
		t.Errorf("MarketData.DownloadTpex returned %v, want %v", data, want)
	}
}

func TestMarketDataService_DownloadTpexError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusBadRequest)
	})

	_, err := client.MarketData.DownloadTpex(2022, 8)
	if err == nil {
		t.Error("MarketData.DownloadTpex returned nil; expected error")
	}
	testErrorContains(t, err, ": 400")

	_, err = client.MarketData.DownloadTpex(1998, 12)
	if err == nil {
		t.Error("MarketData.DownloadTpex returned nil; expected error")
	}
}

func TestMarketDataService_DownloadTpexErrNoData(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"date": "20220801",
			"stat": "ok",
			"tables": []
		}`)
	})

	_, err := client.MarketData.DownloadTpex(2022, 8)
	if err == nil {
		t.Error("MarketData.DownloadTpex returned nil; expected error")
	}
	if !errors.Is(err, ErrNoData) {
		t.Errorf("MarketData.DownloadTpex returned %v, want %v", err, ErrNoData)
	}
}

func TestMarketDataService_DownloadTpexErrNoData2(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"date": "20220801",
			"stat": "ok",
			"tables": [
				{
					"data": [],
					"totalCount": 0
				}
			]
		}`)
	})

	_, err := client.MarketData.DownloadTpex(2022, 8)
	if err == nil {
		t.Error("MarketData.DownloadTpex returned nil; expected error")
	}
	if !errors.Is(err, ErrNoData) {
		t.Errorf("MarketData.DownloadTpex returned %v, want %v", err, ErrNoData)
	}
}

func TestMarketDataService_DownloadTpexBadDataLength(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"tables": [
				{
					"data": [],
					"totalCount": 1
				}
			],
			"date": "20220801",
			"stat": "ok"
		}`)
	})

	_, err := client.MarketData.DownloadTpex(2022, 8)
	if err == nil {
		t.Error("MarketData.DownloadTpex returned nil; expected error")
	}
}

func TestMarketDataService_DownloadTpexBadData(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc(tpexMarketDataPath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"tables": [
				{
					"title": "日成交量值指數",
					"date": "20220801",
					"data": [
						[
							"111/08/01",
							"BADDATA",
							"46,240,795",
							"436,953",
							182.75,
							-0.83
						]
					],
					"fields": [
						"日期",
						"成交股數（仟股）",
						"金額（仟元）",
						"筆數",
						"櫃買指數",
						"漲/跌"
					],
					"notes": [
						"上表為於等價、零股、盤後定價等交易系統交易之上櫃股票成交資訊。",
						"每日下午6:00另行產製於等價、零股、盤後定價、鉅額等交易系統交易之上櫃股票、權證、TDR、ETF、ETN、受益證券等上櫃有價證券之成交資訊，但不含轉(交)換公司債之成交統計報表，如<a href=\"daily-indices-rpk.html\">連結</a>"
					],
					"totalCount": 1,
					"summary": []
				}
			],
			"date": "20220801",
			"stat": "ok"
		}`)
	})

	_, err := client.MarketData.DownloadTpex(2022, 8)
	if err == nil {
		t.Error("MarketData.DownloadTpex returned nil; expected error")
	}
}

func TestMarketDataService_parse(t *testing.T) {
	client, _, teardown := setup()
	defer teardown()

	var testCases = []([]string){
		[]string{},
		[]string{"2022/50/01", "", "", "", "", ""},
		[]string{"2022/08/01", "1B", "", "", "", ""},
		[]string{"2022/08/01", "1", "1B", "", "", ""},
		[]string{"2022/08/01", "1", "1", "1B", "", ""},
		[]string{"2022/08/01", "1", "1", "1", "1B", ""},
		[]string{"2022/08/01", "1", "1", "1", "1", "1B"},
	}

	for _, test := range testCases {
		t.Run("parse", func(t *testing.T) {
			_, err := client.MarketData.parse(test)
			if err == nil {
				t.Error("client.MarketData.parse returned nil; expected error")
			}
		})
	}
}
