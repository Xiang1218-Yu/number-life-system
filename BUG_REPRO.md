# Bug 复现说明

## Bug 是什么
订阅趋势的历史月份和周期金额不稳定，年付服务的月度支出被高估。

## 如何触发
计算年付和季付订阅的趋势月度金额，并比较不同查询月份范围下的历史结果。

## 根因
根因涉及文件：internal/service/subscription.go、internal/service/stats.go；符号：monthlyAmount、StatsService.SubscriptionTrend。年付金额没有按 12 个月均摊，趋势又以当前时刻而不是自然月边界计算并忽略取消时间，跨层数据流导致历史月份和金额口径随查询范围变化。

## 修复状态
Claude 主轨迹已统一周期金额归一化、自然月边界和取消时间判断。

## 运行指令
以下命令在含 Bug 的基线快照中运行：

```bash
go test -v ./internal/service -run '^TestV6MonthlyAmountNormalizesCycles$' -count=1
```

## 错误信息
目标验证用例应失败，下面保留该次运行的原始输出。

## 错误堆栈
```text
$ go test -v ./internal/service -run '^TestV6MonthlyAmountNormalizesCycles$' -count=1
returncode=1
=== RUN   TestV6MonthlyAmountNormalizesCycles
    v6_verify_test.go:7: annual monthly amount = 120, want 10
--- FAIL: TestV6MonthlyAmountNormalizesCycles (0.00s)
FAIL
FAIL	number-life-system/internal/service	0.457s
FAIL
```
