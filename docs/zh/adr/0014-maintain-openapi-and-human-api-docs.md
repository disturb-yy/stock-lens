# 维护 OpenAPI 和人工 API 文档

第一阶段维护 `specs/openapi.yaml` 作为机器可读的 HTTP 契约，并维护 `specs/market-api.md` 作为人工可读的 API 指南。OpenAPI 负责路径、参数、schema、响应形状和错误码；Markdown 指南解释字段含义、市场数据单位、默认值、限制和同步行为，这些内容仅靠 schema 难以表达。
