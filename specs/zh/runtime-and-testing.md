# 运行时和测试

本文档冻结第一阶段运行时默认值、就绪语义和单元测试约定。

## 服务

默认 HTTP 服务端口为 `30078`。

本地默认 base URL：

```text
http://localhost:30078
```

公共 API 端点位于 `/api/v1` 下，market 端点位于 `/api/v1/market` 下。

健康检查端点：

```text
GET /healthz
GET /readyz
```

`/healthz` 表示进程存活，并能响应 HTTP。

`/readyz` 表示服务已经就绪，可以承接流量。第一阶段 readiness 应包括启动配置校验成功和 MySQL 连通性。

## 配置期望

服务应暴露显式配置：

- server 地址和端口，默认端口 `30078`
- 业务时区，默认 `Asia/Shanghai`
- MySQL 连接设置
- Admin Token
- market provider 选择，`mock` 或 `tushare`
- 使用 Tushare provider 时的 Tushare token 和 base URL
- 同步批大小
- 日志级别

Mock provider 模式可以在没有 Tushare token 的情况下启动。Tushare provider 模式要求提供 Tushare token。

默认配置路径：

```text
configs/config.yaml
```

测试配置路径：

```text
configs/config.test.yaml
```

推荐配置形状：

```yaml
server:
  addr: "0.0.0.0"
  port: 30078
  timezone: "Asia/Shanghai"

database:
  driver: "mysql"
  dsn: "${MYSQL_DSN}"

auth:
  admin_token: "${ADMIN_TOKEN}"

market:
  provider: "mock"
  batch_size: 500

tushare:
  base_url: "https://api.tushare.pro"
  token: "${TUSHARE_TOKEN}"

log:
  level: "info"
```

YAML 值支持环境变量展开，例如 `${TUSHARE_TOKEN}`。必填值缺失会导致启动失败。

配置路径优先级：

```text
explicit config path flag > configs/config.yaml
```

环境变量只用于 YAML 展开。不要在代码库中分散直接读取环境变量。

校验规则：

- `server.addr` 默认 `0.0.0.0`。
- `server.port` 默认 `30078`。
- `server.timezone` 默认 `Asia/Shanghai`；无效时区导致启动失败。
- `database.dsn` 必填。
- `auth.admin_token` 必填，因为同步 API 需要 Admin Token。
- `market.provider` 默认 `mock`，且必须为 `mock` 或 `tushare`。
- `market.batch_size` 默认 `500`；非正数导致启动失败。
- `market.provider=mock` 不要求 `tushare.token`。
- `market.provider=tushare` 要求 `tushare.token`。
- `log.level` 默认 `info`，且必须为 `debug`、`info`、`warn` 或 `error`。

Admin 保护的同步 API 使用：

```text
Authorization: Bearer <token>
```

## 启动和关闭

启动顺序：

1. 从显式配置路径或 `configs/config.yaml` 加载配置。
2. 展开 YAML 值中的环境变量。
3. 校验配置。
4. 初始化日志。
5. 连接 MySQL 并 ping。
6. 将 stale `PENDING` 和 `RUNNING` 同步任务标记为 `FAILED`。
7. 启动 HTTP server。

配置校验失败会在 HTTP server 启动前停止启动。

MySQL ping 失败会停止启动。

服务启动时不自动执行 migration。Migration 通过命令或 Makefile target 显式执行。

`/readyz` 只有在配置校验、MySQL 连通性和 stale task recovery 都完成后才返回 ready。否则返回非 200。

`/healthz` 不检查 MySQL。只要进程能通过 HTTP 响应，就返回 `200`。

收到 `SIGINT` 或 `SIGTERM` 时，服务应停止接收新请求，并在超时时间内执行 HTTP graceful shutdown。

第一阶段 graceful shutdown 时不支持取消正在运行的同步任务。如果进程退出时任务仍在运行，下次启动会将该任务标记为 `FAILED`。

## 单元测试

第一阶段优先使用单元测试，而不是宽泛集成测试。

当一个函数有多组 input/output case 时，默认使用表格驱动测试。

优先选择最小有用单元测试范围。直接测试函数或方法行为，除非被测行为就是边界本身，否则避免引入 handler、repository、provider 和数据库。

当必须拦截某个依赖函数且普通接口替换不实际时，单元测试可以使用 `gomonkey`。

使用 `gomonkey` 的测试必须考虑 Go 内联。如果 patch 依赖替换一个可能被内联的函数，用关闭内联的方式运行该测试：

```sh
go test ./... -gcflags=all="-N -l"
```

默认测试命令不应关闭内联。无内联命令只保留给确实需要 monkey patch 的测试或本地调试流程。

Provider 单元测试必须使用构造的原始响应，而不是真实 Tushare 网络调用。

Repository 集成测试添加后，应继续放在 `integration` build tag 后面。
