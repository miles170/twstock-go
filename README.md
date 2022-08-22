# twstock-go

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/miles170/twstock-go/twstock)
[![Test Status](https://github.com/miles170/twstock-go/workflows/tests/badge.svg)](https://github.com/miles170/twstock-go/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/miles170/twstock-go/branch/main/graph/badge.svg)](https://codecov.io/gh/miles170/twstock-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/miles170/twstock-go)](https://goreportcard.com/report/github.com/miles170/twstock-go)
[![Code Climate](https://codeclimate.com/github/miles170/twstock-go/badges/gpa.svg)](https://codeclimate.com/github/miles170/twstock-go)

從[證券交易所 (TWSE)](https://www.twse.com.tw/zh/) 及[證券櫃台買賣中心 (TPEX)](https://www.tpex.org.tw/web/) 抓取有價證卷資料

## Installation

twstock-go is compatible with modern Go releases in module mode, with Go installed:

```bash
go get github.com/miles170/twstock-go/twstock
```

## Usage

```go
import "github.com/miles170/twstock-go/twstock"

client := twstock.NewClient()
```

### 從台灣證卷交易所下載上市及上櫃國際證券資料

```go
securities, err := client.Security.Download()
```

### 從台灣證卷交易所下載下市的國際證券資料

```go
securities, err := client.Security.DownloadTwseDelisted()
```

### 從證券櫃檯買賣中心下載下櫃的國際證券資料

> [櫃買中心查詢下櫃證卷資料](https://www.tpex.org.tw/web/regular_emerging/deListed/de-listed_companies.php?l=zh-tw)需要指定頁數

```go
securities, err := client.Security.DownloadTpexDelisted(0)
```

### 從台灣證卷交易所或證券櫃檯買賣中心下載盤後個股日成交資訊

```go
quotes, err := client.Quote.DownloadTwse("2330", 2022, 8)
```

### 從台灣證卷交易所下載即時個股成交資訊

```go
quotes, err := client.Quote.Realtime([]string{"2330"})
```

## License

[BSD-3-Clause](LICENSE)
