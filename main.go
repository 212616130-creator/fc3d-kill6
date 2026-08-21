// fc3d-kill6 — 福彩3D 百十个杀码预测 CLI（V9.3 六杀制，Go 实现）
//
// 数据流：抓数(灰鸟/17500 双源) → 期号校验 → 追加CSV → V9引擎回测
//
//	→ kill6监控+升级检测 → 生成 index.html
//
// 部署形态兼容：GitHub Pages(静态产物) / Docker(单二进制) / Serverless(引擎纯函数)。
package main

import (
	"flag"
	"fmt"
	"os"

	"fc3d-kill6/backtest"
	"fc3d-kill6/data"
	"fc3d-kill6/fetch"
	"fc3d-kill6/report"
)

func main() {
	csvPath := flag.String("csv", "fc3d-history.csv", "历史开奖 CSV 路径")
	htmlPath := flag.String("html", "index.html", "输出 HTML 路径")
	kill6Path := flag.String("kill6", "kill6_history.json", "kill6 监控历史 JSON 路径")
	flag.Parse()

	fmt.Println("=" + repeat("=", 30))
	fmt.Println("福彩3D 百十个杀码 · 云端更新 (Go)")
	fmt.Println("=" + repeat("=", 30))

	// Step 1: 获取最新开奖
	fmt.Println("📡 获取最新开奖...")
	newData, dataAlive := fetch.FetchLatest(*csvPath)
	if newData != nil {
		added, err := data.AppendCSV(*csvPath, data.Draw{
			Issue: newData.Issue, Date: newData.Date,
			B: newData.B, S: newData.S, G: newData.G,
		})
		if err != nil {
			fmt.Printf("  ❌ 追加失败: %v\n", err)
		} else if added == 1 {
			fmt.Printf("  ✅ 已追加第%s期 (%s) %d%d%d\n", newData.Issue, newData.Date, newData.B, newData.S, newData.G)
		} else {
			fmt.Printf("  ℹ️ 第%s期已存在, 无需追加\n", newData.Issue)
		}
	} else if !dataAlive {
		fmt.Println("\n🚨🚨🚨 所有数据源均失败! 页面将显示旧数据, 请检查数据源 🚨🚨🚨")
	} else {
		fmt.Println("  ℹ️ 数据源正常但无新一期(开奖前运行), 继续用现有数据")
	}

	// Step 2: 加载数据
	draws, err := data.LoadCSV(*csvPath)
	if err != nil || len(draws) < 100 {
		fmt.Printf("❌ 数据不足或读取失败: %v (%d期)\n", err, len(draws))
		os.Exit(1)
	}

	// Step 3: 回测
	bt := backtest.RunAll(draws)
	m := bt.Meta
	fmt.Printf("\n📊 回测 %d期: 百%.1f%% 十%.1f%% 个%.1f%% 综合%.1f%%\n",
		m.BacktestN, m.AccH, m.AccT, m.AccO, m.AccAll)
	fmt.Printf("   kill2: 百%.1f%% 十%.1f%% 个%.1f%%\n", m.AccH2, m.AccT2, m.AccO2)
	fmt.Printf("   6杀全中: %d/%d = %.1f%%\n", m.All6, m.BacktestN, m.All6Pct)
	fmt.Printf("   近100期综合(按期): %d/%d = %.1f%%\n", m.PeriodCorrect100, m.PeriodN100, m.AccPeriod100)
	fmt.Printf("   近100期6杀全中: %d/%d = %.1f%%\n", m.Period6Correct100, m.PeriodN100, m.Period6Pct100)

	// Step 3.5: 多窗口回测（本地诊断：窗口独立重置状态机，含个位自适应）
	fmt.Printf("\n📈 多窗口回测 (3杀综合/6杀全中/最大连错):\n")
	for w, ws := range backtest.MultiWindow(draws, []int{100, 200, 300, 500}) {
		fmt.Printf("   %s: 综合%.1f%% | 6杀全中%.1f%% | 最大连错%d期\n",
			w, ws.Overall, ws.All6Pct, ws.MaxConsecutive)
	}

	// Step 4: 升级触发检测
	hist, _ := backtest.RecordKill6(*kill6Path, m.Period6Pct100, m.LatestIssue, m.LatestDate)
	triggered, reasons, monthDrop := backtest.CheckUpgradeTrigger(m.Period6Pct100, hist)
	if triggered {
		fmt.Println("\n🚨🚨🚨 升级触发！建议重新穷举6个算法 🚨🚨🚨")
		for _, r := range reasons {
			fmt.Printf("   ⚠️ %s\n", r)
		}
	} else {
		fmt.Printf("   ✅ 升级触发器: 正常 (单月%+.1fpp, 阈值跌破70%%/月降8pp)\n", monthDrop)
	}

	// Step 5: 生成 HTML
	nextIssue := fetch.NextIssueCalc(m.LatestIssue, m.LatestDate, nextIssueHint(newData))
	banners := report.Banners{DataUpgrade: triggered, UpgradeReasons: reasons, DataFailed: !dataAlive}
	html, err := report.GenerateHTML(m, bt.Pred, bt.Rows, banners, nextIssue)
	if err != nil {
		fmt.Printf("❌ HTML 生成失败: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*htmlPath, []byte(html), 0o644); err != nil {
		fmt.Printf("❌ 写入 HTML 失败: %v\n", err)
		os.Exit(1)
	}

	p := bt.Pred
	fmt.Printf("\n🔮 下一期: %s | 百杀%d,%d 十杀%d,%d 个杀%d,%d\n",
		nextIssue, p.H, p.H2, p.T, p.T2, p.O, p.O2)
	fmt.Printf("✅ HTML已生成 (%s, %d字节)\n", *htmlPath, len(html))
}

// nextIssueHint 透传数据源 next_code（跨年安全）
func nextIssueHint(lt *fetch.Latest) string {
	if lt != nil {
		return lt.NextIssue
	}
	return ""
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
