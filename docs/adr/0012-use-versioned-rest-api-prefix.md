# Use versioned REST API prefix

Phase 1 exposes public HTTP APIs under `/api/v1`, with market endpoints under `/api/v1/market`, while health checks remain at `/healthz` and `/readyz`. The versioned prefix avoids collisions with future API surfaces and allows later `/api/v2` endpoints to coexist without breaking existing clients.
