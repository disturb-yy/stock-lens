# 分离原始 Tushare Client 和 Market Provider

第一阶段将 `pkg/tushare` 中的原始 Tushare HTTP client 与 `domain/market/provider_tushare.go` 中的 market 领域适配器分离。原始 client 负责 token 处理、`api_name`、字段、重试、原始响应解析等协议细节；market provider 负责把 Tushare 专用字段转换为领域模型，便于未来新增 provider，同时避免 Tushare 语义泄漏到 market 领域。
