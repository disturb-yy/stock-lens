# Use layered tests with table-driven unit tests

Phase 1 uses fast unit tests as the default test suite and keeps MySQL repository integration tests behind the `integration` build tag. Unit tests should be table-driven where practical, provider tests use constructed raw responses instead of real Tushare network calls, and test commands do not disable Go inlining by default; `-gcflags=all="-N -l"` is reserved for local debugging.
