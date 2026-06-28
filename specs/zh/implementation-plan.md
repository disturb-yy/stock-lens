# 实施计划

本计划按顺序推进第一阶段实现，从基础设施底座到 market 领域行为和 API 文档。

## 原则

- 第一阶段聚焦 `domain/market`。
- 以薄纵切推进，但先建立共享基础设施。
- 先跑通 mock provider/repository 流程，再接真实 Tushare。
- 测试贴近各层，默认使用表格驱动单元测试。
- 英文文档是事实来源；英文文档变更时同步更新中文镜像。

## 里程碑 1：项目骨架和运行时

交付：

- 使用 Go 1.26 的 `go.mod`
- `cmd/server` 下的命令入口
- 支持 YAML 和环境变量展开的显式 config loader
- 对 server、database、auth、market、Tushare 和 log 配置做校验
- `slog` logger 初始化
- 带 `/healthz` 和 `/readyz` 的 HTTP router
- 默认服务端口 `30078`
- 收到 `SIGINT` 和 `SIGTERM` 时 graceful shutdown

测试：

- 表格驱动的配置加载和校验测试
- health/readiness handler 测试

## 里程碑 2：共享 HTTP 和 Middleware

交付：

- 包含 `code`、`message`、`request_id` 和 `data` 的统一响应 helper
- request ID middleware
- access log middleware
- 使用 `Authorization: Bearer <token>` 的 Admin Token middleware
- 通用错误和 market 错误映射 helper

测试：

- 响应形状测试
- request ID 传播测试
- Admin Token middleware 测试
- 错误映射表格测试

## 里程碑 3：Market 领域模型和端口

交付：

- `domain/market/model.go`
- `domain/market/port.go`
- `domain/market/errors.go`
- 领域常量和枚举
- repository 接口
- provider 接口
- transaction manager 接口
- service input/query struct

测试：

- 枚举校验测试
- symbol/date/range 校验测试
- 领域错误分类测试

## 里程碑 4：Mock Repository 和 Provider

交付：

- 内存 mock repository
- 使用 `DataSource = MOCK` 的 mock provider
- 用于同步任务测试的可控失败场景
- mock repository 的类 upsert 行为
- 接近 MySQL 语义的分页和过滤行为

测试：

- repository upsert 和 query 表格测试
- mock provider 成功、空结果和失败测试

## 里程碑 5：Query Service

交付：

- 股票列表查询
- 带 `latest_daily_k_line` 的股票详情查询
- 日 K 列表查询
- 最新日 K 查询
- 交易日历列表查询
- 最近开市日查询
- is-open 查询
- 枚举元数据查询

测试：

- 使用 mock repository 的表格驱动 service 测试
- 空结果和 not-found 行为测试
- 默认日期范围测试

## 里程碑 6：Sync Service

交付：

- 同步任务创建和全局 active task 冲突检查
- 后台 goroutine 执行
- 启动时 stale task recovery
- 同步股票主数据
- 同步交易日历
- 同步日 K，包括单股票和全市场
- 任务状态流转
- 进度计数器
- 同步日志
- panic recovery 到失败任务

测试：

- 任务冲突测试
- 状态流转测试
- 前置条件失败测试
- 全市场日 K 同步部分成功测试
- panic recovery 测试
- 只有当接口替换不实际时才使用 gomonkey；这些测试必须考虑 no-inline 测试命令

## 里程碑 7：HTTP Handler 和路由

交付：

- market handler 和路由注册
- `/api/v1/market` 下的读取端点
- Admin Token 保护的同步端点
- 带 snake_case JSON tag 的 DTO 转换
- 分页响应形状
- date/datetime 格式化
- error 到 HTTP 的映射

测试：

- 成功和错误响应 handler 测试
- DTO 格式化测试
- Admin 保护路由测试

## 里程碑 8：MySQL 持久化

交付：

- `migrations/` 下的 SQL migration
- 只在 repository 内部使用 GORM 的 MySQL repository 实现
- MySQL transaction manager
- 批量 upsert 行为
- 与表结构规格一致的索引
- Makefile migration target

测试：

- `integration` build tag 后的 repository 集成测试
- 可行时做 migration smoke check

## 里程碑 9：Tushare 集成

交付：

- 原始 `pkg/tushare` client
- 显式 Tushare `fields` 请求
- `domain/market/provider_tushare.go` 中的 Tushare provider 适配器
- `ts_code` 到 `symbol` 的映射
- exchange、board、status、date、decimal 和 unit 映射
- 不记录 token 的 provider 错误包装

测试：

- 使用构造响应的原始 response 解析测试
- provider 映射表格测试
- 无效映射错误测试
- 不做真实 Tushare 网络测试

## 里程碑 10：API 契约和本地开发体验

交付：

- `specs/openapi.yaml`
- 必要时增加人工 API 指南
- `configs/config.yaml`
- `configs/config.test.yaml`
- 用于本地 MySQL 的 `docker-compose.yml`
- 用于 run、test、integration test、migration up/down 和 formatting 的 `Makefile` target
- README 更新
- 包含 formatting、vet 和单元测试的最小 CI

测试：

- `go test ./...`
- `go vet ./...`
- formatting check

## 推荐构建顺序

```text
runtime/config/logger/router
-> HTTP response and middleware
-> market models/errors/ports
-> mock repositories/providers
-> query service
-> sync service
-> handlers/routes
-> MySQL repositories/migrations
-> Tushare raw client/provider
-> OpenAPI/docs/CI polish
```
