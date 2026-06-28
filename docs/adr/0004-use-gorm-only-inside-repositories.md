# Use GORM only inside repositories

Phase 1 uses GORM for MySQL access only inside repository implementations, while handlers and services depend on repository interfaces instead of `*gorm.DB`. Database schema changes are controlled by goose SQL migrations rather than GORM AutoMigrate so table structure, indexes, and upsert behavior remain explicit and reviewable.
