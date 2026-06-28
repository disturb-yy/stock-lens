# 第一阶段战略

Stock Lens 第一阶段是市场数据底座，不是完整股票分析产品。它的成功标准是稳定同步、存储和查询 A 股市场数据。

## 优先级

第一阶段优先级为：

```text
data correctness > sync recoverability > API stability > performance optimization > analysis capability
```

主要用户是下游前端、分析服务和开发者，而不是终端投资者。第一阶段中，API 清晰度和数据契约比页面级用户体验更重要。

## 范围

第一阶段只实现市场数据后端底座：

- 股票主数据同步
- 交易日历同步
- 原始不复权日 K 同步
- 股票列表和详情查询
- 日 K 和最新日 K 查询
- 同步任务和同步日志查询
- 健康检查
- 枚举元数据

第一阶段明确排除：

- 前端页面
- 登录、RBAC、用户、账户和租户系统
- 自选股
- 告警
- 技术指标
- AI 分析
- 回测
- 分钟级数据
- 实时行情
- 定时同步
- 分布式任务执行
- 多数据源仲裁

第一阶段不要引入 `domain/analysis`、`domain/strategy` 或 `domain/indicator`。原始数据底座稳定性优先。

## 数据源战略

Tushare 是第一阶段真实数据源，但领域模型不能耦合到 Tushare 语义。

Provider 边界存在的目的，是让未来可以加入 AkShare、JoinQuant 或内部数据源，而不需要重写 market service。

## 存储战略

MySQL 是第一阶段唯一主存储。

不要在出现明确查询、吞吐、搜索、缓存或分析压力前引入 ClickHouse、Elasticsearch、Redis 或其他存储系统。

## 同步战略

第一阶段同步是手动触发、单实例、全局互斥的。

推荐演进路径：

```text
manual sync -> scheduled sync -> cancelable tasks -> queue/workers -> distributed scheduling
```

## 市场数据战略

第一阶段只存储原始不复权日 K。

未来扩展可以包括：

- 复权因子
- 前复权和后复权视图
- 分钟级 K 线
- 指数、ETF、基金和其他资产类型
- 实时行情

## API 战略

`/api/v1` 必须保持稳定，不应为了内部实现便利而变更。

不向后兼容的公共 API 变更属于未来的 `/api/v2`。

## 文档战略

英文文档是事实来源。

中文文档是翻译镜像，不得作为决策依据。

英文文档变更时，更新对应中文翻译。

## 测试战略

第一阶段默认测试套件使用快速单元测试。

Repository 集成测试可以放在 `integration` build tag 后面。真实 Tushare 网络测试不属于第一阶段。

## 性能战略

第一阶段关注合理的批量 upsert 行为和查询索引。

在出现明确压力前，不添加缓存系统、搜索引擎或列式存储。

## 可观测性战略

第一阶段使用 `request_id`、结构化 `slog` 和 `sync_logs`。

OpenTelemetry、指标栈和分布式追踪推迟到服务拆分或生产排障压力能证明其必要性时再引入。

## 安全战略

第一阶段使用 Admin Token 保护写入型同步操作。

查询 API 默认开放。如果服务未来面向公网，可以再引入认证、授权和限流。

## 部署战略

第一阶段目标是单 Go 服务加 MySQL。

Docker Compose 只用于本地依赖。在真正需要前，不设计 Kubernetes 或分布式部署拓扑。

## 领域边界战略

第一阶段只实现 `domain/market`。

只有当真实用例和独立领域语言出现时，才创建新领域。不要基于技术想象创建领域。

## 推迟能力触发条件

被推迟的能力需要具体触发条件：

```text
scheduled sync: 需要无人值守日更
Redis: 热点查询对 MySQL 产生可度量压力
ClickHouse: 历史 K 线分析拖慢 OLTP 工作负载
user system: 出现多用户权限需求
indicator domain: 需要明确的技术指标 API 或计算
```

## 完成标准

第一阶段完成时，系统应能够：

- 同步股票主数据
- 同步交易日历
- 同步日 K
- 查询股票列表和股票详情
- 查询日 K 和最新日 K
- 查询同步任务和同步日志
- 暴露稳定健康检查
- 暴露枚举元数据
- 维护稳定错误码、API 文档和表结构文档
- 运行基础测试和 CI
