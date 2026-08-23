# Bug 复现说明

## Bug 是什么
列表结果刚好存在尾部不足整页的数据时，总页数少一页，页面没有继续翻页入口。

## 如何触发
以 total=21、page_size=20 计算列表分页元数据，并读取下一页判断。

## 根因
根因涉及文件：internal/service/pagination.go；符号：TotalPages、PageResult.HasNext。分页计算先取整除，再把有余数的结果减一，把尾部不完整页错误地丢弃，导致总页数和导航状态不一致。

## 修复状态
Claude 主轨迹已修正尾页计数，并补充前后页导航判断。

## 运行指令
以下命令在含 Bug 的基线快照中运行：

```bash
go test -v ./internal/service -run '^TestV6PaginationIncludesPartialPage$' -count=1
```

## 错误信息
目标验证用例应失败，下面保留该次运行的原始输出。

## 错误堆栈
```text
$ go test -v ./internal/service -run '^TestV6PaginationIncludesPartialPage$' -count=1
returncode=1
=== RUN   TestV6PaginationIncludesPartialPage
--- FAIL: TestV6PaginationIncludesPartialPage (0.00s)
panic: pagination total pages mismatch [recovered, repanicked]

goroutine 7 [running]:
testing.tRunner.func1.2({0x1048111e0, 0x104859ee0})
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/testing/testing.go:1974 +0x1a0
testing.tRunner.func1()
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/testing/testing.go:1977 +0x318
panic({0x1048111e0?, 0x104859ee0?})
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/runtime/panic.go:860 +0x12c
number-life-system/internal/service.TestV6PaginationIncludesPartialPage(0x37804f488d88?)
	/Users/tog/Documents/ChatGPT/新建go/2026-08-23/number-life-system__010/.base_snapshot/internal/service/v6_verify_test.go:7 +0x2c
testing.tRunner(0x37804f488d88, 0x104859048)
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/testing/testing.go:2036 +0xc4
created by testing.(*T).Run in goroutine 1
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/testing/testing.go:2101 +0x3a8
FAIL	number-life-system/internal/service	0.454s
FAIL
```
