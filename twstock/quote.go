package twstock

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-sql/civil"
	"github.com/shopspring/decimal"
)

type QuoteService struct {
	client *Client
}

type Quote struct {
	Date   civil.Date      // 本筆資料所屬日期
	Open   decimal.Decimal // 開盤價
	High   decimal.Decimal // 最高價
	Low    decimal.Decimal // 最低價
	Close  decimal.Decimal // 收盤價
	Volume int             // 成交量
}

const (
	// 上市個股日成交資訊
	twseQuotesPath = "/rwd/zh/afterTrading/STOCK_DAY"

	// 上櫃個股日成交資訊
	tpexQuotesPath = "/www/zh-tw/afterTrading/tradingStock"

	// 個股即時交易行情
	realtimeQuotesPath = "/stock/api/getStockInfo.jsp"
)

type twseOptions struct {
	Response string `url:"response"`
	Date     string `url:"date"`
	Code     string `url:"stockNo,omitempty"`
}

type twseResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
	Notes  []string   `json:"notes"`
}

var (
	errSuspendedTrading = errors.New("parse: suspended trading")

	// 當查詢不到個股日成交資訊丟出此錯誤
	ErrNoData = errors.New("no data found")

	// 當查詢日期超出限制的時候丟出此錯誤
	ErrDateOutOffRange = errors.New("date out of range")
)

func parseDate(s string) (civil.Date, error) {
	var date civil.Date
	rawDate := strings.Split(strings.TrimSpace(s), "/")
	if len(rawDate) != 3 {
		return date, fmt.Errorf("failed parsing quote date: %s", s)
	}
	year, err := strconv.Atoi(rawDate[0])
	if err != nil {
		return date, fmt.Errorf("failed parsing quote date: %w", err)
	}
	month, err := strconv.Atoi(rawDate[1])
	if err != nil {
		return date, fmt.Errorf("failed parsing quote date: %w", err)
	}
	// 櫃買中心的日期在 IPO 那天結果會有＊
	day, err := strconv.Atoi(strings.TrimRight(rawDate[2], "＊*"))
	if err != nil {
		return date, fmt.Errorf("failed parsing quote date: %w", err)
	}
	// 需要將民國年轉成西元年
	date = civil.Date{Year: year + 1911, Month: time.Month(month), Day: day}
	if !date.IsValid() {
		return date, fmt.Errorf("failed parsing quote date: %s", s)
	}
	return date, nil
}

func parsePrice(s string) (decimal.Decimal, error) {
	var v decimal.Decimal
	f, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
	if err != nil {
		return v, err
	}
	return decimal.NewFromFloat(f), nil
}

func parseVolume(s string) (int, error) {
	v, err := strconv.Atoi(strings.ReplaceAll(s, ",", ""))
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (*QuoteService) parse(data []string) (Quote, error) {
	var quote Quote
	if len(data) < 7 {
		return quote, fmt.Errorf("failed parsing quote data")
	}
	// 暫停交易
	if data[3] == "--" ||
		data[4] == "--" ||
		data[5] == "--" ||
		data[6] == "--" {
		return quote, errSuspendedTrading
	}
	date, err := parseDate(data[0])
	if err != nil {
		return quote, err
	}
	open, err := parsePrice(data[3])
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote open: %w", err)
	}
	high, err := parsePrice(data[4])
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote high: %w", err)
	}
	low, err := parsePrice(data[5])
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote low: %w", err)
	}
	close, err := parsePrice(data[6])
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote close: %w", err)
	}
	volume, err := parseVolume(data[1])
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote volume: %w", err)
	}
	quote.Date = date
	quote.Open = open
	quote.High = high
	quote.Low = low
	quote.Close = close
	quote.Volume = volume
	return quote, nil
}

// 台灣證卷交易所或是證券櫃檯買賣中心有最小查詢日期的限制
func (s *QuoteService) MinimumDate(m Market) civil.Date {
	if m == TWSE {
		// 台灣證卷交易所個股日成交資訊最早到民國99年1月
		return civil.Date{Year: 2010, Month: time.January, Day: 1}
	}
	// 證券櫃檯買賣中心個股日成交資訊最早到民國83年1月
	return civil.Date{Year: 1994, Month: time.January, Day: 1}
}

// 從台灣證卷交易所下載盤後個股日成交資訊
func (s *QuoteService) DownloadTwse(code string, year int, month time.Month) ([]Quote, error) {
	date := civil.Date{Year: year, Month: month, Day: 1}
	if date.Before(s.MinimumDate(TWSE)) {
		return nil, fmt.Errorf("invalid date: %s", fmt.Sprintf("%04d-%02d", date.Year, date.Month))
	}
	url, _ := s.client.twseBaseURL.Parse(twseQuotesPath)
	opts := twseOptions{
		Response: "json",
		Date:     fmt.Sprintf("%04d%02d%02d", date.Year, date.Month, date.Day),
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
		switch resp.Stat {
		case "很抱歉，沒有符合條件的資料!":
			return nil, ErrNoData
		case "查詢日期大於今日，請重新查詢!":
			return nil, ErrDateOutOffRange
		default:
			return nil, fmt.Errorf("invalid state: %s", resp.Stat)
		}
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
		quote, err := s.parse(data)
		if err != nil {
			if errors.Is(err, errSuspendedTrading) {
				continue
			}
			return nil, err
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

type tpexOptions struct {
	Response string `url:"response"`
	Date     string `url:"date"`
	Code     string `url:"code,omitempty"`
}

type StringOrNumber string

func (s *StringOrNumber) UnmarshalJSON(data []byte) error {
	// Try unmarshaling as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = StringOrNumber(str)
		return nil
	}

	// If that fails, try as number
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*s = StringOrNumber(strconv.FormatFloat(num, 'f', -1, 64))
		return nil
	}

	return fmt.Errorf("value must be string or number")
}

type tpexResponse struct {
	Stat   string `json:"stat"`
	Date   string `json:"date"`
	Code   string `json:"code"`
	Tables []struct {
		Title      string             `json:"title"`
		Date       string             `json:"date"`
		Data       [][]StringOrNumber `json:"data"`
		Fields     []string           `json:"fields"`
		Notes      []string           `json:"notes"`
		TotalCount int                `json:"totalCount"`
	} `json:"tables"`
}

// 從證券櫃檯買賣中心下載盤後個股日成交資訊
func (s *QuoteService) DownloadTpex(code string, year int, month time.Month) ([]Quote, error) {
	date := civil.Date{Year: year, Month: month, Day: 1}
	if date.Before(s.MinimumDate(TPEx)) {
		return nil, fmt.Errorf("invalid date: %s", fmt.Sprintf("%04d-%02d", date.Year, date.Month))
	}
	url, _ := s.client.tpexBaseURL.Parse(tpexQuotesPath)
	opts := tpexOptions{
		Response: "json",
		Date:     fmt.Sprintf("%04d/%02d/%02d", date.Year, date.Month, date.Day),
		Code:     code,
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
	if len(resp.Tables) != 1 || resp.Tables[0].TotalCount == 0 {
		return nil, ErrNoData
	}
	if resp.Tables[0].TotalCount != len(resp.Tables[0].Data) {
		return nil, fmt.Errorf("failed parsing quote data length returned %d, want %d", resp.Tables[0].TotalCount, len(resp.Tables[0].Data))
	}
	quotes := []Quote{}
	for _, data := range resp.Tables[0].Data {
		if len(data) != 9 {
			return nil, fmt.Errorf("failed parsing quote fields")
		}
		stringData := make([]string, len(data))
		for i, v := range data {
			stringData[i] = string(v)
		}
		quote, err := s.parse(stringData)
		if err != nil {
			if errors.Is(err, errSuspendedTrading) {
				continue
			}
			return nil, err
		}
		// 成交仟股
		quote.Volume *= 1000
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

// 從台灣證卷交易所或證券櫃檯買賣中心下載盤後個股日成交資訊
func (s *QuoteService) Download(code string, year int, month time.Month) ([]Quote, error) {
	//nolint:typecheck
	if security, ok := Securities[code]; ok {
		switch security.Market {
		case TWSE:
			return s.DownloadTwse(code, year, month)
		case TPEx:
			return s.DownloadTpex(code, year, month)
		default:
			return nil, fmt.Errorf("invalid market: %s", security.Market)
		}
	}
	return nil, fmt.Errorf("invalid code: %s", code)
}

type BidAsk struct {
	Price  decimal.Decimal // 價格
	Volume int             // 數量
}

type RealtimeQuote struct {
	At       time.Time       // 最新一筆成交時間
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
		price, err := parsePrice(prices[i])
		if err != nil {
			return nil, fmt.Errorf("failed parsing quote price: %w", err)
		}
		volume, err := parseVolume(volumes[i])
		if err != nil {
			return nil, fmt.Errorf("failed parsing quote volume: %w", err)
		}
		v = append(v, BidAsk{price, volume})
	}
	return v, nil
}

func parseRealtimeData(data realtimeData) (RealtimeQuote, error) {
	var quote RealtimeQuote
	price, err := parsePrice(data.Price)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote price: %w", err)
	}
	open, err := parsePrice(data.Open)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote open: %w", err)
	}
	high, err := parsePrice(data.High)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote high: %w", err)
	}
	low, err := parsePrice(data.Low)
	if err != nil {
		return quote, fmt.Errorf("failed parsing quote low: %w", err)
	}
	volume, err := parseVolume(data.Volume)
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

	quote.At = data.Timestamp.Time
	quote.Code = data.Code
	quote.Name = data.Name
	quote.FullName = data.FullName
	quote.Price = price
	quote.Open = open
	quote.High = high
	quote.Low = low
	quote.Volume = volume
	quote.Bids = bids
	quote.Asks = asks

	return quote, nil
}

// 從台灣證卷交易所下載即時個股成交資訊
func (s *QuoteService) Realtime(codes ...string) (map[string]RealtimeQuote, error) {
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
