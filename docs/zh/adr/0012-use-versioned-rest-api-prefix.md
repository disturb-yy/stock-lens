# 使用带版本的 REST API 前缀

第一阶段在 `/api/v1` 下暴露公共 HTTP API，market 端点位于 `/api/v1/market` 下，而健康检查保留在 `/healthz` 和 `/readyz`。带版本的前缀可以避免与未来 API 表面冲突，并允许后续 `/api/v2` 端点与现有客户端共存。
