# Use a single market domain package

Phase 1 keeps the market domain in one `domain/market` Go package and separates responsibilities by files such as `model.go`, `port.go`, `service.go`, `handler.go`, repository files, and provider files. This follows the project's Domain-Driven Flattening approach: the market domain is self-contained, while avoiding premature subpackages, import cycles, and duplicated boundary types before the domain grows enough to justify finer package splits.
