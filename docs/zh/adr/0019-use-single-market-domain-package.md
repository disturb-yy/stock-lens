# 使用单一 Market 领域包

第一阶段将 market 领域保持在一个 `domain/market` Go package 中，并通过 `model.go`、`port.go`、`service.go`、`handler.go`、repository 文件和 provider 文件等文件区分职责。这遵循项目的 Domain-Driven Flattening 方法：market 领域自包含，同时避免在领域还没有增长到需要更细包拆分前，过早引入子包、import cycle 和重复边界类型。
