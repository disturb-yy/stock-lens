# Use Gin only at the HTTP boundary

Phase 1 uses Gin for routing, handlers, and HTTP middleware, but `gin.Context` must not enter application services or repositories. This keeps the market services testable and prevents HTTP framework concerns from leaking into sync jobs, repository code, or future non-HTTP entry points.
