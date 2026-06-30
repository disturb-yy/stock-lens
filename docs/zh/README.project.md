# Stock Lens 第一阶段

Stock Lens 第一阶段是面向 A 股数据的行情数据后端底座。它提供配置驱动的运行时启动、MySQL 持久化、行情数据同步任务，以及 `/api/v1/market` 查询 API。

## 本地运行

启动本地 MySQL：

```sh
make dev-up
```

执行数据库迁移：

```sh
MYSQL_DSN='stock_lens:stock_lens@tcp(127.0.0.1:33306)/stock_lens?parseTime=true&loc=Asia%2FShanghai' make migrate-up
```

启动服务：

```sh
MYSQL_DSN='stock_lens:stock_lens@tcp(127.0.0.1:33306)/stock_lens?parseTime=true&loc=Asia%2FShanghai' \
ADMIN_TOKEN='local-admin-token' \
make run
```

默认 HTTP 端口是 `30078`。

## API 与配置

- 健康检查：`GET /healthz`
- 就绪检查：`GET /readyz`
- Market API 前缀：`/api/v1/market`
- OpenAPI 契约：`specs/openapi.yaml`
- 默认配置：`configs/config.yaml`
- 测试配置：`configs/config.test.yaml`

需要 Admin Token 的同步 API 使用：

```text
Authorization: Bearer <admin-token>
```

## 验证

运行单元测试：

```sh
make test
```

运行格式检查和 vet：

```sh
make fmt-check
make vet
```

在真实 MySQL 数据库完成迁移后运行集成测试：

```sh
MYSQL_DSN='stock_lens:stock_lens@tcp(127.0.0.1:33306)/stock_lens?parseTime=true&loc=Asia%2FShanghai' make test-integration
```
