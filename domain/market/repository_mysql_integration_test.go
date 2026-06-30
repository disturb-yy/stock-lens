//go:build integration

package market

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestMySQLRepositoriesIntegration(t *testing.T) {
	db := openIntegrationDB(t)
	truncateMarketTables(t, db)

	ctx := context.Background()
	stocks := NewMySQLStockRepository(db)
	kLines := NewMySQLKLineRepository(db)
	calendars := NewMySQLTradeCalendarRepository(db)
	tasks := NewMySQLSyncTaskRepository(db)
	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	tradeDate := dateOnly(now)

	if err := stocks.UpsertStocks(ctx, []Stock{{
		Market:     MarketCN,
		AssetType:  AssetTypeStock,
		Symbol:     "600519",
		TSCode:     "600519.SH",
		Name:       "Old Name",
		Exchange:   ExchangeSSE,
		Board:      BoardMain,
		Status:     StockStatusListed,
		DataSource: DataSourceMock,
		SyncedAt:   now,
	}}); err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}
	if err := stocks.UpsertStocks(ctx, []Stock{{
		Market:     MarketCN,
		AssetType:  AssetTypeStock,
		Symbol:     "600519",
		TSCode:     "600519.SH",
		Name:       "Kweichow Moutai",
		Exchange:   ExchangeSSE,
		Board:      BoardMain,
		Status:     StockStatusListed,
		DataSource: DataSourceMock,
		SyncedAt:   now,
	}}); err != nil {
		t.Fatalf("UpsertStocks() update error = %v", err)
	}
	stock, err := stocks.FindStock(ctx, FindStockQuery{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519"})
	if err != nil {
		t.Fatalf("FindStock() error = %v", err)
	}
	if stock.Name != "Kweichow Moutai" {
		t.Fatalf("stock name = %q, want updated name", stock.Name)
	}

	if err := kLines.UpsertDailyKLines(ctx, []DailyKLine{{
		Market:     MarketCN,
		AssetType:  AssetTypeStock,
		Symbol:     "600519",
		TradeDate:  tradeDate,
		OpenPrice:  decimal.NewFromInt(10),
		HighPrice:  decimal.NewFromInt(12),
		LowPrice:   decimal.NewFromInt(9),
		ClosePrice: decimal.NewFromInt(11),
		DataSource: DataSourceMock,
		SyncedAt:   now,
	}}); err != nil {
		t.Fatalf("UpsertDailyKLines() error = %v", err)
	}
	latest, err := kLines.LatestDailyKLine(ctx, MarketCN, AssetTypeStock, "600519")
	if err != nil {
		t.Fatalf("LatestDailyKLine() error = %v", err)
	}
	if !latest.TradeDate.Equal(tradeDate) {
		t.Fatalf("latest date = %s, want %s", latest.TradeDate, tradeDate)
	}

	if err := calendars.UpsertTradeCalendars(ctx, []TradeCalendar{{
		Market:     MarketCN,
		Exchange:   ExchangeSSE,
		CalDate:    tradeDate,
		IsOpen:     true,
		DataSource: DataSourceMock,
		SyncedAt:   now,
	}}); err != nil {
		t.Fatalf("UpsertTradeCalendars() error = %v", err)
	}
	openDay, err := calendars.LatestOpenDay(ctx, MarketCN, ExchangeSSE)
	if err != nil {
		t.Fatalf("LatestOpenDay() error = %v", err)
	}
	if !openDay.Equal(tradeDate) {
		t.Fatalf("open day = %s, want %s", openDay, tradeDate)
	}

	task := SyncTask{
		UID:        "01J2Z3ABCDEF123456789XYZAB",
		TaskType:   SyncTaskTypeStockMaster,
		Market:     MarketCN,
		AssetType:  AssetTypeStock,
		DataSource: DataSourceMock,
		Status:     SyncTaskStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := tasks.CreateSyncTask(ctx, task); err != nil {
		t.Fatalf("CreateSyncTask() error = %v", err)
	}
	task.Status = SyncTaskStatusSuccess
	if err := tasks.UpdateSyncTask(ctx, task); err != nil {
		t.Fatalf("UpdateSyncTask() error = %v", err)
	}
	if err := tasks.AppendSyncLogs(ctx, []SyncLog{{
		TaskUID:      task.UID,
		Step:         "stock_master",
		Status:       SyncLogStatusSuccess,
		Market:       MarketCN,
		AssetType:    AssetTypeStock,
		DataSource:   DataSourceMock,
		Message:      "ok",
		AffectedRows: 1,
		CreatedAt:    now,
	}}); err != nil {
		t.Fatalf("AppendSyncLogs() error = %v", err)
	}
	logs, err := tasks.ListSyncLogs(ctx, ListSyncLogsQuery{
		PageQuery: PageQuery{Page: 1, PageSize: 20},
		TaskUID:   task.UID,
	})
	if err != nil {
		t.Fatalf("ListSyncLogs() error = %v", err)
	}
	if logs.Pagination.Total != 1 || len(logs.Items) != 1 {
		t.Fatalf("logs = %+v, want one log", logs)
	}
}

func openIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN is required for integration tests")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	return db
}

func truncateMarketTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	for _, table := range []string{"sync_logs", "sync_tasks", "trade_calendars", "daily_k_lines", "stocks"} {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
}
