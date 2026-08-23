# Bug 复现说明

## Bug 是什么
旧档案导入后，子表记录的账户关联消失或指向错误编号。

## 如何触发
用旧账户 ID 建立导出档案，导入到会重新分配账户 ID 的环境，再检查子表关联映射。

## 根因
根因涉及文件：internal/handler/handler.go、internal/service/export.go；符号：remapAccountID、Handler.ImportData。导入关联恢复对映射值做了额外偏移，且导出字段默认不参与 JSON 传输，跨表数据流因此可能指向其他账户或变成空关联。

## 修复状态
Claude 主轨迹已让导入直接使用旧 ID 到新 ID 的映射值，并同步修正导入边界信息。

## 运行指令
以下命令在含 Bug 的基线快照中运行：

```bash
go test -v ./internal/service -run '^TestV6ImportRemapsAccountRelation$' -count=1
```

## 错误信息
目标验证用例应失败，下面保留该次运行的原始输出。

## 错误堆栈
```text
$ go test -v ./internal/service -run '^TestV6ImportRemapsAccountRelation$' -count=1
returncode=1
=== RUN   TestV6ImportRemapsAccountRelation
    v6_verify_test.go:9: remapped account id = 0x506bab5186e0, want 19
--- FAIL: TestV6ImportRemapsAccountRelation (0.00s)
FAIL
FAIL	number-life-system/internal/service	0.552s
FAIL
```
