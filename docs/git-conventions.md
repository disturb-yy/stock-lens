# Git Conventions

This document defines the Phase 1 Git workflow and commit conventions.

## Default Branch

The default branch is:

```text
main
```

Do not force-push `main`.

## Branch Naming

Use short lowercase branch names with a type prefix:

```text
feature/<short-name>
fix/<short-name>
docs/<short-name>
chore/<short-name>
refactor/<short-name>
test/<short-name>
```

Examples:

```text
feature/market-query-service
docs/api-contract
chore/init-ci
```

## Commit Message Format

Use Conventional Commits:

```text
type(scope): summary
```

Examples:

```text
docs(specs): add phase1 implementation plan
feat(market): add stock query service
fix(sync): handle stale running tasks
```

Allowed types:

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

Recommended scopes:

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

## Commit Scope

Each commit should express one intent.

Do not mix unrelated documentation, formatting, and feature code in the same commit.

When English documentation changes, update the corresponding Chinese translation in the same commit.

## Ignored Files

Do not commit local workspace files, local secrets, logs, build outputs, or coverage artifacts.

Examples:

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

The project `.gitignore` should enforce these defaults.

## Pull Requests

Use pull requests for changes that are not trivial local setup.

Recommended merge strategy:

```text
squash merge
```

Squash merge keeps `main` history grouped by coherent change while Phase 1 implementation commits are still evolving.

Personal feature branches may be rebased or force-pushed before merge. Do not force-push `main`.

## Local Checks Before Commit

At minimum, run:

```sh
gofmt
go test ./...
```

If the default Go build cache is not writable in the local environment, use:

```sh
GOCACHE=/tmp/stock-lens-go-cache go test ./...
```

When tests rely on `gomonkey` and require inlining to be disabled, run:

```sh
go test ./... -gcflags=all="-N -l"
```

Do not use the no-inline test command as the default unless the specific test flow requires monkey patching.

## CI

Phase 1 CI runs:

```text
formatting check
go vet ./...
go test ./...
```

CI does not run Docker-based MySQL integration tests, real Tushare tests, Docker image builds, or deployment jobs in Phase 1.

## Tags

Phase 1 does not require release tags.

If milestone tags are needed, use semantic version tags:

```text
v0.1.0
v0.2.0
```

## Initial Commit

Recommended initial commit message:

```text
chore: initialize project documentation and go module
```
