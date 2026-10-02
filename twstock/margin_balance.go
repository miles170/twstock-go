package twstock

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/golang-sql/civil"
	"github.com/shopspring/decimal"
)

// MarginBalance contains financing and short-selling balances for one market.
type MarginBalance struct {
	Date             civil.Date      // 日期
	FinancingAmount  decimal.Decimal // 融資金額餘額（元）
	ShortSellingLots int             // 融券張數餘額（張）
}

type marginBalanceRow struct {
	date   string
	amount *decimal.Decimal
	lots   *int
}

// DownloadMarginBalance downloads up to days trading days ending on date for
// TWSE or TPEx. Dashboard data is available from 2026-08-03. FinancingAmount is
// in TWD and ShortSellingLots is in lots. TPEx amounts have the source's 0.01
// hundred-million TWD precision (one million TWD).
func (s *MarketDataService) DownloadMarginBalance(market Market, date civil.Date, days int) ([]MarginBalance, error) {
	if err := validateDashboardQuery(date, days); err != nil {
		return nil, err
	}
	params := url.Values{"days": {strconv.Itoa(days)}}
	var stat, expectedStat string
	var rows []marginBalanceRow
	unit := decimal.NewFromInt(1000)
	switch market {
	case TWSE:
		params.Set("response", "json")
		params.Set("date", fmt.Sprintf("%04d%02d%02d", date.Year, date.Month, date.Day))
		var resp struct {
			Stat string `json:"stat"`
			Data []struct {
				Date   string           `json:"date"`
				Amount *decimal.Decimal `json:"marginAmt"`
				Lots   *int             `json:"shortShr"`
			} `json:"data"`
		}
		if err := s.client.downloadDashboardJSON(s.client.twseBaseURL, "/rwd/zh/marginTrading/MI_MARGN_TREND", params, &resp); err != nil {
			return nil, err
		}
		stat, expectedStat = resp.Stat, "OK"
		for _, row := range resp.Data {
			rows = append(rows, marginBalanceRow{row.Date, row.Amount, row.Lots})
		}
	case TPEx:
		params.Set("lang", "zh-tw")
		params.Set("date", fmt.Sprintf("%03d/%02d/%02d", date.Year-1911, date.Month, date.Day))
		var resp struct {
			Stat  string `json:"stat"`
			Trend []struct {
				Date   string           `json:"date"`
				Amount *decimal.Decimal `json:"marginBalance"`
				Lots   *int             `json:"shortBalance"`
			} `json:"trend"`
		}
		if err := s.client.downloadDashboardJSON(s.client.tpexBaseURL, "/www/zh-tw/dashboardOtc/marginTrend", params, &resp); err != nil {
			return nil, err
		}
		stat, expectedStat, unit = resp.Stat, "ok", decimal.NewFromInt(100000000)
		for _, row := range resp.Trend {
			rows = append(rows, marginBalanceRow{row.Date, row.Amount, row.Lots})
		}
	default:
		return nil, fmt.Errorf("invalid margin balance market: %s", market)
	}
	if err := checkDashboardState(stat, expectedStat); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNoData
	}
	result := make([]MarginBalance, 0, len(rows))
	for _, row := range rows {
		parsedDate, err := parseDashboardDate(row.date)
		if err != nil {
			return nil, err
		}
		if row.amount == nil || row.lots == nil {
			return nil, fmt.Errorf("missing margin balance for %s", row.date)
		}
		result = append(result, MarginBalance{parsedDate, row.amount.Mul(unit), *row.lots})
	}
	return result, nil
}
