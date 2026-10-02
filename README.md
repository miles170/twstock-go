# twstock-go

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/miles170/twstock-go/twstock)
[![Test Status](https://github.com/miles170/twstock-go/workflows/tests/badge.svg)](https://github.com/miles170/twstock-go/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/miles170/twstock-go/branch/main/graph/badge.svg)](https://codecov.io/gh/miles170/twstock-go)
[![Code Climate](https://codeclimate.com/github/miles170/twstock-go/badges/gpa.svg)](https://codeclimate.com/github/miles170/twstock-go)

從 [台灣證券交易所 (TWSE)](https://www.twse.com.tw/zh/) 及 [證券櫃檯買賣中心 (TPEx)](https://www.tpex.org.tw/web/) 下載台股資料的 Go 套件。

## Installation

```bash
go get github.com/miles170/twstock-go/twstock
```

## Usage

```go
import "github.com/miles170/twstock-go/twstock"

client := twstock.NewClient()
```

### 證券資料

#### 下載上市及上櫃國際證券識別碼 (ISIN)

```go
securities, err := client.Security.Download()
```

#### 下載已下市的上市證券資料

```go
securities, err := client.Security.DownloadTwseDelisted()
```

#### 下載已下櫃的上櫃證券資料

> 櫃買中心需指定頁數（從第 0 頁開始）

```go
securities, err := client.Security.DownloadTpexDelisted(0)
```

### 個股行情

#### 下載上市個股盤後日成交資訊

```go
quotes, err := client.Quote.DownloadTwse("2330", 2022, 8)
```

#### 下載上櫃個股盤後日成交資訊

```go
quotes, err := client.Quote.DownloadTpex("3374", 2022, 8)
```

#### 下載個股即時成交資訊

```go
quotes, err := client.Quote.Realtime("2330", "3374")
```

### 大盤成交資訊

#### 下載上市盤後每日市場成交資訊

```go
marketData, err := client.MarketData.DownloadTwse(2022, 8)
```

#### 下載上櫃盤後每日市場成交資訊

```go
marketData, err := client.MarketData.DownloadTpex(2022, 8)
```

### 指數資料

#### 下載發行量加權股價指數 (TAIEX) 歷史資料

> 最早資料：1999 年 1 月

```go
indices, err := client.MarketData.DownloadTAIEX(1999, 1)
```

#### 下載櫃買指數歷史資料

> 最早資料：1999 年 9 月

```go
indices, err := client.MarketData.DownloadTPExIndex(1999, 9)
```

### 全市場擔保維持率

從[證交所臺股儀表板](https://www.twse.com.tw/dashboard/zh/credit/margin.html)下載上市與上櫃合計的擔保維持率。資料自 **2026/08/03** 起提供。

指定截止日期與往前查詢的交易日筆數；資料不足時會回傳實際可取得的筆數。`Ratio` 為百分比數值，例如 `178.76` 表示 `178.76%`。

```go
// 取得首日資料。
ratios, err := client.MarketData.DownloadMarginMaintenanceRatio(
    civil.Date{Year: 2026, Month: time.August, Day: 3}, 1,
)

// 取得截至 2026/10/01 往前 60 個交易日的可用資料，包含 2026/08/03。
ratios, err = client.MarketData.DownloadMarginMaintenanceRatio(
    civil.Date{Year: 2026, Month: time.October, Day: 1}, 60,
)
```

無效日期、早於首日的日期或非正數筆數會回傳錯誤。無資料時可用 `errors.Is(err, twstock.ErrNoData)` 判斷；日期超出 API 可查範圍時回傳 `twstock.ErrDateOutOffRange`。HTTP、JSON 與資料格式錯誤也會回傳 error。

### 上市／上櫃融資融券餘額

指定市場、截止日期與往前查詢的交易日筆數，一次取得「融資金額餘額」及「融券張數餘額」。截止日期須自 `2026/08/03` 起，筆數須為正數。

```go
date := civil.Date{Year: 2026, Month: time.October, Day: 1}
listed, err := client.MarketData.DownloadMarginBalance(twstock.TWSE, date, 5)
otc, err := client.MarketData.DownloadMarginBalance(twstock.TPEx, date, 5)
```

回傳 `MarginBalance`：`Date` 為交易日期，`FinancingAmount` 統一為新臺幣「元」，`ShortSellingLots` 為「張」。上市來源以仟元提供金額，上櫃來源以億元提供金額，換算使用 decimal；上櫃原始金額保留兩位小數，換算後精度為百萬元。

上市 API 可回傳早於 `2026/08/03` 的歷史餘額；上櫃儀表板資料自該日起提供。資料不足時回傳實際筆數，不補零或改查其他日期。無資料、日期超出 API 查詢範圍、HTTP／連線、JSON 及欄位格式錯誤的處理與擔保維持率方法相同。

## License

[BSD-3-Clause](LICENSE)
