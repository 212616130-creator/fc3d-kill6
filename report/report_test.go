package report

import (
	"strings"
	"testing"

	"fc3d-kill6/backtest"
	"fc3d-kill6/monitor"
)

func TestGenerateHTMLCore(t *testing.T) {
	m := backtest.Meta{Total: 8730, LatestIssue: "2026222", LatestDate: "2026-08-20", BacktestN: 100, Period6Pct100: 81.0}
	pred := backtest.Predict{H: 2, T: 3, O: 0, H2: 4, T2: 0, O2: 3}
	rows := []backtest.Row{{Issue: "2026222", Date: "2026-08-20", Open: "380", HK: 3, TK: 4, OK: 2}}
	wf := []backtest.WFWindow{{Label: "100期", N: 100, All6: 81, All6Pct: 81.0, BeatPP: 29.8, Z: 6.0, PVal: 0.0001}}
	hist := []monitor.Entry{{Issue: "2026222", Date: "2026-08-20", Pct: 81.0}}

	html, err := GenerateHTML(m, pred, rows, Banners{}, "2026223", hist, wf)
	if err != nil {
		t.Fatalf("GenerateHTML: %v", err)
	}
	for _, want := range []string{
		"福彩3D 百十个杀码预测",
		"2026222", "380",
		"wu529778790/fc3d-kill6", // GitHub 图标
		"wx-auth-sdk",            // 认证接入
		"Walk-forward 滚动验证",      // walk-forward 摘要
		"polyline",               // 趋势图折线
	} {
		if !strings.Contains(html, want) {
			t.Errorf("HTML 缺少关键内容: %q", want)
		}
	}
}

func TestTrendSVG(t *testing.T) {
	entries := []monitor.Entry{
		{Issue: "1", Pct: 60}, {Issue: "2", Pct: 75}, {Issue: "3", Pct: 81},
	}
	svg := trendSVG(entries)
	if !strings.Contains(svg, "51.2") || !strings.Contains(svg, "70%") {
		t.Errorf("趋势图缺少基线/预警线: %s", svg)
	}
	if !strings.Contains(svg, "81.0%") {
		t.Errorf("趋势图缺少最新值标签: %s", svg)
	}
	if trendSVG(nil) != "" {
		t.Errorf("空数据应返回空串")
	}
}

func TestWFNote(t *testing.T) {
	if wfNote(nil) != "" {
		t.Errorf("空 walk-forward 应返回空串")
	}
	note := wfNote([]backtest.WFWindow{{Label: "100期", N: 100, All6: 81, All6Pct: 81.0, BeatPP: 29.8, Z: 6.0, PVal: 0.00001}})
	if !strings.Contains(note, "p=<0.001") {
		t.Errorf("p<0.001 应显示为 <0.001, got: %s", note)
	}
}
