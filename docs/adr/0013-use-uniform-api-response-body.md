# Use uniform API response body

Phase 1 returns a uniform JSON response body with `code`, `message`, `request_id`, and `data` for both success and error responses, while HTTP status codes still carry protocol semantics. This gives clients a stable domain error contract without hiding failures from gateways, logs, monitors, or standard HTTP tooling.
