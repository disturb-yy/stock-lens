package market

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestMockStockRepositoryUpsertAndList(t *testing.T) {
	ctx := context.Background()
	repo := NewMockStockRepository()

	err := repo.UpsertStocks(ctx, []Stock{
		{
			Market:     MarketCN,
			AssetType:  AssetTypeStock,
			Symbol:     "600519",
			Name:       "Old Name",
			Exchange:   ExchangeSSE,
			Status:     StockStatusListed,
			DataSource: DataSourceMock,
		},
		{
			Market:     MarketCN,
			AssetType:  AssetTypeStock,
			Symbol:     "000001",
			Name:       "Ping An Bank",
			Exchange:   ExchangeSZSE,
			Status:     StockStatusListed,
			DataSource: DataSourceMock,
		},
	})
	if err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	err = repo.UpsertStocks(ctx, []Stock{
		{
			Market:     MarketCN,
			AssetType:  AssetTypeStock,
			Symbol:     "600519",
			Name:       "Kweichow Moutai",
			Exchange:   ExchangeSSE,
			Status:     StockStatusListed,
			DataSource: DataSourceMock,
		},
	})
	if err != nil {
		t.Fatalf("UpsertStocks() update error = %v", err)
	}

	got, err := repo.ListStocks(ctx, ListStocksQuery{
		PageQuery: PageQuery{Page: 1, PageSize: 10},
		Market:    MarketCN,
		AssetType: AssetTypeStock,
		Exchange:  ExchangeSSE,
		Status:    StockStatusListed,
	})
	if err != nil {
		t.Fatalf("ListStocks() error = %v", err)
	}

	if got.Pagination.Total != 1 {
		t.Fatalf("total = %d, want %d", got.Pagination.Total, 1)
	}
	if len(got.Items) != 1 {
		t.Fatalf("len(items) = %d, want %d", len(got.Items), 1)
	}
	if got.Items[0].Name != "Kweichow Moutai" {
		t.Fatalf("name = %q, want updated name", got.Items[0].Name)
	}
}

func TestMockStockRepositoryListReturnsEmptySlice(t *testing.T) {
	repo := NewMockStockRepository()

	got, err := repo.ListStocks(context.Background(), ListStocksQuery{
		PageQuery: PageQuery{Page: 1, PageSize: 20},
		Market:    MarketCN,
		AssetType: AssetTypeStock,
		Keyword:   "missing",
	})
	if err != nil {
		t.Fatalf("ListStocks() error = %v", err)
	}
	if got.Items == nil {
		t.Fatalf("items = nil, want empty slice")
	}
	if len(got.Items) != 0 {
		t.Fatalf("len(items) = %d, want 0", len(got.Items))
	}
}

func TestMockKLineRepositoryUpsertListAndLatest(t *testing.T) {
	ctx := context.Background()
	repo := NewMockKLineRepository()
	day1 := mustDate(t, "2026-06-29")
	day2 := mustDate(t, "2026-06-30")

	err := repo.UpsertDailyKLines(ctx, []DailyKLine{
		{
			Market:     MarketCN,
			AssetType:  AssetTypeStock,
			Symbol:     "600519",
			TradeDate:  day1,
			ClosePrice: decimal.NewFromInt(10),
			DataSource: DataSourceMock,
		},
		{
			Market:     MarketCN,
			AssetType:  AssetTypeStock,
			Symbol:     "600519",
			TradeDate:  day2,
			ClosePrice: decimal.NewFromInt(11),
			DataSource: DataSourceMock,
		},
	})
	if err != nil {
		t.Fatalf("UpsertDailyKLines() error = %v", err)
	}

	err = repo.UpsertDailyKLines(ctx, []DailyKLine{
		{
			Market:     MarketCN,
			AssetType:  AssetTypeStock,
			Symbol:     "600519",
			TradeDate:  day1,
			ClosePrice: decimal.NewFromInt(12),
			DataSource: DataSourceMock,
		},
	})
	if err != nil {
		t.Fatalf("UpsertDailyKLines() update error = %v", err)
	}

	got, err := repo.ListDailyKLines(ctx, ListDailyKLinesQuery{
		Market:    MarketCN,
		AssetType: AssetTypeStock,
		Symbol:    "600519",
		StartDate: day1,
		EndDate:   day2,
	})
	if err != nil {
		t.Fatalf("ListDailyKLines() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(lines) = %d, want 2", len(got))
	}
	if !got[0].TradeDate.Equal(day1) || !got[0].ClosePrice.Equal(decimal.NewFromInt(12)) {
		t.Fatalf("first line = %+v, want updated day1 line", got[0])
	}

	latest, err := repo.LatestDailyKLine(ctx, MarketCN, AssetTypeStock, "600519")
	if err != nil {
		t.Fatalf("LatestDailyKLine() error = %v", err)
	}
	if !latest.TradeDate.Equal(day2) {
		t.Fatalf("latest date = %s, want %s", latest.TradeDate.Format(DateLayout), day2.Format(DateLayout))
	}
}

func TestMockKLineRepositoryListReturnsEmptySlice(t *testing.T) {
	repo := NewMockKLineRepository()

	got, err := repo.ListDailyKLines(context.Background(), ListDailyKLinesQuery{
		Market:    MarketCN,
		AssetType: AssetTypeStock,
		Symbol:    "600519",
		StartDate: mustDate(t, "2026-06-01"),
		EndDate:   mustDate(t, "2026-06-30"),
	})
	if err != nil {
		t.Fatalf("ListDailyKLines() error = %v", err)
	}
	if got == nil {
		t.Fatalf("lines = nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("len(lines) = %d, want 0", len(got))
	}
}

func TestMockTradeCalendarRepositoryUpsertQueries(t *testing.T) {
	ctx := context.Background()
	repo := NewMockTradeCalendarRepository()
	openDay := mustDate(t, "2026-06-29")
	closedDay := mustDate(t, "2026-06-30")

	err := repo.UpsertTradeCalendars(ctx, []TradeCalendar{
		{
			Market:     MarketCN,
			Exchange:   ExchangeSSE,
			CalDate:    openDay,
			IsOpen:     true,
			DataSource: DataSourceMock,
		},
		{
			Market:     MarketCN,
			Exchange:   ExchangeSSE,
			CalDate:    closedDay,
			IsOpen:     true,
			DataSource: DataSourceMock,
		},
	})
	if err != nil {
		t.Fatalf("UpsertTradeCalendars() error = %v", err)
	}

	err = repo.UpsertTradeCalendars(ctx, []TradeCalendar{
		{
			Market:     MarketCN,
			Exchange:   ExchangeSSE,
			CalDate:    closedDay,
			IsOpen:     false,
			DataSource: DataSourceMock,
		},
	})
	if err != nil {
		t.Fatalf("UpsertTradeCalendars() update error = %v", err)
	}

	items, err := repo.ListTradeCalendars(ctx, ListTradeCalendarsQuery{
		Market:    MarketCN,
		Exchange:  ExchangeSSE,
		StartDate: openDay,
		EndDate:   closedDay,
	})
	if err != nil {
		t.Fatalf("ListTradeCalendars() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(calendars) = %d, want 2", len(items))
	}
	if items[1].IsOpen {
		t.Fatalf("closed day IsOpen = true, want false")
	}

	latest, err := repo.LatestOpenDay(ctx, MarketCN, ExchangeSSE)
	if err != nil {
		t.Fatalf("LatestOpenDay() error = %v", err)
	}
	if !latest.Equal(openDay) {
		t.Fatalf("latest open day = %s, want %s", latest.Format(DateLayout), openDay.Format(DateLayout))
	}

	isOpen, err := repo.IsOpenDay(ctx, MarketCN, ExchangeSSE, closedDay)
	if err != nil {
		t.Fatalf("IsOpenDay() error = %v", err)
	}
	if isOpen {
		t.Fatalf("IsOpenDay() = true, want false")
	}
}

func TestMockTradeCalendarRepositoryMissingDataErrors(t *testing.T) {
	repo := NewMockTradeCalendarRepository()

	_, err := repo.LatestOpenDay(context.Background(), MarketCN, ExchangeSSE)
	if !IsCode(err, CodeTradeCalendarNotInitialized) {
		t.Fatalf("LatestOpenDay() error = %v, want %s", err, CodeTradeCalendarNotInitialized)
	}

	_, err = repo.IsOpenDay(context.Background(), MarketCN, ExchangeSSE, mustDate(t, "2026-06-30"))
	if !IsCode(err, CodeTradeCalendarNotFound) {
		t.Fatalf("IsOpenDay() error = %v, want %s", err, CodeTradeCalendarNotFound)
	}
}

func TestMockSyncTaskRepositoryLifecycleAndLogs(t *testing.T) {
	ctx := context.Background()
	repo := NewMockSyncTaskRepository()
	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)

	err := repo.CreateSyncTask(ctx, SyncTask{
		UID:        "task-1",
		TaskType:   SyncTaskTypeStockMaster,
		Market:     MarketCN,
		AssetType:  AssetTypeStock,
		DataSource: DataSourceMock,
		Status:     SyncTaskStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("CreateSyncTask() error = %v", err)
	}

	active, err := repo.FindActiveSyncTask(ctx)
	if err != nil {
		t.Fatalf("FindActiveSyncTask() error = %v", err)
	}
	if active.UID != "task-1" {
		t.Fatalf("active UID = %q, want task-1", active.UID)
	}

	active.Status = SyncTaskStatusSuccess
	finishedAt := now.Add(time.Minute)
	active.FinishedAt = &finishedAt
	if err := repo.UpdateSyncTask(ctx, active); err != nil {
		t.Fatalf("UpdateSyncTask() error = %v", err)
	}

	err = repo.AppendSyncLogs(ctx, []SyncLog{
		{TaskUID: "task-1", Step: "stock_master", Status: SyncLogStatusSuccess, CreatedAt: now},
	})
	if err != nil {
		t.Fatalf("AppendSyncLogs() error = %v", err)
	}

	tasks, err := repo.ListSyncTasks(ctx, ListSyncTasksQuery{
		PageQuery: PageQuery{Page: 1, PageSize: 10},
		Status:    SyncTaskStatusSuccess,
	})
	if err != nil {
		t.Fatalf("ListSyncTasks() error = %v", err)
	}
	if tasks.Pagination.Total != 1 || len(tasks.Items) != 1 {
		t.Fatalf("tasks = %+v, want one success task", tasks)
	}

	logs, err := repo.ListSyncLogs(ctx, ListSyncLogsQuery{
		PageQuery: PageQuery{Page: 1, PageSize: 10},
		TaskUID:   "task-1",
		Status:    SyncLogStatusSuccess,
	})
	if err != nil {
		t.Fatalf("ListSyncLogs() error = %v", err)
	}
	if logs.Pagination.Total != 1 || len(logs.Items) != 1 {
		t.Fatalf("logs = %+v, want one success log", logs)
	}
}

func TestMockSyncTaskRepositoryMarksStaleTasksFailed(t *testing.T) {
	ctx := context.Background()
	repo := NewMockSyncTaskRepository()

	for _, task := range []SyncTask{
		{UID: "pending", Status: SyncTaskStatusPending},
		{UID: "running", Status: SyncTaskStatusRunning},
		{UID: "success", Status: SyncTaskStatusSuccess},
	} {
		if err := repo.CreateSyncTask(ctx, task); err != nil {
			t.Fatalf("CreateSyncTask(%s) error = %v", task.UID, err)
		}
	}

	if err := repo.MarkStaleTasksFailed(ctx); err != nil {
		t.Fatalf("MarkStaleTasksFailed() error = %v", err)
	}

	tests := map[string]SyncTaskStatus{
		"pending": SyncTaskStatusFailed,
		"running": SyncTaskStatusFailed,
		"success": SyncTaskStatusSuccess,
	}
	for uid, want := range tests {
		got, err := repo.FindSyncTask(ctx, uid)
		if err != nil {
			t.Fatalf("FindSyncTask(%s) error = %v", uid, err)
		}
		if got.Status != want {
			t.Fatalf("%s status = %s, want %s", uid, got.Status, want)
		}
	}
}

func TestMockProviderSuccessEmptyAndFailure(t *testing.T) {
	ctx := context.Background()
	day := mustDate(t, "2026-06-30")

	stocks, err := NewMockInstrumentProvider([]Stock{{Market: MarketCN, Symbol: "600519"}}).FetchStocks(ctx, FetchStocksRequest{Market: MarketCN})
	if err != nil {
		t.Fatalf("FetchStocks() error = %v", err)
	}
	if len(stocks) != 1 || stocks[0].DataSource != DataSourceMock {
		t.Fatalf("stocks = %+v, want mock stock", stocks)
	}

	calendars, err := NewMockCalendarProvider([]TradeCalendar{}).FetchTradeCalendars(ctx, FetchTradeCalendarsRequest{
		Market:    MarketCN,
		Exchange:  ExchangeSSE,
		StartDate: day,
		EndDate:   day,
	})
	if err != nil {
		t.Fatalf("FetchTradeCalendars() empty error = %v", err)
	}
	if calendars == nil || len(calendars) != 0 {
		t.Fatalf("calendars = %+v, want empty slice", calendars)
	}

	wantErr := errors.New("provider unavailable")
	_, err = NewFailingMockMarketDataProvider(wantErr).FetchDailyKLines(ctx, FetchDailyKLinesRequest{
		Market:    MarketCN,
		AssetType: AssetTypeStock,
		Symbol:    "600519",
		StartDate: day,
		EndDate:   day,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("FetchDailyKLines() error = %v, want %v", err, wantErr)
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()

	date, err := ParseDate(value)
	if err != nil {
		t.Fatalf("ParseDate(%q) error = %v", value, err)
	}
	return date
}
