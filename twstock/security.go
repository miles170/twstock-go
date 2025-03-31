package twstock

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang-sql/civil"
	"golang.org/x/text/transform"
)

type SecurityService struct {
	client *Client
}

type Market string

const (
	TWSE Market = "tse" // 臺灣證券交易所
	TPEx Market = "otc" // 證券櫃檯買賣中心
)

// 有價證券
type Security struct {
	Type     string     // 有價證卷類別
	Code     string     // 有價證券代號
	Name     string     // 有價證券名稱
	ISIN     string     // 國際證卷辨識號碼
	IPO      civil.Date // 上市日
	Market   Market     // 市場別
	Industry string     // 產業
	CFI      string     // CFICode
	Remark   string     // 備註
}

// 下市的有價證卷
type DelistedSecurity struct {
	Code   string // 有價證券代號
	Name   string // 有價證券名稱
	Market Market // 市場別
}

func (s *SecurityService) download(url string, t transform.Transformer) ([]Security, error) {
	req, err := s.client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	doc, err := s.client.DoTransformToDocument(req, t)
	if err != nil {
		return nil, err
	}
	securities := []Security{}
	var securityType = ""
	doc.Find("tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// 跳過標題
		if i == 0 {
			return true
		}
		elements := s.Find("td")
		if len(elements.Nodes) == 1 {
			// 有價證卷類型
			securityType = strings.TrimSpace(elements.Find("b").First().Text())
		} else if len(elements.Nodes) == 7 {
			// 有價證券代號及名稱
			codeAndName := strings.Fields(elements.Eq(0).Text())
			code := codeAndName[0]
			name := codeAndName[1]
			isin := strings.TrimSpace(elements.Eq(1).Text())
			ipo, parseErr := time.Parse("2006/01/02", strings.TrimSpace(elements.Eq(2).Text()))
			if parseErr != nil {
				err = parseErr
				return false
			}
			var market Market
			marketText := strings.TrimSpace(elements.Eq(3).Text())
			switch marketText {
			case "上市":
				market = TWSE
			case "上市臺灣創新板":
				market = TWSE
			case "上櫃":
				market = TPEx
			default:
				err = fmt.Errorf("failed parsing security market: %s", marketText)
				return false
			}
			industry := strings.TrimSpace(elements.Eq(4).Text())
			cfi := strings.TrimSpace(elements.Eq(5).Text())
			remark := strings.TrimSpace(elements.Eq(6).Text())
			securities = append(securities,
				Security{securityType, code, name, isin, civil.DateOf(ipo), market, industry, cfi, remark})
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return securities, nil
}

const (
	// 上市證券國際證券辨識號碼一覽表
	twseSecuritiesPath = "/isin/C_public.jsp?strMode=2"
	// 終止上市公司
	twseDelistedSecuritiesPath = "/zh/company/suspendListing"

	// 上櫃證券國際證券辨識號碼一覽表
	tpexSecuritiesPath = "/isin/C_public.jsp?strMode=4"
	// 終止上櫃公司
	tpexDelistedSecuritiesPath = "/web/regular_emerging/deListed/de-listed_companies.php"
)

// 從台灣證卷交易所下載上市及上櫃國際證券資料
func (s *SecurityService) Download() ([]Security, error) {
	securities := []Security{}
	for _, path := range []string{twseSecuritiesPath, tpexSecuritiesPath} {
		url, _ := s.client.isinTwseBaseURL.Parse(path)
		s, err := s.download(url.String(), s.client.isinTwseDecoder)
		if err != nil {
			return nil, err
		}
		securities = append(securities, s...)
	}
	return securities, nil
}

// 從台灣證卷交易所下載下市的國際證券資料
func (s *SecurityService) DownloadTwseDelisted() ([]DelistedSecurity, error) {
	url, _ := s.client.twseBaseURL.Parse(twseDelistedSecuritiesPath)
	req, _ := s.client.NewRequest("POST", url.String(), "maxLength=-1&selectYear=&submitBtn=%E6%9F%A5%E8%A9%A2")
	doc, err := s.client.DoTransformToDocument(req, s.client.twseDecoder)
	if err != nil {
		return nil, err
	}
	delistedSecurities := []DelistedSecurity{}
	doc.Find("tbody tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		elements := s.Find("td")
		if len(elements.Nodes) != 3 {
			err = fmt.Errorf("failed parsing security fields")
			return false
		}
		name := strings.TrimSpace(elements.Eq(1).Text())
		code := strings.TrimSpace(elements.Eq(2).Text())
		delistedSecurities = append(delistedSecurities, DelistedSecurity{code, name, TWSE})
		return true
	})
	if err != nil {
		return nil, err
	}
	return delistedSecurities, nil
}

// 從證券櫃檯買賣中心下載下櫃的國際證券資料
func (s *SecurityService) DownloadTpexDelisted(page int) ([]DelistedSecurity, error) {
	url, _ := s.client.tpexBaseURL.Parse(tpexDelistedSecuritiesPath)
	req, _ := s.client.NewRequest("POST", url.String(), fmt.Sprintf("stk_code=&select_year=ALL&topage=%d&DELIST_REASON=-1", page+1))
	doc, err := s.client.DoTransformToDocument(req, s.client.twseDecoder)
	if err != nil {
		return nil, err
	}
	delistedSecurities := []DelistedSecurity{}
	doc.Find("table").First().Find("tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		elements := s.Find("td")
		length := len(elements.Nodes)
		// 標題
		if length == 3 {
			return true
		}
		if length != 4 {
			err = fmt.Errorf("failed parsing security fields")
			return false
		}
		code := strings.TrimSpace(elements.Eq(0).Text())
		name := strings.TrimSpace(elements.Eq(1).Find("a").Text())
		delistedSecurities = append(delistedSecurities, DelistedSecurity{code, name, TPEx})
		return true
	})
	if err != nil {
		return nil, err
	}
	return delistedSecurities, nil
}
