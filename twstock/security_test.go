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

	_, err = client.Security.download("test-url", errDecoder{})
	if err == nil {
		t.Error("Security.download returned nil; expected error")
	}
}
