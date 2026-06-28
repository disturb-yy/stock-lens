# 使用 MySQL 作为第一阶段主存储

第一阶段使用 MySQL 作为股票、日 K、交易日历、同步任务和同步日志的唯一主存储。MySQL 能满足所需的 OLTP 查询和批量 upsert，不需要引入单独的分析型数据库；ClickHouse、Redis、Elasticsearch 以及其他存储会推迟到产品出现明确的规模压力或查询形态压力时再考虑。
