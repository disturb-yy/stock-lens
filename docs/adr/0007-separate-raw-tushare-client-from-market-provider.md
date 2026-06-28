# Separate raw Tushare client from market provider

Phase 1 separates the raw Tushare HTTP client in `pkg/tushare` from the market-domain adapter in `domain/market/provider_tushare.go`. The raw client owns protocol details such as token handling, `api_name`, fields, retries, and raw response parsing, while the market provider converts Tushare-specific fields into domain models so future providers can be added without leaking Tushare semantics into the market domain.
