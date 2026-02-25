# twstock-go

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/miles170/twstock-go/twstock)
[![Test Status](https://github.com/miles170/twstock-go/workflows/tests/badge.svg)](https://github.com/miles170/twstock-go/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/miles170/twstock-go/branch/main/graph/badge.svg)](https://codecov.io/gh/miles170/twstock-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/miles170/twstock-go)](https://goreportcard.com/report/github.com/miles170/twstock-go)
[![Code Climate](https://codeclimate.com/github/miles170/twstock-go/badges/gpa.svg)](https://codeclimate.com/github/miles170/twstock-go)

從[台灣證券交易所 (TWSE)](https://www.twse.com.tw/zh/) 及[證券櫃檯買賣中心 (TPEx)](https://www.tpex.org.tw/web/) 下載台股資料的 Go 套件。

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

## License

[BSD-3-Clause](LICENSE)
