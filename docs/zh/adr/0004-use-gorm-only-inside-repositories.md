# 仅在 Repository 内使用 GORM

第一阶段只在 repository 实现内部使用 GORM 访问 MySQL，handler 和 service 依赖 repository 接口，而不是依赖 `*gorm.DB`。数据库结构变更由 goose SQL migration 控制，而不是使用 GORM AutoMigrate，这样表结构、索引和 upsert 行为都保持显式且可审查。
