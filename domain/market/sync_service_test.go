package market

import (
	"context"
	"errors"
	"testing"
)

func TestSyncServiceSyncStocksCreatesAndCompletesTask(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	tasks := NewMockSyncTaskRepository()
	service := NewSyncServiceWithRunner(
		stocks,
		NewMockKLineRepository(),
		NewMockTradeCalendarRepository(),
		tasks,
		NewMockInstrumentProvider([]Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", Name: "Kweichow Moutai"}}),
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider(nil),
		runSyncForTest,
	)

	created, err := service.SyncStocks(ctx, SyncStocksInput{RequestID: "req_test"})
	if err != nil {
		t.Fatalf("SyncStocks() error = %v", err)
	}
	if created.TaskUID == "" {
		t.Fatalf("TaskUID is empty")
	}

	task, err := tasks.FindSyncTask(ctx, created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusSuccess {
		t.Fatalf("task status = %s, want %s", task.Status, SyncTaskStatusSuccess)
	}
	if task.TotalItems != 1 || task.ProcessedItems != 1 || task.SuccessItems != 1 || task.FailedItems != 0 {
		t.Fatalf("task counters = %+v, want one success", task)
	}

	stock, err := stocks.FindStock(ctx, FindStockQuery{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519"})
	if err != nil {
		t.Fatalf("FindStock() error = %v", err)
	}
	if stock.DataSource != DataSourceMock {
		t.Fatalf("DataSource = %s, want %s", stock.DataSource, DataSourceMock)
	}
}

func TestSyncServiceDetachedTaskIgnoresRequestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	stocks := NewMockStockRepository()
	tasks := NewMockSyncTaskRepository()
	var run func()
	service := NewSyncServiceWithRunner(
		stocks,
		NewMockKLineRepository(),
		NewMockTradeCalendarRepository(),
		tasks,
		contextCheckingInstrumentProvider{
			source: DataSourceTushare,
			stocks: []Stock{{
				Market:    MarketCN,
				AssetType: AssetTypeStock,
				Symbol:    "600519",
				Name:      "Kweichow Moutai",
			}},
		},
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider(nil),
		func(fn func()) {
			run = fn
		},
	)

	created, err := service.SyncStocks(ctx, SyncStocksInput{})
	if err != nil {
		t.Fatalf("SyncStocks() error = %v", err)
	}
	if run == nil {
		t.Fatalf("sync runner was not called")
	}

	cancel()
	run()

	task, err := tasks.FindSyncTask(context.Background(), created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusSuccess {
		t.Fatalf("task status = %s, want %s; error=%q", task.Status, SyncTaskStatusSuccess, task.ErrorMsg)
	}
	if task.DataSource != DataSourceTushare {
		t.Fatalf("task data source = %s, want %s", task.DataSource, DataSourceTushare)
	}
}

func TestSyncServiceRejectsActiveTaskConflict(t *testing.T) {
	ctx := context.Background()
	tasks := NewMockSyncTaskRepository()
	service := NewSyncServiceWithRunner(
		NewMockStockRepository(),
		NewMockKLineRepository(),
		NewMockTradeCalendarRepository(),
		tasks,
		NewMockInstrumentProvider(nil),
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider(nil),
		runSyncForTest,
	)

	err := tasks.CreateSyncTask(ctx, SyncTask{
		UID:        "active",
		Status:     SyncTaskStatusRunning,
		TaskType:   SyncTaskTypeStockMaster,
		Market:     MarketCN,
		AssetType:  AssetTypeStock,
		DataSource: DataSourceMock,
	})
	if err != nil {
		t.Fatalf("CreateSyncTask() error = %v", err)
	}

	_, err = service.SyncStocks(ctx, SyncStocksInput{})
	if !IsCode(err, CodeSyncTaskConflict) {
		t.Fatalf("SyncStocks() error = %v, want %s", err, CodeSyncTaskConflict)
	}
}

func TestSyncServiceSyncStocksMarksProviderFailure(t *testing.T) {
	ctx := context.Background()
	tasks := NewMockSyncTaskRepository()
	service := NewSyncServiceWithRunner(
		NewMockStockRepository(),
		NewMockKLineRepository(),
		NewMockTradeCalendarRepository(),
		tasks,
		NewFailingMockInstrumentProvider(errors.New("provider down")),
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider(nil),
		runSyncForTest,
	)

	created, err := service.SyncStocks(ctx, SyncStocksInput{})
	if err != nil {
		t.Fatalf("SyncStocks() error = %v", err)
	}

	task, err := tasks.FindSyncTask(ctx, created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusFailed {
		t.Fatalf("task status = %s, want %s", task.Status, SyncTaskStatusFailed)
	}
	if task.FailedItems != 1 || task.ErrorMsg == "" {
		t.Fatalf("failed task = %+v, want failed counter and message", task)
	}
}

func TestSyncServiceSyncTradeCalendarsSuccessAndEmptyProviderFailure(t *testing.T) {
	ctx := context.Background()
	start := serviceTestDate(t, "2026-06-29")
	end := serviceTestDate(t, "2026-06-30")

	tests := []struct {
		name       string
		provider   CalendarProvider
		wantStatus SyncTaskStatus
		wantRows   int64
	}{
		{
			name: "success",
			provider: NewMockCalendarProvider([]TradeCalendar{
				{Market: MarketCN, Exchange: ExchangeSSE, CalDate: start, IsOpen: true},
				{Market: MarketCN, Exchange: ExchangeSSE, CalDate: end, IsOpen: false, PretradeDate: &start},
			}),
			wantStatus: SyncTaskStatusSuccess,
			wantRows:   2,
		},
		{
			name:       "empty provider result fails",
			provider:   NewMockCalendarProvider([]TradeCalendar{}),
			wantStatus: SyncTaskStatusFailed,
			wantRows:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calendars := NewMockTradeCalendarRepository()
			tasks := NewMockSyncTaskRepository()
			service := NewSyncServiceWithRunner(
				NewMockStockRepository(),
				NewMockKLineRepository(),
				calendars,
				tasks,
				NewMockInstrumentProvider(nil),
				tt.provider,
				NewMockMarketDataProvider(nil),
				runSyncForTest,
			)

			created, err := service.SyncTradeCalendars(ctx, SyncTradeCalendarsInput{
				StartDate: start,
				EndDate:   end,
			})
			if err != nil {
				t.Fatalf("SyncTradeCalendars() error = %v", err)
			}

			task, err := tasks.FindSyncTask(ctx, created.TaskUID)
			if err != nil {
				t.Fatalf("FindSyncTask() error = %v", err)
			}
			if task.Status != tt.wantStatus {
				t.Fatalf("task status = %s, want %s", task.Status, tt.wantStatus)
			}
			if task.SuccessItems != tt.wantRows {
				t.Fatalf("success items = %d, want %d", task.SuccessItems, tt.wantRows)
			}
		})
	}
}

func TestSyncServiceSyncSingleStockDailyKLines(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	lines := NewMockKLineRepository()
	tasks := NewMockSyncTaskRepository()
	tradeDate := serviceTestDate(t, "2026-06-30")
	service := NewSyncServiceWithRunner(
		stocks,
		lines,
		NewMockTradeCalendarRepository(),
		tasks,
		NewMockInstrumentProvider(nil),
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider([]DailyKLine{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", TradeDate: tradeDate}}),
		runSyncForTest,
	)

	if err := stocks.UpsertStocks(ctx, []Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", Status: StockStatusListed}}); err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	created, err := service.SyncDailyKLines(ctx, SyncDailyKLinesInput{
		Symbol:    "600519",
		StartDate: tradeDate,
		EndDate:   tradeDate,
	})
	if err != nil {
		t.Fatalf("SyncDailyKLines() error = %v", err)
	}

	task, err := tasks.FindSyncTask(ctx, created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusSuccess {
		t.Fatalf("task status = %s, want %s", task.Status, SyncTaskStatusSuccess)
	}
	if task.TotalItems != 1 || task.SuccessItems != 1 || task.FailedItems != 0 {
		t.Fatalf("task counters = %+v, want one stock success", task)
	}

	got, err := lines.ListDailyKLines(ctx, ListDailyKLinesQuery{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", StartDate: tradeDate, EndDate: tradeDate})
	if err != nil {
		t.Fatalf("ListDailyKLines() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(got))
	}
}

func TestSyncServiceSyncSingleStockDailyKLinesMissingStockFailsCreatedTask(t *testing.T) {
	ctx := context.Background()
	tasks := NewMockSyncTaskRepository()
	service := NewSyncServiceWithRunner(
		NewMockStockRepository(),
		NewMockKLineRepository(),
		NewMockTradeCalendarRepository(),
		tasks,
		NewMockInstrumentProvider(nil),
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider(nil),
		runSyncForTest,
	)

	created, err := service.SyncDailyKLines(ctx, SyncDailyKLinesInput{
		Symbol:    "600519",
		StartDate: serviceTestDate(t, "2026-06-30"),
		EndDate:   serviceTestDate(t, "2026-06-30"),
	})
	if err != nil {
		t.Fatalf("SyncDailyKLines() error = %v", err)
	}

	task, err := tasks.FindSyncTask(ctx, created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusFailed {
		t.Fatalf("task status = %s, want %s", task.Status, SyncTaskStatusFailed)
	}
	if task.ErrorMsg == "" || task.FailedItems != 1 {
		t.Fatalf("task = %+v, want failed missing-stock task", task)
	}
}

func TestSyncServiceSyncFullMarketDailyKLinesPartialSuccess(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	lines := NewMockKLineRepository()
	tasks := NewMockSyncTaskRepository()
	tradeDate := serviceTestDate(t, "2026-06-30")
	service := NewSyncServiceWithRunner(
		stocks,
		lines,
		NewMockTradeCalendarRepository(),
		tasks,
		NewMockInstrumentProvider(nil),
		NewMockCalendarProvider(nil),
		symbolMarketDataProvider{
			lines: map[string][]DailyKLine{
				"600519": {{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", TradeDate: tradeDate}},
			},
			errs: map[string]error{
				"000001": errors.New("provider failed for stock"),
			},
		},
		runSyncForTest,
	)

	if err := stocks.UpsertStocks(ctx, []Stock{
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", Status: StockStatusListed},
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "000001", Status: StockStatusListed},
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "000002", Status: StockStatusPaused},
	}); err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	created, err := service.SyncDailyKLines(ctx, SyncDailyKLinesInput{
		StartDate: tradeDate,
		EndDate:   tradeDate,
	})
	if err != nil {
		t.Fatalf("SyncDailyKLines() error = %v", err)
	}

	task, err := tasks.FindSyncTask(ctx, created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusPartialSuccess {
		t.Fatalf("task status = %s, want %s", task.Status, SyncTaskStatusPartialSuccess)
	}
	if task.TotalItems != 2 || task.ProcessedItems != 2 || task.SuccessItems != 1 || task.FailedItems != 1 {
		t.Fatalf("task counters = %+v, want one success and one failure", task)
	}
}

func TestSyncServiceSyncFullMarketDailyKLinesWithoutStocksFailsTask(t *testing.T) {
	ctx := context.Background()
	tasks := NewMockSyncTaskRepository()
	service := NewSyncServiceWithRunner(
		NewMockStockRepository(),
		NewMockKLineRepository(),
		NewMockTradeCalendarRepository(),
		tasks,
		NewMockInstrumentProvider(nil),
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider(nil),
		runSyncForTest,
	)

	created, err := service.SyncDailyKLines(ctx, SyncDailyKLinesInput{
		StartDate: serviceTestDate(t, "2026-06-30"),
		EndDate:   serviceTestDate(t, "2026-06-30"),
	})
	if err != nil {
		t.Fatalf("SyncDailyKLines() error = %v", err)
	}

	task, err := tasks.FindSyncTask(ctx, created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusFailed {
		t.Fatalf("task status = %s, want %s", task.Status, SyncTaskStatusFailed)
	}
	if task.ErrorMsg == "" || task.FailedItems != 1 {
		t.Fatalf("task = %+v, want stocks-not-initialized failure", task)
	}
}

func TestSyncServiceSyncDailyKLinesDefaultsEndDateToLatestOpenDay(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	calendars := NewMockTradeCalendarRepository()
	tasks := NewMockSyncTaskRepository()
	latestOpenDay := serviceTestDate(t, "2026-06-30")
	service := NewSyncServiceWithRunner(
		stocks,
		NewMockKLineRepository(),
		calendars,
		tasks,
		NewMockInstrumentProvider(nil),
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider([]DailyKLine{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", TradeDate: latestOpenDay}}),
		runSyncForTest,
	)

	if err := stocks.UpsertStocks(ctx, []Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", Status: StockStatusListed}}); err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}
	if err := calendars.UpsertTradeCalendars(ctx, []TradeCalendar{{Market: MarketCN, Exchange: ExchangeSSE, CalDate: latestOpenDay, IsOpen: true}}); err != nil {
		t.Fatalf("UpsertTradeCalendars() error = %v", err)
	}

	created, err := service.SyncDailyKLines(ctx, SyncDailyKLinesInput{Symbol: "600519"})
	if err != nil {
		t.Fatalf("SyncDailyKLines() error = %v", err)
	}

	task, err := tasks.FindSyncTask(ctx, created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusSuccess {
		t.Fatalf("task status = %s, want %s", task.Status, SyncTaskStatusSuccess)
	}
}

func TestSyncServiceSyncFullMarketDailyKLinesAllFailed(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	tasks := NewMockSyncTaskRepository()
	tradeDate := serviceTestDate(t, "2026-06-30")
	service := NewSyncServiceWithRunner(
		stocks,
		NewMockKLineRepository(),
		NewMockTradeCalendarRepository(),
		tasks,
		NewMockInstrumentProvider(nil),
		NewMockCalendarProvider(nil),
		symbolMarketDataProvider{
			errs: map[string]error{
				"600519": errors.New("provider failed for 600519"),
				"000001": errors.New("provider failed for 000001"),
			},
		},
		runSyncForTest,
	)

	if err := stocks.UpsertStocks(ctx, []Stock{
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", Status: StockStatusListed},
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "000001", Status: StockStatusListed},
	}); err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	created, err := service.SyncDailyKLines(ctx, SyncDailyKLinesInput{StartDate: tradeDate, EndDate: tradeDate})
	if err != nil {
		t.Fatalf("SyncDailyKLines() error = %v", err)
	}

	task, err := tasks.FindSyncTask(ctx, created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusFailed {
		t.Fatalf("task status = %s, want %s", task.Status, SyncTaskStatusFailed)
	}
	if task.TotalItems != 2 || task.ProcessedItems != 2 || task.SuccessItems != 0 || task.FailedItems != 2 {
		t.Fatalf("task counters = %+v, want all failed", task)
	}
}

func TestSyncServiceRecoversPanicAsFailedTask(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	tasks := NewMockSyncTaskRepository()
	tradeDate := serviceTestDate(t, "2026-06-30")
	service := NewSyncServiceWithRunner(
		stocks,
		NewMockKLineRepository(),
		NewMockTradeCalendarRepository(),
		tasks,
		NewMockInstrumentProvider(nil),
		NewMockCalendarProvider(nil),
		panicMarketDataProvider{},
		runSyncForTest,
	)

	if err := stocks.UpsertStocks(ctx, []Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", Status: StockStatusListed}}); err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	created, err := service.SyncDailyKLines(ctx, SyncDailyKLinesInput{
		Symbol:    "600519",
		StartDate: tradeDate,
		EndDate:   tradeDate,
	})
	if err != nil {
		t.Fatalf("SyncDailyKLines() error = %v", err)
	}

	task, err := tasks.FindSyncTask(ctx, created.TaskUID)
	if err != nil {
		t.Fatalf("FindSyncTask() error = %v", err)
	}
	if task.Status != SyncTaskStatusFailed {
		t.Fatalf("task status = %s, want %s", task.Status, SyncTaskStatusFailed)
	}
	if task.ErrorMsg == "" {
		t.Fatalf("ErrorMsg is empty, want panic summary")
	}
}

func TestSyncServiceRecoverStaleTasks(t *testing.T) {
	ctx := context.Background()
	tasks := NewMockSyncTaskRepository()
	service := NewSyncServiceWithRunner(
		NewMockStockRepository(),
		NewMockKLineRepository(),
		NewMockTradeCalendarRepository(),
		tasks,
		NewMockInstrumentProvider(nil),
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider(nil),
		runSyncForTest,
	)

	for _, task := range []SyncTask{
		{UID: "pending", Status: SyncTaskStatusPending},
		{UID: "running", Status: SyncTaskStatusRunning},
	} {
		if err := tasks.CreateSyncTask(ctx, task); err != nil {
			t.Fatalf("CreateSyncTask(%s) error = %v", task.UID, err)
		}
	}

	if err := service.RecoverStaleTasks(ctx); err != nil {
		t.Fatalf("RecoverStaleTasks() error = %v", err)
	}

	for _, uid := range []string{"pending", "running"} {
		task, err := tasks.FindSyncTask(ctx, uid)
		if err != nil {
			t.Fatalf("FindSyncTask(%s) error = %v", uid, err)
		}
		if task.Status != SyncTaskStatusFailed {
			t.Fatalf("%s status = %s, want %s", uid, task.Status, SyncTaskStatusFailed)
		}
	}
}

func runSyncForTest(fn func()) {
	fn()
}

type symbolMarketDataProvider struct {
	lines map[string][]DailyKLine
	errs  map[string]error
}

func (p symbolMarketDataProvider) FetchDailyKLines(_ context.Context, req FetchDailyKLinesRequest) ([]DailyKLine, error) {
	if err := p.errs[req.Symbol]; err != nil {
		return nil, err
	}
	return append([]DailyKLine(nil), p.lines[req.Symbol]...), nil
}

type panicMarketDataProvider struct{}

func (p panicMarketDataProvider) FetchDailyKLines(_ context.Context, _ FetchDailyKLinesRequest) ([]DailyKLine, error) {
	panic("provider panic")
}

type contextCheckingInstrumentProvider struct {
	source DataSource
	stocks []Stock
}

func (p contextCheckingInstrumentProvider) DataSource() DataSource {
	return p.source
}

func (p contextCheckingInstrumentProvider) FetchStocks(ctx context.Context, _ FetchStocksRequest) ([]Stock, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	stocks := make([]Stock, 0, len(p.stocks))
	for _, stock := range p.stocks {
		stock.DataSource = p.source
		stocks = append(stocks, stock)
	}
	return stocks, nil
}
