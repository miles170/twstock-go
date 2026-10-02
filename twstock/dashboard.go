package twstock

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-sql/civil"
)

func validateDashboardQuery(date civil.Date, days int) error {
	minimumDate := civil.Date{Year: 2026, Month: time.August, Day: 3}
	if !date.IsValid() || date.Before(minimumDate) {
		return fmt.Errorf("invalid dashboard date: %s; earliest date is %s", date, minimumDate)
	}
	if days <= 0 {
		return fmt.Errorf("invalid trading days: %d; must be positive", days)
	}
	return nil
}

func (c *Client) downloadDashboardJSON(baseURL *url.URL, path string, params url.Values, result any) error {
	u, err := baseURL.Parse(path)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()
	req, err := c.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.Do(req, result)
	return err
}

func checkDashboardState(stat, expected string) error {
	if stat == expected {
		return nil
	}
	if isDateOutOfRangeStat(stat) {
		return ErrDateOutOffRange
	}
	if stat == "很抱歉，沒有符合條件的資料!" {
		return ErrNoData
	}
	return fmt.Errorf("invalid dashboard state: %s", stat)
}

func parseDashboardDate(rawDate string) (civil.Date, error) {
	date, err := time.Parse("20060102", rawDate)
	if err != nil {
		return civil.Date{}, fmt.Errorf("failed parsing dashboard date: %w", err)
	}
	return civil.DateOf(date), nil
}
