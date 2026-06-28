# 市场时间值使用 DATE 和 DATETIME(3)

第一阶段将交易日期存储为 MySQL `DATE`，将审计时间戳存储为 `DATETIME(3)`；HTTP API 使用 `YYYY-MM-DD`，Tushare 适配器在 provider 边界处理 `YYYYMMDD`。系统避免使用 MySQL `TIMESTAMP` 和 Unix 秒级时间戳，因为交易日期是日历日期，审计时间戳需要毫秒精度，而隐式时区转换会增加同步排障难度。
