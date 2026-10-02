package twstock

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-sql/civil"
	"github.com/shopspring/decimal"
)

const twseMarginMaintenanceRatioPath = "/rwd/zh/marginTrading/BFIJ3U_TREND"

// MarginMaintenanceRatio is the combined TWSE and TPEx market margin maintenance ratio.
type MarginMaintenanceRatio struct {
	Date  civil.Date      // 日期
	Ratio decimal.Decimal // 全市場擔保維持率（百分比，例如 178.76 表示 178.76%）
}

// DownloadMarginMaintenanceRatio downloads up to days trading days ending on date.
// Data is available from 2026-08-03. The API may return fewer days when less data
// is available. Ratio values are percentages, not fractions.
func (s *MarketDataService) DownloadMarginMaintenanceRatio(date civil.Date, days int) ([]MarginMaintenanceRatio, error) {
	minimumDate := civil.Date{Year: 2026, Month: time.August, Day: 3}
	if !date.IsValid() || date.Before(minimumDate) {
		return nil, fmt.Errorf("invalid margin maintenance ratio date: %s; earliest date is %s", date, minimumDate)
	}
	if days <= 0 {
		return nil, fmt.Errorf("invalid trading days: %d; must be positive", days)
	}

	u, err := s.client.twseBaseURL.Parse(twseMarginMaintenanceRatioPath)
	if err != nil {
		return nil, err
	}
	u.RawQuery = url.Values{
		"response": {"json"},
		"date":     {fmt.Sprintf("%04d%02d%02d", date.Year, date.Month, date.Day)},
		"days":     {strconv.Itoa(days)},
	}.Encode()
	req, err := s.client.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Stat string `json:"stat"`
		Data []struct {
			Date     string           `json:"date"`
			KeepRate *decimal.Decimal `json:"keepRate"`
		} `json:"data"`
	}
	if _, err := s.client.Do(req, &resp); err != nil {
		return nil, err
	}
	if resp.Stat != "OK" {
		if isDateOutOfRangeStat(resp.Stat) {
			return nil, ErrDateOutOffRange
		}
		if resp.Stat == "很抱歉，沒有符合條件的資料!" {
			return nil, ErrNoData
		}
		return nil, fmt.Errorf("invalid margin maintenance ratio state: %s", resp.Stat)
	}
	if len(resp.Data) == 0 {
		return nil, ErrNoData
	}
	result := make([]MarginMaintenanceRatio, 0, len(resp.Data))
	for _, row := range resp.Data {
		parsedDate, err := time.Parse("20060102", row.Date)
		if err != nil {
			return nil, fmt.Errorf("failed parsing margin maintenance ratio date: %w", err)
		}
		if row.KeepRate == nil {
			return nil, fmt.Errorf("missing margin maintenance ratio for %s", row.Date)
		}
		result = append(result, MarginMaintenanceRatio{
			Date:  civil.DateOf(parsedDate),
			Ratio: *row.KeepRate,
		})
	}
	return result, nil
}
