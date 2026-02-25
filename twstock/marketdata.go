package twstock

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/golang-sql/civil"
	"github.com/shopspring/decimal"
)

type MarketDataService struct {
	client *Client
}

const (
	// 上市每日市場成交資訊
	twseMarketDataPath = "/exchangeReport/FMTQIK"

	// 上櫃每日市場成交資訊
	tpexMarketDataPath = "/www/zh-tw/afterTrading/tradingIndex"

	// 發行量加權股價指數歷史資料
	twseTAIEXPath = "/rwd/zh/TAIEX/MI_5MINS_HIST"

	// 櫃買指數歷史資料
	tpexIndexPath = "/www/zh-tw/indexInfo/inx"
)

type MarketData struct {
	Date        civil.Date      // 日期
	TradeVolume int             // 總成交股數
	TradeValue  decimal.Decimal // 總成交金額
	Transaction int             // 總成交筆數
	Index       decimal.Decimal // 發行量加權股價指數或櫃買指數
	Change      decimal.Decimal // 漲跌點數
}

type TAIEXIndex struct {
	Date  civil.Date      // 日期
	Open  decimal.Decimal // 開盤指數
	High  decimal.Decimal // 最高指數
	Low   decimal.Decimal // 最低指數
	Close decimal.Decimal // 收盤指數
}

type TPExIndex struct {
	Date   civil.Date      // 日期
	Open   decimal.Decimal // 開市
	High   decimal.Decimal // 最高
	Low    decimal.Decimal // 最低
	Close  decimal.Decimal // 收市
	Change decimal.Decimal // 漲跌
}

func (*MarketDataService) parse(data []string) (MarketData, error) {
	var marketData MarketData
	if len(data) < 6 {
		return marketData, fmt.Errorf("failed parsing market data fields")
	}
	date, err := parseDate(data[0])
	if err != nil {
		return marketData, err
	}
	tradeVolume, err := parseVolume(data[1])
	if err != nil {
		return marketData, fmt.Errorf("failed parsing market trade volume: %w", err)
	}
	tradeValue, err := parsePrice(data[2])
	if err != nil {
		return marketData, fmt.Errorf("failed parsing market trade value: %w", err)
	}
	transaction, err := parseVolume(data[3])
	if err != nil {
		return marketData, fmt.Errorf("failed parsing market transaction: %w", err)
	}
	index, err := parsePrice(data[4])
	if err != nil {
		return marketData, fmt.Errorf("failed parsing market index: %w", err)
	}
	change, err := parsePrice(data[5])
	if err != nil {
		return marketData, fmt.Errorf("failed parsing market change: %w", err)
	}
	marketData.Date = date
	marketData.TradeVolume = tradeVolume
	marketData.TradeValue = tradeValue
	marketData.Transaction = transaction
	marketData.Index = index
	marketData.Change = change
	return marketData, nil
}

func (*MarketDataService) parseTAIEX(data []string) (TAIEXIndex, error) {
	var index TAIEXIndex
	if len(data) < 5 {
		return index, fmt.Errorf("failed parsing TAIEX index fields")
	}
	date, err := parseDate(data[0])
	if err != nil {
		return index, err
	}
	open, err := parsePrice(data[1])
	if err != nil {
		return index, fmt.Errorf("failed parsing TAIEX open: %w", err)
	}
	high, err := parsePrice(data[2])
	if err != nil {
		return index, fmt.Errorf("failed parsing TAIEX high: %w", err)
	}
	low, err := parsePrice(data[3])
	if err != nil {
		return index, fmt.Errorf("failed parsing TAIEX low: %w", err)
	}
	close, err := parsePrice(data[4])
	if err != nil {
		return index, fmt.Errorf("failed parsing TAIEX close: %w", err)
	}
	index.Date = date
	index.Open = open
	index.High = high
	index.Low = low
	index.Close = close
	return index, nil
}

func (*MarketDataService) parseTPExIndex(data []string) (TPExIndex, error) {
	var index TPExIndex
	if len(data) < 6 {
		return index, fmt.Errorf("failed parsing TPEx index fields")
	}
	date, err := parseWesternDate(data[0])
	if err != nil {
		return index, err
	}
	index.Date = date
	fields := []struct {
		name  string
		value *decimal.Decimal
	}{
		{"open", &index.Open},
		{"high", &index.High},
		{"low", &index.Low},
		{"close", &index.Close},
		{"change", &index.Change},
	}
	for i, f := range fields {
		*f.value, err = parsePrice(data[i+1])
		if err != nil {
			return index, fmt.Errorf("failed parsing TPEx index %s: %w", f.name, err)
		}
	}
	return index, nil
}

// 台灣證卷交易所或是證券櫃檯買賣中心有最小查詢日期的限制
func (s *MarketDataService) MinimumDate(m Market) civil.Date {
	if m == TWSE {
		// 台灣證卷交易所每日市場成交資訊最早到民國79年1月
		return civil.Date{Year: 1990, Month: time.January, Day: 1}
	}
	// 證券櫃檯買賣中心每日市場成交資訊最早到民國88年1月
	return civil.Date{Year: 1999, Month: time.January, Day: 1}
}

// 從台灣證卷交易所下載盤後每日市場成交資訊
func (s *MarketDataService) DownloadTwse(year int, month time.Month) ([]MarketData, error) {
	date := civil.Date{Year: year, Month: month, Day: 1}
	if date.Before(s.MinimumDate(TWSE)) {
		return nil, fmt.Errorf("invalid date: %s", fmt.Sprintf("%04d-%02d", date.Year, date.Month))
	}
	url, _ := s.client.twseBaseURL.Parse(twseMarketDataPath)
	opts := twseOptions{
		Response: "json",
		Date:     fmt.Sprintf("%04d%02d%02d", date.Year, date.Month, date.Day),
	}
	url, _ = addOptions(url, opts)
	req, _ := s.client.NewRequest("GET", url.String(), nil)
	resp := &twseResponse{}
	_, err := s.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Stat != "OK" {
		if resp.Stat == "查詢日期大於今日，請重新查詢!" {
			return nil, ErrDateOutOffRange
		}
		return nil, fmt.Errorf("invalid state: %s", resp.Stat)
	}
	if len(resp.Fields) != 6 ||
		resp.Fields[0] != "日期" ||
		resp.Fields[1] != "成交股數" ||
		resp.Fields[2] != "成交金額" ||
		resp.Fields[3] != "成交筆數" ||
		resp.Fields[4] != "發行量加權股價指數" ||
		resp.Fields[5] != "漲跌點數" {
		return nil, fmt.Errorf("failed parsing quote fields: %s", strings.Join(resp.Fields, ","))
	}
	result := []MarketData{}
	for _, data := range resp.Data {
		marketData, err := s.parse(data)
		if err != nil {
			return nil, err
		}
		result = append(result, marketData)
	}
	return result, nil
}

// 從證券櫃檯買賣中心下載盤後每日市場成交資訊
func (s *MarketDataService) DownloadTpex(year int, month time.Month) ([]MarketData, error) {
	date := civil.Date{Year: year, Month: month, Day: 1}
	if date.Before(s.MinimumDate(TPEx)) {
		return nil, fmt.Errorf("invalid date: %s", fmt.Sprintf("%04d-%02d", date.Year, date.Month))
	}
	url, _ := s.client.tpexBaseURL.Parse(tpexMarketDataPath)
	opts := tpexOptions{
		Response: "json",
		Date:     fmt.Sprintf("%04d/%02d/%02d", date.Year, date.Month, date.Day),
	}
	url, _ = addOptions(url, opts)
	req, _ := s.client.NewRequest("GET", url.String(), nil)
	resp := &tpexResponse{}
	_, err := s.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Tables) != 1 || resp.Tables[0].TotalCount == 0 {
		return nil, ErrNoData
	}
	if resp.Tables[0].TotalCount != len(resp.Tables[0].Data) {
		return nil, fmt.Errorf("failed parsing market data length returned %d, want %d", resp.Tables[0].TotalCount, len(resp.Tables[0].Data))
	}
	result := []MarketData{}
	for _, data := range resp.Tables[0].Data {
		stringData := make([]string, len(data))
		for i, v := range data {
			stringData[i] = string(v)
		}
		marketData, err := s.parse(stringData)
		if err != nil {
			return nil, err
		}
		// 成交股數（仟股）
		marketData.TradeVolume *= 1000
		// 金額（仟元）
		marketData.TradeValue = marketData.TradeValue.Mul(decimal.NewFromInt(1000))
		result = append(result, marketData)
	}
	return result, nil
}

// 從台灣證卷交易所下載發行量加權股價指數歷史資料，最早到民國88年1月
func (s *MarketDataService) DownloadTAIEX(year int, month time.Month) ([]TAIEXIndex, error) {
	date := civil.Date{Year: year, Month: month, Day: 1}
	minimumDate := civil.Date{Year: 1999, Month: time.January, Day: 1}
	if date.Before(minimumDate) {
		return nil, fmt.Errorf("invalid date: %s", fmt.Sprintf("%04d-%02d", date.Year, date.Month))
	}
	u, _ := s.client.twseBaseURL.Parse(twseTAIEXPath)
	opts := twseOptions{
		Response: "json",
		Date:     fmt.Sprintf("%04d%02d%02d", date.Year, date.Month, date.Day),
	}
	u, _ = addOptions(u, opts)
	req, _ := s.client.NewRequest("GET", u.String(), nil)
	resp := &twseResponse{}
	_, err := s.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Stat != "OK" {
		if resp.Stat == "查詢日期大於今日，請重新查詢!" {
			return nil, ErrDateOutOffRange
		}
		return nil, fmt.Errorf("invalid state: %s", resp.Stat)
	}
	if len(resp.Fields) != 5 ||
		resp.Fields[0] != "日期" ||
		resp.Fields[1] != "開盤指數" ||
		resp.Fields[2] != "最高指數" ||
		resp.Fields[3] != "最低指數" ||
		resp.Fields[4] != "收盤指數" {
		return nil, fmt.Errorf("failed parsing TAIEX fields: %s", strings.Join(resp.Fields, ","))
	}
	result := []TAIEXIndex{}
	for _, data := range resp.Data {
		index, err := s.parseTAIEX(data)
		if err != nil {
			return nil, err
		}
		result = append(result, index)
	}
	return result, nil
}

// 從證券櫃檯買賣中心下載櫃買指數歷史資料，最早到民國88年9月
func (s *MarketDataService) DownloadTPExIndex(year int, month time.Month) ([]TPExIndex, error) {
	date := civil.Date{Year: year, Month: month, Day: 1}
	minimumDate := civil.Date{Year: 1999, Month: time.September, Day: 1}
	if date.Before(minimumDate) {
		return nil, fmt.Errorf("invalid date: %s", fmt.Sprintf("%04d-%02d", date.Year, date.Month))
	}
	u, _ := s.client.tpexBaseURL.Parse(tpexIndexPath)
	body := url.Values{
		"response": {"json"},
		"date":     {fmt.Sprintf("%04d/%02d/%02d", date.Year, date.Month, date.Day)},
	}.Encode()
	req, _ := s.client.NewRequest("POST", u.String(), body)
	resp := &tpexResponse{}
	_, err := s.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Tables) != 1 || resp.Tables[0].TotalCount == 0 {
		return nil, ErrNoData
	}
	if resp.Tables[0].TotalCount != len(resp.Tables[0].Data) {
		return nil, fmt.Errorf("failed parsing TPEx index data length returned %d, want %d", resp.Tables[0].TotalCount, len(resp.Tables[0].Data))
	}
	if len(resp.Tables[0].Fields) != 6 ||
		resp.Tables[0].Fields[0] != "日期" ||
		resp.Tables[0].Fields[1] != "開市" ||
		resp.Tables[0].Fields[2] != "最高" ||
		resp.Tables[0].Fields[3] != "最低" ||
		resp.Tables[0].Fields[4] != "收市" ||
		resp.Tables[0].Fields[5] != "漲/跌" {
		return nil, fmt.Errorf("failed parsing TPEx index fields: %s", strings.Join(resp.Tables[0].Fields, ","))
	}
	result := []TPExIndex{}
	for _, data := range resp.Tables[0].Data {
		stringData := make([]string, len(data))
		for i, v := range data {
			stringData[i] = string(v)
		}
		index, err := s.parseTPExIndex(stringData)
		if err != nil {
			return nil, err
		}
		result = append(result, index)
	}
	return result, nil
}
