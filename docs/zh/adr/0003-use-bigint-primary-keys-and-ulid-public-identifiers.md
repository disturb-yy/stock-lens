# 使用 BIGINT 主键和 ULID 公开标识

第一阶段对数据库表内部主键使用 `BIGINT UNSIGNED AUTO_INCREMENT`，只对需要公开不透明标识的资源使用 ULID，例如同步任务和请求 ID。市场数据通过 `market + asset_type + symbol`、`market + asset_type + symbol + trade_date` 等业务键保持领域身份，因此股票不再拥有单独的公开 `uid`。
