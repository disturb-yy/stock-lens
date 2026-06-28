# Use DATE and DATETIME(3) for market time values

Phase 1 stores trade dates as MySQL `DATE` values and audit timestamps as `DATETIME(3)`, while HTTP APIs use `YYYY-MM-DD` and Tushare adapters handle `YYYYMMDD` at the provider boundary. The system avoids MySQL `TIMESTAMP` and Unix-second timestamps because trade dates are calendar dates, audit timestamps need millisecond precision, and implicit time-zone conversion would make sync troubleshooting harder.
