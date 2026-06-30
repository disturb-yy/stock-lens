package market

import (
	"context"
	"time"
)

type PageQuery struct {
	Page     int
	PageSize int
}

type PageResult struct {
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
}

type ListStocksQuery struct {
	PageQuery
	Market    Market
	AssetType AssetType
	Keyword   string
	Exchange  Exchange
	Status    StockStatus
}

type FindStockQuery struct {
	Market    Market
	AssetType AssetType
	Symbol    string
}

type ListStocksResult struct {
	Items      []Stock
	Pagination PageResult
}

type ListDailyKLinesQuery struct {
	Market    Market
	AssetType AssetType
	Symbol    string
	StartDate time.Time
	EndDate   time.Time
}

type ListTradeCalendarsQuery struct {
	Market    Market
	Exchange  Exchange
	StartDate time.Time
	EndDate   time.Time
}

type ListSyncTasksQuery struct {
	PageQuery
	Status SyncTaskStatus
}

type ListSyncTasksResult struct {
	Items      []SyncTask
	Pagination PageResult
}

type ListSyncLogsQuery struct {
	PageQuery
	TaskUID string
	Status  SyncLogStatus
	Symbol  string
}

type ListSyncLogsResult struct {
	Items      []SyncLog
	Pagination PageResult
}

type StockRepository interface {
	UpsertStocks(ctx context.Context, stocks []Stock) error
	ListStocks(ctx context.Context, query ListStocksQuery) (ListStocksResult, error)
	FindStock(ctx context.Context, query FindStockQuery) (Stock, error)
}

type KLineRepository interface {
	UpsertDailyKLines(ctx context.Context, lines []DailyKLine) error
	ListDailyKLines(ctx context.Context, query ListDailyKLinesQuery) ([]DailyKLine, error)
	LatestDailyKLine(ctx context.Context, market Market, assetType AssetType, symbol string) (DailyKLine, error)
}

type TradeCalendarRepository interface {
	UpsertTradeCalendars(ctx context.Context, calendars []TradeCalendar) error
	ListTradeCalendars(ctx context.Context, query ListTradeCalendarsQuery) ([]TradeCalendar, error)
	LatestOpenDay(ctx context.Context, market Market, exchange Exchange) (time.Time, error)
	IsOpenDay(ctx context.Context, market Market, exchange Exchange, date time.Time) (bool, error)
}

type SyncTaskRepository interface {
	CreateSyncTask(ctx context.Context, task SyncTask) error
	UpdateSyncTask(ctx context.Context, task SyncTask) error
	FindSyncTask(ctx context.Context, uid string) (SyncTask, error)
	FindActiveSyncTask(ctx context.Context) (SyncTask, error)
	ListSyncTasks(ctx context.Context, query ListSyncTasksQuery) (ListSyncTasksResult, error)
	MarkStaleTasksFailed(ctx context.Context) error
	AppendSyncLogs(ctx context.Context, logs []SyncLog) error
	ListSyncLogs(ctx context.Context, query ListSyncLogsQuery) (ListSyncLogsResult, error)
}

type TxRepositories interface {
	Stocks() StockRepository
	KLines() KLineRepository
	TradeCalendars() TradeCalendarRepository
	SyncTasks() SyncTaskRepository
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, repos TxRepositories) error) error
}

type FetchStocksRequest struct {
	Market Market
}

type FetchTradeCalendarsRequest struct {
	Market    Market
	Exchange  Exchange
	StartDate time.Time
	EndDate   time.Time
}

type FetchDailyKLinesRequest struct {
	Market    Market
	AssetType AssetType
	Symbol    string
	StartDate time.Time
	EndDate   time.Time
}

type InstrumentProvider interface {
	FetchStocks(ctx context.Context, req FetchStocksRequest) ([]Stock, error)
}

type CalendarProvider interface {
	FetchTradeCalendars(ctx context.Context, req FetchTradeCalendarsRequest) ([]TradeCalendar, error)
}

type MarketDataProvider interface {
	FetchDailyKLines(ctx context.Context, req FetchDailyKLinesRequest) ([]DailyKLine, error)
}
