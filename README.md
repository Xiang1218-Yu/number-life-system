# Number Life
个人数字生活档案系统，使用 Go、Gin、GORM、PostgreSQL 和原生 HTML/CSS/JavaScript 构建。
## 功能
- 账户记录：分类、密码强度、两步验证、泄露风险、登录时间和归档
- 订阅追踪：月费/年费、下一次扣费自动计算、近期扣费列表
- 数字足迹：注册、改密、换绑等事件时间线
- 数据分布：平台、数据类型、容量和隐私等级
- 安全评分：密码、两步验证、活跃度和泄露标记综合评分
- 备份计划：按月、季度或年度计算下一次备份时间
- JSON 导入导出和统计看板
## 启动
```bash
docker compose up -d postgres
go run ./cmd/server --config ./config/config.yaml
```
浏览器打开 `http://localhost:8080`，先注册账户。
## 构建
```bash
go build -o digital-life ./cmd/server
```
数据库启动参数在 `config/config.yaml`，服务启动时会自动执行 GORM 迁移并初始化账户分类。
