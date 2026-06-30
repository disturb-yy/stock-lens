package market

import (
	"context"
	"time"
)

const (
	maxPageSize             = 100
	dailyKLineDefaultYears  = 1
	dailyKLineMaxYears      = 5
	tradeCalendarRangeYears = 1
	tradeCalendarMaxYears   = 30
	defaultQueryServiceTZ   = "Asia/Shanghai"
)

type QueryService struct {
	stocks         StockRepository
	kLines         KLineRepository
	tradeCalendars TradeCalendarRepository
	syncTasks      SyncTaskRepository
	now            func() time.Time
}

type StockDetailResult struct {
	Stock            Stock
	LatestDailyKLine *DailyKLine
}

type IsOpenDayQuery struct {
	Market   Market
	Exchange Exchange
	Date     time.Time
}

type IsOpenDayResult struct {
	Market       Market
	Exchange     Exchange
	Date         time.Time
	IsOpen       bool
	PretradeDate *time.Time
}

type EnumMetadataResult struct {
	Markets          []EnumValue
	AssetTypes       []EnumValue
	Exchanges        []EnumValue
	Boards           []EnumValue
	StockStatuses    []EnumValue
	DataSources      []EnumValue
	SyncTaskTypes    []EnumValue
	SyncTaskStatuses []EnumValue
}

type EnumValue struct {
	Value string
}

func NewQueryService(stocks StockRepository, kLines KLineRepository, tradeCalendars TradeCalendarRepository, syncTasks SyncTaskRepository) *QueryService {
	return NewQueryServiceWithClock(stocks, kLines, tradeCalendars, syncTasks, defaultNow)
}

func NewQueryServiceWithClock(stocks StockRepository, kLines KLineRepository, tradeCalendars TradeCalendarRepository, syncTasks SyncTaskRepository, now func() time.Time) *QueryService {
	if now == nil {
		now = defaultNow
	}
	return &QueryService{
		stocks:         stocks,
		kLines:         kLines,
		tradeCalendars: tradeCalendars,
		syncTasks:      syncTasks,
		now:            now,
	}
}

func (s *QueryService) ListStocks(ctx context.Context, query ListStocksQuery) (ListStocksResult, error) {
	query.Market = defaultMarket(query.Market)
	query.AssetType = defaultAssetType(query.AssetType)

	if err := validateMarketAsset(query.Market, query.AssetType); err != nil {
		return ListStocksResult{}, err
	}
	if err := validatePageQuery(query.PageQuery); err != nil {
		return ListStocksResult{}, err
	}
	if query.Exchange != "" && !query.Exchange.Valid() {
		return ListStocksResult{}, NewError(CodeInvalidExchange, "invalid exchange")
	}
	if query.Status != "" && !query.Status.Valid() {
		return ListStocksResult{}, NewError(CodeInvalidStatus, "invalid status")
	}

	return s.stocks.ListStocks(ctx, query)
}

func (s *QueryService) GetStockDetail(ctx context.Context, query FindStockQuery) (StockDetailResult, error) {
	query, err := s.normalizeFindStockQuery(query)
	if err != nil {
		return StockDetailResult{}, err
	}

	stock, err := s.stocks.FindStock(ctx, query)
	if err != nil {
		return StockDetailResult{}, err
	}

	latest, err := s.kLines.LatestDailyKLine(ctx, query.Market, query.AssetType, query.Symbol)
	if err != nil {
		return StockDetailResult{}, err
	}

	var latestPtr *DailyKLine
	if !latest.TradeDate.IsZero() {
		latestPtr = &latest
	}
	return StockDetailResult{
		Stock:            stock,
		LatestDailyKLine: latestPtr,
	}, nil
}

func (s *QueryService) ListDailyKLines(ctx context.Context, query ListDailyKLinesQuery) ([]DailyKLine, error) {
	normalized, err := s.normalizeDailyKLineQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	if normalized.EndDate.IsZero() {
		return []DailyKLine{}, nil
	}

	return s.kLines.ListDailyKLines(ctx, normalized)
}

func (s *QueryService) LatestDailyKLine(ctx context.Context, query FindStockQuery) (*DailyKLine, error) {
	query, err := s.normalizeFindStockQuery(query)
	if err != nil {
		return nil, err
	}
	if _, err := s.stocks.FindStock(ctx, query); err != nil {
		return nil, err
	}

	latest, err := s.kLines.LatestDailyKLine(ctx, query.Market, query.AssetType, query.Symbol)
	if err != nil {
		return nil, err
	}
	if latest.TradeDate.IsZero() {
		return nil, nil
	}
	return &latest, nil
}

func (s *QueryService) ListTradeCalendars(ctx context.Context, query ListTradeCalendarsQuery) ([]TradeCalendar, error) {
	query.Market = defaultMarket(query.Market)
	query.Exchange = defaultExchange(query.Exchange)
	if err := validateMarketExchange(query.Market, query.Exchange); err != nil {
		return nil, err
	}

	// API 默认查询最近一年，锚点使用业务时区下的“今天”，而不是数据库最新日期。
	if query.EndDate.IsZero() {
		query.EndDate = dateOnly(s.now())
	}
	if query.StartDate.IsZero() {
		query.StartDate = query.EndDate.AddDate(-tradeCalendarRangeYears, 0, 0)
	}
	if err := validateDateRange(query.StartDate, query.EndDate, tradeCalendarMaxYears); err != nil {
		return nil, err
	}

	return s.tradeCalendars.ListTradeCalendars(ctx, query)
}

func (s *QueryService) LatestOpenDay(ctx context.Context, market Market, exchange Exchange) (time.Time, error) {
	market = defaultMarket(market)
	exchange = defaultExchange(exchange)
	if err := validateMarketExchange(market, exchange); err != nil {
		return time.Time{}, err
	}
	return s.tradeCalendars.LatestOpenDay(ctx, market, exchange)
}

func (s *QueryService) IsOpenDay(ctx context.Context, query IsOpenDayQuery) (IsOpenDayResult, error) {
	query.Market = defaultMarket(query.Market)
	query.Exchange = defaultExchange(query.Exchange)
	if query.Date.IsZero() {
		query.Date = dateOnly(s.now())
	}
	if err := validateMarketExchange(query.Market, query.Exchange); err != nil {
		return IsOpenDayResult{}, err
	}

	calendars, err := s.tradeCalendars.ListTradeCalendars(ctx, ListTradeCalendarsQuery{
		Market:    query.Market,
		Exchange:  query.Exchange,
		StartDate: query.Date,
		EndDate:   query.Date,
	})
	if err != nil {
		return IsOpenDayResult{}, err
	}
	if len(calendars) == 0 {
		return IsOpenDayResult{}, NewError(CodeTradeCalendarNotFound, "trade calendar not found")
	}

	calendar := calendars[0]
	return IsOpenDayResult{
		Market:       calendar.Market,
		Exchange:     calendar.Exchange,
		Date:         calendar.CalDate,
		IsOpen:       calendar.IsOpen,
		PretradeDate: calendar.PretradeDate,
	}, nil
}

func (s *QueryService) EnumMetadata() EnumMetadataResult {
	return EnumMetadataResult{
		Markets:          enumValues(string(MarketCN)),
		AssetTypes:       enumValues(string(AssetTypeStock)),
		Exchanges:        enumValues(string(ExchangeSSE), string(ExchangeSZSE), string(ExchangeBSE)),
		Boards:           enumValues(string(BoardMain), string(BoardGEM), string(BoardSTAR), string(BoardBSE), string(BoardUnknown)),
		StockStatuses:    enumValues(string(StockStatusListed), string(StockStatusDelisted), string(StockStatusPaused)),
		DataSources:      enumValues(string(DataSourceMock), string(DataSourceTushare)),
		SyncTaskTypes:    enumValues(string(SyncTaskTypeStockMaster), string(SyncTaskTypeTradeCalendars), string(SyncTaskTypeDailyKLines)),
		SyncTaskStatuses: enumValues(string(SyncTaskStatusPending), string(SyncTaskStatusRunning), string(SyncTaskStatusSuccess), string(SyncTaskStatusFailed), string(SyncTaskStatusPartialSuccess)),
	}
}

func (s *QueryService) GetSyncTask(ctx context.Context, uid string) (SyncTask, error) {
	return s.syncTasks.FindSyncTask(ctx, uid)
}

func (s *QueryService) ListSyncTasks(ctx context.Context, query ListSyncTasksQuery) (ListSyncTasksResult, error) {
	if err := validatePageQuery(query.PageQuery); err != nil {
		return ListSyncTasksResult{}, err
	}
	if query.Status != "" && !query.Status.Valid() {
		return ListSyncTasksResult{}, NewError(CodeInvalidTaskStatus, "invalid task status")
	}
	return s.syncTasks.ListSyncTasks(ctx, query)
}

func (s *QueryService) ListSyncLogs(ctx context.Context, query ListSyncLogsQuery) (ListSyncLogsResult, error) {
	if _, err := s.syncTasks.FindSyncTask(ctx, query.TaskUID); err != nil {
		return ListSyncLogsResult{}, err
	}
	if err := validatePageQuery(query.PageQuery); err != nil {
		return ListSyncLogsResult{}, err
	}
	if query.Status != "" && !query.Status.Valid() {
		return ListSyncLogsResult{}, NewError(CodeInvalidTaskStatus, "invalid task status")
	}
	if query.Symbol != "" && !ValidSymbol(query.Symbol) {
		return ListSyncLogsResult{}, NewError(CodeInvalidSymbol, "invalid symbol")
	}
	return s.syncTasks.ListSyncLogs(ctx, query)
}

func (s *QueryService) normalizeFindStockQuery(query FindStockQuery) (FindStockQuery, error) {
	query.Market = defaultMarket(query.Market)
	query.AssetType = defaultAssetType(query.AssetType)
	if err := validateMarketAsset(query.Market, query.AssetType); err != nil {
		return FindStockQuery{}, err
	}
	if !ValidSymbol(query.Symbol) {
		return FindStockQuery{}, NewError(CodeInvalidSymbol, "invalid symbol")
	}
	return query, nil
}

func (s *QueryService) normalizeDailyKLineQuery(ctx context.Context, query ListDailyKLinesQuery) (ListDailyKLinesQuery, error) {
	findQuery, err := s.normalizeFindStockQuery(FindStockQuery{
		Market:    query.Market,
		AssetType: query.AssetType,
		Symbol:    query.Symbol,
	})
	if err != nil {
		return ListDailyKLinesQuery{}, err
	}
	if _, err := s.stocks.FindStock(ctx, findQuery); err != nil {
		return ListDailyKLinesQuery{}, err
	}

	query.Market = findQuery.Market
	query.AssetType = findQuery.AssetType
	query.Symbol = findQuery.Symbol
	if query.EndDate.IsZero() {
		latest, err := s.kLines.LatestDailyKLine(ctx, query.Market, query.AssetType, query.Symbol)
		if err != nil {
			return ListDailyKLinesQuery{}, err
		}
		if latest.TradeDate.IsZero() {
			return query, nil
		}
		query.EndDate = latest.TradeDate
	}
	if query.StartDate.IsZero() {
		query.StartDate = query.EndDate.AddDate(-dailyKLineDefaultYears, 0, 0)
	}
	if err := validateDateRange(query.StartDate, query.EndDate, dailyKLineMaxYears); err != nil {
		return ListDailyKLinesQuery{}, err
	}
	return query, nil
}

func validatePageQuery(query PageQuery) error {
	if query.Page < 0 {
		return NewError(CodeInvalidPageSize, "invalid page")
	}
	if query.PageSize > maxPageSize {
		return NewError(CodeInvalidPageSize, "invalid page size")
	}
	return nil
}

func validateMarketAsset(market Market, assetType AssetType) error {
	if !market.Valid() {
		return NewError(CodeInvalidMarket, "invalid market")
	}
	if !assetType.Valid() {
		return NewError(CodeInvalidAssetType, "invalid asset type")
	}
	return nil
}

func validateMarketExchange(market Market, exchange Exchange) error {
	if !market.Valid() {
		return NewError(CodeInvalidMarket, "invalid market")
	}
	if !exchange.Valid() {
		return NewError(CodeInvalidExchange, "invalid exchange")
	}
	return nil
}

func validateDateRange(start time.Time, end time.Time, maxYears int) error {
	if !ValidDateRange(start, end) {
		return NewError(CodeInvalidDateRange, "invalid date range")
	}
	if start.AddDate(maxYears, 0, 0).Before(end) {
		return NewError(CodeDateRangeTooLarge, "date range too large")
	}
	return nil
}

func defaultMarket(market Market) Market {
	if market == "" {
		return MarketCN
	}
	return market
}

func defaultAssetType(assetType AssetType) AssetType {
	if assetType == "" {
		return AssetTypeStock
	}
	return assetType
}

func defaultExchange(exchange Exchange) Exchange {
	if exchange == "" {
		return ExchangeSSE
	}
	return exchange
}

func defaultNow() time.Time {
	loc, err := time.LoadLocation(defaultQueryServiceTZ)
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

func enumValues(values ...string) []EnumValue {
	items := make([]EnumValue, 0, len(values))
	for _, value := range values {
		items = append(items, EnumValue{Value: value})
	}
	return items
}
