package market

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const DateLayout = "2006-01-02"

type Market string

const (
	MarketCN Market = "CN"
)

func (m Market) Valid() bool {
	return m == MarketCN
}

type AssetType string

const (
	AssetTypeStock AssetType = "STOCK"
)

func (a AssetType) Valid() bool {
	return a == AssetTypeStock
}

type Exchange string

const (
	ExchangeSSE  Exchange = "SSE"
	ExchangeSZSE Exchange = "SZSE"
	ExchangeBSE  Exchange = "BSE"
)

func (e Exchange) Valid() bool {
	switch e {
	case ExchangeSSE, ExchangeSZSE, ExchangeBSE:
		return true
	default:
		return false
	}
}

type Board string

const (
	BoardMain    Board = "MAIN"
	BoardGEM     Board = "GEM"
	BoardSTAR    Board = "STAR"
	BoardBSE     Board = "BSE"
	BoardUnknown Board = "UNKNOWN"
)

type StockStatus string

const (
	StockStatusListed   StockStatus = "LISTED"
	StockStatusDelisted StockStatus = "DELISTED"
	StockStatusPaused   StockStatus = "PAUSED"
)

func (s StockStatus) Valid() bool {
	switch s {
	case StockStatusListed, StockStatusDelisted, StockStatusPaused:
		return true
	default:
		return false
	}
}

type DataSource string

const (
	DataSourceMock    DataSource = "MOCK"
	DataSourceTushare DataSource = "TUSHARE"
)

func (d DataSource) Valid() bool {
	switch d {
	case DataSourceMock, DataSourceTushare:
		return true
	default:
		return false
	}
}

type SyncTaskStatus string

const (
	SyncTaskStatusPending        SyncTaskStatus = "PENDING"
	SyncTaskStatusRunning        SyncTaskStatus = "RUNNING"
	SyncTaskStatusSuccess        SyncTaskStatus = "SUCCESS"
	SyncTaskStatusFailed         SyncTaskStatus = "FAILED"
	SyncTaskStatusPartialSuccess SyncTaskStatus = "PARTIAL_SUCCESS"
)

func (s SyncTaskStatus) Valid() bool {
	switch s {
	case SyncTaskStatusPending, SyncTaskStatusRunning, SyncTaskStatusSuccess, SyncTaskStatusFailed, SyncTaskStatusPartialSuccess:
		return true
	default:
		return false
	}
}

type SyncTaskType string

const (
	SyncTaskTypeStockMaster    SyncTaskType = "SYNC_STOCK_MASTER"
	SyncTaskTypeTradeCalendars SyncTaskType = "SYNC_TRADE_CALENDARS"
	SyncTaskTypeDailyKLines    SyncTaskType = "SYNC_DAILY_K_LINES"
)

type SyncLogStatus string

const (
	SyncLogStatusSuccess SyncLogStatus = "SUCCESS"
	SyncLogStatusFailed  SyncLogStatus = "FAILED"
	SyncLogStatusWarning SyncLogStatus = "WARNING"
)

func (s SyncLogStatus) Valid() bool {
	switch s {
	case SyncLogStatusSuccess, SyncLogStatusFailed, SyncLogStatusWarning:
		return true
	default:
		return false
	}
}

type Stock struct {
	Market     Market
	AssetType  AssetType
	Symbol     string
	TSCode     string
	Name       string
	Exchange   Exchange
	Board      Board
	Area       string
	Industry   string
	Status     StockStatus
	ListDate   *time.Time
	DelistDate *time.Time
	DataSource DataSource
	SyncedAt   time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type DailyKLine struct {
	Market     Market
	AssetType  AssetType
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
	DataSource DataSource
	SyncedAt   time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type TradeCalendar struct {
	Market       Market
	Exchange     Exchange
	CalDate      time.Time
	IsOpen       bool
	PretradeDate *time.Time
	DataSource   DataSource
	SyncedAt     time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SyncTask struct {
	UID            string
	TaskType       SyncTaskType
	Market         Market
	AssetType      AssetType
	DataSource     DataSource
	Status         SyncTaskStatus
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

type SyncLog struct {
	TaskUID      string
	Step         string
	Status       SyncLogStatus
	Market       Market
	AssetType    AssetType
	Symbol       string
	DataSource   DataSource
	Message      string
	ErrorDetail  string
	AffectedRows int64
	CreatedAt    time.Time
}

func ValidSymbol(symbol string) bool {
	if len(symbol) != 6 {
		return false
	}
	for _, r := range symbol {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func ParseDate(value string) (time.Time, error) {
	date, err := time.Parse(DateLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", value, err)
	}
	return date, nil
}

func ValidDateRange(start time.Time, end time.Time) bool {
	return !start.After(end)
}
