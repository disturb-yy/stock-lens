package market

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const mysqlUpsertBatchSize = 500

type mysqlRepositories struct {
	stocks         StockRepository
	kLines         KLineRepository
	tradeCalendars TradeCalendarRepository
	syncTasks      SyncTaskRepository
}

type mysqlTxManager struct {
	db *gorm.DB
}

type mysqlStockRepository struct {
	db *gorm.DB
}

type mysqlKLineRepository struct {
	db *gorm.DB
}

type mysqlTradeCalendarRepository struct {
	db *gorm.DB
}

type mysqlSyncTaskRepository struct {
	db *gorm.DB
}

type stockRecord struct {
	ID         uint64 `gorm:"primaryKey;column:id"`
	Market     string `gorm:"column:market"`
	AssetType  string `gorm:"column:asset_type"`
	Symbol     string `gorm:"column:symbol"`
	TSCode     string `gorm:"column:ts_code"`
	Name       string `gorm:"column:name"`
	Exchange   string `gorm:"column:exchange"`
	Board      string `gorm:"column:board"`
	Area       string `gorm:"column:area"`
	Industry   string `gorm:"column:industry"`
	Status     string `gorm:"column:status"`
	ListDate   *time.Time
	DelistDate *time.Time
	DataSource string    `gorm:"column:data_source"`
	SyncedAt   time.Time `gorm:"column:synced_at"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

type dailyKLineRecord struct {
	ID         uint64 `gorm:"primaryKey;column:id"`
	Market     string `gorm:"column:market"`
	AssetType  string `gorm:"column:asset_type"`
	Symbol     string `gorm:"column:symbol"`
	TradeDate  time.Time
	OpenPrice  decimal.Decimal `gorm:"column:open_price"`
	HighPrice  decimal.Decimal `gorm:"column:high_price"`
	LowPrice   decimal.Decimal `gorm:"column:low_price"`
	ClosePrice decimal.Decimal `gorm:"column:close_price"`
	PreClose   decimal.Decimal `gorm:"column:pre_close"`
	ChangeAmt  decimal.Decimal `gorm:"column:change_amt"`
	PctChange  decimal.Decimal `gorm:"column:pct_change"`
	Volume     decimal.Decimal `gorm:"column:volume"`
	Amount     decimal.Decimal `gorm:"column:amount"`
	DataSource string          `gorm:"column:data_source"`
	SyncedAt   time.Time       `gorm:"column:synced_at"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
}

type tradeCalendarRecord struct {
	ID           uint64 `gorm:"primaryKey;column:id"`
	Market       string `gorm:"column:market"`
	Exchange     string `gorm:"column:exchange"`
	CalDate      time.Time
	IsOpen       bool `gorm:"column:is_open"`
	PretradeDate *time.Time
	DataSource   string    `gorm:"column:data_source"`
	SyncedAt     time.Time `gorm:"column:synced_at"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

type syncTaskRecord struct {
	ID             uint64 `gorm:"primaryKey;column:id"`
	UID            string `gorm:"column:uid"`
	TaskType       string `gorm:"column:task_type"`
	Market         string `gorm:"column:market"`
	AssetType      string `gorm:"column:asset_type"`
	DataSource     string `gorm:"column:data_source"`
	Status         string `gorm:"column:status"`
	TotalItems     int64  `gorm:"column:total_items"`
	ProcessedItems int64  `gorm:"column:processed_items"`
	SuccessItems   int64  `gorm:"column:success_items"`
	FailedItems    int64  `gorm:"column:failed_items"`
	RequestID      string `gorm:"column:request_id"`
	StartedAt      *time.Time
	FinishedAt     *time.Time
	ErrorMsg       string    `gorm:"column:error_msg"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

type syncLogRecord struct {
	ID           uint64 `gorm:"primaryKey;column:id"`
	TaskID       uint64 `gorm:"column:task_id"`
	TaskUID      string `gorm:"column:task_uid"`
	Step         string `gorm:"column:step"`
	Status       string `gorm:"column:status"`
	Market       string `gorm:"column:market"`
	AssetType    string `gorm:"column:asset_type"`
	Symbol       string `gorm:"column:symbol"`
	DataSource   string `gorm:"column:data_source"`
	Message      string `gorm:"column:message"`
	ErrorDetail  string `gorm:"column:error_detail"`
	AffectedRows int64  `gorm:"column:affected_rows"`
	CreatedAt    time.Time
}

func (stockRecord) TableName() string         { return "stocks" }
func (dailyKLineRecord) TableName() string    { return "daily_k_lines" }
func (tradeCalendarRecord) TableName() string { return "trade_calendars" }
func (syncTaskRecord) TableName() string      { return "sync_tasks" }
func (syncLogRecord) TableName() string       { return "sync_logs" }

func NewMySQLRepositories(db *gorm.DB) TxRepositories {
	return mysqlRepositories{
		stocks:         NewMySQLStockRepository(db),
		kLines:         NewMySQLKLineRepository(db),
		tradeCalendars: NewMySQLTradeCalendarRepository(db),
		syncTasks:      NewMySQLSyncTaskRepository(db),
	}
}

func NewMySQLTxManager(db *gorm.DB) TxManager {
	return &mysqlTxManager{db: db}
}

func NewMySQLStockRepository(db *gorm.DB) StockRepository {
	return &mysqlStockRepository{db: db}
}

func NewMySQLKLineRepository(db *gorm.DB) KLineRepository {
	return &mysqlKLineRepository{db: db}
}

func NewMySQLTradeCalendarRepository(db *gorm.DB) TradeCalendarRepository {
	return &mysqlTradeCalendarRepository{db: db}
}

func NewMySQLSyncTaskRepository(db *gorm.DB) SyncTaskRepository {
	return &mysqlSyncTaskRepository{db: db}
}

func (r mysqlRepositories) Stocks() StockRepository {
	return r.stocks
}

func (r mysqlRepositories) KLines() KLineRepository {
	return r.kLines
}

func (r mysqlRepositories) TradeCalendars() TradeCalendarRepository {
	return r.tradeCalendars
}

func (r mysqlRepositories) SyncTasks() SyncTaskRepository {
	return r.syncTasks
}

func (m *mysqlTxManager) WithTx(ctx context.Context, fn func(ctx context.Context, repos TxRepositories) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, NewMySQLRepositories(tx))
	})
}

func upsertMySQLRecords[T any](ctx context.Context, db *gorm.DB, records []T, conflict clause.OnConflict) error {
	// MySQL prepared statement 有占位符上限，所有批量 upsert 都从这里统一拆批。
	for _, batch := range splitMySQLUpsertBatches(records) {
		if err := db.WithContext(ctx).Clauses(conflict).Create(&batch).Error; err != nil {
			return err
		}
	}
	return nil
}

func splitMySQLUpsertBatches[T any](records []T) [][]T {
	if len(records) == 0 {
		return nil
	}

	batches := make([][]T, 0, (len(records)+mysqlUpsertBatchSize-1)/mysqlUpsertBatchSize)
	for start := 0; start < len(records); start += mysqlUpsertBatchSize {
		end := start + mysqlUpsertBatchSize
		if end > len(records) {
			end = len(records)
		}
		batches = append(batches, records[start:end])
	}
	return batches
}

func (r *mysqlStockRepository) UpsertStocks(ctx context.Context, stocks []Stock) error {
	records := make([]stockRecord, 0, len(stocks))
	for _, stock := range stocks {
		records = append(records, toStockRecord(stock))
	}
	if len(records) == 0 {
		return nil
	}

	err := upsertMySQLRecords(ctx, r.db, records, clause.OnConflict{
		Columns: []clause.Column{{Name: "market"}, {Name: "asset_type"}, {Name: "symbol"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"ts_code", "name", "exchange", "board", "area", "industry", "status", "list_date", "delist_date",
			"data_source", "synced_at", "updated_at",
		}),
	})
	if err != nil {
		return fmt.Errorf("upsert stocks: %w", err)
	}
	return nil
}

func (r *mysqlStockRepository) ListStocks(ctx context.Context, query ListStocksQuery) (ListStocksResult, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := applyStockFilters(r.db.WithContext(ctx).Model(&stockRecord{}), query)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListStocksResult{}, fmt.Errorf("count stocks: %w", err)
	}

	var records []stockRecord
	if err := db.Order("symbol ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
		return ListStocksResult{}, fmt.Errorf("list stocks: %w", err)
	}

	items := make([]Stock, 0, len(records))
	for _, record := range records {
		items = append(items, record.toDomain())
	}
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

func (r *mysqlStockRepository) FindStock(ctx context.Context, query FindStockQuery) (Stock, error) {
	var record stockRecord
	err := r.db.WithContext(ctx).
		Where("market = ? AND asset_type = ? AND symbol = ?", query.Market, query.AssetType, query.Symbol).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Stock{}, NewError(CodeStockNotFound, "stock not found")
	}
	if err != nil {
		return Stock{}, fmt.Errorf("find stock: %w", err)
	}
	return record.toDomain(), nil
}

func (r *mysqlKLineRepository) UpsertDailyKLines(ctx context.Context, lines []DailyKLine) error {
	records := make([]dailyKLineRecord, 0, len(lines))
	for _, line := range lines {
		records = append(records, toDailyKLineRecord(line))
	}
	if len(records) == 0 {
		return nil
	}

	err := upsertMySQLRecords(ctx, r.db, records, clause.OnConflict{
		Columns: []clause.Column{{Name: "market"}, {Name: "asset_type"}, {Name: "symbol"}, {Name: "trade_date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"open_price", "high_price", "low_price", "close_price", "pre_close", "change_amt",
			"pct_change", "volume", "amount", "data_source", "synced_at", "updated_at",
		}),
	})
	if err != nil {
		return fmt.Errorf("upsert daily k lines: %w", err)
	}
	return nil
}

func (r *mysqlKLineRepository) ListDailyKLines(ctx context.Context, query ListDailyKLinesQuery) ([]DailyKLine, error) {
	db := r.db.WithContext(ctx).Model(&dailyKLineRecord{}).
		Where("market = ? AND asset_type = ? AND symbol = ?", query.Market, query.AssetType, query.Symbol)
	if !query.StartDate.IsZero() {
		db = db.Where("trade_date >= ?", dateOnly(query.StartDate))
	}
	if !query.EndDate.IsZero() {
		db = db.Where("trade_date <= ?", dateOnly(query.EndDate))
	}

	var records []dailyKLineRecord
	if err := db.Order("trade_date ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list daily k lines: %w", err)
	}
	items := make([]DailyKLine, 0, len(records))
	for _, record := range records {
		items = append(items, record.toDomain())
	}
	return items, nil
}

func (r *mysqlKLineRepository) LatestDailyKLine(ctx context.Context, market Market, assetType AssetType, symbol string) (DailyKLine, error) {
	var record dailyKLineRecord
	err := r.db.WithContext(ctx).
		Where("market = ? AND asset_type = ? AND symbol = ?", market, assetType, symbol).
		Order("trade_date DESC").
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DailyKLine{}, nil
	}
	if err != nil {
		return DailyKLine{}, fmt.Errorf("latest daily k line: %w", err)
	}
	return record.toDomain(), nil
}

func (r *mysqlTradeCalendarRepository) UpsertTradeCalendars(ctx context.Context, calendars []TradeCalendar) error {
	records := make([]tradeCalendarRecord, 0, len(calendars))
	for _, calendar := range calendars {
		records = append(records, toTradeCalendarRecord(calendar))
	}
	if len(records) == 0 {
		return nil
	}

	err := upsertMySQLRecords(ctx, r.db, records, clause.OnConflict{
		Columns: []clause.Column{{Name: "market"}, {Name: "exchange"}, {Name: "cal_date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"is_open", "pretrade_date", "data_source", "synced_at", "updated_at",
		}),
	})
	if err != nil {
		return fmt.Errorf("upsert trade calendars: %w", err)
	}
	return nil
}

func (r *mysqlTradeCalendarRepository) ListTradeCalendars(ctx context.Context, query ListTradeCalendarsQuery) ([]TradeCalendar, error) {
	db := r.db.WithContext(ctx).Model(&tradeCalendarRecord{}).
		Where("market = ? AND exchange = ?", query.Market, query.Exchange)
	if !query.StartDate.IsZero() {
		db = db.Where("cal_date >= ?", dateOnly(query.StartDate))
	}
	if !query.EndDate.IsZero() {
		db = db.Where("cal_date <= ?", dateOnly(query.EndDate))
	}

	var records []tradeCalendarRecord
	if err := db.Order("cal_date ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list trade calendars: %w", err)
	}
	items := make([]TradeCalendar, 0, len(records))
	for _, record := range records {
		items = append(items, record.toDomain())
	}
	return items, nil
}

func (r *mysqlTradeCalendarRepository) LatestOpenDay(ctx context.Context, market Market, exchange Exchange) (time.Time, error) {
	var record tradeCalendarRecord
	err := r.db.WithContext(ctx).
		Where("market = ? AND exchange = ? AND is_open = ?", market, exchange, true).
		Order("cal_date DESC").
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return time.Time{}, NewError(CodeTradeCalendarNotInitialized, "trade calendar not initialized")
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("latest open day: %w", err)
	}
	return record.CalDate, nil
}

func (r *mysqlTradeCalendarRepository) IsOpenDay(ctx context.Context, market Market, exchange Exchange, date time.Time) (bool, error) {
	var record tradeCalendarRecord
	err := r.db.WithContext(ctx).
		Where("market = ? AND exchange = ? AND cal_date = ?", market, exchange, dateOnly(date)).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, NewError(CodeTradeCalendarNotFound, "trade calendar not found")
	}
	if err != nil {
		return false, fmt.Errorf("is open day: %w", err)
	}
	return record.IsOpen, nil
}

func (r *mysqlSyncTaskRepository) CreateSyncTask(ctx context.Context, task SyncTask) error {
	record := toSyncTaskRecord(task)
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("create sync task: %w", err)
	}
	return nil
}

func (r *mysqlSyncTaskRepository) UpdateSyncTask(ctx context.Context, task SyncTask) error {
	record := toSyncTaskRecord(task)
	result := r.db.WithContext(ctx).Model(&syncTaskRecord{}).
		Where("uid = ?", task.UID).
		Updates(map[string]any{
			"task_type":       record.TaskType,
			"market":          record.Market,
			"asset_type":      record.AssetType,
			"data_source":     record.DataSource,
			"status":          record.Status,
			"total_items":     record.TotalItems,
			"processed_items": record.ProcessedItems,
			"success_items":   record.SuccessItems,
			"failed_items":    record.FailedItems,
			"request_id":      record.RequestID,
			"started_at":      record.StartedAt,
			"finished_at":     record.FinishedAt,
			"error_msg":       record.ErrorMsg,
			"updated_at":      record.UpdatedAt,
		})
	if result.Error != nil {
		return fmt.Errorf("update sync task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewError(CodeSyncTaskNotFound, "sync task not found")
	}
	return nil
}

func (r *mysqlSyncTaskRepository) FindSyncTask(ctx context.Context, uid string) (SyncTask, error) {
	record, err := r.findSyncTaskRecord(ctx, uid)
	if err != nil {
		return SyncTask{}, err
	}
	return record.toDomain(), nil
}

func (r *mysqlSyncTaskRepository) FindActiveSyncTask(ctx context.Context) (SyncTask, error) {
	var record syncTaskRecord
	err := r.db.WithContext(ctx).
		Where("status IN ?", []string{string(SyncTaskStatusPending), string(SyncTaskStatusRunning)}).
		Order("created_at ASC").
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return SyncTask{}, NewError(CodeSyncTaskNotFound, "sync task not found")
	}
	if err != nil {
		return SyncTask{}, fmt.Errorf("find active sync task: %w", err)
	}
	return record.toDomain(), nil
}

func (r *mysqlSyncTaskRepository) ListSyncTasks(ctx context.Context, query ListSyncTasksQuery) (ListSyncTasksResult, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Model(&syncTaskRecord{})
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListSyncTasksResult{}, fmt.Errorf("count sync tasks: %w", err)
	}

	var records []syncTaskRecord
	if err := db.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
		return ListSyncTasksResult{}, fmt.Errorf("list sync tasks: %w", err)
	}
	items := make([]SyncTask, 0, len(records))
	for _, record := range records {
		items = append(items, record.toDomain())
	}
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

func (r *mysqlSyncTaskRepository) MarkStaleTasksFailed(ctx context.Context) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&syncTaskRecord{}).
		Where("status IN ?", []string{string(SyncTaskStatusPending), string(SyncTaskStatusRunning)}).
		Updates(map[string]any{
			"status":      string(SyncTaskStatusFailed),
			"finished_at": now,
			"error_msg":   "service restarted before task completed",
			"updated_at":  now,
		}).Error
	if err != nil {
		return fmt.Errorf("mark stale sync tasks failed: %w", err)
	}
	return nil
}

func (r *mysqlSyncTaskRepository) AppendSyncLogs(ctx context.Context, logs []SyncLog) error {
	if len(logs) == 0 {
		return nil
	}
	records := make([]syncLogRecord, 0, len(logs))
	for _, log := range logs {
		task, err := r.findSyncTaskRecord(ctx, log.TaskUID)
		if err != nil {
			return err
		}
		records = append(records, toSyncLogRecord(log, task.ID))
	}
	if err := r.db.WithContext(ctx).Create(&records).Error; err != nil {
		return fmt.Errorf("append sync logs: %w", err)
	}
	return nil
}

func (r *mysqlSyncTaskRepository) ListSyncLogs(ctx context.Context, query ListSyncLogsQuery) (ListSyncLogsResult, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Model(&syncLogRecord{}).Where("task_uid = ?", query.TaskUID)
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Symbol != "" {
		db = db.Where("symbol = ?", query.Symbol)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListSyncLogsResult{}, fmt.Errorf("count sync logs: %w", err)
	}

	var records []syncLogRecord
	if err := db.Order("created_at ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
		return ListSyncLogsResult{}, fmt.Errorf("list sync logs: %w", err)
	}
	items := make([]SyncLog, 0, len(records))
	for _, record := range records {
		items = append(items, record.toDomain())
	}
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

func (r *mysqlSyncTaskRepository) findSyncTaskRecord(ctx context.Context, uid string) (syncTaskRecord, error) {
	var record syncTaskRecord
	err := r.db.WithContext(ctx).Where("uid = ?", uid).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return syncTaskRecord{}, NewError(CodeSyncTaskNotFound, "sync task not found")
	}
	if err != nil {
		return syncTaskRecord{}, fmt.Errorf("find sync task: %w", err)
	}
	return record, nil
}

func applyStockFilters(db *gorm.DB, query ListStocksQuery) *gorm.DB {
	if query.Market != "" {
		db = db.Where("market = ?", query.Market)
	}
	if query.AssetType != "" {
		db = db.Where("asset_type = ?", query.AssetType)
	}
	if query.Exchange != "" {
		db = db.Where("exchange = ?", query.Exchange)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("symbol LIKE ? OR name LIKE ?", like, like)
	}
	return db
}

func toStockRecord(stock Stock) stockRecord {
	return stockRecord{
		Market:     string(stock.Market),
		AssetType:  string(stock.AssetType),
		Symbol:     stock.Symbol,
		TSCode:     stock.TSCode,
		Name:       stock.Name,
		Exchange:   string(stock.Exchange),
		Board:      string(stock.Board),
		Area:       stock.Area,
		Industry:   stock.Industry,
		Status:     string(stock.Status),
		ListDate:   datePtr(stock.ListDate),
		DelistDate: datePtr(stock.DelistDate),
		DataSource: string(stock.DataSource),
		SyncedAt:   stock.SyncedAt,
		CreatedAt:  stock.CreatedAt,
		UpdatedAt:  stock.UpdatedAt,
	}
}

func (r stockRecord) toDomain() Stock {
	return Stock{
		Market:     Market(r.Market),
		AssetType:  AssetType(r.AssetType),
		Symbol:     r.Symbol,
		TSCode:     r.TSCode,
		Name:       r.Name,
		Exchange:   Exchange(r.Exchange),
		Board:      Board(r.Board),
		Area:       r.Area,
		Industry:   r.Industry,
		Status:     StockStatus(r.Status),
		ListDate:   datePtr(r.ListDate),
		DelistDate: datePtr(r.DelistDate),
		DataSource: DataSource(r.DataSource),
		SyncedAt:   r.SyncedAt,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func toDailyKLineRecord(line DailyKLine) dailyKLineRecord {
	return dailyKLineRecord{
		Market:     string(line.Market),
		AssetType:  string(line.AssetType),
		Symbol:     line.Symbol,
		TradeDate:  dateOnly(line.TradeDate),
		OpenPrice:  line.OpenPrice,
		HighPrice:  line.HighPrice,
		LowPrice:   line.LowPrice,
		ClosePrice: line.ClosePrice,
		PreClose:   line.PreClose,
		ChangeAmt:  line.ChangeAmt,
		PctChange:  line.PctChange,
		Volume:     line.Volume,
		Amount:     line.Amount,
		DataSource: string(line.DataSource),
		SyncedAt:   line.SyncedAt,
		CreatedAt:  line.CreatedAt,
		UpdatedAt:  line.UpdatedAt,
	}
}

func (r dailyKLineRecord) toDomain() DailyKLine {
	return DailyKLine{
		Market:     Market(r.Market),
		AssetType:  AssetType(r.AssetType),
		Symbol:     r.Symbol,
		TradeDate:  r.TradeDate,
		OpenPrice:  r.OpenPrice,
		HighPrice:  r.HighPrice,
		LowPrice:   r.LowPrice,
		ClosePrice: r.ClosePrice,
		PreClose:   r.PreClose,
		ChangeAmt:  r.ChangeAmt,
		PctChange:  r.PctChange,
		Volume:     r.Volume,
		Amount:     r.Amount,
		DataSource: DataSource(r.DataSource),
		SyncedAt:   r.SyncedAt,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func toTradeCalendarRecord(calendar TradeCalendar) tradeCalendarRecord {
	return tradeCalendarRecord{
		Market:       string(calendar.Market),
		Exchange:     string(calendar.Exchange),
		CalDate:      dateOnly(calendar.CalDate),
		IsOpen:       calendar.IsOpen,
		PretradeDate: datePtr(calendar.PretradeDate),
		DataSource:   string(calendar.DataSource),
		SyncedAt:     calendar.SyncedAt,
		CreatedAt:    calendar.CreatedAt,
		UpdatedAt:    calendar.UpdatedAt,
	}
}

func (r tradeCalendarRecord) toDomain() TradeCalendar {
	return TradeCalendar{
		Market:       Market(r.Market),
		Exchange:     Exchange(r.Exchange),
		CalDate:      r.CalDate,
		IsOpen:       r.IsOpen,
		PretradeDate: datePtr(r.PretradeDate),
		DataSource:   DataSource(r.DataSource),
		SyncedAt:     r.SyncedAt,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func toSyncTaskRecord(task SyncTask) syncTaskRecord {
	return syncTaskRecord{
		UID:            task.UID,
		TaskType:       string(task.TaskType),
		Market:         string(task.Market),
		AssetType:      string(task.AssetType),
		DataSource:     string(task.DataSource),
		Status:         string(task.Status),
		TotalItems:     task.TotalItems,
		ProcessedItems: task.ProcessedItems,
		SuccessItems:   task.SuccessItems,
		FailedItems:    task.FailedItems,
		RequestID:      task.RequestID,
		StartedAt:      dateTimePtr(task.StartedAt),
		FinishedAt:     dateTimePtr(task.FinishedAt),
		ErrorMsg:       task.ErrorMsg,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}
}

func (r syncTaskRecord) toDomain() SyncTask {
	return SyncTask{
		UID:            r.UID,
		TaskType:       SyncTaskType(r.TaskType),
		Market:         Market(r.Market),
		AssetType:      AssetType(r.AssetType),
		DataSource:     DataSource(r.DataSource),
		Status:         SyncTaskStatus(r.Status),
		TotalItems:     r.TotalItems,
		ProcessedItems: r.ProcessedItems,
		SuccessItems:   r.SuccessItems,
		FailedItems:    r.FailedItems,
		RequestID:      r.RequestID,
		StartedAt:      dateTimePtr(r.StartedAt),
		FinishedAt:     dateTimePtr(r.FinishedAt),
		ErrorMsg:       r.ErrorMsg,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

func toSyncLogRecord(log SyncLog, taskID uint64) syncLogRecord {
	return syncLogRecord{
		TaskID:       taskID,
		TaskUID:      log.TaskUID,
		Step:         log.Step,
		Status:       string(log.Status),
		Market:       string(log.Market),
		AssetType:    string(log.AssetType),
		Symbol:       log.Symbol,
		DataSource:   string(log.DataSource),
		Message:      log.Message,
		ErrorDetail:  log.ErrorDetail,
		AffectedRows: log.AffectedRows,
		CreatedAt:    log.CreatedAt,
	}
}

func (r syncLogRecord) toDomain() SyncLog {
	return SyncLog{
		TaskUID:      r.TaskUID,
		Step:         r.Step,
		Status:       SyncLogStatus(r.Status),
		Market:       Market(r.Market),
		AssetType:    AssetType(r.AssetType),
		Symbol:       r.Symbol,
		DataSource:   DataSource(r.DataSource),
		Message:      r.Message,
		ErrorDetail:  r.ErrorDetail,
		AffectedRows: r.AffectedRows,
		CreatedAt:    r.CreatedAt,
	}
}

func datePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	date := dateOnly(*value)
	return &date
}

func dateTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
