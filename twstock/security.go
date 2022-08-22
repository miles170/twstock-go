package twstock

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/traditionalchinese"
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
	IPO      string // 上市日
	Market   Market // 市場別
	Industry string // 產業
	CFI      string // CFICode
	Remark   string // 備註
}

func (s *SecurityService) download(url string, t transform.Transformer) ([]Security, error) {
	req, err := s.client.NewRequest("GET", url)
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

// 從台灣證卷交易所下載國際證券資料
func (s *SecurityService) Download() ([]Security, error) {
	securities := []Security{}
	decoder := traditionalchinese.Big5.NewDecoder()
	for _, path := range []string{twseSecuritiesPath, tpexSecuritiesPath} {
		url, _ := s.client.isinTwseBaseURL.Parse(path)
		s, err := s.download(url.String(), decoder)
		if err != nil {
			return nil, err
		}
		securities = append(securities, s...)
	}
	return securities, nil
}
