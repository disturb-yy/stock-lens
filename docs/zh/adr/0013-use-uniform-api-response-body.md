# 使用统一 API 响应体

第一阶段对成功和错误响应都返回统一 JSON 响应体，包含 `code`、`message`、`request_id` 和 `data`，同时 HTTP status code 仍然承载协议语义。这样客户端可以获得稳定的领域错误契约，同时不会向网关、日志、监控或标准 HTTP 工具隐藏失败。
