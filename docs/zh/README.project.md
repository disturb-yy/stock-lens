# Stock Lens

Stock Lens 是第一阶段 A 股市场数据后端底座。它同步股票主数据、交易日历和原始不复权日 K，并提供稳定的查询 API 和同步任务 API。

## 当前范围

第一阶段聚焦：

- A 股股票主数据同步
- A 股交易日历同步
- 原始不复权日 K 同步
- 股票列表和股票详情查询
- 日 K 和最新日 K 查询
- 同步任务和同步日志查询
- 健康检查和枚举元数据

第一阶段不包含前端页面、用户登录、RBAC、自选股、告警、技术指标、AI 分析、回测、实时数据、定时同步或分布式任务执行。

## 运行时默认值

默认本地服务：

```text
http://localhost:30078
```

健康检查端点：

```text
GET /healthz
GET /readyz
```

`/healthz` 表示 HTTP 进程存活。`/readyz` 表示服务已经就绪，包括必要的启动校验和 MySQL 连通性。

API 前缀：

```text
/api/v1
```

Market API 前缀：

```text
/api/v1/market
```

## 配置

实现应使用显式配置：

- server 地址和端口，默认 `30078`
- 业务时区，默认 `Asia/Shanghai`
- MySQL 连接设置
- 用于同步 API 的 Admin Token
- market provider，`mock` 或 `tushare`
- 使用 Tushare provider 时的 Tushare token 和 base URL
- 同步批大小
- 日志级别

Mock provider 模式可以在没有 Tushare token 的情况下运行。Tushare provider 模式要求提供 Tushare token。

## 主要 API

读取 API：

```text
GET /api/v1/market/stocks
GET /api/v1/market/stocks/{symbol}
GET /api/v1/market/stocks/{symbol}/daily-k-lines
GET /api/v1/market/stocks/{symbol}/latest-daily-k-line
GET /api/v1/market/trade-calendars
GET /api/v1/market/trade-calendars/latest-open-day
GET /api/v1/market/trade-calendars/is-open
GET /api/v1/market/meta/enums
```

同步 API 需要 Admin Token：

```text
POST /api/v1/market/sync/stocks
POST /api/v1/market/sync/trade-calendars
POST /api/v1/market/sync/daily-k-lines
GET  /api/v1/market/sync/tasks/{task_uid}
GET  /api/v1/market/sync/tasks/{task_uid}/logs
```

## 测试

默认测试套件使用快速单元测试：

```sh
go test ./...
```

单元测试应优先使用表格驱动写法，并选择最小有用测试范围。

依赖 `gomonkey` 的测试可能需要关闭内联：

```sh
go test ./... -gcflags=all="-N -l"
```

除非特定测试流程确实需要 monkey patch，否则默认测试运行不要关闭内联。

Repository 集成测试添加后，应使用 `integration` build tag。

## 文档

英文文档是事实来源。`zh/` 目录下的中文文档是翻译镜像；对应英文文档变更时应同步更新。

关键文档：

- `specs/phase1-strategy.md`
- `specs/api-contract-decisions.md`
- `specs/phase1-table-structure.md`
- `specs/sync-task-behavior.md`
- `specs/provider-tushare-adapter.md`
- `specs/runtime-and-testing.md`
- `specs/implementation-plan.md`
- `docs/git-conventions.md`
- `domain/market/INDEX.md`
