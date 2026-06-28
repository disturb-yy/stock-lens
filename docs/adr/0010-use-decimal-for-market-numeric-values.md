# Use decimal for market numeric values

Phase 1 represents market numeric values with `decimal.Decimal` in Go, `DECIMAL` columns in MySQL, and strings in HTTP responses. Prices, volume, amount, and percentage change must not use floating-point types because market data will be displayed, synchronized repeatedly, and later reused by analysis features where deterministic formatting and arithmetic matter.
