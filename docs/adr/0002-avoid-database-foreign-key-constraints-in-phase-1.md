# Avoid database foreign key constraints in Phase 1

Phase 1 does not use database-level foreign key constraints between market tables. The system favors business unique keys, repository-level consistency, and focused tests because market data is synchronized in batches, re-run often, and must tolerate partial failures without making recovery or upsert flows brittle.
