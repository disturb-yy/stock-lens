# Use MySQL as the Phase 1 primary store

Phase 1 uses MySQL as the only primary store for stocks, daily K lines, trade calendars, sync tasks, and sync logs. MySQL fits the required OLTP queries and batch upserts without introducing a separate analytical database, while ClickHouse, Redis, Elasticsearch, and other stores remain deferred until the product has concrete scale or query-shape pressure that justifies them.
