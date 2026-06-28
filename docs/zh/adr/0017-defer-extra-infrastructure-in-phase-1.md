# 第一阶段推迟额外基础设施

第一阶段只要求 Go 服务、MySQL 和外部 Tushare API 作为基础设施，本地 Docker Compose 仅限 MySQL。Redis、消息队列、ClickHouse、Elasticsearch、指标栈和分布式调度器都推迟到项目出现明确的缓存、吞吐、分析查询、搜索、可观测性或可靠性压力，足以证明额外运维复杂度合理时再引入。
