# Defer extra infrastructure in Phase 1

Phase 1 runs with the Go service, MySQL, and the external Tushare API as its only required infrastructure, with local Docker Compose limited to MySQL. Redis, message queues, ClickHouse, Elasticsearch, metrics stacks, and distributed schedulers are deferred until the project has concrete cache, throughput, analytical-query, search, observability, or reliability pressure that justifies the additional operational complexity.
