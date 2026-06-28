# Use slog and request_id correlation

Phase 1 uses Go `log/slog` for structured logs and uses `request_id` as the single correlation identifier across HTTP requests, background sync tasks, sync logs, providers, and repositories. Distributed tracing fields such as `trace_id` and `span_id` are deferred until the system has multiple services or OpenTelemetry infrastructure that justifies them.
