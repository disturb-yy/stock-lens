# 第一阶段使用最小 CI

第一阶段 CI 运行格式检查、`go vet ./...` 和 `go test ./...`，但不运行基于 Docker 的 MySQL 集成测试、真实 Tushare 测试、Docker 镜像构建或部署任务。Repository 集成测试仍可通过 `integration` build tag 运行，并可在项目需要每个 pull request 都强化数据库契约时再加入 CI。
