# 股票分析系统第一阶段后端设计文档

## 1. 阶段目标

第一阶段目标是实现股票分析系统的后端数据底座和基础展示 API。

核心链路：

```text
Tushare 数据源
    ↓
数据源适配器
    ↓
domain/market 领域服务
    ↓
MySQL 存储
    ↓
HTTP API 查询展示
```

第一阶段重点实现：

```text
1. A 股股票基础信息同步
2. A 股交易日历同步
3. A 股日 K 历史行情同步
4. 股票列表查询
5. 股票详情查询
6. 日 K 查询
7. 最新日 K 查询
8. 同步任务状态查询
9. 同步日志查询
10. 健康检查
11. 枚举字典接口
```

---

## 2. 非目标范围

第一阶段暂不实现：

```text
1. 前端页面
2. 用户注册 / 登录 / JWT / RBAC
3. 多用户体系
4. 复杂技术指标
5. AI 分析报告
6. 量化策略
7. 策略回测
8. 自选股
9. 告警系统
10. WebSocket 实时行情
11. 分钟级行情
12. 前复权 / 后复权
13. 定时同步
14. 任务取消
15. 分布式任务队列
16. Redis / ClickHouse / Elasticsearch
17. 真实 Tushare 网络测试
```

---

## 3. 总体架构原则

项目采用 DDF 架构：

```text
DDF = Domain-Driven Flattening
```

核心原则：

```text
1. 每个领域自包含
   模型、接口、服务、HTTP handler、repository、provider 放在同一个领域 package 内。

2. 领域之间禁止直接依赖
   market 不直接 import auth / anomaly / notify / daily。

3. 跨领域协作由 Composition Root 组装
   cmd/server/main.go 是唯一组装入口。

4. Mock-First
   外部依赖先提供 Mock 实现，再提供真实实现。

5. pkg 只放真正通用组件
   config / logger / middleware / httputil / tushare 等。
```

---

## 4. 项目目录结构

第一阶段推荐目录：

```text
stock-monitor/
├── cmd/server/
│   └── main.go
├── api/
│   └── route.go
├── domain/
│   └── market/
│       ├── model.go
│       ├── port.go
│       ├── service.go
│       ├── handler.go
│       ├── dto.go
│       ├── errors.go
│       ├── repository_mysql.go
│       ├── repository_mock.go
│       ├── provider_tushare.go
│       ├── provider_mock.go
│       └── INDEX.md
├── pkg/
│   ├── config/
│   │   ├── config.go
│   │   └── loader.go
│   ├── httputil/
│   │   └── response.go
│   ├── logger/
│   │   └── logger.go
│   ├── middleware/
│   │   ├── request_id.go
│   │   ├── admin_token.go
│   │   └── access_log.go
│   ├── router/
│   │   └── router.go
│   └── tushare/
│       ├── client.go
│       ├── request.go
│       ├── response.go
│       └── errors.go
├── configs/
│   ├── config.yaml
│   └── config.test.yaml
├── migrations/
│   ├── 000001_create_market_tables.up.sql
│   └── 000001_create_market_tables.down.sql
├── specs/
│   ├── phase1-backend-design.md
│   ├── openapi.yaml
│   └── market-api.md
├── docker-compose.yml
├── Makefile
├── go.mod
└── .github/
    └── workflows/
        └── ci.yml
```

第一阶段只重点实现：

```text
domain/market
```

暂不展开：

```text
domain/auth
domain/anomaly
domain/notify
domain/daily
```

---

## 5. 第一阶段市场和数据源范围

第一阶段市场范围：

```text
只支持 A 股
market = CN
asset_type = STOCK
```

但模型层预留：

```text
market
asset_type
exchange
data_source
```

后续可扩展：

```text
港股
美股
ETF
指数
基金
可转债
期货
```

第一阶段数据源：

```text
TUSHARE
```

但 provider 层预留多数据源适配能力。

---

## 6. 核心枚举

### 6.1 Market

```text
CN = A 股
```

第一阶段只允许：

```text
CN
```

### 6.2 AssetType

```text
STOCK = 股票
```

第一阶段只允许：

```text
STOCK
```

### 6.3 Exchange

```text
SSE  = 上海证券交易所
SZSE = 深圳证券交易所
BSE  = 北京证券交易所
```

### 6.4 Board

```text
MAIN    = 主板
GEM     = 创业板
STAR    = 科创板
BSE     = 北交所
UNKNOWN = 未知
```

推导规则：

```text
688xxx           → STAR
300xxx           → GEM
8xxxxx / 4xxxxx  → BSE
其余             → MAIN
无法判断         → UNKNOWN
```

### 6.5 StockStatus

```text
LISTED   = 上市
DELISTED = 退市
PAUSED   = 暂停上市
```

Tushare 状态转换：

```text
L → LISTED
D → DELISTED
P → PAUSED
```

### 6.6 DataSource

```text
TUSHARE
MOCK
```

第一阶段真实数据源：

```text
TUSHARE
```

---

## 7. 核心领域模型

### 7.1 Stock

```go
type Stock struct {
    ID         uint64
    Market     string
    AssetType  string
    Symbol     string
    TSCode     string
    Name       string
    Exchange   string
    Board      string
    Area       string
    Industry   string
    Status     string
    ListDate   *time.Time
    DelistDate *time.Time
    DataSource string
    SyncedAt   time.Time
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

说明：

```text
symbol = 系统内部股票代码，例如 600519
ts_code = Tushare 专用代码，例如 600519.SH
```

系统业务唯一键：

```text
market + asset_type + symbol
```

### 7.2 DailyKLine

```go
type DailyKLine struct {
    ID         uint64
    Market     string
    AssetType  string
    Symbol     string
    TradeDate  time.Time

    OpenPrice  decimal.Decimal
    HighPrice  decimal.Decimal
    LowPrice   decimal.Decimal
    ClosePrice decimal.Decimal
    PreClose   decimal.Decimal

    ChangeAmt  decimal.Decimal
    PctChange  decimal.Decimal
    Volume     decimal.Decimal
    Amount     decimal.Decimal

    DataSource string
    SyncedAt   time.Time
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

说明：

```text
daily_k_lines 不保存：
- ts_code
- stock name
- exchange
```

这些信息通过 `stocks` 表获取。

日 K 第一阶段只保存：

```text
不复权原始行情
```

暂不保存：

```text
前复权
后复权
adjust_type
```

### 7.3 TradeCalendar

```go
type TradeCalendar struct {
    ID           uint64
    Market       string
    Exchange     string
    CalDate      time.Time
    IsOpen       bool
    PretradeDate *time.Time
    DataSource   string
    SyncedAt     time.Time
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

唯一业务键：

```text
market + exchange + cal_date
```

第一阶段默认：

```text
market = CN
exchange = SSE
```

### 7.4 SyncTask

```go
type SyncTask struct {
    ID             uint64
    UID            string
    TaskType       string
    Market         string
    AssetType      string
    DataSource     string
    Status         string
    TotalItems     int64
    ProcessedItems int64
    SuccessItems   int64
    FailedItems    int64
    RequestID      string
    StartedAt      *time.Time
    FinishedAt     *time.Time
    ErrorMsg       string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

任务状态：

```text
PENDING
RUNNING
SUCCESS
FAILED
PARTIAL_SUCCESS
```

第一阶段暂不实现：

```text
CANCELING
CANCELED
```

### 7.5 SyncLog

```go
type SyncLog struct {
    ID           uint64
    TaskID       uint64
    TaskUID      string
    Step         string
    Status       string
    Market       string
    AssetType    string
    Symbol       string
    DataSource   string
    Message      string
    ErrorDetail  string
    AffectedRows int64
    CreatedAt    time.Time
}
```

日志粒度：

```text
任务步骤级
股票级成功 / 失败摘要
```

不记录：

```text
每一条 K 线级日志
```

---

## 8. 时间和数值类型规范

### 8.1 日期类型

边界规则：

```text
Tushare 边界：YYYYMMDD 字符串
HTTP API：YYYY-MM-DD 字符串
领域模型：time.Time
MySQL 交易日期：DATE
MySQL 审计时间：DATETIME(3)
```

使用：

```sql
created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
```

不使用：

```text
MySQL TIMESTAMP
INT Unix 秒级时间戳
```

原因：

```text
DATETIME(3) 不存在 2038 问题
TIMESTAMP / 32 位 Unix 秒时间戳存在 2038 风险
```

### 8.2 数值类型

Go 领域模型：

```go
decimal.Decimal
```

MySQL 精度：

```sql
open_price   DECIMAL(20,4) NOT NULL
high_price   DECIMAL(20,4) NOT NULL
low_price    DECIMAL(20,4) NOT NULL
close_price  DECIMAL(20,4) NOT NULL
pre_close    DECIMAL(20,4) NOT NULL
change_amt   DECIMAL(20,4) NOT NULL
pct_change   DECIMAL(20,4) NOT NULL
volume       DECIMAL(24,4) NOT NULL
amount       DECIMAL(24,4) NOT NULL
```

HTTP decimal 字段：

```text
统一返回 string
```

示例：

```json
{
  "open": "10.2300",
  "close": "10.4500",
  "pct_change": "2.1500"
}
```

---

## 9. 行情单位规范

数据库保持 Tushare 原始单位：

```text
volume     = 成交量，单位：手
amount     = 成交额，单位：千元
pct_change = 涨跌幅，单位：百分比
```

例如：

```text
pct_change = 2.35 表示 2.35%，不是 0.0235
```

HTTP 日 K 响应返回单位字段：

```json
{
  "volume": "525152.7700",
  "volume_unit": "lot",
  "amount": "460697.3770",
  "amount_unit": "thousand_cny",
  "pct_change": "-0.2300",
  "pct_change_unit": "percent"
}
```

数据库不额外存单位字段，单位作为领域常量和 DTO 输出。

---

## 10. Provider 接口设计

第一阶段 provider 按业务能力分组，不按 Tushare API 拆分。

```go
type InstrumentProvider interface {
    FetchStocks(ctx context.Context, req FetchStocksRequest) ([]Stock, error)
}

type MarketDataProvider interface {
    FetchDailyKLines(ctx context.Context, req FetchDailyKLinesRequest) ([]DailyKLine, error)
}

type CalendarProvider interface {
    FetchTradeCalendar(ctx context.Context, req FetchTradeCalendarRequest) ([]TradeCalendar, error)
}
```

第一阶段实现：

```text
MockInstrumentProvider
MockMarketDataProvider
MockCalendarProvider

TushareInstrumentProvider
TushareMarketDataProvider
TushareCalendarProvider
```

运行时通过配置切换：

```yaml
market:
  provider: "mock" # mock / tushare
```

---

## 11. Repository 接口设计

Service 层只依赖 Repository 接口，不直接依赖 GORM / MySQL。

建议接口：

```go
type StockRepository interface {
    UpsertStocks(ctx context.Context, stocks []Stock) error
    ListStocks(ctx context.Context, query ListStocksQuery) ([]Stock, int64, error)
    FindStock(ctx context.Context, query FindStockQuery) (*Stock, error)
    ListSyncableStocks(ctx context.Context, query ListSyncableStocksQuery) ([]Stock, error)
}

type KLineRepository interface {
    UpsertDailyKLines(ctx context.Context, lines []DailyKLine) error
    ListDailyKLines(ctx context.Context, query ListDailyKLinesQuery) ([]DailyKLine, error)
    FindLatestDailyKLine(ctx context.Context, query FindLatestDailyKLineQuery) (*DailyKLine, error)
}

type TradeCalendarRepository interface {
    UpsertTradeCalendars(ctx context.Context, calendars []TradeCalendar) error
    ListTradeCalendars(ctx context.Context, query ListTradeCalendarsQuery) ([]TradeCalendar, error)
    FindLatestOpenDay(ctx context.Context, query LatestOpenDayQuery) (*TradeCalendar, error)
    FindTradeCalendar(ctx context.Context, query FindTradeCalendarQuery) (*TradeCalendar, error)
}

type SyncTaskRepository interface {
    CreateTask(ctx context.Context, task *SyncTask) error
    UpdateTask(ctx context.Context, task *SyncTask) error
    FindTaskByUID(ctx context.Context, uid string) (*SyncTask, error)
    HasActiveTask(ctx context.Context) (bool, error)
    MarkStaleTasksFailed(ctx context.Context, reason string) error
    CreateLog(ctx context.Context, log *SyncLog) error
    ListLogs(ctx context.Context, query ListSyncLogsQuery) ([]SyncLog, int64, error)
}
```

### 11.1 TxManager

同步任务需要短事务。

```go
type TxManager interface {
    WithTx(ctx context.Context, fn func(ctx context.Context, repos TxRepositories) error) error
}

type TxRepositories struct {
    Stocks    StockRepository
    KLines    KLineRepository
    Calendars TradeCalendarRepository
    SyncTasks SyncTaskRepository
}
```

原则：

```text
Service 层不直接使用 *gorm.DB
GORM 事务封装在 repository_mysql.go 内部
QueryService 不需要事务
SyncService 使用 TxManager
```

---

## 12. Service 设计

`market` 领域拆成两个轻量 Service，但仍然放在 `service.go`。

```go
type QueryService struct {
    stocks    StockRepository
    klines    KLineRepository
    calendars TradeCalendarRepository
}

type SyncService struct {
    stocks    StockRepository
    klines    KLineRepository
    calendars TradeCalendarRepository
    tasks     SyncTaskRepository
    tx        TxManager

    instrumentProvider InstrumentProvider
    marketDataProvider MarketDataProvider
    calendarProvider    CalendarProvider

    appCtx context.Context
}
```

职责：

```text
QueryService:
- 股票列表
- 股票详情
- 日 K 查询
- 最新日 K 查询
- 交易日历查询
- 枚举字典

SyncService:
- 同步股票基础信息
- 同步日 K
- 同步交易日历
- 创建同步任务
- 启动后台 goroutine
- 更新任务状态
- 写同步日志
- 处理任务冲突
- recover panic
```

Handler 持有：

```go
type Handler struct {
    query *QueryService
    sync  *SyncService
}
```

---

## 13. 数据库设计

### 13.1 stocks

```sql
CREATE TABLE stocks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,

    market VARCHAR(16) NOT NULL,
    asset_type VARCHAR(32) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    ts_code VARCHAR(32) NOT NULL,

    name VARCHAR(128) NOT NULL,
    exchange VARCHAR(32) NOT NULL,
    board VARCHAR(32) NOT NULL DEFAULT 'UNKNOWN',
    area VARCHAR(64) NOT NULL DEFAULT '',
    industry VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL,

    list_date DATE NULL,
    delist_date DATE NULL,

    data_source VARCHAR(32) NOT NULL,
    synced_at DATETIME(3) NOT NULL,

    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    UNIQUE KEY uk_stock_identity (market, asset_type, symbol),
    INDEX idx_stocks_filter (market, asset_type, exchange, status),
    INDEX idx_stocks_symbol (symbol),
    INDEX idx_stocks_name (name)
);
```

说明：

```text
不保存中文展示名。
不使用 deleted_at。
不删除本地股票。
股票生命周期通过 status 表达。
```

### 13.2 daily_k_lines

```sql
CREATE TABLE daily_k_lines (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,

    market VARCHAR(16) NOT NULL,
    asset_type VARCHAR(32) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    trade_date DATE NOT NULL,

    open_price DECIMAL(20,4) NOT NULL,
    high_price DECIMAL(20,4) NOT NULL,
    low_price DECIMAL(20,4) NOT NULL,
    close_price DECIMAL(20,4) NOT NULL,
    pre_close DECIMAL(20,4) NOT NULL,

    change_amt DECIMAL(20,4) NOT NULL,
    pct_change DECIMAL(20,4) NOT NULL,
    volume DECIMAL(24,4) NOT NULL,
    amount DECIMAL(24,4) NOT NULL,

    data_source VARCHAR(32) NOT NULL,
    synced_at DATETIME(3) NOT NULL,

    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    UNIQUE KEY uk_daily_kline_identity (market, asset_type, symbol, trade_date),
    INDEX idx_daily_kline_query (market, asset_type, symbol, trade_date)
);
```

说明：

```text
不保存 ts_code。
不保存 name。
不保存 exchange。
不保存 adjust_type。
只保存不复权原始日 K。
```

### 13.3 trade_calendars

```sql
CREATE TABLE trade_calendars (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,

    market VARCHAR(16) NOT NULL,
    exchange VARCHAR(32) NOT NULL,
    cal_date DATE NOT NULL,
    is_open TINYINT(1) NOT NULL,
    pretrade_date DATE NULL,

    data_source VARCHAR(32) NOT NULL,
    synced_at DATETIME(3) NOT NULL,

    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    UNIQUE KEY uk_trade_calendar_identity (market, exchange, cal_date),
    INDEX idx_trade_calendar_open_day (market, exchange, is_open, cal_date)
);
```

### 13.4 sync_tasks

```sql
CREATE TABLE sync_tasks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    uid CHAR(26) NOT NULL,

    task_type VARCHAR(64) NOT NULL,
    market VARCHAR(16) NOT NULL,
    asset_type VARCHAR(32) NOT NULL DEFAULT 'STOCK',
    data_source VARCHAR(32) NOT NULL,

    status VARCHAR(32) NOT NULL,

    total_items BIGINT NOT NULL DEFAULT 0,
    processed_items BIGINT NOT NULL DEFAULT 0,
    success_items BIGINT NOT NULL DEFAULT 0,
    failed_items BIGINT NOT NULL DEFAULT 0,

    request_id VARCHAR(128) NOT NULL DEFAULT '',

    started_at DATETIME(3) NULL,
    finished_at DATETIME(3) NULL,
    error_msg TEXT NULL,

    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    UNIQUE KEY uk_sync_tasks_uid (uid),
    INDEX idx_sync_tasks_status (status),
    INDEX idx_sync_tasks_created_at (created_at)
);
```

### 13.5 sync_logs

```sql
CREATE TABLE sync_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,

    task_id BIGINT UNSIGNED NOT NULL,
    task_uid CHAR(26) NOT NULL,

    step VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,

    market VARCHAR(16) NOT NULL DEFAULT '',
    asset_type VARCHAR(32) NOT NULL DEFAULT '',
    symbol VARCHAR(32) NOT NULL DEFAULT '',
    data_source VARCHAR(32) NOT NULL DEFAULT '',

    message TEXT NULL,
    error_detail TEXT NULL,
    affected_rows BIGINT NOT NULL DEFAULT 0,

    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    INDEX idx_sync_logs_task_id (task_id),
    INDEX idx_sync_logs_task_uid (task_uid),
    INDEX idx_sync_logs_created_at (created_at)
);
```

不加数据库外键约束。

关联关系靠：

```text
业务唯一键
内部 id
Repository 查询逻辑
```

---

## 14. ID 策略

采用三层 ID 策略：

```text
1. 内部物理主键：
   BIGINT UNSIGNED AUTO_INCREMENT

2. 对外暴露 ID：
   ULID，CHAR(26)

3. 业务唯一键：
   stocks: market + asset_type + symbol
   daily_k_lines: market + asset_type + symbol + trade_date
```

规则：

```text
API 不暴露内部自增 id。
sync_tasks 对外使用 uid。
request_id 可以使用 ULID。
```

示例：

```json
{
  "task_id": "01J2Z3ABCDEF123456789XYZAB",
  "status": "PENDING"
}
```

---

## 15. Migration 策略

使用：

```text
goose + SQL migration 文件
```

不使用：

```text
GORM AutoMigrate 作为正式建表方案
```

服务启动时：

```text
不自动执行 migration
只检查数据库连接
```

手动命令：

```bash
make migrate-up
make migrate-down
make migrate-status
```

---

## 16. Upsert 策略

所有同步写入使用：

```text
批量 upsert
batch_size 默认 500
```

配置：

```yaml
market:
  sync:
    batch_size: 500
```

冲突时更新：

### 16.1 stocks 冲突更新

更新：

```text
name
exchange
board
area
industry
status
list_date
delist_date
data_source
synced_at
updated_at
```

不更新：

```text
id
market
asset_type
symbol
created_at
```

### 16.2 daily_k_lines 冲突更新

更新：

```text
open_price
high_price
low_price
close_price
pre_close
change_amt
pct_change
volume
amount
data_source
synced_at
updated_at
```

不更新：

```text
id
market
asset_type
symbol
trade_date
created_at
```

### 16.3 trade_calendars 冲突更新

更新：

```text
is_open
pretrade_date
data_source
synced_at
updated_at
```

不更新：

```text
id
market
exchange
cal_date
created_at
```

---

## 17. HTTP API 规范

### 17.1 通用响应结构

响应体：

```json
{
  "code": "OK",
  "message": "success",
  "request_id": "req_xxxxx",
  "data": {}
}
```

错误响应：

```json
{
  "code": "MARKET_STOCK_NOT_FOUND",
  "message": "stock not found",
  "request_id": "req_xxxxx",
  "data": null
}
```

规则：

```text
HTTP 状态码走协议层。
业务 code 使用字符串。
响应 body 不额外返回 http_status。
```

### 17.2 Decimal 返回规则

所有 decimal 字段在 HTTP 中返回 string。

### 17.3 API 前缀

```text
/api/v1
```

### 17.4 market / asset_type 参数

API path 不显式带 market / asset_type。

使用 query 参数：

```text
market 默认 CN
asset_type 默认 STOCK
```

示例：

```http
GET /api/v1/market/stocks/600519
```

等价于：

```http
GET /api/v1/market/stocks/600519?market=CN&asset_type=STOCK
```

---

## 18. API 清单

### 18.1 股票查询

```http
GET /api/v1/market/stocks
GET /api/v1/market/stocks/{symbol}
GET /api/v1/market/stocks/{symbol}/daily-k-lines
GET /api/v1/market/stocks/{symbol}/latest-daily-k-line
```

#### 股票列表

```http
GET /api/v1/market/stocks?page=1&page_size=20&keyword=茅台&exchange=SSE&status=LISTED
```

参数：

```text
page 默认 1
page_size 默认 20
page_size 最大 100
keyword 支持 symbol / name LIKE 搜索
exchange 可选
status 可选
market 默认 CN
asset_type 默认 STOCK
```

返回：

```json
{
  "items": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

股票列表第一阶段只返回股票基础信息，不默认返回最新行情。

#### 股票详情

```http
GET /api/v1/market/stocks/600519
```

返回：

```json
{
  "symbol": "600519",
  "name": "贵州茅台",
  "market": "CN",
  "asset_type": "STOCK",
  "exchange": "SSE",
  "board": "MAIN",
  "area": "贵州",
  "industry": "白酒",
  "status": "LISTED",
  "latest_daily_k_line": {
    "trade_date": "2026-06-25",
    "open": "1500.0000",
    "high": "1510.0000",
    "low": "1490.0000",
    "close": "1505.0000",
    "pre_close": "1498.0000",
    "change": "7.0000",
    "pct_change": "0.4673",
    "pct_change_unit": "percent",
    "volume": "123456.0000",
    "volume_unit": "lot",
    "amount": "987654321.0000",
    "amount_unit": "thousand_cny"
  }
}
```

如果未同步日 K：

```json
{
  "latest_daily_k_line": null
}
```

#### 日 K 查询

```http
GET /api/v1/market/stocks/600519/daily-k-lines?start_date=2024-01-01&end_date=2026-06-25
```

规则：

```text
默认返回最近 1 年。
最大查询范围 5 年。
不分页。
按 trade_date ASC 正序返回。
```

超过 5 年返回：

```text
MARKET_DATE_RANGE_TOO_LARGE
```

#### 最新日 K 查询

```http
GET /api/v1/market/stocks/600519/latest-daily-k-line
```

内部查询：

```sql
ORDER BY trade_date DESC
LIMIT 1
```

---

### 18.2 交易日历查询

```http
GET /api/v1/market/trade-calendars
GET /api/v1/market/trade-calendars/latest-open-day
GET /api/v1/market/trade-calendars/is-open
```

#### 交易日历列表

```http
GET /api/v1/market/trade-calendars?start_date=2026-01-01&end_date=2026-12-31
```

#### 最近开市交易日

```http
GET /api/v1/market/trade-calendars/latest-open-day
```

#### 某日是否开市

```http
GET /api/v1/market/trade-calendars/is-open?date=2026-06-25
```

不传 `date` 时默认今天。

---

### 18.3 同步接口

同步接口需要 Admin Token。

```http
POST /api/v1/market/sync/trade-calendars
POST /api/v1/market/sync/stocks
POST /api/v1/market/sync/daily-k-lines
GET  /api/v1/market/sync/tasks/{task_uid}
GET  /api/v1/market/sync/tasks/{task_uid}/logs
```

暂不实现：

```http
POST /api/v1/market/sync/tasks/{task_uid}/cancel
```

#### 同步交易日历

```http
POST /api/v1/market/sync/trade-calendars
```

默认范围：

```text
最近 5 年 + 未来 1 年
```

#### 同步股票基础信息

```http
POST /api/v1/market/sync/stocks
```

同步：

```text
LISTED
DELISTED
PAUSED
```

拉取策略：

```text
分别按 Tushare L / D / P 拉取
转换为 LISTED / DELISTED / PAUSED
合并后统一 upsert
```

任意一个状态拉取失败：

```text
任务 FAILED
```

不使用：

```text
PARTIAL_SUCCESS
```

同步策略：

```text
只 upsert
不 delete
不 soft delete
```

#### 同步日 K

```http
POST /api/v1/market/sync/daily-k-lines
```

请求体示例：

```json
{
  "market": "CN",
  "asset_type": "STOCK",
  "symbol": "600519",
  "start_date": "2023-01-01",
  "end_date": "2026-06-25"
}
```

规则：

```text
symbol 为空：
全市场同步

symbol 非空：
单只股票同步
```

默认范围：

```text
start_date 为空：最近 3 年
end_date 为空：最近一个已开市交易日
```

依赖：

```text
必须先同步 trade_calendars。
必须先同步 stocks。
```

如果交易日历为空：

```text
MARKET_TRADE_CALENDAR_NOT_INITIALIZED
```

如果全市场同步时 stocks 为空：

```text
MARKET_STOCKS_NOT_INITIALIZED
```

如果单只股票不存在：

```text
MARKET_STOCK_NOT_FOUND
```

范围限制：

```text
全市场同步：
默认最近 3 年
最大 5 年

单只股票同步：
默认最近 3 年
最大 20 年
```

全市场日 K 同步：

```text
默认只同步 LISTED 股票
串行按股票同步
单只股票失败不阻断整体任务
最终可为 PARTIAL_SUCCESS
```

---

## 19. 同步任务执行模型

同步任务采用：

```text
HTTP 提交
    ↓
创建 sync_tasks
    ↓
后台 goroutine 异步执行
    ↓
写 sync_logs
    ↓
更新任务状态和进度
```

Handler 不直接启动 goroutine。

由 `SyncService` 负责：

```text
检查任务冲突
创建任务
启动 goroutine
执行同步
recover panic
更新状态
写日志
```

### 19.1 全局单任务限制

第一阶段全局只允许一个同步任务执行。

如果已有：

```text
PENDING
RUNNING
```

则新任务返回：

```text
MARKET_SYNC_TASK_CONFLICT
```

实现方式：

```text
数据库状态控制 + 进程内 mutex
```

不使用：

```text
Redis 分布式锁
etcd lock
MySQL GET_LOCK
任务队列
```

### 19.2 服务启动清理

服务启动时：

```text
PENDING → FAILED
RUNNING → FAILED
```

错误原因：

```text
service restarted during task execution
```

### 19.3 Context 策略

异步任务不直接使用 HTTP request ctx。

使用：

```text
appCtx 派生 context
```

服务关闭时：

```text
cancel appCtx
后台任务尽量停止
任务标记 FAILED
```

### 19.4 Panic 处理

同步 goroutine 必须 recover panic。

panic 后：

```text
任务状态 FAILED
写 sync_logs
记录 error_msg
日志包含 task_uid / request_id
```

### 19.5 事务边界

全市场日 K 同步不使用全任务大事务。

采用：

```text
单只股票级事务
```

流程：

```text
拉取某只股票日 K
转换领域模型
开启短事务
批量 upsert
更新任务进度
写 sync_logs
提交事务
```

---

## 20. Tushare Client 设计

### 20.1 pkg/tushare 边界

`pkg/tushare` 只做 Tushare 原始 HTTP Client。

负责：

```text
构造请求
处理 token
处理 base_url
处理 timeout
调用 api_name
解析 fields/items
返回原始响应结构
```

不负责：

```text
转换 market
转换 exchange
转换 board
转换 status
转换 time.Time
转换 decimal.Decimal
返回 market.Stock
```

领域转换放在：

```text
domain/market/provider_tushare.go
```

依赖方向：

```text
domain/market/provider_tushare.go
        ↓
pkg/tushare
```

### 20.2 Client 方法

底层：

```go
func (c *Client) Call(ctx context.Context, req CallRequest) (*CallResponse, error)
```

第一阶段薄封装：

```go
func (c *Client) StockBasic(ctx context.Context, req StockBasicRequest) (*CallResponse, error)
func (c *Client) Daily(ctx context.Context, req DailyRequest) (*CallResponse, error)
func (c *Client) TradeCal(ctx context.Context, req TradeCalRequest) (*CallResponse, error)
```

### 20.3 显式 fields

调用 Tushare 时必须显式指定 fields，不依赖默认返回字段。

第一阶段字段：

```text
StockBasic:
ts_code, symbol, name, area, industry, market, list_date, delist_date, list_status

Daily:
ts_code, trade_date, open, high, low, close, pre_close, change, pct_chg, vol, amount

TradeCal:
exchange, cal_date, is_open, pretrade_date
```

### 20.4 重试与限速

配置：

```yaml
tushare:
  base_url: "https://api.tushare.pro"
  token: "${TUSHARE_TOKEN}"
  timeout_seconds: 10
  max_retries: 3
  retry_backoff_seconds: 1
  requests_per_minute: 40
```

规则：

```text
最多重试 3 次
退避 1s / 2s / 4s
重试也经过限速器
默认每分钟最多 40 次请求
后续根据账号积分手动调大
```

不可重试错误：

```text
token 无效
参数错误
权限不足
接口不存在
```

可重试错误：

```text
网络超时
HTTP 5xx
临时连接失败
临时服务异常
```

### 20.5 Token 管理

真实 token 不写入 Git。

配置：

```yaml
tushare:
  token: "${TUSHARE_TOKEN}"
```

运行环境：

```bash
export TUSHARE_TOKEN="真实 token"
```

规则：

```text
market.provider = mock：
不要求 TUSHARE_TOKEN

market.provider = tushare：
TUSHARE_TOKEN 为空则启动失败
```

---

## 21. HTTP 中间件

### 21.1 Request ID

由 middleware 统一处理：

```text
读取 X-Request-ID
没有则自动生成
写入响应 Header
写入响应 body
写入日志上下文
```

响应 Header：

```http
X-Request-ID: req_xxx
```

### 21.2 Admin Token

同步接口需要 Admin Token。

配置：

```yaml
admin:
  token: "${ADMIN_TOKEN}"
```

请求头：

```http
Authorization: Bearer your-admin-token
```

第一阶段不实现完整 auth 领域。

### 21.3 Access Log

使用 slog 记录结构化请求日志。

字段：

```text
request_id
method
path
status
latency
client_ip
```

---

## 22. 错误码规范

错误码按领域管理。

Market 错误码放在：

```text
domain/market/errors.go
```

示例：

```text
MARKET_STOCK_NOT_FOUND
MARKET_STOCKS_NOT_INITIALIZED
MARKET_INVALID_SYMBOL
MARKET_INVALID_DATE_RANGE
MARKET_DATE_RANGE_TOO_LARGE
MARKET_TRADE_CALENDAR_NOT_INITIALIZED
MARKET_SYNC_TASK_CONFLICT
MARKET_SYNC_TASK_NOT_FOUND
MARKET_PROVIDER_FAILED
MARKET_REPOSITORY_FAILED
```

通用错误：

```text
OK
INVALID_ARGUMENT
UNAUTHORIZED
INTERNAL_ERROR
```

HTTP 状态码映射：

```text
OK                                  → 200
INVALID_ARGUMENT                    → 400
MARKET_INVALID_SYMBOL               → 400
MARKET_INVALID_DATE_RANGE           → 400
MARKET_DATE_RANGE_TOO_LARGE         → 400
UNAUTHORIZED                        → 401
MARKET_STOCK_NOT_FOUND              → 404
MARKET_SYNC_TASK_NOT_FOUND          → 404
MARKET_SYNC_TASK_CONFLICT           → 409
MARKET_PROVIDER_FAILED              → 502
MARKET_REPOSITORY_FAILED            → 500
INTERNAL_ERROR                      → 500
```

---

## 23. 配置加载

第一阶段不引入 Viper。

使用：

```text
显式 Config struct
YAML 配置
os.ExpandEnv 替换 ${ENV_NAME}
gopkg.in/yaml.v3
```

示例配置：

```yaml
server:
  addr: ":8080"

database:
  dsn: "${MYSQL_DSN}"

admin:
  token: "${ADMIN_TOKEN}"

market:
  provider: "mock"
  sync:
    batch_size: 500

tushare:
  base_url: "https://api.tushare.pro"
  token: "${TUSHARE_TOKEN}"
  timeout_seconds: 10
  max_retries: 3
  retry_backoff_seconds: 1
  requests_per_minute: 40
```

---

## 24. 日志方案

使用 Go 标准库：

```text
log/slog
```

封装在：

```text
pkg/logger
```

统一字段：

```text
request_id
task_uid
market
asset_type
symbol
data_source
error
```

禁止在业务代码中大量使用：

```text
fmt.Println
log.Println
```

注意：

```text
Tushare client 不允许打印 token。
```

---

## 25. HTTP 框架和数据库访问

HTTP 框架：

```text
Gin
```

MySQL 访问层：

```text
GORM
```

边界：

```text
GORM 只允许出现在 repository_mysql.go / 数据库初始化代码中。
service / handler 不直接依赖 *gorm.DB。
```

---

## 26. 健康检查

提供：

```http
GET /healthz
GET /readyz
```

规则：

```text
/healthz:
只判断服务进程是否活着
不检查 MySQL
不检查 Tushare

/readyz:
检查服务是否准备好
第一阶段只检查 MySQL
不调用 Tushare
```

---

## 27. 枚举字典接口

提供：

```http
GET /api/v1/market/meta/enums
```

返回：

```json
{
  "markets": [
    { "code": "CN", "name": "A股" }
  ],
  "asset_types": [
    { "code": "STOCK", "name": "股票" }
  ],
  "exchanges": [
    { "code": "SSE", "name": "上交所" },
    { "code": "SZSE", "name": "深交所" },
    { "code": "BSE", "name": "北交所" }
  ],
  "boards": [
    { "code": "MAIN", "name": "主板" },
    { "code": "GEM", "name": "创业板" },
    { "code": "STAR", "name": "科创板" },
    { "code": "BSE", "name": "北交所" },
    { "code": "UNKNOWN", "name": "未知" }
  ],
  "stock_statuses": [
    { "code": "LISTED", "name": "上市" },
    { "code": "DELISTED", "name": "退市" },
    { "code": "PAUSED", "name": "暂停上市" }
  ]
}
```

实现方式：

```text
使用 domain/market 常量
不建数据库字典表
不做后台维护
不做多语言
```

---

## 28. 测试策略

第一阶段需要测试，但不追求 100% 覆盖率。

### 28.1 Service 单元测试

使用 Mock Repository / Mock Provider。

测试重点：

```text
查询参数默认值
日期范围校验
market / asset_type 校验
symbol 校验
任务冲突
股票不存在
交易日历未初始化
PARTIAL_SUCCESS
```

### 28.2 Provider 转换测试

不访问真实 Tushare。

使用构造好的原始响应样例。

测试：

```text
ts_code → symbol / exchange
L/D/P → LISTED/DELISTED/PAUSED
board 推导
YYYYMMDD → time.Time
decimal 字段转换
Tushare 单位保持
```

### 28.3 Repository 集成测试

使用：

```text
Docker MySQL + goose migration + integration build tag
```

命令：

```bash
go test -tags=integration ./...
```

验证：

```text
stocks upsert
daily_k_lines upsert
trade_calendars upsert
sync_tasks / sync_logs 查询
DECIMAL
DATE
DATETIME(3)
UNIQUE KEY
ON DUPLICATE KEY UPDATE
```

普通单元测试：

```bash
go test ./...
```

---

## 29. 本地开发环境

提供：

```text
docker-compose.yml
```

第一阶段只包含：

```text
mysql
```

暂不加入：

```text
redis
prometheus
grafana
adminer
phpMyAdmin
```

本地启动流程：

```bash
make docker-up
make migrate-up
make run
```

---

## 30. Makefile

提供命令：

```makefile
run
test
test-integration
migrate-up
migrate-down
migrate-status
docker-up
docker-down
```

推荐流程：

```bash
make docker-up
make migrate-up
make run
make test
make test-integration
```

---

## 31. GitHub Actions / CI

第一阶段提供最小 CI：

```text
.github/workflows/ci.yml
```

执行：

```text
go test ./...
go vet ./...
gofmt 检查
```

暂不执行：

```text
集成测试
真实 MySQL 测试
真实 Tushare 测试
Docker 镜像构建
自动部署
```

---

## 32. OpenAPI 和接口文档

第一阶段必须维护：

```text
specs/openapi.yaml
specs/market-api.md
```

`openapi.yaml`：

```text
机器可读
给前端、测试工具、AI 使用
描述路径、参数、响应、错误码
```

`market-api.md`：

```text
人类可读
解释接口设计、字段含义、单位、默认值、限制
```

必须写清楚：

```text
decimal 字段返回 string
volume 单位是 lot / 手
amount 单位是 thousand_cny / 千元
pct_change 单位是 percent
symbol 是 6 位 A 股代码，不是 ts_code
market 默认 CN
asset_type 默认 STOCK
统一响应 code/message/request_id/data
```

---

## 33. 第一阶段实现顺序

推荐按以下顺序实现：

```text
1. 项目骨架
   cmd/server
   api/route.go
   pkg/config
   pkg/logger
   pkg/httputil
   pkg/middleware

2. migration
   stocks
   daily_k_lines
   trade_calendars
   sync_tasks
   sync_logs

3. domain/market 模型和接口
   model.go
   port.go
   errors.go
   dto.go

4. Repository
   repository_mysql.go
   repository_mock.go
   TxManager

5. Provider
   provider_mock.go
   pkg/tushare
   provider_tushare.go

6. QueryService + Handler 查询接口
   股票列表
   股票详情
   日 K
   最新日 K
   交易日历
   枚举字典

7. SyncService + 同步接口
   sync trade calendars
   sync stocks
   sync daily k lines
   sync tasks
   sync logs

8. 测试
   service 单元测试
   provider 转换测试
   repository 集成测试

9. 规约文档
   openapi.yaml
   market-api.md

10. Makefile / Docker / CI
```

---

## 34. 后续扩展点

第一阶段预留但不实现：

```text
1. 多数据源
   后续新增 AKShareProvider / BaoStockProvider。
   多数据源映射表可设计 instrument_source_mappings。

2. 多市场
   后续 market 支持 HK / US。

3. 多资产类型
   后续 asset_type 支持 ETF / INDEX / FUND / BOND / FUTURE。

4. 复权行情
   后续新增 adj_factors 表和 adjust=qfq/hfq 查询参数。

5. 批量最新行情
   后续新增 GET /api/v1/market/stocks/quotes?symbols=...

6. 定时同步
   后续新增每日收盘后自动同步。

7. 任务取消
   后续新增 CANCELING / CANCELED。

8. 并发同步
   后续从串行同步升级为固定 worker pool + 全局 Tushare limiter。

9. ClickHouse
   后续如果行情和回测数据量变大，引入 ClickHouse 作为分析库。
   MySQL 继续保存主数据、任务、配置。

10. 技术指标
    后续新增 indicator 相关模型和接口。

11. 量化策略
    后续新增 strategy / backtest / signal 相关领域。
```

---

## 35. 冻结决策摘要

第一阶段最终冻结为：

```text
架构：
DDF，domain/market 单领域优先。

市场：
只支持 A 股，market=CN。

资产：
只支持股票，asset_type=STOCK。

数据源：
先接 Tushare，Provider 层预留扩展。

存储：
MySQL + GORM Repository。

迁移：
goose + SQL migration，服务启动不自动迁移。

时间：
DATE + DATETIME(3) + Go time.Time，不使用 TIMESTAMP。

数值：
decimal.Decimal + MySQL DECIMAL，HTTP 返回 string。

ID：
内部自增 id + 对外 ULID + 业务唯一键。

同步：
手动触发，后台 goroutine 异步执行，全局单任务。

任务：
支持 PENDING / RUNNING / SUCCESS / FAILED / PARTIAL_SUCCESS。

日志：
任务步骤级 + 股票级日志，不记录每条 K 线。

Tushare：
显式 fields，轻量重试，默认 requests_per_minute=40。

API：
/api/v1，统一 code/message/request_id/data。

认证：
查询接口开放，同步接口 Admin Token。

测试：
Service 单测 + Provider 转换测试 + Repository 集成测试。

文档：
openapi.yaml + market-api.md。

CI：
最小 GitHub Actions。
```
