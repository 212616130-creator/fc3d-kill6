// Package report 生成预测页面 HTML（移动端优先，内联样式零外部依赖）。
package report

import (
	"bytes"
	"text/template"

	"fc3d-kill6/backtest"
)

// Data 模板数据
type Data struct {
	Meta    backtest.Meta
	Pred    backtest.Predict
	Rows    []backtest.Row
	Banners Banners
}

// Banners 页面顶部横幅
type Banners struct {
	DataUpgrade    bool // 算法升级触发（红条）
	UpgradeReasons []string
	DataFailed     bool // 数据源全挂（橙条）
}

const tmplSrc = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>福彩3D 百十个杀码预测</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC",sans-serif;background:#f5f7fa;color:#333;padding:12px;max-width:600px;margin:0 auto}
h1{font-size:18px;text-align:center;color:#1a237e;margin:8px 0 12px}
.pred{background:linear-gradient(135deg,#1a237e,#283593);border-radius:14px;padding:16px;color:#fff;margin-bottom:14px}
.pred .badge{font-size:11px;opacity:.8;margin-bottom:4px}
.pred .issue{font-size:13px;margin-bottom:12px}
.poses{display:flex;gap:10px}
.pos{flex:1;text-align:center;background:rgba(255,255,255,.12);border-radius:10px;padding:12px 6px}
.pos-label{font-size:11px;opacity:.8;margin-bottom:4px}
.pos-num{font-size:32px;font-weight:800;line-height:1}
.section-title{font-size:14px;font-weight:700;color:#455a64;margin:16px 0 8px;display:flex;align-items:center;gap:8px}
.section-title .dot{width:8px;height:8px;border-radius:50%;background:#1a237e;display:inline-block}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin-bottom:14px}
.stat{background:#fff;border-radius:10px;padding:12px 8px;text-align:center;box-shadow:0 1px 4px rgba(0,0,0,.06)}
.stat .sv{font-size:24px;font-weight:800;color:#1a237e}
.stat .sl{font-size:11px;color:#78909c;margin-top:2px}
.stat .se{font-size:10px;color:#90a4ae}
.period-stat{background:#fff;border-radius:10px;padding:12px;text-align:center;box-shadow:0 1px 4px rgba(0,0,0,.06);margin-bottom:14px}
.period-stat .pv{font-size:22px;font-weight:800;color:#e65100}
.period-stat .pl{font-size:11px;color:#78909c}
.info-card{background:#fff;border-radius:10px;padding:14px;margin-bottom:12px;box-shadow:0 1px 4px rgba(0,0,0,.06);font-size:12px;line-height:1.7}
.info-card h3{font-size:13px;color:#37474f;margin-bottom:6px}
.warn{background:#fff3e0;border-left:3px solid #ff9800;padding:10px 12px;border-radius:0 8px 8px 0;font-size:11px;margin-top:8px;color:#e65100}
.upgrade-alert{background:linear-gradient(135deg,#b71c1c,#d32f2f);color:#fff;border-radius:12px;padding:14px 16px;margin-bottom:14px;font-size:13px;line-height:1.7;box-shadow:0 2px 8px rgba(183,28,28,.3)}
.upgrade-alert .ua-title{font-size:15px;font-weight:800;margin-bottom:4px}
.data-alert{background:linear-gradient(135deg,#e65100,#f57c00);color:#fff;border-radius:12px;padding:14px 16px;margin-bottom:14px;font-size:13px;line-height:1.7;box-shadow:0 2px 8px rgba(230,81,0,.3)}
.data-alert .da-title{font-size:15px;font-weight:800;margin-bottom:4px}
table{width:100%;border-collapse:collapse;font-size:11px;background:#fff;border-radius:10px;overflow:hidden;box-shadow:0 1px 4px rgba(0,0,0,.06)}
th{background:#eceff1;padding:8px 6px;text-align:center;font-weight:600;color:#455a64;position:sticky;top:0}
td{padding:6px;text-align:center;border-bottom:1px solid #f0f0f0}
.ok{color:#2e7d32;font-weight:700}
.bad{color:#c62828;font-weight:700}
.table-wrap{max-height:60vh;overflow-y:auto;-webkit-overflow-scrolling:touch;border-radius:10px;box-shadow:0 1px 4px rgba(0,0,0,.06)}
.disclaimer{text-align:center;font-size:10px;color:#90a4ae;margin-top:16px;padding:8px}
@media(max-width:380px){.pos-num{font-size:26px}.stats{grid-template-columns:repeat(3,1fr);gap:6px}.stat{padding:8px 4px}.stat .sv{font-size:20px}}
</style>
</head>
<body>
<h1>福彩3D 百十个杀码预测</h1>

<div class="pred">
<div class="badge">🔮 下一期预测（6杀制）</div>
<div class="issue">第 <strong>{{.NextIssue}}</strong> 期</div>
<div class="poses">
<div class="pos"><div class="pos-label">百位杀码</div><div class="pos-num">{{.Pred.H}},{{.Pred.H2}}</div></div>
<div class="pos"><div class="pos-label">十位杀码</div><div class="pos-num">{{.Pred.T}},{{.Pred.T2}}</div></div>
<div class="pos"><div class="pos-label">个位杀码</div><div class="pos-num">{{.Pred.O}},{{.Pred.O2}}</div></div>
</div>
</div>
{{if .Banners.DataFailed}}
<div class="data-alert"><div class="da-title">⚠️ 数据源异常</div>所有数据源获取失败，页面为最后一次成功数据，请检查数据源（灰鸟/17500）。</div>
{{end}}
{{if .Banners.DataUpgrade}}
<div class="upgrade-alert"><div class="ua-title">🚨 算法升级触发</div>6杀全中率已触及升级阈值，建议重新穷举6个算法：<br>{{range .Banners.UpgradeReasons}}• {{.}}<br>{{end}}<span style="font-size:11px;opacity:.85">触发条件：滚动100期跌破70% 或 单月下滑超 8pp</span></div>
{{end}}

<div class="section-title"><span class="dot"></span>近{{.Meta.BacktestN}}期回测（3杀+6杀）</div>
<div class="stats">
<div class="stat"><div class="sv">{{printf "%.1f" .Meta.AccH}}%</div><div class="sl">百位</div><div class="se">错{{.Meta.ErrH}}期</div></div>
<div class="stat"><div class="sv">{{printf "%.1f" .Meta.AccT}}%</div><div class="sl">十位</div><div class="se">错{{.Meta.ErrT}}期</div></div>
<div class="stat"><div class="sv">{{printf "%.1f" .Meta.AccO}}%</div><div class="sl">个位</div><div class="se">错{{.Meta.ErrO}}期</div></div>
</div>

<div class="stats" style="grid-template-columns:repeat(4,1fr)">
<div class="stat"><div class="sv" style="font-size:20px;color:#0f3460">{{printf "%.1f" .Meta.AccH2}}%</div><div class="sl">百kill2</div></div>
<div class="stat"><div class="sv" style="font-size:20px;color:#533483">{{printf "%.1f" .Meta.AccT2}}%</div><div class="sl">十kill2</div></div>
<div class="stat"><div class="sv" style="font-size:20px;color:#16a085">{{printf "%.1f" .Meta.AccO2}}%</div><div class="sl">个kill2</div></div>
<div class="stat" style="border-top:3px solid #e94560"><div class="sv" style="font-size:20px;color:#e94560">{{printf "%.1f" .Meta.All6Pct}}%</div><div class="sl">6杀全中</div></div>
</div>

<div class="info-card">
<h3>📋 六杀引擎</h3>
<p><strong>每位置双杀码：</strong>kill1（条件决策树）+ kill2（独立算术公式）<br>
<strong>kill2公式：</strong>百=(b-span+9)%10 · 十=(s-mid+5)%10 · 个=(g²+|b-g|)%10<br>
<strong>6杀全中：</strong>近100期 <strong>{{printf "%.1f" .Meta.All6Pct}}%</strong> · 全量≈53%（基线51.2%）</p>
<div class="warn">
⚠️ <strong>重要提示：</strong>彩票本质是随机游戏。近{{.Meta.BacktestN}}期3杀综合<strong>{{printf "%.1f" .Meta.AccAll}}%</strong>，6杀全中<strong>{{printf "%.1f" .Meta.All6Pct}}%</strong>。6杀全中理论上限≈66%。历史回测不代表未来表现，请理性参考，勿沉迷投注。
</div>
</div>

<div class="section-title"><span class="dot"></span>近{{.Meta.BacktestN}}期回测明细（6杀码）</div>
<div class="table-wrap">
<table>
<thead><tr><th>期号</th><th>日期</th><th>开奖</th><th>百杀</th><th>十杀</th><th>个杀</th><th>6杀</th></tr></thead>
<tbody>
{{range .Rows}}
<tr>
<td>{{.Issue}}</td><td>{{.Date}}</td><td>{{.Open}}</td>
<td>{{if and .HOK .H2OK}}<span class="ok">✅{{.HK}},{{.HK2}}</span>{{else}}<span class="bad">❌{{.HK}},{{.HK2}}</span>{{end}}</td>
<td>{{if and .TOK .T2OK}}<span class="ok">✅{{.TK}},{{.TK2}}</span>{{else}}<span class="bad">❌{{.TK}},{{.TK2}}</span>{{end}}</td>
<td>{{if and .OOK .O2OK}}<span class="ok">✅{{.OK}},{{.OK2}}</span>{{else}}<span class="bad">❌{{.OK}},{{.OK2}}</span>{{end}}</td>
<td>{{if .All6OK}}✅{{else}}❌{{end}}</td>
</tr>
{{end}}
</tbody>
</table>
</div>

<div class="disclaimer">
数据来源：福彩3D历史开奖数据 | 算法严格不含未来信息 | 仅供研究参考<br>
数据截止 {{.Meta.LatestDate}} · 共{{.Meta.Total}}期
</div>
</body>
</html>`

// GenerateHTML 渲染完整页面
func GenerateHTML(m backtest.Meta, pred backtest.Predict, rows []backtest.Row, b Banners, nextIssue string) (string, error) {
	t, err := template.New("page").Parse(tmplSrc)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	data := struct {
		Meta      backtest.Meta
		Pred      backtest.Predict
		Rows      []backtest.Row
		Banners   Banners
		NextIssue string
	}{
		Meta: m, Pred: pred, Rows: rows, Banners: b, NextIssue: nextIssue,
	}
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
