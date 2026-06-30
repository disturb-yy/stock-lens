package market

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

const (
	syncDailyKLineYears     = 3
	syncSingleDailyKLineMax = 20
	syncFullDailyKLineMax   = 5
)

type syncRunner func(fn func())

type dataSourceProvider interface {
	DataSource() DataSource
}

type SyncService struct {
	stocks             StockRepository
	kLines             KLineRepository
	tradeCalendars     TradeCalendarRepository
	syncTasks          SyncTaskRepository
	instrumentProvider InstrumentProvider
	calendarProvider   CalendarProvider
	marketDataProvider MarketDataProvider
	runner             syncRunner
}

type SyncTaskCreationResult struct {
	TaskUID string
	Status  SyncTaskStatus
}

type SyncStocksInput struct {
	Market    Market
	AssetType AssetType
	RequestID string
}

type SyncTradeCalendarsInput struct {
	Market    Market
	Exchange  Exchange
	StartDate time.Time
	EndDate   time.Time
	RequestID string
}

type SyncDailyKLinesInput struct {
	Market    Market
	AssetType AssetType
	Symbol    string
	StartDate time.Time
	EndDate   time.Time
	RequestID string
}

func NewSyncService(
	stocks StockRepository,
	kLines KLineRepository,
	tradeCalendars TradeCalendarRepository,
	syncTasks SyncTaskRepository,
	instrumentProvider InstrumentProvider,
	calendarProvider CalendarProvider,
	marketDataProvider MarketDataProvider,
) *SyncService {
	return NewSyncServiceWithRunner(stocks, kLines, tradeCalendars, syncTasks, instrumentProvider, calendarProvider, marketDataProvider, runAsync)
}

func NewSyncServiceWithRunner(
	stocks StockRepository,
	kLines KLineRepository,
	tradeCalendars TradeCalendarRepository,
	syncTasks SyncTaskRepository,
	instrumentProvider InstrumentProvider,
	calendarProvider CalendarProvider,
	marketDataProvider MarketDataProvider,
	runner syncRunner,
) *SyncService {
	if runner == nil {
		runner = runAsync
	}
	return &SyncService{
		stocks:             stocks,
		kLines:             kLines,
		tradeCalendars:     tradeCalendars,
		syncTasks:          syncTasks,
		instrumentProvider: instrumentProvider,
		calendarProvider:   calendarProvider,
		marketDataProvider: marketDataProvider,
		runner:             runner,
	}
}

func (s *SyncService) SyncStocks(ctx context.Context, input SyncStocksInput) (SyncTaskCreationResult, error) {
	input.Market = defaultMarket(input.Market)
	input.AssetType = defaultAssetType(input.AssetType)
	if err := validateMarketAsset(input.Market, input.AssetType); err != nil {
		return SyncTaskCreationResult{}, err
	}

	task, err := s.createPendingTask(ctx, SyncTask{
		TaskType:   SyncTaskTypeStockMaster,
		Market:     input.Market,
		AssetType:  input.AssetType,
		DataSource: s.dataSource(),
		RequestID:  input.RequestID,
	})
	if err != nil {
		return SyncTaskCreationResult{}, err
	}

	s.runTask(ctx, task, func(taskCtx context.Context, task SyncTask) SyncTask {
		return s.executeStockSync(taskCtx, task)
	})
	return SyncTaskCreationResult{TaskUID: task.UID, Status: SyncTaskStatusPending}, nil
}

func (s *SyncService) SyncTradeCalendars(ctx context.Context, input SyncTradeCalendarsInput) (SyncTaskCreationResult, error) {
	input.Market = defaultMarket(input.Market)
	input.Exchange = defaultExchange(input.Exchange)
	if err := validateMarketExchange(input.Market, input.Exchange); err != nil {
		return SyncTaskCreationResult{}, err
	}
	if input.StartDate.IsZero() || input.EndDate.IsZero() {
		return SyncTaskCreationResult{}, NewError(CodeInvalidDateRange, "invalid date range")
	}
	if err := validateDateRange(input.StartDate, input.EndDate, tradeCalendarMaxYears); err != nil {
		return SyncTaskCreationResult{}, err
	}

	task, err := s.createPendingTask(ctx, SyncTask{
		TaskType:   SyncTaskTypeTradeCalendars,
		Market:     input.Market,
		AssetType:  AssetTypeStock,
		DataSource: s.dataSource(),
		RequestID:  input.RequestID,
	})
	if err != nil {
		return SyncTaskCreationResult{}, err
	}

	s.runTask(ctx, task, func(taskCtx context.Context, task SyncTask) SyncTask {
		return s.executeTradeCalendarSync(taskCtx, task, input)
	})
	return SyncTaskCreationResult{TaskUID: task.UID, Status: SyncTaskStatusPending}, nil
}

func (s *SyncService) SyncDailyKLines(ctx context.Context, input SyncDailyKLinesInput) (SyncTaskCreationResult, error) {
	normalized, err := s.normalizeSyncDailyKLinesInput(ctx, input)
	if err != nil {
		return SyncTaskCreationResult{}, err
	}

	task, err := s.createPendingTask(ctx, SyncTask{
		TaskType:   SyncTaskTypeDailyKLines,
		Market:     normalized.Market,
		AssetType:  normalized.AssetType,
		DataSource: s.dataSource(),
		RequestID:  normalized.RequestID,
	})
	if err != nil {
		return SyncTaskCreationResult{}, err
	}

	s.runTask(ctx, task, func(taskCtx context.Context, task SyncTask) SyncTask {
		return s.executeDailyKLineSync(taskCtx, task, normalized)
	})
	return SyncTaskCreationResult{TaskUID: task.UID, Status: SyncTaskStatusPending}, nil
}

func (s *SyncService) RecoverStaleTasks(ctx context.Context) error {
	return s.syncTasks.MarkStaleTasksFailed(ctx)
}

func (s *SyncService) createPendingTask(ctx context.Context, task SyncTask) (SyncTask, error) {
	if err := s.rejectActiveTask(ctx); err != nil {
		return SyncTask{}, err
	}

	now := time.Now()
	task.UID = newTaskUID(now)
	task.Status = SyncTaskStatusPending
	task.CreatedAt = now
	task.UpdatedAt = now
	if err := s.syncTasks.CreateSyncTask(ctx, task); err != nil {
		return SyncTask{}, err
	}
	return task, nil
}

func (s *SyncService) rejectActiveTask(ctx context.Context) error {
	active, err := s.syncTasks.FindActiveSyncTask(ctx)
	if err == nil {
		return NewError(CodeSyncTaskConflict, fmt.Sprintf("sync task already running: %s", active.UID))
	}
	if IsCode(err, CodeSyncTaskNotFound) {
		return nil
	}
	return err
}

func (s *SyncService) runTask(ctx context.Context, task SyncTask, execute func(context.Context, SyncTask) SyncTask) {
	// 同步任务不应因为触发请求结束而取消，重启后的残留任务由 stale recovery 兜底。
	taskCtx := context.WithoutCancel(ctx)
	s.runner(func() {
		defer s.recoverTaskPanic(taskCtx, task)

		running := markTaskRunning(task)
		if err := s.syncTasks.UpdateSyncTask(taskCtx, running); err != nil {
			return
		}
		finished := execute(taskCtx, running)
		_ = s.syncTasks.UpdateSyncTask(taskCtx, finished)
	})
}

func (s *SyncService) dataSource() DataSource {
	// 任务数据源应反映实际 provider，避免 Tushare 同步任务在审计时显示为 MOCK。
	for _, provider := range []any{s.instrumentProvider, s.calendarProvider, s.marketDataProvider} {
		if source, ok := provider.(dataSourceProvider); ok {
			return source.DataSource()
		}
	}
	return DataSourceMock
}

func (s *SyncService) executeStockSync(ctx context.Context, task SyncTask) SyncTask {
	stocks, err := s.instrumentProvider.FetchStocks(ctx, FetchStocksRequest{Market: task.Market})
	if err != nil {
		return s.failTask(ctx, task, "fetch stocks", 1, err)
	}
	if len(stocks) == 0 {
		return s.failTask(ctx, task, "fetch stocks", 1, NewError(CodeProviderError, "empty stock list"))
	}
	if err := s.stocks.UpsertStocks(ctx, stocks); err != nil {
		return s.failTask(ctx, task, "upsert stocks", int64(len(stocks)), err)
	}

	task.TotalItems = int64(len(stocks))
	task.ProcessedItems = int64(len(stocks))
	task.SuccessItems = int64(len(stocks))
	task.FailedItems = 0
	task.Status = SyncTaskStatusSuccess
	task.ErrorMsg = ""
	task = markTaskFinished(task)
	_ = s.appendTaskLog(ctx, task, "stock_master", SyncLogStatusSuccess, "stock master sync succeeded", "", int64(len(stocks)))
	return task
}

func (s *SyncService) executeTradeCalendarSync(ctx context.Context, task SyncTask, input SyncTradeCalendarsInput) SyncTask {
	calendars, err := s.calendarProvider.FetchTradeCalendars(ctx, FetchTradeCalendarsRequest{
		Market:    input.Market,
		Exchange:  input.Exchange,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
	})
	if err != nil {
		return s.failTask(ctx, task, "fetch trade calendars", 1, err)
	}
	if len(calendars) == 0 {
		return s.failTask(ctx, task, "fetch trade calendars", 1, NewError(CodeProviderError, "empty trade calendar"))
	}
	if err := s.tradeCalendars.UpsertTradeCalendars(ctx, calendars); err != nil {
		return s.failTask(ctx, task, "upsert trade calendars", int64(len(calendars)), err)
	}

	task.TotalItems = int64(len(calendars))
	task.ProcessedItems = int64(len(calendars))
	task.SuccessItems = int64(len(calendars))
	task.FailedItems = 0
	task.Status = SyncTaskStatusSuccess
	task.ErrorMsg = ""
	task = markTaskFinished(task)
	_ = s.appendTaskLog(ctx, task, "trade_calendars", SyncLogStatusSuccess, "trade calendar sync succeeded", "", int64(len(calendars)))
	return task
}

func (s *SyncService) executeDailyKLineSync(ctx context.Context, task SyncTask, input SyncDailyKLinesInput) SyncTask {
	if input.Symbol != "" {
		return s.executeSingleStockDailyKLineSync(ctx, task, input)
	}
	return s.executeFullMarketDailyKLineSync(ctx, task, input)
}

func (s *SyncService) executeSingleStockDailyKLineSync(ctx context.Context, task SyncTask, input SyncDailyKLinesInput) SyncTask {
	if _, err := s.stocks.FindStock(ctx, FindStockQuery{Market: input.Market, AssetType: input.AssetType, Symbol: input.Symbol}); err != nil {
		return s.failTask(ctx, task, "daily_k_lines_preflight", 1, err)
	}

	lines, err := s.marketDataProvider.FetchDailyKLines(ctx, FetchDailyKLinesRequest{
		Market:    input.Market,
		AssetType: input.AssetType,
		Symbol:    input.Symbol,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
	})
	if err != nil {
		return s.failTask(ctx, task, "fetch_daily_k_lines", 1, err)
	}
	if err := s.kLines.UpsertDailyKLines(ctx, lines); err != nil {
		return s.failTask(ctx, task, "upsert_daily_k_lines", 1, err)
	}

	task.TotalItems = 1
	task.ProcessedItems = 1
	task.SuccessItems = 1
	task.FailedItems = 0
	task.Status = SyncTaskStatusSuccess
	task.ErrorMsg = ""
	task = markTaskFinished(task)
	_ = s.appendTaskLogWithSymbol(ctx, task, "daily_k_lines", input.Symbol, SyncLogStatusSuccess, "daily k line sync succeeded", "", int64(len(lines)))
	return task
}

func (s *SyncService) executeFullMarketDailyKLineSync(ctx context.Context, task SyncTask, input SyncDailyKLinesInput) SyncTask {
	stocks, err := s.listAllListedStocks(ctx, input.Market, input.AssetType)
	if err != nil {
		return s.failTask(ctx, task, "list_stocks", 1, err)
	}
	if len(stocks) == 0 {
		return s.failTask(ctx, task, "list_stocks", 1, NewError(CodeStocksNotInitialized, "stocks not initialized"))
	}

	task.TotalItems = int64(len(stocks))
	for _, stock := range stocks {
		lines, err := s.marketDataProvider.FetchDailyKLines(ctx, FetchDailyKLinesRequest{
			Market:    input.Market,
			AssetType: input.AssetType,
			Symbol:    stock.Symbol,
			StartDate: input.StartDate,
			EndDate:   input.EndDate,
		})
		if err != nil {
			task.FailedItems++
			task.ProcessedItems++
			_ = s.appendTaskLogWithSymbol(ctx, task, "daily_k_lines", stock.Symbol, SyncLogStatusFailed, "daily k line sync failed", err.Error(), 0)
			continue
		}
		if err := s.kLines.UpsertDailyKLines(ctx, lines); err != nil {
			task.FailedItems++
			task.ProcessedItems++
			_ = s.appendTaskLogWithSymbol(ctx, task, "daily_k_lines", stock.Symbol, SyncLogStatusFailed, "daily k line upsert failed", err.Error(), 0)
			continue
		}
		task.SuccessItems++
		task.ProcessedItems++
		_ = s.appendTaskLogWithSymbol(ctx, task, "daily_k_lines", stock.Symbol, SyncLogStatusSuccess, "daily k line sync succeeded", "", int64(len(lines)))
	}

	task.Status = finalDailyKLineTaskStatus(task.SuccessItems, task.FailedItems)
	if task.Status == SyncTaskStatusFailed {
		task.ErrorMsg = "all daily k line sync items failed"
	}
	task = markTaskFinished(task)
	return task
}

func (s *SyncService) listAllListedStocks(ctx context.Context, market Market, assetType AssetType) ([]Stock, error) {
	items := make([]Stock, 0)
	for page := 1; ; page++ {
		result, err := s.stocks.ListStocks(ctx, ListStocksQuery{
			PageQuery: PageQuery{Page: page, PageSize: maxPageSize},
			Market:    market,
			AssetType: assetType,
			Status:    StockStatusListed,
		})
		if err != nil {
			return nil, err
		}
		items = append(items, result.Items...)
		if page >= result.Pagination.TotalPages || len(result.Items) == 0 {
			break
		}
	}
	return items, nil
}

func (s *SyncService) failTask(ctx context.Context, task SyncTask, step string, failedItems int64, err error) SyncTask {
	task.TotalItems = maxInt64(task.TotalItems, failedItems)
	task.ProcessedItems = failedItems
	task.SuccessItems = 0
	task.FailedItems = failedItems
	task.Status = SyncTaskStatusFailed
	task.ErrorMsg = err.Error()
	task = markTaskFinished(task)
	_ = s.appendTaskLog(ctx, task, step, SyncLogStatusFailed, "sync step failed", err.Error(), 0)
	return task
}

func (s *SyncService) recoverTaskPanic(ctx context.Context, task SyncTask) {
	recovered := recover()
	if recovered == nil {
		return
	}
	task.Status = SyncTaskStatusFailed
	task.ErrorMsg = fmt.Sprintf("panic: %v", recovered)
	task.FailedItems = maxInt64(task.FailedItems, 1)
	task.ProcessedItems = maxInt64(task.ProcessedItems, task.FailedItems)
	task = markTaskFinished(task)
	_ = s.appendTaskLog(ctx, task, "panic", SyncLogStatusFailed, "sync task panicked", task.ErrorMsg, 0)
	_ = s.syncTasks.UpdateSyncTask(ctx, task)
}

func (s *SyncService) appendTaskLog(ctx context.Context, task SyncTask, step string, status SyncLogStatus, message string, detail string, affectedRows int64) error {
	return s.appendTaskLogWithSymbol(ctx, task, step, "", status, message, detail, affectedRows)
}

func (s *SyncService) appendTaskLogWithSymbol(ctx context.Context, task SyncTask, step string, symbol string, status SyncLogStatus, message string, detail string, affectedRows int64) error {
	return s.syncTasks.AppendSyncLogs(ctx, []SyncLog{
		{
			TaskUID:      task.UID,
			Step:         step,
			Status:       status,
			Market:       task.Market,
			AssetType:    task.AssetType,
			Symbol:       symbol,
			DataSource:   task.DataSource,
			Message:      message,
			ErrorDetail:  detail,
			AffectedRows: affectedRows,
			CreatedAt:    time.Now(),
		},
	})
}

func (s *SyncService) normalizeSyncDailyKLinesInput(ctx context.Context, input SyncDailyKLinesInput) (SyncDailyKLinesInput, error) {
	input.Market = defaultMarket(input.Market)
	input.AssetType = defaultAssetType(input.AssetType)
	if err := validateMarketAsset(input.Market, input.AssetType); err != nil {
		return SyncDailyKLinesInput{}, err
	}
	if input.Symbol != "" && !ValidSymbol(input.Symbol) {
		return SyncDailyKLinesInput{}, NewError(CodeInvalidSymbol, "invalid symbol")
	}
	if input.EndDate.IsZero() {
		latest, err := s.tradeCalendars.LatestOpenDay(ctx, input.Market, ExchangeSSE)
		if err != nil {
			return SyncDailyKLinesInput{}, err
		}
		input.EndDate = latest
	}
	if input.StartDate.IsZero() {
		input.StartDate = input.EndDate.AddDate(-syncDailyKLineYears, 0, 0)
	}

	maxYears := syncFullDailyKLineMax
	if input.Symbol != "" {
		maxYears = syncSingleDailyKLineMax
	}
	if err := validateDateRange(input.StartDate, input.EndDate, maxYears); err != nil {
		return SyncDailyKLinesInput{}, err
	}
	return input, nil
}

func finalDailyKLineTaskStatus(successItems int64, failedItems int64) SyncTaskStatus {
	if failedItems == 0 {
		return SyncTaskStatusSuccess
	}
	if successItems == 0 {
		return SyncTaskStatusFailed
	}
	return SyncTaskStatusPartialSuccess
}

func markTaskRunning(task SyncTask) SyncTask {
	now := time.Now()
	task.Status = SyncTaskStatusRunning
	task.StartedAt = &now
	task.UpdatedAt = now
	return task
}

func markTaskFinished(task SyncTask) SyncTask {
	now := time.Now()
	task.FinishedAt = &now
	task.UpdatedAt = now
	return task
}

func newTaskUID(now time.Time) string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(now), entropy).String()
}

func runAsync(fn func()) {
	go fn()
}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
