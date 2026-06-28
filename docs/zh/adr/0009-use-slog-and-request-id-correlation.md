# 使用 slog 和 request_id 关联

第一阶段使用 Go `log/slog` 记录结构化日志，并把 `request_id` 作为贯穿 HTTP 请求、后台同步任务、同步日志、provider 和 repository 的唯一关联标识。`trace_id`、`span_id` 等分布式追踪字段推迟到系统拥有多个服务或 OpenTelemetry 基础设施后再引入。
