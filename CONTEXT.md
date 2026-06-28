# Stock Lens

Stock Lens 为股票分析系统提供行情数据同步与查询的后端数据底座。
This context describes the market-data language used by the project.

## Language

**Market**:
市场。证券标的上市所在的市场。第一阶段只支持 `CN`，表示 A 股市场。
The securities market in which an instrument is listed. Phase 1 only supports `CN`, meaning A-share markets.
_Avoid_: Region, locale

**Asset Type**:
资产类型。一条记录代表的金融资产类别。第一阶段只支持 `STOCK`，表示股票。
The class of financial instrument represented by a record. Phase 1 only supports `STOCK`.
_Avoid_: Product type, instrument kind

**Stock**:
股票。系统跟踪的一只已上市或曾上市的 A 股权益类标的；第一阶段由 `market + asset_type + symbol` 唯一识别。
A listed or previously listed A-share equity instrument tracked by the system; in Phase 1 it is uniquely identified by `market + asset_type + symbol`.
_Avoid_: Security, instrument, share

**Symbol**:
股票代码。系统内部使用的标准股票代码，例如 `600519`。它不是数据源专用的 `TS Code`。
The system's canonical stock code, such as `600519`. It is not the provider-specific `TS Code`.
_Avoid_: Code, ticker

**TS Code**:
Tushare 股票代码。Tushare 专用的股票标识，例如 `600519.SH`。它是数据源标识，不是系统标准股票身份。
The Tushare-specific stock identifier, such as `600519.SH`. It is a provider identifier, not the system's canonical stock identity.
_Avoid_: Symbol, stock code

**Exchange**:
交易所。股票上市所在的交易场所，例如 `SSE`、`SZSE` 或 `BSE`。
The exchange on which a stock is listed, such as `SSE`, `SZSE`, or `BSE`.
_Avoid_: Market

**Board**:
板块。股票所属的上市板块分类，例如 `MAIN`、`GEM`、`STAR` 或 `BSE`。
The listing board classification for a stock, such as `MAIN`, `GEM`, `STAR`, or `BSE`.
_Avoid_: Exchange, market segment

**Stock Status**:
股票状态。股票在市场中的生命周期状态，例如 `LISTED`、`DELISTED` 或 `PAUSED`。
The lifecycle state of a stock in the market, such as `LISTED`, `DELISTED`, or `PAUSED`.
_Avoid_: Availability, active flag

**Daily K Line**:
日 K。某只股票在某个交易日的一条原始、不复权 OHLCV 行情记录；第一阶段由 `market + asset_type + symbol + trade_date` 唯一识别。
The raw, unadjusted daily OHLCV market-data record for one stock on one trade date; in Phase 1 it is uniquely identified by `market + asset_type + symbol + trade_date`.
_Avoid_: Daily quote, candle

**Trade Calendar**:
交易日历。记录某个交易所在自然日是否开市的市场日历。
The market calendar that records whether a calendar date is an open trading day for an exchange.
_Avoid_: Holiday calendar, schedule

**Latest Open Day**:
最近开市交易日。交易日历中最近一个标记为开市的日期。
The most recent date in the trade calendar that is marked as open for trading.
_Avoid_: Today, latest day

**Data Source**:
数据源。行情数据同步进入系统时对应的外部或模拟来源，例如 `TUSHARE` 或 `MOCK`。
The external or mock source from which market data was synchronized, such as `TUSHARE` or `MOCK`.
_Avoid_: Provider when referring to stored data provenance

**Provider**:
数据源适配器。把某个数据源的原始协议、字段和值转换成市场领域模型的组件。
A data-source adapter that converts one source's raw protocol, fields, and values into market domain models.
_Avoid_: Data Source when referring to adapter code

**Sync Task**:
同步任务。由用户触发、在后台执行的行情数据同步作业；对外使用 ULID 格式的 `uid` 识别。
A user-triggered background job that synchronizes market data into the system; externally identified by a ULID `uid`.
_Avoid_: Job, worker task

**Sync Log**:
同步日志。记录同步任务在任务级或股票级的进度与失败信息。
A task-level or stock-level record of sync progress or failure.
_Avoid_: Audit log, application log

**Request ID**:
请求关联 ID。第一阶段用于关联 HTTP 请求、后台同步任务、同步日志和结构化应用日志。
A correlation identifier used in Phase 1 to connect HTTP requests, background sync tasks, sync logs, and structured application logs.
_Avoid_: Trace ID, span ID
