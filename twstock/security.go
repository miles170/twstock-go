package twstock

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
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
	Type     string // 有價證卷類別
	Code     string // 有價證券代號
	Name     string // 有價證券名稱
	ISIN     string // 國際證卷辨識號碼
	IPO      string // 上市日 (格式為 2016/01/02)
	Market   Market // 市場別
	Industry string // 產業
	CFI      string // CFICode
	Remark   string // 備註
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
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		// 跳過標題
		if i == 0 {
			return
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
			isin := elements.Eq(1).Text()
			ipo := elements.Eq(2).Text()
			var market Market
			marketText := elements.Eq(3).Text()
			if marketText == "上市" {
				market = TWSE
			} else if marketText == "上櫃" {
				market = TPEx
			} else {
				return
			}
			industry := elements.Eq(4).Text()
			cfi := elements.Eq(5).Text()
			remark := elements.Eq(6).Text()
			securities = append(securities,
				Security{securityType, code, name, isin, ipo, market, industry, cfi, remark})
		}
	})
	return securities, nil
}

const (
	// 本國上市證券國際證券辨識號碼一覽表
	twseSecuritiesPath = "/isin/C_public.jsp?strMode=2"

	// 本國上櫃證券國際證券辨識號碼一覽表
	tpexSecuritiesPath = "/isin/C_public.jsp?strMode=4"
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
	url, _ := s.client.twseBaseURL.Parse("/zh/company/suspendListing")
	req, _ := s.client.NewRequest("POST", url.String(), "maxLength=-1&selectYear=&submitBtn=%E6%9F%A5%E8%A9%A2")
	doc, err := s.client.DoTransformToDocument(req, s.client.twseDecoder)
	if err != nil {
		return nil, err
	}
	delistedSecurities := []DelistedSecurity{}
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		elements := s.Find("td")
		if len(elements.Nodes) != 3 {
			return
		}
		name := elements.Eq(1).Text()
		code := elements.Eq(2).Text()
		delistedSecurities = append(delistedSecurities, DelistedSecurity{code, name, TWSE})
	})
	return delistedSecurities, nil
}
