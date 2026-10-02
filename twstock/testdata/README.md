# Margin maintenance ratio fixtures

These JSON responses were downloaded directly from the TWSE dashboard API on
2026-10-02 and preserved without modification. The dashboard's full-market view
uses `BFIJ3U_TREND`; `keepRate` is the combined TWSE and TPEx maintenance ratio
in percent. Data starts on 2026-08-03.

- `margin_maintenance_ratio.json`: https://www.twse.com.tw/rwd/zh/marginTrading/BFIJ3U_TREND?response=json&date=20261001&days=5
- `margin_maintenance_ratio_first.json`: https://www.twse.com.tw/rwd/zh/marginTrading/BFIJ3U_TREND?response=json&date=20260803&days=1
- `margin_maintenance_ratio_empty.json`: https://www.twse.com.tw/rwd/zh/marginTrading/BFIJ3U_TREND?response=json&date=20260802&days=1

Source page: https://www.twse.com.tw/dashboard/zh/credit/margin.html

# Margin balance fixtures

The following unmodified responses were also downloaded on 2026-10-02:

- `margin_balance_twse.json`: https://www.twse.com.tw/rwd/zh/marginTrading/MI_MARGN_TREND?response=json&date=20261001&days=5
- `margin_balance_twse_first.json`: https://www.twse.com.tw/rwd/zh/marginTrading/MI_MARGN_TREND?response=json&date=20260803&days=1
- `margin_balance_tpex.json`: https://www.tpex.org.tw/www/zh-tw/dashboardOtc/marginTrend?lang=zh-tw&date=115/10/01&days=5
- `margin_balance_tpex_first.json`: https://www.tpex.org.tw/www/zh-tw/dashboardOtc/marginTrend?lang=zh-tw&date=115/08/03&days=1
- `margin_balance_tpex_empty.json`: https://www.tpex.org.tw/www/zh-tw/dashboardOtc/marginTrend?lang=zh-tw&date=115/08/02&days=1

TWSE `marginAmt` is in thousands of TWD and `shortShr` is in lots. The
2026-10-01 values were checked against the `MI_MARGN` credit trading report.
TPEx `marginBalance` is in hundreds of millions of TWD (two decimal places)
and `shortBalance` is in lots. Its 2026-10-01 short balance of 36,462 was
verified by summing the current short balances in the official report at
https://www.tpex.org.tw/www/zh-tw/margin/balance?response=json&date=2026/10/01 .

The TPEx trend chart labels short balances as thousands of lots, but the
snapshot and the detailed report confirm that the JSON values are in lots.
TWSE has historical records before 2026-08-03; TPEx dashboard records start
on that date. Large `days` queries can therefore return different date ranges.
