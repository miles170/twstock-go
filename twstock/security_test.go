package twstock

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/text/encoding/traditionalchinese"
)

func TestSecurityService_Download(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/isin/C_public.jsp", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		enc := traditionalchinese.Big5.NewEncoder()
		raw := ""
		mode := r.FormValue("strMode")
		if mode == "2" {
			raw = `
			<link rel="stylesheet" href="http://isin.twse.com.tw/isin/style1.css" type="text/css">
			<body>
			<table  align=center>
				<h2><strong><font class='h1'>本國上市證券國際證券辨識號碼一覽表</font></strong></h2>
				<h2><strong><font class='h1'><center>最近更新日期:2022/08/21  </center> </font></strong></h2>
				<h2><font color='red'><center>掛牌日以正式公告為準</center></font></h2>
			</table>
			<TABLE class='h4' align=center cellSpacing=3 cellPadding=2 width=750 border=0>
				<tr align=center>
					<td bgcolor=#D5FFD5>有價證券代號及名稱 </td>
					<td bgcolor=#D5FFD5>國際證券辨識號碼(ISIN Code)</td>
					<td bgcolor=#D5FFD5>上市日</td>
					<td bgcolor=#D5FFD5>市場別</td>
					<td bgcolor=#D5FFD5>產業別</td>
					<td bgcolor=#D5FFD5>CFICode</td>
					<td bgcolor=#D5FFD5>備註</td>
				</tr>
				<tr><td bgcolor=#FAFAD2 colspan=7 ><B> 股票 <B> </td></tr>
				<tr>
					<td bgcolor=#FAFAD2>1101　台泥</td>
					<td bgcolor=#FAFAD2>TW0001101004</td>
					<td bgcolor=#FAFAD2>1962/02/09</td>
					<td bgcolor=#FAFAD2>上市</td>
					<td bgcolor=#FAFAD2>水泥工業</td>
					<td bgcolor=#FAFAD2>ESVUFR</td>
					<td bgcolor=#FAFAD2></td>
				</tr>
				<tr>
					<td bgcolor=#FAFAD2>1102　亞泥</td>
					<td bgcolor=#FAFAD2>TW0001102002</td>
					<td bgcolor=#FAFAD2>1962/06/08</td>
					<td bgcolor=#FAFAD2>上市</td>
					<td bgcolor=#FAFAD2>水泥工業</td>
					<td bgcolor=#FAFAD2>ESVUFR</td>
					<td bgcolor=#FAFAD2></td>
				</tr>
			</table>
			<font color='red'><center>掛牌日以正式公告為準</center></font>`
		} else if mode == "4" {
			raw = `
			<link rel="stylesheet" href="http://isin.twse.com.tw/isin/style1.css" type="text/css">
			<body>
			<table  align=center>
				<h2><strong><font class='h1'>本國上櫃證券國際證券辨識號碼一覽表</font></strong></h2>
				<h2><strong><font class='h1'><center>最近更新日期:2022/08/21  </center> </font></strong></h2>
				<h2><font color='red'><center>掛牌日以正式公告為準</center></font></h2>
			</table>
			<TABLE class='h4' align=center cellSpacing=3 cellPadding=2 width=750 border=0>
				<tr align=center>
					<td bgcolor=#D5FFD5>有價證券代號及名稱 </td>
					<td bgcolor=#D5FFD5>國際證券辨識號碼(ISIN Code)</td>
					<td bgcolor=#D5FFD5>上市日</td>
					<td bgcolor=#D5FFD5>市場別</td>
					<td bgcolor=#D5FFD5>產業別</td>
					<td bgcolor=#D5FFD5>CFICode</td>
					<td bgcolor=#D5FFD5>備註</td>
				</tr>
				<tr><td bgcolor=#FAFAD2 colspan=7 ><B> 上櫃認購(售)權證 <B> </td></tr>
				<tr>
					<td bgcolor=#FAFAD2>70286P　驊訊元富18售01</td>
					<td bgcolor=#FAFAD2>TW21Z70286P0</td>
					<td bgcolor=#FAFAD2>2021/11/23</td>
					<td bgcolor=#FAFAD2>上櫃</td>
					<td bgcolor=#FAFAD2></td>
					<td bgcolor=#FAFAD2>RWSCPE</td>
					<td bgcolor=#FAFAD2></td>
				</tr>
				<tr>
					<td bgcolor=#FAFAD2>70299P　合晶元富18售03</td>
					<td bgcolor=#FAFAD2>TW21Z70299P3</td>
					<td bgcolor=#FAFAD2>2021/11/26</td>
					<td bgcolor=#FAFAD2>上櫃</td>
					<td bgcolor=#FAFAD2></td>
					<td bgcolor=#FAFAD2>RWSCPE</td>
					<td bgcolor=#FAFAD2></td>
				</tr>
			</table>
			<font color='red'><center>掛牌日以正式公告為準</center></font>`
		}
		s, err := enc.String(raw)
		if err == nil {
			fmt.Fprint(w, s)
		}
	})

	securities, err := client.Security.Download()
	if err != nil {
		t.Errorf("Security.Download returned error: %v", err)
	}
	want := []Security{
		{"股票", "1101", "台泥", "TW0001101004", "1962/02/09", "tse", "水泥工業", "ESVUFR", ""},
		{"股票", "1102", "亞泥", "TW0001102002", "1962/06/08", "tse", "水泥工業", "ESVUFR", ""},
		{"上櫃認購(售)權證", "70286P", "驊訊元富18售01", "TW21Z70286P0", "2021/11/23", "otc", "", "RWSCPE", ""},
		{"上櫃認購(售)權證", "70299P", "合晶元富18售03", "TW21Z70299P3", "2021/11/26", "otc", "", "RWSCPE", ""},
	}
	if !cmp.Equal(securities, want) {
		t.Errorf("Security.Download returned %v, want %v", securities, want)
	}
}

func TestSecurityService_DownloadBadContent(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/isin/C_public.jsp", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		enc := traditionalchinese.Big5.NewEncoder()
		s, err := enc.String(`
		<link rel="stylesheet" href="http://isin.twse.com.tw/isin/style1.css" type="text/css">
		<body>
		<table  align=center>
			<h2><strong><font class='h1'>本國上市證券國際證券辨識號碼一覽表</font></strong></h2>
			<h2><strong><font class='h1'><center>最近更新日期:2022/08/21  </center> </font></strong></h2>
			<h2><font color='red'><center>掛牌日以正式公告為準</center></font></h2>
		</table>
		<TABLE class='h4' align=center cellSpacing=3 cellPadding=2 width=750 border=0>
			<tr align=center>
				<td bgcolor=#D5FFD5>有價證券代號及名稱 </td>
				<td bgcolor=#D5FFD5>國際證券辨識號碼(ISIN Code)</td>
				<td bgcolor=#D5FFD5>上市日</td>
				<td bgcolor=#D5FFD5>市場別</td>
				<td bgcolor=#D5FFD5>產業別</td>
				<td bgcolor=#D5FFD5>CFICode</td>
				<td bgcolor=#D5FFD5>備註</td>
			</tr>
			<tr><td bgcolor=#FAFAD2 colspan=7 ><B> 股票 <B> </td></tr>
			<tr>
				<td bgcolor=#FAFAD2>1101　台泥</td>
				<td bgcolor=#FAFAD2>TW0001101004</td>
				<td bgcolor=#FAFAD2>1962/02/09</td>
				<td bgcolor=#FAFAD2>======</td>
				<td bgcolor=#FAFAD2>水泥工業</td>
				<td bgcolor=#FAFAD2>ESVUFR</td>
				<td bgcolor=#FAFAD2></td>
			</tr>
		</table>
		<font color='red'><center>掛牌日以正式公告為準</center></font>`)
		if err == nil {
			fmt.Fprint(w, s)
		}
	})

	securities, err := client.Security.Download()
	if err != nil {
		t.Errorf("Security.Download returned error: %v", err)
	}
	want := []Security{}
	if !cmp.Equal(securities, want) {
		t.Errorf("Security.Download returned %v, want %v", securities, want)
	}
}

func TestSecurityService_DownloadError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/isin/C_public.jsp", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusBadRequest)
	})

	_, err := client.Security.Download()
	if err == nil {
		t.Error("Security.Download returned nil; expected error")
	}
	testErrorContains(t, err, ": 400")

	_, err = client.Security.download("\n", nil)
	if err == nil {
		t.Error("Security.download returned nil; expected error")
	}

	decoder := errDecoder{}
	client.isinTwseDecoder = &decoder
	_, err = client.Security.Download()
	if err == nil {
		t.Error("Security.Download returned nil; expected error")
	}
}

func TestSecurityService_DownloadTwseDelisted(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/zh/company/suspendListing", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		raw := `
<!doctype html>
<html lang="zh-hant">
    <body class="zh ">
        <div id="layout">
            <div class="body">
                <div class="wp">
                    <div class="content">
                        <aside id="sidebar"></aside>
                        <main id="main">
                            <a id="accesskey-c" accesskey="c" href="#accesskey-c" style="text-decoration: none;" title="主要內容(C)">:::</a>
                            <div id="breadcrumbs"></div>
                            <div id="main-form">
                                <div class="outer">
                                    <h2>終止上市公司</h2>
                                    <div class="body">
                                        <form action="/zh/company/suspendListing" method="post" class="main" name="search_form" id="search_form">
                                            <input type="hidden" name="maxLength" value="10" id="maxLength"/>
                                            請選擇年份︰
            
                                            <select name="selectYear" id="selectYear">
                                                <option value="">全部</option>
                                                <option value="2022" selected="selected">111</option>
                                                <option value="2021">110</option>
                                                <option value="2020">109</option>
                                                <option value="2019">108</option>
                                                <option value="2018">107</option>
                                                <option value="2017">106</option>
                                                <option value="2016">105</option>
                                                <option value="2015">104</option>
                                                <option value="2014">103</option>
                                                <option value="2013">102</option>
                                                <option value="2012">101</option>
                                                <option value="2011">100</option>
                                                <option value="2010">99</option>
                                                <option value="2009">98</option>
                                                <option value="2008">97</option>
                                                <option value="2007">96</option>
                                                <option value="2006">95</option>
                                                <option value="2005">94</option>
                                                <option value="2004">93</option>
                                                <option value="2003">92</option>
                                                <option value="2002">91</option>
                                                <option value="2001">90</option>
                                            </select>
                                            <input type="submit" name="submitBtn" id="submitBtn" value="查詢"/>
                                        </form>
                                    </div>
                                </div>
                            </div>
                            <div id="reports" style="display: block;">
                                <div class="tools">
                                    <a data-et="終止上市公司" href="/company/suspendListingCsvAndHtml?type=html&amp;selectYear=2022&amp;lang=zh" class="html" target="_blank">列印 / HTML</a>
                                    <a data-et="終止上市公司" href="/company/suspendListingCsvAndHtml?type=csv&amp;selectYear=2022&amp;lang=zh" class="csv">CSV 下載</a>
                                </div>
                                <div class="title">
                                    <h1>終止上市公司</h1>
                                </div>
                            </div>
                            <article>
                                <form action="/zh/company/suspendListing" method="post" mehtod="POST" class="changeLength">
                                    <input type="hidden" name="selectYear" value="2022" id="selectYear"/>
                                    <label>
                                        每頁 
                                        <select id="maxLength" name="maxLength">
                                            <option value="10" selected="selected">10</option>
                                            <option value="25">25</option>
                                            <option value="50">50</option>
                                            <option value="100">100</option>
                                            <option value="-1">全部</option>
                                        </select>
                                        筆
                                    </label>
                                </form>
                                <table class="grid" width="100%">
                                    <thead>
                                        <tr>
                                            <th>終止上市日期</th>
                                            <th>公司名稱</th>
                                            <th>上市編號</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr>
                                            <td>111年08月04日</td>
                                            <td>台開</td>
                                            <td>2841</td>
                                        </tr>
                                        <tr>
                                            <td>111年06月29日</td>
                                            <td>互億</td>
                                            <td>6172</td>
                                        </tr>
                                        <tr>
                                            <td>111年06月27日</td>
                                            <td>客思達-KY</td>
                                            <td>2936</td>
                                        </tr>
                                        <tr>
                                            <td>111年05月03日</td>
                                            <td>龍燈-KY</td>
                                            <td>4141</td>
                                        </tr>
                                        <tr>
                                            <td>111年04月21日</td>
                                            <td>永大</td>
                                            <td>1507</td>
                                        </tr>
                                        <tr>
                                            <td>111年03月18日</td>
                                            <td>精熙-DR</td>
                                            <td>9188</td>
                                        </tr>
                                        <tr>
                                            <td>111年03月03日</td>
                                            <td>基勝-KY</td>
                                            <td>8427</td>
                                        </tr>
                                        <tr>
                                            <td>111年01月27日</td>
                                            <td>英瑞-KY</td>
                                            <td>1592</td>
                                        </tr>
                                        <tr>
                                            <td>111年01月05日</td>
                                            <td>奇力新</td>
                                            <td>2456</td>
                                        </tr>
                                    </tbody>
                                </table>
                                <div class="pagination"></div>
                            </article>
                        </main>
                    </div>
                </div>
            </div>
        </div>
    </body>
</html>
		`
		fmt.Fprint(w, raw)
	})

	securities, err := client.Security.DownloadTwseDelisted()
	if err != nil {
		t.Errorf("Security.DownloadTwseDelisted returned error: %v", err)
	}
	want := []DelistedSecurity{
		{"2841", "台開", TWSE},
		{"6172", "互億", TWSE},
		{"2936", "客思達-KY", TWSE},
		{"4141", "龍燈-KY", TWSE},
		{"1507", "永大", TWSE},
		{"9188", "精熙-DR", TWSE},
		{"8427", "基勝-KY", TWSE},
		{"1592", "英瑞-KY", TWSE},
		{"2456", "奇力新", TWSE},
	}
	if !cmp.Equal(securities, want) {
		t.Errorf("Security.Download returned %v, want %v", securities, want)
	}
}

func TestSecurityService_DownloadTwseDelistedBadContent(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/zh/company/suspendListing", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		raw := `
<!doctype html>
<html lang="zh-hant">
    <body class="zh ">
        <div id="layout">
            <div class="body">
                <div class="wp">
                    <div class="content">
                        <aside id="sidebar"></aside>
                        <main id="main">
                            <a id="accesskey-c" accesskey="c" href="#accesskey-c" style="text-decoration: none;" title="主要內容(C)">:::</a>
                            <div id="breadcrumbs"></div>
                            <div id="main-form">
                                <div class="outer">
                                    <h2>終止上市公司</h2>
                                    <div class="body">
                                        <form action="/zh/company/suspendListing" method="post" class="main" name="search_form" id="search_form">
                                            <input type="hidden" name="maxLength" value="10" id="maxLength"/>
                                            請選擇年份︰
            
                                            <select name="selectYear" id="selectYear">
                                                <option value="">全部</option>
                                                <option value="2022" selected="selected">111</option>
                                                <option value="2021">110</option>
                                                <option value="2020">109</option>
                                                <option value="2019">108</option>
                                                <option value="2018">107</option>
                                                <option value="2017">106</option>
                                                <option value="2016">105</option>
                                                <option value="2015">104</option>
                                                <option value="2014">103</option>
                                                <option value="2013">102</option>
                                                <option value="2012">101</option>
                                                <option value="2011">100</option>
                                                <option value="2010">99</option>
                                                <option value="2009">98</option>
                                                <option value="2008">97</option>
                                                <option value="2007">96</option>
                                                <option value="2006">95</option>
                                                <option value="2005">94</option>
                                                <option value="2004">93</option>
                                                <option value="2003">92</option>
                                                <option value="2002">91</option>
                                                <option value="2001">90</option>
                                            </select>
                                            <input type="submit" name="submitBtn" id="submitBtn" value="查詢"/>
                                        </form>
                                    </div>
                                </div>
                            </div>
                            <div id="reports" style="display: block;">
                                <div class="tools">
                                    <a data-et="終止上市公司" href="/company/suspendListingCsvAndHtml?type=html&amp;selectYear=2022&amp;lang=zh" class="html" target="_blank">列印 / HTML</a>
                                    <a data-et="終止上市公司" href="/company/suspendListingCsvAndHtml?type=csv&amp;selectYear=2022&amp;lang=zh" class="csv">CSV 下載</a>
                                </div>
                                <div class="title">
                                    <h1>終止上市公司</h1>
                                </div>
                            </div>
                            <article>
                                <form action="/zh/company/suspendListing" method="post" mehtod="POST" class="changeLength">
                                    <input type="hidden" name="selectYear" value="2022" id="selectYear"/>
                                    <label>
                                        每頁 
                                        <select id="maxLength" name="maxLength">
                                            <option value="10" selected="selected">10</option>
                                            <option value="25">25</option>
                                            <option value="50">50</option>
                                            <option value="100">100</option>
                                            <option value="-1">全部</option>
                                        </select>
                                        筆
                                    </label>
                                </form>
                                <table class="grid" width="100%">
                                    <thead>
                                        <tr>
                                            <th>終止上市日期</th>
                                            <th>公司名稱</th>
                                            <th>上市編號</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr>
                                            <td>台開</td>
                                            <td>2841</td>
                                        </tr>
                                        <tr>
                                            <td>111年06月29日</td>
                                            <td>6172</td>
                                        </tr>
                                        <tr>
                                            <td>111年06月27日</td>
                                            <td>客思達-KY</td>
                                        </tr>
                                        <tr>
                                            <td>111年05月03日</td>
                                            <td>4141</td>
                                        </tr>
                                        <tr>
                                            <td>永大</td>
                                            <td>1507</td>
                                        </tr>
                                        <tr>
                                            <td>111年03月18日</td>
                                            <td>精熙-DR</td>
                                        </tr>
                                        <tr>
                                            <td>111年03月03日</td>
                                            <td>8427</td>
                                        </tr>
                                        <tr>
                                            <td>英瑞-KY</td>
                                            <td>1592</td>
                                        </tr>
                                        <tr>
                                            <td>111年01月05日</td>
                                            <td>奇力新</td>
                                        </tr>
                                    </tbody>
                                </table>
                                <div class="pagination"></div>
                            </article>
                        </main>
                    </div>
                </div>
            </div>
        </div>
    </body>
</html>
		`
		fmt.Fprint(w, raw)
	})

	securities, err := client.Security.DownloadTwseDelisted()
	if err != nil {
		t.Errorf("Security.DownloadTwseDelisted returned error: %v", err)
	}
	want := []DelistedSecurity{}
	if !cmp.Equal(securities, want) {
		t.Errorf("Security.DownloadTwseDelisted returned %v, want %v", securities, want)
	}
}

func TestSecurityService_DownloadTwseDelistedError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/zh/company/suspendListing", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusBadRequest)
	})

	_, err := client.Security.DownloadTwseDelisted()
	if err == nil {
		t.Error("Security.DownloadTwseDelisted returned nil; expected error")
	}
	testErrorContains(t, err, ": 400")

	decoder := errDecoder{}
	client.twseDecoder = &decoder
	_, err = client.Security.DownloadTwseDelisted()
	if err == nil {
		t.Error("Security.DownloadTwseDelisted returned nil; expected error")
	}
}
