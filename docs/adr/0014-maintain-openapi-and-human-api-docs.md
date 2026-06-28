# Maintain OpenAPI and human API docs

Phase 1 maintains `specs/openapi.yaml` as the machine-readable HTTP contract and `specs/market-api.md` as the human-readable API guide. OpenAPI owns paths, parameters, schemas, response shapes, and error codes, while the Markdown guide explains field meaning, market-data units, defaults, limits, and sync behavior that are awkward to capture as schema alone.
