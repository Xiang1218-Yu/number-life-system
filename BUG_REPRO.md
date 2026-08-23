# Bug 复现说明

## Bug 是什么
迁移用户携带字符串 sub 的有效 JWT 通过认证入口后，后续请求被解析成错误用户而被拒绝。

## 如何触发
签发 sub 为字符串 42 的 HS256 JWT，再调用认证解析入口读取用户 ID。

## 根因
根因涉及文件：internal/service/auth.go；符号：ParseUserID。字符串形式的 sub 在解析时被转换后额外加一，令牌签名的用户身份因此发生偏移，认证入口和后续授权请求看到的用户不一致。

## 修复状态
Claude 主轨迹已统一 JWT subject 的数值和字符串解析，并保留签名中的原始用户 ID。

## 运行指令
以下命令在含 Bug 的基线快照中运行：

```bash
go test -v ./internal/service -run '^TestV6StringSubjectKeepsUserIdentity$' -count=1
```

## 错误信息
目标验证用例应失败，下面保留该次运行的原始输出。

## 错误堆栈
```text
$ go test -v ./internal/service -run '^TestV6StringSubjectKeepsUserIdentity$' -count=1
returncode=1
=== RUN   TestV6StringSubjectKeepsUserIdentity
    v6_verify_test.go:20: parsed user id = 43, want 42
--- FAIL: TestV6StringSubjectKeepsUserIdentity (0.00s)
FAIL
FAIL	number-life-system/internal/service	0.491s
FAIL
```
