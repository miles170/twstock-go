package twstock

import (
	"fmt"
	"net/url"
	"strconv"

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
	if err := validateDashboardQuery(date, days); err != nil {
		return nil, err
	}
	params := url.Values{
		"response": {"json"},
		"date":     {fmt.Sprintf("%04d%02d%02d", date.Year, date.Month, date.Day)},
		"days":     {strconv.Itoa(days)},
	}
	var resp struct {
		Stat string `json:"stat"`
		Data []struct {
			Date     string           `json:"date"`
			KeepRate *decimal.Decimal `json:"keepRate"`
		} `json:"data"`
	}
	if err := s.client.downloadDashboardJSON(s.client.twseBaseURL, twseMarginMaintenanceRatioPath, params, &resp); err != nil {
		return nil, err
	}
	if err := checkDashboardState(resp.Stat, "OK"); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, ErrNoData
	}
	result := make([]MarginMaintenanceRatio, 0, len(resp.Data))
	for _, row := range resp.Data {
		parsedDate, err := parseDashboardDate(row.Date)
		if err != nil {
			return nil, err
		}
		if row.KeepRate == nil {
			return nil, fmt.Errorf("missing margin maintenance ratio for %s", row.Date)
		}
		result = append(result, MarginMaintenanceRatio{
			Date:  parsedDate,
			Ratio: *row.KeepRate,
		})
	}
	return result, nil
}
