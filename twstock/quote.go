package twstock

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type QuoteService struct {
	client *Client
}

type Quote struct {
	Date   time.Time       // 本筆資料所屬日期
	Open   decimal.Decimal // 開盤價
	High   decimal.Decimal // 最高價
	Low    decimal.Decimal // 最低價
	Close  decimal.Decimal // 收盤價
	Volume int             // 成交量
}

// 台灣證卷交易所或是證券櫃檯買賣中心有最小查詢日期的限制
func (m Market) MinimumDate() time.Time {
	if m == TWSE {
		// 台灣證卷交易所個股日成交資訊最早到民國99年1月
		return time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	// 證券櫃檯買賣中心個股日成交資訊最早到民國83年1月
	return time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC)
}

const (
	// 上市個股日成交資訊
	twseQuotesPath = "/exchangeReport/STOCK_DAY"

	// 上櫃個股日成交資訊
	tpexQuotesPath = "/web/stock/aftertrading/daily_trading_info/st43_result.php"

	// 個股即時交易行情
	realtimeQuotesPath = "/stock/api/getStockInfo.jsp"
)

type twseOptions struct {
	Response string `url:"response"`
	Date     string `url:"date"`
	Code     string `url:"stockNo"`
}

type twseResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
	Notes  []string   `json:"notes"`
}

var ErrSuspendedTrading = errors.New("parse: suspended trading")

func parse(data []string) (Quote, error) {
	var quote Quote
	if len(data) < 7 {
		return quote, fmt.Errorf("failed parsing quote data")
	}
	// 暫停交易
	if data[3] == "--" ||
		data[4] == "--" ||
		data[5] == "--" ||
		data[6] == "--" {
		return quote, ErrSuspendedTrading
	}
	rawDate := strings.Split(strings.TrimSpace(data[0]), "/")
	if len(rawDate) != 3 {
		return quote, fmt.Errorf("failed parsing quote date: %s", data[0])
	}
	year, err := strconv.Atoi(rawDate[0])
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote date: %w", err)
	}
	month, err := strconv.Atoi(rawDate[1])
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote date: %w", err)
	}
	// 櫃買中心的日期在 IPO 那天結果會有＊
	day, err := strconv.Atoi(strings.TrimRight(rawDate[2], "＊"))
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote date: %w", err)
	}
	// 需要將民國年轉成西元年
	date := time.Date(year+1911, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	open, err := strconv.ParseFloat(strings.ReplaceAll(data[3], ",", ""), 64)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote open: %w", err)
	}
	high, err := strconv.ParseFloat(strings.ReplaceAll(data[4], ",", ""), 64)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote high: %w", err)
	}
	low, err := strconv.ParseFloat(strings.ReplaceAll(data[5], ",", ""), 64)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote low: %w", err)
	}
	close, err := strconv.ParseFloat(strings.ReplaceAll(data[6], ",", ""), 64)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote close: %w", err)
	}
	volume, err := strconv.Atoi(strings.ReplaceAll(data[1], ",", ""))
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote volume: %w", err)
	}
	quote.Date = date
	quote.Open = decimal.NewFromFloat(open)
	quote.High = decimal.NewFromFloat(high)
	quote.Low = decimal.NewFromFloat(low)
	quote.Close = decimal.NewFromFloat(close)
	quote.Volume = volume
	return quote, nil
}

// 從台灣證卷交易所下載盤後個股日成交資訊
func (s *QuoteService) DownloadTwse(code string, year int, month time.Month) ([]Quote, error) {
	//nolint:typecheck
	if security, ok := Securities[code]; !ok || security.Market != TWSE {
		return nil, fmt.Errorf("invalid twse code: %s", code)
	}
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	if date.Before(TWSE.MinimumDate()) {
		return nil, fmt.Errorf("invalid date: %s", date.Format("2006-01"))
	}
	url, _ := s.client.twseBaseURL.Parse(twseQuotesPath)
	opts := twseOptions{
		Response: "json",
		Date:     date.Format("20060102"),
		Code:     code,
	}
	url, _ = addOptions(url, opts)
	req, _ := s.client.NewRequest("GET", url.String(), nil)
	resp := &twseResponse{}
	_, err := s.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Stat != "OK" {
		return nil, fmt.Errorf("invalid state: %s", resp.Stat)
	}
	if len(resp.Fields) != 9 ||
		resp.Fields[0] != "日期" ||
		resp.Fields[1] != "成交股數" ||
		resp.Fields[2] != "成交金額" ||
		resp.Fields[3] != "開盤價" ||
		resp.Fields[4] != "最高價" ||
		resp.Fields[5] != "最低價" ||
		resp.Fields[6] != "收盤價" ||
		resp.Fields[7] != "漲跌價差" ||
		resp.Fields[8] != "成交筆數" {
		return nil, fmt.Errorf("failed parsing quote fields: %s", strings.Join(resp.Fields, ","))
	}
	quotes := []Quote{}
	for _, data := range resp.Data {
		quote, err := parse(data)
		if err != nil {
			if errors.Is(err, ErrSuspendedTrading) {
				continue
			}
			return nil, err
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

type tpexOptions struct {
	Date string `url:"d"`
	Code string `url:"stkno"`
}

type tpexResponse struct {
	Code       string     `json:"stkNo"`
	Name       string     `json:"stkName"`
	Date       string     `json:"reportDate"`
	DataLength int        `json:"iTotalRecords"`
	Data       [][]string `json:"aaData"`
}

// 從證券櫃檯買賣中心下載盤後個股日成交資訊
func (s *QuoteService) DownloadTpex(code string, year int, month time.Month) ([]Quote, error) {
	//nolint:typecheck
	if security, ok := Securities[code]; !ok || security.Market != TPEx {
		return nil, fmt.Errorf("invalid tpex code: %s", code)
	}
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	if date.Before(TPEx.MinimumDate()) {
		return nil, fmt.Errorf("invalid date: %s", date.Format("2006-01"))
	}
	url, _ := s.client.tpexBaseURL.Parse(tpexQuotesPath)
	opts := tpexOptions{
		// 需要將西元年轉為民國年
		Date: fmt.Sprintf("%d/%s", date.Year()-1911, date.Format("01")),
		Code: code,
	}
	url, _ = addOptions(url, opts)
	req, _ := s.client.NewRequest("GET", url.String(), nil)
	resp := &tpexResponse{}
	_, err := s.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != code {
		return nil, fmt.Errorf("invalid tpex code returned %s, want %s", resp.Code, code)
	}
	if resp.DataLength == 0 || resp.DataLength != len(resp.Data) {
		return nil, fmt.Errorf("failed parsing quote data length returned %d, want %d", resp.DataLength, len(resp.Data))
	}
	quotes := []Quote{}
	for _, data := range resp.Data {
		if len(data) != 9 {
			return nil, fmt.Errorf("failed parsing quote fields")
		}
		quote, err := parse(data)
		if err != nil {
			if errors.Is(err, ErrSuspendedTrading) {
				continue
			}
			return nil, err
		}
		// 成交仟股
		quote.Volume = quote.Volume * 1000
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

// 從台灣證卷交易所或證券櫃檯買賣中心下載盤後個股日成交資訊
func (s *QuoteService) Download(code string, year int, month time.Month) ([]Quote, error) {
	//nolint:typecheck
	if security, ok := Securities[code]; ok {
		if security.Market == TWSE {
			return s.DownloadTwse(code, year, month)
		} else if security.Market == TPEx {
			return s.DownloadTpex(code, year, month)
		}
	}
	return nil, fmt.Errorf("invalid code: %s", code)
}

type BidAsk struct {
	Price  decimal.Decimal // 價格
	Volume int             // 數量
}

type RealtimeQuote struct {
	Date     time.Time       // 最新一筆成交時間
	Code     string          // 股票代號
	Name     string          // 簡稱
	FullName string          // 全名
	Price    decimal.Decimal // 最新一筆成交價
	Open     decimal.Decimal // 開盤價
	High     decimal.Decimal // 最高價
	Low      decimal.Decimal // 最低價
	Volume   int             // 總成交量
	Bids     []BidAsk        // 最佳五檔委買資料
	Asks     []BidAsk        // 最佳五檔委賣資料
}

type realtimeOptions struct {
	Codes string `url:"ex_ch"`
}

type timestamp struct {
	time.Time
}

// UnmarshalJSON handles incoming JSON.
func (p *timestamp) UnmarshalJSON(bytes []byte) error {
	var s string
	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	// realtime returns Unix timestamp in milliseconds
	p.Time = time.Unix(i/1000, (i % 1000 * 1000000)).In(time.UTC)
	return nil
}

type realtimeData struct {
	Timestamp  timestamp `json:"tlong"`
	Code       string    `json:"c"`
	Price      string    `json:"z"`
	BidPrices  string    `json:"b"`
	BidVolumes string    `json:"g"`
	AskPrices  string    `json:"a"`
	AskVolumes string    `json:"f"`
	Open       string    `json:"o"`
	High       string    `json:"h"`
	Low        string    `json:"l"`
	Volume     string    `json:"v"`
	Name       string    `json:"n"`
	FullName   string    `json:"nf"`
}

type realtimeResponse struct {
	Stat string         `json:"rtmessage"`
	Data []realtimeData `json:"msgArray"`
}

func parseBidAsk(pricesStr string, volumesStr string) ([]BidAsk, error) {
	split := func(v string) []string { return strings.Split(strings.Trim(v, "_"), "_") }

	prices := split(pricesStr)
	volumes := split(volumesStr)
	if len(prices) != len(volumes) {
		return nil, fmt.Errorf("failed parsing bid-ask")
	}

	v := []BidAsk{}
	for i := 0; i < len(prices); i++ {
		price, err := strconv.ParseFloat(prices[i], 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing quote price: %w", err)
		}
		volume, err := strconv.Atoi(volumes[i])
		if err != nil {
			return nil, fmt.Errorf("failed parsing quote volume: %w", err)
		}
		v = append(v, BidAsk{decimal.NewFromFloat(price), volume})
	}
	return v, nil
}

func parseRealtimeData(data realtimeData) (RealtimeQuote, error) {
	var quote RealtimeQuote
	price, err := strconv.ParseFloat(data.Price, 64)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote price: %w", err)
	}
	open, err := strconv.ParseFloat(data.Open, 64)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote open: %w", err)
	}
	high, err := strconv.ParseFloat(data.High, 64)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote high: %w", err)
	}
	low, err := strconv.ParseFloat(data.Low, 64)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote low: %w", err)
	}
	volume, err := strconv.Atoi(data.Volume)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote volume: %w", err)
	}
	bids, err := parseBidAsk(data.BidPrices, data.BidVolumes)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote bids: %w", err)
	}
	asks, err := parseBidAsk(data.AskPrices, data.AskVolumes)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote asks: %w", err)
	}

	quote.Date = data.Timestamp.Time
	quote.Code = data.Code
	quote.Name = data.Name
	quote.FullName = data.FullName
	quote.Price = decimal.NewFromFloat(price)
	quote.Open = decimal.NewFromFloat(open)
	quote.High = decimal.NewFromFloat(high)
	quote.Low = decimal.NewFromFloat(low)
	quote.Volume = volume
	quote.Bids = bids
	quote.Asks = asks

	return quote, nil
}

func (s *QuoteService) Realtime(codes []string) (map[string]RealtimeQuote, error) {
	for i, v := range codes {
		//nolint:typecheck
		if security, ok := Securities[v]; ok {
			if security.Market == TWSE {
				codes[i] = fmt.Sprintf("%s_%s.tw", TWSE, v)
				continue
			} else if security.Market == TPEx {
				codes[i] = fmt.Sprintf("%s_%s.tw", TPEx, v)
				continue
			}
		}
		return nil, fmt.Errorf("invalid code: %s", v)
	}

	url, _ := s.client.misTwseBaseURL.Parse(realtimeQuotesPath)
	opts := realtimeOptions{
		Codes: strings.Join(codes, "|"),
	}
	url, _ = addOptions(url, opts)
	req, _ := s.client.NewRequest("GET", url.String(), nil)
	resp := &realtimeResponse{}
	_, err := s.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Stat != "OK" {
		return nil, fmt.Errorf("invalid state: %s", resp.Stat)
	}

	quotes := map[string]RealtimeQuote{}
	for _, data := range resp.Data {
		quote, err := parseRealtimeData(data)
		if err != nil {
			return nil, err
		}
		quotes[data.Code] = quote
	}

	return quotes, nil
}
