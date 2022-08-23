package twstock

import (
	"fmt"
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
	tpexMarketDataPath = "/web/stock/aftertrading/daily_trading_index/st41_result.php"
)

type MarketData struct {
	Date        civil.Date      // 日期
	TradeVolume int             // 總成交股數
	TradeValue  decimal.Decimal // 總成交金額
	Transaction int             // 總成交筆數
	Index       decimal.Decimal // 發行量加權股價指數或櫃買指數
	Change      decimal.Decimal // 漲跌點數
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
		if resp.Stat == "查詢日期大於今日，請重新查詢!" || resp.Stat == "查詢日期小於79年1月4日，請重新查詢!" {
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
		// 需要將西元年轉為民國年
		Date: fmt.Sprintf("%d/%02d", date.Year-1911, date.Month),
	}
	url, _ = addOptions(url, opts)
	req, _ := s.client.NewRequest("GET", url.String(), nil)
	resp := &tpexResponse{}
	_, err := s.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.DataLength != len(resp.Data) {
		return nil, fmt.Errorf("failed parsing market data length returned %d, want %d", resp.DataLength, len(resp.Data))
	}
	if resp.DataLength == 0 {
		return nil, ErrNoData
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
