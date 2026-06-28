# Use minimal CI for Phase 1

Phase 1 CI runs formatting checks, `go vet ./...`, and `go test ./...`, but does not run Docker-based MySQL integration tests, real Tushare tests, Docker image builds, or deployment jobs. Repository integration tests remain available through the `integration` build tag and can be added to CI later when the project needs stronger database contract enforcement in every pull request.
