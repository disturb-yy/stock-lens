package market

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

type mockStockRepository struct {
	mu     sync.RWMutex
	stocks map[stockKey]Stock
}

type stockKey struct {
	market    Market
	assetType AssetType
	symbol    string
}

type dailyKLineKey struct {
	market    Market
	assetType AssetType
	symbol    string
	tradeDate time.Time
}

type tradeCalendarKey struct {
	market   Market
	exchange Exchange
	calDate  time.Time
}

type mockKLineRepository struct {
	mu    sync.RWMutex
	lines map[dailyKLineKey]DailyKLine
}

type mockTradeCalendarRepository struct {
	mu        sync.RWMutex
	calendars map[tradeCalendarKey]TradeCalendar
}

type mockSyncTaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]SyncTask
	logs  []SyncLog
}

func NewMockStockRepository() StockRepository {
	return &mockStockRepository{
		stocks: make(map[stockKey]Stock),
	}
}

func NewMockKLineRepository() KLineRepository {
	return &mockKLineRepository{
		lines: make(map[dailyKLineKey]DailyKLine),
	}
}

func NewMockTradeCalendarRepository() TradeCalendarRepository {
	return &mockTradeCalendarRepository{
		calendars: make(map[tradeCalendarKey]TradeCalendar),
	}
}

func NewMockSyncTaskRepository() SyncTaskRepository {
	return &mockSyncTaskRepository{
		tasks: make(map[string]SyncTask),
		logs:  make([]SyncLog, 0),
	}
}

func (r *mockStockRepository) UpsertStocks(_ context.Context, stocks []Stock) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, stock := range stocks {
		key := stockKey{
			market:    stock.Market,
			assetType: stock.AssetType,
			symbol:    stock.Symbol,
		}
		r.stocks[key] = stock
	}
	return nil
}

func (r *mockStockRepository) ListStocks(_ context.Context, query ListStocksQuery) (ListStocksResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]Stock, 0)
	for _, stock := range r.stocks {
		if !matchStock(stock, query) {
			continue
		}
		items = append(items, stock)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Symbol < items[j].Symbol
	})

	page, pageSize := normalizePage(query.Page, query.PageSize)
	total := int64(len(items))
	items = paginateStocks(items, page, pageSize)

	return ListStocksResult{
		Items: items,
		Pagination: PageResult{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages(total, pageSize),
		},
	}, nil
}

func (r *mockStockRepository) FindStock(_ context.Context, query FindStockQuery) (Stock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stock, ok := r.stocks[stockKey{
		market:    query.Market,
		assetType: query.AssetType,
		symbol:    query.Symbol,
	}]
	if !ok {
		return Stock{}, NewError(CodeStockNotFound, "stock not found")
	}
	return stock, nil
}

func (r *mockKLineRepository) UpsertDailyKLines(_ context.Context, lines []DailyKLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, line := range lines {
		r.lines[dailyKLineKey{
			market:    line.Market,
			assetType: line.AssetType,
			symbol:    line.Symbol,
			tradeDate: dateOnly(line.TradeDate),
		}] = line
	}
	return nil
}

func (r *mockKLineRepository) ListDailyKLines(_ context.Context, query ListDailyKLinesQuery) ([]DailyKLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]DailyKLine, 0)
	for _, line := range r.lines {
		if !matchDailyKLine(line, query) {
			continue
		}
		items = append(items, line)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].TradeDate.Before(items[j].TradeDate)
	})
	return items, nil
}

func (r *mockKLineRepository) LatestDailyKLine(_ context.Context, market Market, assetType AssetType, symbol string) (DailyKLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest DailyKLine
	found := false
	for _, line := range r.lines {
		if line.Market != market || line.AssetType != assetType || line.Symbol != symbol {
			continue
		}
		if !found || latest.TradeDate.Before(line.TradeDate) {
			latest = line
			found = true
		}
	}
	return latest, nil
}

func (r *mockTradeCalendarRepository) UpsertTradeCalendars(_ context.Context, calendars []TradeCalendar) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, calendar := range calendars {
		r.calendars[tradeCalendarKey{
			market:   calendar.Market,
			exchange: calendar.Exchange,
			calDate:  dateOnly(calendar.CalDate),
		}] = calendar
	}
	return nil
}

func (r *mockTradeCalendarRepository) ListTradeCalendars(_ context.Context, query ListTradeCalendarsQuery) ([]TradeCalendar, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]TradeCalendar, 0)
	for _, calendar := range r.calendars {
		if !matchTradeCalendar(calendar, query) {
			continue
		}
		items = append(items, calendar)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CalDate.Before(items[j].CalDate)
	})
	return items, nil
}

func (r *mockTradeCalendarRepository) LatestOpenDay(_ context.Context, market Market, exchange Exchange) (time.Time, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest time.Time
	found := false
	for _, calendar := range r.calendars {
		if calendar.Market != market || calendar.Exchange != exchange || !calendar.IsOpen {
			continue
		}
		if !found || latest.Before(calendar.CalDate) {
			latest = calendar.CalDate
			found = true
		}
	}
	if !found {
		return time.Time{}, NewError(CodeTradeCalendarNotInitialized, "trade calendar not initialized")
	}
	return latest, nil
}

func (r *mockTradeCalendarRepository) IsOpenDay(_ context.Context, market Market, exchange Exchange, date time.Time) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	calendar, ok := r.calendars[tradeCalendarKey{
		market:   market,
		exchange: exchange,
		calDate:  dateOnly(date),
	}]
	if !ok {
		return false, NewError(CodeTradeCalendarNotFound, "trade calendar not found")
	}
	return calendar.IsOpen, nil
}

func (r *mockSyncTaskRepository) CreateSyncTask(_ context.Context, task SyncTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[task.UID] = task
	return nil
}

func (r *mockSyncTaskRepository) UpdateSyncTask(_ context.Context, task SyncTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[task.UID]; !ok {
		return NewError(CodeSyncTaskNotFound, "sync task not found")
	}
	r.tasks[task.UID] = task
	return nil
}

func (r *mockSyncTaskRepository) FindSyncTask(_ context.Context, uid string) (SyncTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[uid]
	if !ok {
		return SyncTask{}, NewError(CodeSyncTaskNotFound, "sync task not found")
	}
	return task, nil
}

func (r *mockSyncTaskRepository) FindActiveSyncTask(_ context.Context) (SyncTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	active := make([]SyncTask, 0)
	for _, task := range r.tasks {
		if task.Status == SyncTaskStatusPending || task.Status == SyncTaskStatusRunning {
			active = append(active, task)
		}
	}
	if len(active) == 0 {
		return SyncTask{}, NewError(CodeSyncTaskNotFound, "sync task not found")
	}
	sortSyncTasks(active)
	return active[0], nil
}

func (r *mockSyncTaskRepository) ListSyncTasks(_ context.Context, query ListSyncTasksQuery) (ListSyncTasksResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]SyncTask, 0)
	for _, task := range r.tasks {
		if query.Status != "" && task.Status != query.Status {
			continue
		}
		items = append(items, task)
	}
	sortSyncTasks(items)

	page, pageSize := normalizePage(query.Page, query.PageSize)
	total := int64(len(items))
	items = paginateSyncTasks(items, page, pageSize)

	return ListSyncTasksResult{
		Items: items,
		Pagination: PageResult{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages(total, pageSize),
		},
	}, nil
}

func (r *mockSyncTaskRepository) MarkStaleTasksFailed(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for uid, task := range r.tasks {
		if task.Status != SyncTaskStatusPending && task.Status != SyncTaskStatusRunning {
			continue
		}
		task.Status = SyncTaskStatusFailed
		task.FinishedAt = &now
		task.ErrorMsg = "stale sync task failed on startup"
		task.UpdatedAt = now
		r.tasks[uid] = task
	}
	return nil
}

func (r *mockSyncTaskRepository) AppendSyncLogs(_ context.Context, logs []SyncLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logs = append(r.logs, logs...)
	return nil
}

func (r *mockSyncTaskRepository) ListSyncLogs(_ context.Context, query ListSyncLogsQuery) (ListSyncLogsResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]SyncLog, 0)
	for _, log := range r.logs {
		if query.TaskUID != "" && log.TaskUID != query.TaskUID {
			continue
		}
		if query.Status != "" && log.Status != query.Status {
			continue
		}
		if query.Symbol != "" && log.Symbol != query.Symbol {
			continue
		}
		items = append(items, log)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	page, pageSize := normalizePage(query.Page, query.PageSize)
	total := int64(len(items))
	items = paginateSyncLogs(items, page, pageSize)

	return ListSyncLogsResult{
		Items: items,
		Pagination: PageResult{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages(total, pageSize),
		},
	}, nil
}

func matchStock(stock Stock, query ListStocksQuery) bool {
	if query.Market != "" && stock.Market != query.Market {
		return false
	}
	if query.AssetType != "" && stock.AssetType != query.AssetType {
		return false
	}
	if query.Exchange != "" && stock.Exchange != query.Exchange {
		return false
	}
	if query.Status != "" && stock.Status != query.Status {
		return false
	}
	if query.Keyword != "" && !strings.Contains(stock.Symbol, query.Keyword) && !strings.Contains(stock.Name, query.Keyword) {
		return false
	}
	return true
}

func matchDailyKLine(line DailyKLine, query ListDailyKLinesQuery) bool {
	if query.Market != "" && line.Market != query.Market {
		return false
	}
	if query.AssetType != "" && line.AssetType != query.AssetType {
		return false
	}
	if query.Symbol != "" && line.Symbol != query.Symbol {
		return false
	}
	return withinDateRange(line.TradeDate, query.StartDate, query.EndDate)
}

func matchTradeCalendar(calendar TradeCalendar, query ListTradeCalendarsQuery) bool {
	if query.Market != "" && calendar.Market != query.Market {
		return false
	}
	if query.Exchange != "" && calendar.Exchange != query.Exchange {
		return false
	}
	return withinDateRange(calendar.CalDate, query.StartDate, query.EndDate)
}

func normalizePage(page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}

func sortSyncTasks(items []SyncTask) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].UID < items[j].UID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func paginateStocks(items []Stock, page int, pageSize int) []Stock {
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []Stock{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func paginateSyncTasks(items []SyncTask, page int, pageSize int) []SyncTask {
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []SyncTask{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func paginateSyncLogs(items []SyncLog, page int, pageSize int) []SyncLog {
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []SyncLog{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func totalPages(total int64, pageSize int) int {
	if total == 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

func withinDateRange(value time.Time, start time.Time, end time.Time) bool {
	date := dateOnly(value)
	if !start.IsZero() && date.Before(dateOnly(start)) {
		return false
	}
	if !end.IsZero() && date.After(dateOnly(end)) {
		return false
	}
	return true
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
