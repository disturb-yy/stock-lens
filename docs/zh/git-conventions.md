# Git 规范

本文档定义第一阶段 Git 工作流和提交约定。

## 默认分支

默认分支为：

```text
main
```

不要 force-push `main`。

## 分支命名

使用带 type 前缀的短小写分支名：

```text
feature/<short-name>
fix/<short-name>
docs/<short-name>
chore/<short-name>
refactor/<short-name>
test/<short-name>
```

示例：

```text
feature/market-query-service
docs/api-contract
chore/init-ci
```

## Commit Message 格式

使用 Conventional Commits：

```text
type(scope): summary
```

示例：

```text
docs(specs): add phase1 implementation plan
feat(market): add stock query service
fix(sync): handle stale running tasks
```

允许的 type：

```text
feat
fix
docs
test
refactor
chore
ci
build
perf
```

推荐 scope：

```text
market
sync
api
config
db
tushare
docs
ci
runtime
```

## Commit 范围

每个 commit 应表达一个意图。

不要在同一个 commit 中混入无关文档、格式化和功能代码。

英文文档变更时，在同一个 commit 中更新对应中文翻译。

## 忽略文件

不要提交本地 workspace 文件、本地 secret、日志、构建产物或覆盖率产物。

示例：

```text
.idea/
.codex/
.agents/
.env
*.log
bin/
dist/
coverage.out
```

项目 `.gitignore` 应执行这些默认规则。

## Pull Request

非平凡本地设置的变更使用 pull request。

推荐合并策略：

```text
squash merge
```

Squash merge 可以在第一阶段实现提交仍在演进时，让 `main` 历史按一致变更聚合。

个人 feature 分支在合并前可以 rebase 或 force-push。不要 force-push `main`。

## Commit 前本地检查

至少运行：

```sh
gofmt
go test ./...
```

如果本地环境默认 Go build cache 不可写，使用：

```sh
GOCACHE=/tmp/stock-lens-go-cache go test ./...
```

当测试依赖 `gomonkey` 且需要关闭内联时，运行：

```sh
go test ./... -gcflags=all="-N -l"
```

除非特定测试流程确实需要 monkey patch，否则不要把 no-inline 测试命令作为默认命令。

## CI

第一阶段 CI 运行：

```text
formatting check
go vet ./...
go test ./...
```

第一阶段 CI 不运行基于 Docker 的 MySQL 集成测试、真实 Tushare 测试、Docker 镜像构建或部署任务。

## Tag

第一阶段不要求 release tag。

如果需要里程碑 tag，使用语义化版本 tag：

```text
v0.1.0
v0.2.0
```

## 初始 Commit

推荐初始 commit message：

```text
chore: initialize project documentation and go module
```
