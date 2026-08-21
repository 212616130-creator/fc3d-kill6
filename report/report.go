// Package report 生成预测页面 HTML（深色数据大屏响应式，内联样式零外部依赖）。
package report

import (
	"bytes"
	"fmt"
	"math"
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

// view 模板视图（Rows 已按最新在前排序）
type view struct {
	Meta      backtest.Meta
	Pred      backtest.Predict
	Rows      []backtest.Row
	MCards    []backtest.Row
	Banners   Banners
	NextIssue string
	Ring      string
	Pct6Beat  float64
}

// ringSVG 生成 6 杀全中率环形进度（path 圆弧，规避 transform 解析问题）
func ringSVG(pct float64) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	const size, r, sw = 96.0, 38.0, 9.0
	cx, cy := size/2, size/2
	theta := pct * 3.6 * math.Pi / 180
	ex := cx + r*math.Cos(theta)
	ey := cy + r*math.Sin(theta)
	large := 0
	if pct > 50 {
		large = 1
	}
	inner := int(math.Round(pct))
	return fmt.Sprintf(`<svg class="ring" width="96" height="96" viewBox="0 0 96 96" fill="none" xmlns="http://www.w3.org/2000/svg"><circle cx="48" cy="48" r="38" stroke="#1E293B" stroke-width="9"/><path d="M86 48A38 38 0 %d 1 %.1f %.1f" stroke="#34D399" stroke-width="9" stroke-linecap="round"/><text x="48" y="55" text-anchor="middle" font-size="21" font-weight="700" fill="#E8EEF9" font-family="SF Mono,monospace">%d%%</text></svg>`, large, ex, ey, inner)
}

// reverseRows 返回 rows 的反转副本（最新在前）
func reverseRows(rows []backtest.Row) []backtest.Row {
	out := make([]backtest.Row, len(rows))
	for i, r := range rows {
		out[len(rows)-1-i] = r
	}
	return out
}

const tmplSrc = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<meta name="color-scheme" content="dark">
<title>福彩3D 百十个杀码预测 · V9.3 六杀制</title>
<style>
:root{--bg1:#0A0E17;--bg2:#0D1424;--surface:#121A2B;--surface2:#0D1424;--border:#1E293B;--border-soft:rgba(148,163,184,.12);--text1:#E8EEF9;--text2:#94A3B8;--text3:#64748B;--cyan:#22D3EE;--cyan-soft:#67E8F9;--violet:#A78BFA;--violet-soft:#C4B5FD;--amber:#FBBF24;--amber-soft:#FCD34D;--blue:#60A5FA;--blue-soft:#93C5FD;--green:#34D399;--green-soft:#6EE7B7;--red:#F87171;--radius:16px;--font-cn:"PingFang SC","Hiragino Sans GB","Microsoft YaHei",-apple-system,sans-serif;--font-num:"SF Mono",ui-monospace,"JetBrains Mono",Menlo,monospace}
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:var(--font-cn);background:linear-gradient(180deg,var(--bg1),var(--bg2));color:var(--text1);min-height:100vh;-webkit-font-smoothing:antialiased}
.bg-grid{position:fixed;inset:0;z-index:0;pointer-events:none;background-image:url("data:image/svg+xml,%3Csvg width='48' height='48' viewBox='0 0 48 48' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M48 0H0V48' fill='none' stroke='%2364748B' stroke-opacity='0.07'/%3E%3C/svg%3E");mask-image:linear-gradient(180deg,#000 0%,#000 55%,transparent 100%);-webkit-mask-image:linear-gradient(180deg,#000 0%,#000 55%,transparent 100%)}
.bg-glow{position:fixed;border-radius:50%;filter:blur(90px);z-index:0;pointer-events:none}
.g1{width:560px;height:560px;top:-210px;left:-170px;background:radial-gradient(circle,rgba(34,64,153,.5),transparent 70%)}
.g2{width:440px;height:440px;top:-170px;right:-120px;background:radial-gradient(circle,rgba(13,166,237,.30),transparent 70%)}
.page{position:relative;z-index:1;max-width:1440px;margin:0 auto;padding:0 64px}
header{display:flex;align-items:center;justify-content:space-between;height:80px;border-bottom:1px solid var(--border)}
.brand{display:flex;align-items:center;gap:12px}
.brand-icon{width:34px;height:34px;flex:none}
.brand-name{font:800 20px var(--font-num);letter-spacing:1px}
.brand-sub{font-size:11px;color:var(--text3);margin-top:2px}
.hdr-right{display:flex;align-items:center;gap:20px}
.hdr-meta{font-size:12px;color:var(--text2)}
.pill{display:inline-flex;align-items:center;gap:8px;padding:8px 14px;border-radius:999px;font-size:12px;font-weight:600}
.pill.green{background:rgba(52,211,153,.12);border:1px solid rgba(52,211,153,.5);color:var(--green-soft)}
.pill .dot{width:8px;height:8px;border-radius:50%;background:var(--green)}
.hero{padding:56px 0 48px;display:flex;flex-direction:column;gap:28px}
.hero-top{display:flex;justify-content:space-between;align-items:flex-end;gap:24px}
.hero-left{display:flex;flex-direction:column;gap:10px}
.kicker{display:flex;align-items:center;gap:12px}
.tag{display:inline-flex;align-items:center;padding:4px 10px;border-radius:6px;background:var(--surface2);border:1px solid rgba(34,211,238,.4);font-size:11px;font-weight:600;color:var(--cyan-soft)}
.kicker-line{font-size:12px;color:var(--text3)}
h1{font-size:44px;font-weight:700;letter-spacing:.5px}
.hero-sub{font-size:15px;color:var(--text2);max-width:640px;line-height:1.7}
.issue-badge{display:flex;flex-direction:column;align-items:center;gap:6px;padding:16px 24px;border-radius:14px;background:var(--surface);border:1px solid rgba(34,211,238,.35);box-shadow:0 0 24px rgba(13,166,237,.15)}
.issue-label{font-size:12px;color:var(--text2)}
.issue-value{font:800 30px var(--font-num);color:var(--cyan-soft)}
.pred-row{display:grid;grid-template-columns:repeat(3,1fr);gap:20px}
.pred-card{display:flex;flex-direction:column;gap:14px;padding:22px 26px;border-radius:var(--radius);border:1px solid var(--border-soft);box-shadow:0 16px 32px -8px rgba(0,0,0,.4);position:relative;overflow:hidden}
.pred-card::before{content:"";position:absolute;inset:0;opacity:.9;z-index:0}
.pred-card>*{position:relative;z-index:1}
.pred-card.cyan{background:linear-gradient(180deg,rgba(23,71,97,.9),rgba(13,20,41,1));border-color:rgba(34,211,238,.35)}
.pred-card.violet{background:linear-gradient(180deg,rgba(66,43,112,.9),rgba(13,20,41,1));border-color:rgba(167,139,250,.35)}
.pred-card.amber{background:linear-gradient(180deg,rgba(107,71,23,.9),rgba(13,20,41,1));border-color:rgba(251,191,36,.35)}
.pred-head{display:flex;justify-content:space-between;align-items:center}
.pred-label{font-size:13px;font-weight:600;color:var(--cyan-soft)}
.pred-card.violet .pred-label{color:var(--violet-soft)}
.pred-card.amber .pred-label{color:var(--amber-soft)}
.mini-tag{padding:4px 8px;border-radius:5px;font-size:10px;font-weight:600;background:rgba(23,71,97,.9);border:1px solid rgba(34,211,238,.4);color:var(--cyan-soft)}
.pred-card.violet .mini-tag{background:rgba(66,43,112,.9);border-color:rgba(167,139,250,.4);color:var(--violet-soft)}
.pred-card.amber .mini-tag{background:rgba(107,71,23,.9);border-color:rgba(251,191,36,.4);color:var(--amber-soft)}
.pred-num{font:800 54px/1.1 var(--font-num);letter-spacing:2px;text-align:center;color:var(--cyan)}
.pred-card.violet .pred-num{color:var(--violet)}
.pred-card.amber .pred-num{color:var(--amber)}
.pred-foot{font-size:11px;color:var(--text3);text-align:center}
.section{padding:0 0 40px}
.section-head{display:flex;justify-content:space-between;align-items:flex-end;margin-bottom:16px}
.section-title{font-size:22px;font-weight:700}
.section-meta{font-size:12px;color:var(--text3)}
.bento{display:grid;grid-template-columns:340px 1fr;gap:20px}
.big-card{display:flex;flex-direction:column;gap:14px;padding:24px;border-radius:var(--radius);background:var(--surface);border:1px solid rgba(52,211,153,.4);box-shadow:0 14px 30px -8px rgba(0,0,0,.35)}
.big-top{display:flex;justify-content:space-between;align-items:center}
.big-label{font-size:14px;font-weight:600;color:var(--green-soft)}
.ring{flex:none;display:block}
.big-value{font:800 46px var(--font-num);color:var(--green)}
.big-sub{font-size:12px;color:var(--text3)}
.bar{height:6px;border-radius:3px;background:var(--border);overflow:hidden}
.bar-fill{height:100%;border-radius:3px;background:linear-gradient(90deg,var(--green),var(--cyan))}
.grid-6{display:grid;grid-template-columns:repeat(3,1fr);gap:16px}
.stat-card{display:flex;flex-direction:column;justify-content:space-between;gap:8px;padding:18px 20px;border-radius:14px;background:var(--surface);border:1px solid var(--border-soft)}
.stat-top{display:flex;justify-content:space-between;align-items:center}
.stat-label{font-size:12px;color:var(--text2)}
.stat-badge{font-size:10px;font-weight:600}
.stat-value{font:800 30px var(--font-num)}
.engine-row{display:grid;grid-template-columns:1fr 1fr 1fr;gap:20px}
.eng-card{display:flex;flex-direction:column;gap:14px;padding:24px 26px;border-radius:var(--radius);background:var(--surface);border:1px solid var(--border-soft)}
.eng-head{display:flex;align-items:center;gap:10px}
.eng-num{width:26px;height:26px;border-radius:8px;flex:none;display:flex;align-items:center;justify-content:center;font:700 14px var(--font-num);background:var(--surface2);border:1px solid rgba(34,211,238,.4);color:var(--cyan-soft)}
.eng-card.violet .eng-num{border-color:rgba(167,139,250,.4);color:var(--violet-soft)}
.eng-card.amber .eng-num{border-color:rgba(251,191,36,.4);color:var(--amber-soft)}
.eng-title{font-size:15px;font-weight:600}
.eng-desc{font-size:13px;line-height:1.8;color:var(--text2)}
.formula-row{display:flex;align-items:center;gap:14px}
.formula-label{width:40px;font-size:12px;color:var(--text3);flex:none}
.formula{font:600 15px var(--font-num);color:var(--violet)}
.cmp-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.cmp-cell{display:flex;flex-direction:column;gap:4px;padding:12px 14px;border-radius:10px;background:var(--surface2);border:1px solid var(--border-soft)}
.cmp-value{font:800 20px var(--font-num)}
.cmp-label{font-size:10px;color:var(--text3)}
.warn{display:flex;align-items:center;gap:12px;padding:18px 22px;margin-top:40px;border-radius:12px;background:rgba(251,146,60,.08);border:1px solid rgba(251,146,60,.4);font-size:13px;color:#FDBA74;line-height:1.7}
.warn-icon{width:22px;height:22px;border-radius:50%;flex:none;display:flex;align-items:center;justify-content:center;background:rgba(251,191,36,.14);border:1.5px solid rgba(251,191,36,.6);font:700 13px var(--font-num);color:var(--amber)}
.upgrade-alert{display:flex;flex-direction:column;gap:6px;padding:16px 22px;border-radius:12px;margin-bottom:24px;background:linear-gradient(135deg,rgba(185,28,28,.85),rgba(220,38,38,.8));border:1px solid rgba(248,113,113,.5);font-size:13px;line-height:1.7;color:#FECACA}
.upgrade-alert .ua-title{font-size:15px;font-weight:800;color:#FEE2E2}
.upgrade-alert .ua-sub{font-size:11px;opacity:.85}
.data-alert{display:flex;flex-direction:column;gap:6px;padding:16px 22px;border-radius:12px;margin-bottom:24px;background:linear-gradient(135deg,rgba(230,81,0,.85),rgba(245,124,0,.8));border:1px solid rgba(251,146,60,.5);font-size:13px;line-height:1.7;color:#FED7AA}
.data-alert .da-title{font-size:15px;font-weight:800;color:#FFEDD5}
.table-wrap{border-radius:14px;border:1px solid var(--border);overflow:hidden;background:rgba(18,26,43,.6)}
table{width:100%;border-collapse:collapse;font-size:13px}
thead th{background:var(--surface2);padding:13px 20px;font-size:11px;font-weight:600;color:var(--text3);text-align:center}
tbody td{padding:14px 20px;text-align:center;border-top:1px solid rgba(30,41,59,.6);color:var(--text2)}
tbody tr:nth-child(odd){background:rgba(13,20,41,.5)}
.issue-no{font:400 13px var(--font-num);color:var(--text2)}
.date{font-size:12px;color:var(--text3)}
.win-num{font:700 15px var(--font-num);color:var(--text1)}
.kill-code{font:600 14px var(--font-num)}
.kill-code.ok{color:var(--green)}
.kill-code.bad{color:var(--red)}
.result{display:inline-block;padding:3px 10px;border-radius:999px;font-size:12px;font-weight:700}
.result.ok{background:rgba(52,211,153,.12);border:1px solid rgba(52,211,153,.5);color:var(--green)}
.result.bad{background:rgba(248,113,113,.12);border:1px solid rgba(248,113,113,.5);color:var(--red)}
.m-detail{display:none;flex-direction:column;gap:12px}
.m-card{display:flex;align-items:center;gap:12px;padding:14px 16px;border-radius:12px;background:var(--surface);border:1px solid var(--border)}
.m-card .left{display:flex;flex-direction:column;gap:3px}
.m-card .issue-no{font-size:13px}
.m-card .date{font-size:10px}
.m-card .win{flex:1;text-align:center;font:700 18px var(--font-num);color:var(--text1)}
.v-cyan{color:var(--cyan)}.v-blue{color:var(--blue)}.v-violet{color:var(--violet)}.v-green{color:var(--green)}.v-amber{color:var(--amber)}.v-text2{color:var(--text2)}
.v-cyan-soft{color:var(--cyan-soft)}.v-blue-soft{color:var(--blue-soft)}.v-violet-soft{color:var(--violet-soft)}.v-green-soft{color:var(--green-soft)}
footer{display:flex;flex-direction:column;align-items:center;gap:14px;padding:56px 0 44px;margin-top:16px;border-top:1px solid var(--border)}
.foot-text{font-size:12px;color:var(--text3)}
.foot-meta{font-size:11px;color:#475569}
.foot-brand{font-size:10px;color:#334155}
@media (max-width:1023px){
.page{padding:0 20px}
header{height:60px}
.hdr-meta{display:none}
.brand-name{font-size:16px}
.brand-sub{display:none}
.brand-icon{width:24px;height:24px}
.pill{padding:6px 10px;font-size:10px}
.hero{padding:28px 0 24px;gap:16px}
.hero-top{flex-direction:column;align-items:stretch;gap:16px}
h1{font-size:26px}
.hero-sub{font-size:13px;line-height:1.6;max-width:none}
.issue-badge{align-items:center;padding:10px 14px;flex-direction:row;justify-content:center;gap:8px;border-radius:999px}
.issue-value{font-size:16px}
.pred-row{grid-template-columns:1fr;gap:12px;display:none}
.pred-grid{display:grid;grid-template-columns:1fr;gap:16px}
.pred-main{display:flex;flex-direction:column;gap:14px;padding:20px 18px;border-radius:16px;background:linear-gradient(180deg,rgba(23,71,97,.9),rgba(13,20,41,1));border:1px solid rgba(34,211,238,.35);box-shadow:0 14px 28px -8px rgba(0,0,0,.4)}
.pos-row{display:grid;grid-template-columns:repeat(3,1fr);gap:10px}
.pos-card{display:flex;flex-direction:column;align-items:center;gap:6px;padding:12px 10px;border-radius:12px;background:rgba(18,26,43,.8);border:1px solid var(--border-soft)}
.pos-label{font-size:10px;color:var(--text2)}
.pos-num{font:800 26px var(--font-num)}
.section{padding-bottom:24px}
.section-title{font-size:18px}
.bento{grid-template-columns:1fr}
.big-card{flex-direction:row;align-items:center;padding:16px;gap:16px}
.big-value{font-size:32px}
.big-top{flex-direction:column;align-items:flex-start;gap:2px}
.ring{width:72px;height:72px}
.grid-6{grid-template-columns:repeat(2,1fr);gap:10px}
.stat-card{padding:12px 14px;border-radius:12px}
.stat-value{font-size:22px}
.engine-row{grid-template-columns:1fr;gap:12px}
.eng-card{padding:16px}
.eng-desc{font-size:11px;line-height:1.7}
.formula{font-size:13px}
.cmp-cell{padding:10px 12px}
.warn{margin-top:24px;padding:14px 16px;font-size:11px;line-height:1.6}
.table-wrap{display:none}
.m-detail{display:flex;max-height:62vh;overflow-y:auto;-webkit-overflow-scrolling:touch;overscroll-behavior:contain;padding:2px 8px 2px 2px;border-radius:14px;border:1px solid var(--border);background:rgba(18,26,43,.4);scrollbar-width:thin;scrollbar-color:rgba(148,163,184,.35) transparent}
.m-detail::-webkit-scrollbar{width:6px}
.m-detail::-webkit-scrollbar-track{background:transparent}
.m-detail::-webkit-scrollbar-thumb{background:rgba(148,163,184,.35);border-radius:3px}
.m-detail::-webkit-scrollbar-thumb:hover{background:rgba(148,163,184,.6)}
footer{padding:22px 0 10px;gap:8px}
.foot-text{font-size:10px}
}
@media (min-width:1024px){.pred-grid{display:none}}
</style>
</head>
<body>
<div class="bg-glow g1"></div>
<div class="bg-glow g2"></div>
<div class="bg-grid"></div>

<div class="page">
  <header>
    <div class="brand">
      <svg class="brand-icon" viewBox="0 0 34 34" fill="none" xmlns="http://www.w3.org/2000/svg"><rect x="1" y="1" width="32" height="32" rx="9" stroke="#22D3EE" stroke-width="2"/><circle cx="17" cy="17" r="7" stroke="#22D3EE" stroke-width="2"/><circle cx="17" cy="17" r="2.5" fill="#22D3EE"/><path d="M17 2v6M17 26v6M2 17h6M26 17h6" stroke="#22D3EE" stroke-width="2" stroke-linecap="round"/></svg>
      <div>
        <div class="brand-name">3D KILL6</div>
        <div class="brand-sub">福彩3D 六杀制预测引擎</div>
      </div>
    </div>
    <div class="hdr-right">
      <span class="hdr-meta">数据截止 {{.Meta.LatestDate}} · 共{{.Meta.Total}}期</span>
      <span class="pill green"><span class="dot"></span>近100期 6杀全中 {{printf "%.1f" .Meta.Period6Pct100}}%</span>
    </div>
  </header>

  <main>
{{if .Banners.DataFailed}}
<div class="data-alert"><div class="da-title">数据源异常</div>所有数据源获取失败，页面为最后一次成功数据，请检查数据源（灰鸟 / 17500.cn）。</div>
{{end}}
{{if .Banners.DataUpgrade}}
<div class="upgrade-alert"><div class="ua-title">算法升级触发</div>6杀全中率已触及升级阈值，建议重新穷举 6 个算法：<br>{{range .Banners.UpgradeReasons}}• {{.}}<br>{{end}}<span class="ua-sub">触发条件：滚动100期跌破 70% 或 单月下滑超 8pp</span></div>
{{end}}

    <section class="hero">
      <div class="hero-top">
        <div class="hero-left">
          <div class="kicker">
            <span class="tag">V9.3 六杀制</span>
            <span class="kicker-line">双引擎独立杀码 · kill1 决策树 + kill2 算术公式</span>
          </div>
          <h1>福彩3D 百十个杀码预测</h1>
          <p class="hero-sub">每期输出百 / 十 / 个三位置双杀码。近 100 期 6 杀全中率 {{printf "%.1f" .Meta.Period6Pct100}}%，显著超越随机基线 51.2%。</p>
        </div>
        <div class="issue-badge">
          <span class="issue-label">下一期预测 · 第 {{.NextIssue}} 期</span>
          <span class="issue-value">6 杀码</span>
        </div>
      </div>

      <div class="pred-row">
        <div class="pred-card cyan">
          <div class="pred-head"><span class="pred-label">百位 · 双杀码</span><span class="mini-tag">kill1 + kill2</span></div>
          <div class="pred-num">{{.Pred.H}}, {{.Pred.H2}}</div>
          <div class="pred-foot">kill1 决策树 + 算术公式 · 近{{.Meta.BacktestN}}期命中 {{printf "%.1f" .Meta.AccH}}%</div>
        </div>
        <div class="pred-card violet">
          <div class="pred-head"><span class="pred-label">十位 · 双杀码</span><span class="mini-tag">kill1 + kill2</span></div>
          <div class="pred-num">{{.Pred.T}}, {{.Pred.T2}}</div>
          <div class="pred-foot">kill1 决策树 + 算术公式 · 近{{.Meta.BacktestN}}期命中 {{printf "%.1f" .Meta.AccT}}%</div>
        </div>
        <div class="pred-card amber">
          <div class="pred-head"><span class="pred-label">个位 · 双杀码</span><span class="mini-tag">kill1 + kill2</span></div>
          <div class="pred-num">{{.Pred.O}}, {{.Pred.O2}}</div>
          <div class="pred-foot">kill1 决策树 + 算术公式 · 近{{.Meta.BacktestN}}期命中 {{printf "%.1f" .Meta.AccO}}%</div>
        </div>
      </div>

      <div class="pred-grid">
        <div class="pred-main">
          <div class="pred-head"><span class="pred-label">下一期预测 · 第 {{.NextIssue}} 期</span><span class="mini-tag">kill1 + kill2</span></div>
          <div class="pos-row">
            <div class="pos-card"><span class="pos-label">百位</span><span class="pos-num v-cyan">{{.Pred.H}}, {{.Pred.H2}}</span></div>
            <div class="pos-card"><span class="pos-label">十位</span><span class="pos-num v-violet">{{.Pred.T}}, {{.Pred.T2}}</span></div>
            <div class="pos-card"><span class="pos-label">个位</span><span class="pos-num v-amber">{{.Pred.O}}, {{.Pred.O2}}</span></div>
          </div>
        </div>
      </div>
    </section>

    <section class="section">
      <div class="section-head">
        <h2 class="section-title">近 {{.Meta.BacktestN}} 期回测</h2>
        <span class="section-meta">3杀综合 {{printf "%.1f" .Meta.AccAll}}% · 6杀全中 {{printf "%.1f" .Meta.Period6Pct100}}% · 随机基线 51.2%</span>
      </div>
      <div class="bento">
        <div class="big-card">
          <div class="big-top">
            <span class="big-label">6 杀全中率</span>
            {{.Ring}}
          </div>
          <div class="big-value">{{printf "%.1f" .Meta.Period6Pct100}}%</div>
          <div class="big-sub">近 100 期 · 超越随机基线 +{{printf "%.1f" .Pct6Beat}}pp</div>
          <div class="bar"><div class="bar-fill" style="width:{{printf "%.0f" .Meta.Period6Pct100}}%"></div></div>
        </div>
        <div class="grid-6">
          <div class="stat-card"><div class="stat-top"><span class="stat-label">百位 · 3杀</span><span class="stat-badge v-cyan-soft">错 {{.Meta.ErrH}}</span></div><div class="stat-value v-cyan">{{printf "%.1f" .Meta.AccH}}%</div></div>
          <div class="stat-card"><div class="stat-top"><span class="stat-label">十位 · 3杀</span><span class="stat-badge v-blue-soft">错 {{.Meta.ErrT}}</span></div><div class="stat-value v-blue">{{printf "%.1f" .Meta.AccT}}%</div></div>
          <div class="stat-card"><div class="stat-top"><span class="stat-label">个位 · 3杀</span><span class="stat-badge v-violet-soft">错 {{.Meta.ErrO}}</span></div><div class="stat-value v-violet">{{printf "%.1f" .Meta.AccO}}%</div></div>
          <div class="stat-card"><div class="stat-top"><span class="stat-label">百位 · kill2</span><span class="stat-badge v-blue-soft">错 {{.Meta.ErrH2}}</span></div><div class="stat-value v-blue">{{printf "%.1f" .Meta.AccH2}}%</div></div>
          <div class="stat-card"><div class="stat-top"><span class="stat-label">十位 · kill2</span><span class="stat-badge v-violet-soft">错 {{.Meta.ErrT2}}</span></div><div class="stat-value v-violet">{{printf "%.1f" .Meta.AccT2}}%</div></div>
          <div class="stat-card"><div class="stat-top"><span class="stat-label">个位 · kill2</span><span class="stat-badge v-green-soft">错 {{.Meta.ErrO2}}</span></div><div class="stat-value v-green">{{printf "%.1f" .Meta.AccO2}}%</div></div>
        </div>
      </div>
    </section>

    <section class="section">
      <div class="section-head">
        <h2 class="section-title">V9 六杀引擎</h2>
        <span class="section-meta">kill1 决策树 + kill2 算术公式 · 双引擎独立运算、互为校验</span>
      </div>
      <div class="engine-row">
        <div class="eng-card">
          <div class="eng-head"><span class="eng-num">1</span><span class="eng-title">kill1 · V8 条件决策树</span></div>
          <div class="eng-desc">百位：10 条件决策树，逐期推导最优杀码。<br>十位：V8a 公式法，历经两处漏洞修复。<br>个位：12 条件决策树 + 自适应备份（5 期失败窗口自动切换）。</div>
        </div>
        <div class="eng-card violet">
          <div class="eng-head"><span class="eng-num">2</span><span class="eng-title">kill2 · 独立算术公式</span></div>
          <div class="formula-row"><span class="formula-label">百位</span><span class="formula">(b − span + 9) mod 10</span></div>
          <div class="formula-row"><span class="formula-label">十位</span><span class="formula">(s − mid + 5) mod 10</span></div>
          <div class="formula-row"><span class="formula-label">个位</span><span class="formula">(g² + |b − g|) mod 10</span></div>
        </div>
        <div class="eng-card amber">
          <div class="eng-head"><span class="eng-num">3</span><span class="eng-title">6杀全中 · 数据对比</span></div>
          <div class="cmp-grid">
            <div class="cmp-cell"><span class="cmp-value v-green">{{printf "%.1f" .Meta.Period6Pct100}}%</span><span class="cmp-label">近 100 期</span></div>
            <div class="cmp-cell"><span class="cmp-value v-blue">≈53%</span><span class="cmp-label">全量 8730 期收敛</span></div>
            <div class="cmp-cell"><span class="cmp-value v-violet">≈66%</span><span class="cmp-label">6杀全中理论上限</span></div>
            <div class="cmp-cell"><span class="cmp-value v-text2">51.2%</span><span class="cmp-label">随机基线</span></div>
          </div>
        </div>
      </div>
      <div class="warn">
        <span class="warn-icon">!</span>
        <span>理性参考提示：彩票本质是随机游戏，杀码结果仅基于历史数据统计，不构成任何投注建议。请理性娱乐。</span>
      </div>
    </section>

    <section class="section">
      <div class="section-head">
        <h2 class="section-title">近 {{.Meta.BacktestN}} 期回测明细</h2>
        <span class="section-meta">完整 {{.Meta.BacktestN}} 期滚动 · 移动端卡片式</span>
      </div>
      <div class="table-wrap">
        <table>
          <thead><tr><th>期号</th><th>日期</th><th>开奖</th><th>百位杀码</th><th>十位杀码</th><th>个位杀码</th><th>6杀结果</th></tr></thead>
          <tbody>
{{range .Rows}}
<tr>
<td class="issue-no">{{.Issue}}</td><td class="date">{{.Date}}</td><td class="win-num">{{.Open}}</td>
<td class="kill-code {{if and .HOK .H2OK}}ok{{else}}bad{{end}}">{{.HK}}, {{.HK2}}</td>
<td class="kill-code {{if and .TOK .T2OK}}ok{{else}}bad{{end}}">{{.TK}}, {{.TK2}}</td>
<td class="kill-code {{if and .OOK .O2OK}}ok{{else}}bad{{end}}">{{.OK}}, {{.OK2}}</td>
<td><span class="result {{if .All6OK}}ok{{else}}bad{{end}}">{{if .All6OK}}全中{{else}}未中{{end}}</span></td>
</tr>
{{end}}
          </tbody>
        </table>
      </div>
      <div class="m-detail">
{{range .MCards}}
<div class="m-card"><div class="left"><span class="issue-no">{{.Issue}}</span><span class="date">{{.Date}}</span></div><span class="win">{{.Open}}</span><span class="result {{if .All6OK}}ok{{else}}bad{{end}}">{{if .All6OK}}全中{{else}}未中{{end}}</span></div>
{{end}}
      </div>
    </section>
  </main>

  <footer>
    <span class="foot-text">数据来源：福彩3D 历史开奖数据 · 算法严格不含未来信息 · 仅供研究参考</span>
    <span class="foot-meta">数据截止 {{.Meta.LatestDate}} · 共{{.Meta.Total}}期 · 每日开奖后自动更新</span>
    <span class="foot-brand">3D KILL6 · V9.3 六杀制预测引擎</span>
  </footer>
</div>
</body>
</html>`

// GenerateHTML 渲染完整页面（Rows 自动转为最新在前）
func GenerateHTML(m backtest.Meta, pred backtest.Predict, rows []backtest.Row, b Banners, nextIssue string) (string, error) {
	t, err := template.New("page").Parse(tmplSrc)
	if err != nil {
		return "", err
	}
	rev := reverseRows(rows)
	mc := rev // 移动端卡片 = 全部 100 期，与 section-meta "完整 100 期滚动" 承诺一致
	data := view{
		Meta: m, Pred: pred,
		Rows: rev, MCards: mc,
		Banners: b, NextIssue: nextIssue,
		Ring:     ringSVG(m.Period6Pct100),
		Pct6Beat: m.Period6Pct100 - 51.2,
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
