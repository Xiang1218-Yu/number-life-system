# Bug 复现说明

## Bug 是什么
CSV 导入的归档账户被转换为 active，导致活跃列表、归档列表和安全审计使用不同口径。

## 如何触发
导入 status 为 archived 的账户 CSV，再读取账户状态和安全审计范围。

## 根因
根因涉及文件：internal/service/csv.go、internal/service/account.go、internal/service/security.go；符号：normalizeAccountStatus、AccountService.ListPage、SecurityService.Report。CSV 状态归一化把 archived 映射成 active，同时不同读取入口使用 active、status <> archived 等不一致条件，形成状态污染并把归档数据重新纳入活跃审计。

## 修复状态
诊断题：Claude 主轨迹仅分析问题，未修改生产代码。

## 运行指令
以下命令在含 Bug 的基线快照中运行：

```bash
go test -v ./internal/service -run '^TestV6CSVImportPreservesArchivedStatus$' -count=1
```

## 错误信息
目标验证用例应失败，下面保留该次运行的原始输出。

## 错误堆栈
```text
$ go test -v ./internal/service -run '^TestV6CSVImportPreservesArchivedStatus$' -count=1
returncode=1
=== RUN   TestV6CSVImportPreservesArchivedStatus

2026/08/23 18:54:58 [31;1m/Users/tog/Documents/ChatGPT/新建go/2026-08-23/number-life-system__004/.base_snapshot/internal/service/v6_verify_test.go:27 [35;1mrecord not found
[0m[33m[0.409ms] [34;1m[rows:0][0m SELECT * FROM "users" WHERE email = 'v6-004@example.test' ORDER BY "users"."id" LIMIT 1
    v6_verify_test.go:45: imported status = "active", want archived
--- FAIL: TestV6CSVImportPreservesArchivedStatus (0.11s)
FAIL
FAIL	number-life-system/internal/service	0.597s
FAIL
```
