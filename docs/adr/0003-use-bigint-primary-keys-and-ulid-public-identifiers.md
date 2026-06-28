# Use BIGINT primary keys and ULID public identifiers

Phase 1 uses `BIGINT UNSIGNED AUTO_INCREMENT` as the internal primary key for database tables and uses ULID only for resources that need a public opaque identifier, such as sync tasks and request IDs. Market data keeps its domain identity in business keys like `market + asset_type + symbol` and `market + asset_type + symbol + trade_date`, so stocks do not get a separate public `uid`.
