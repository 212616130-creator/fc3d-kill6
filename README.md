# fc3d-kill6

福彩3D 百十个杀码预测（V9.3 六杀制）。引擎经 golden 基准测试锁定 8729 期逐期一致 + 1000 组输入穷举一致，Go 单二进制实现。

每位置输出 2 个杀码（kill1 条件决策树 + kill2 独立公式），6 杀全中为命中。输入仅上期开奖号 (b,s,g)，纯函数零依赖。

在线演示：<https://wu529778790.github.io/fc3d-kill6/>

## 使用

```bash
go build -o fc3d-kill6 .
./fc3d-kill6              # 抓数→追加CSV→回测→生成index.html→升级检测
go test ./...             # 引擎基准一致性测试
```

## Docker

```bash
docker build -t fc3d-kill6 .
docker run --rm -v $(pwd):/data fc3d-kill6 \
  -csv /data/fc3d-history.csv -html /data/index.html -kill6 /data/kill6_history.json
```

## 数据源

灰鸟 API（主，带 next_code 跨年）+ 17500.cn（备份，官方级全量 TXT），双源降级 + 期号合理性校验。

## 自动更新

GitHub Actions 每日三重 cron（北京 22:00/23:30/01:00）→ 自动提交 → GitHub Pages 部署。升级触发器：100 期 6 杀率 <70% 或单月下滑 ≥8pp 时页面红条告警。

> 彩票本质是独立随机过程。全量 8700+ 期 6 杀全中收敛于 53.19%（随机基线 53.1%），算法无长期预测能力，仅供研究参考。
