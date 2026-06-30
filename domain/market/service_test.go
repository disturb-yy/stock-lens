package market

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestQueryServiceListStocksDefaultsAndFilters(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	service := NewQueryService(stocks, NewMockKLineRepository(), NewMockTradeCalendarRepository(), NewMockSyncTaskRepository())

	err := stocks.UpsertStocks(ctx, []Stock{
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", Name: "Kweichow Moutai", Exchange: ExchangeSSE, Status: StockStatusListed},
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "000001", Name: "Ping An Bank", Exchange: ExchangeSZSE, Status: StockStatusListed},
	})
	if err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	got, err := service.ListStocks(ctx, ListStocksQuery{
		PageQuery: PageQuery{Page: 1, PageSize: 10},
		Keyword:   "Moutai",
	})
	if err != nil {
		t.Fatalf("ListStocks() error = %v", err)
	}
	if got.Pagination.Total != 1 || len(got.Items) != 1 {
		t.Fatalf("stocks = %+v, want one filtered stock", got)
	}
	if got.Items[0].Symbol != "600519" {
		t.Fatalf("symbol = %q, want 600519", got.Items[0].Symbol)
	}
}

func TestQueryServiceListStocksRejectsInvalidPageSize(t *testing.T) {
	service := NewQueryService(NewMockStockRepository(), NewMockKLineRepository(), NewMockTradeCalendarRepository(), NewMockSyncTaskRepository())

	_, err := service.ListStocks(context.Background(), ListStocksQuery{
		PageQuery: PageQuery{Page: 1, PageSize: 101},
	})
	if !IsCode(err, CodeInvalidPageSize) {
		t.Fatalf("ListStocks() error = %v, want %s", err, CodeInvalidPageSize)
	}
}

func TestQueryServiceStockDetailIncludesLatestDailyKLine(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	lines := NewMockKLineRepository()
	service := NewQueryService(stocks, lines, NewMockTradeCalendarRepository(), NewMockSyncTaskRepository())

	err := stocks.UpsertStocks(ctx, []Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", Name: "Kweichow Moutai"}})
	if err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}
	err = lines.UpsertDailyKLines(ctx, []DailyKLine{
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", TradeDate: serviceTestDate(t, "2026-06-29"), ClosePrice: decimal.NewFromInt(10)},
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", TradeDate: serviceTestDate(t, "2026-06-30"), ClosePrice: decimal.NewFromInt(11)},
	})
	if err != nil {
		t.Fatalf("UpsertDailyKLines() error = %v", err)
	}

	got, err := service.GetStockDetail(ctx, FindStockQuery{Symbol: "600519"})
	if err != nil {
		t.Fatalf("GetStockDetail() error = %v", err)
	}
	if got.Stock.Symbol != "600519" {
		t.Fatalf("stock symbol = %q, want 600519", got.Stock.Symbol)
	}
	if got.LatestDailyKLine == nil {
		t.Fatalf("LatestDailyKLine = nil, want latest line")
	}
	if got.LatestDailyKLine.TradeDate.Format(DateLayout) != "2026-06-30" {
		t.Fatalf("latest date = %s, want 2026-06-30", got.LatestDailyKLine.TradeDate.Format(DateLayout))
	}
}

func TestQueryServiceStockDetailAllowsMissingLatestDailyKLine(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	service := NewQueryService(stocks, NewMockKLineRepository(), NewMockTradeCalendarRepository(), NewMockSyncTaskRepository())

	err := stocks.UpsertStocks(ctx, []Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519"}})
	if err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	got, err := service.GetStockDetail(ctx, FindStockQuery{Symbol: "600519"})
	if err != nil {
		t.Fatalf("GetStockDetail() error = %v", err)
	}
	if got.LatestDailyKLine != nil {
		t.Fatalf("LatestDailyKLine = %+v, want nil", got.LatestDailyKLine)
	}
}

func TestQueryServiceGetStockDetailRejectsInvalidSymbol(t *testing.T) {
	service := NewQueryService(NewMockStockRepository(), NewMockKLineRepository(), NewMockTradeCalendarRepository(), NewMockSyncTaskRepository())

	_, err := service.GetStockDetail(context.Background(), FindStockQuery{Symbol: "600519.SH"})
	if !IsCode(err, CodeInvalidSymbol) {
		t.Fatalf("GetStockDetail() error = %v, want %s", err, CodeInvalidSymbol)
	}
}

func TestQueryServiceListDailyKLinesDefaultsToLatestOneYear(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	lines := NewMockKLineRepository()
	service := NewQueryService(stocks, lines, NewMockTradeCalendarRepository(), NewMockSyncTaskRepository())

	err := stocks.UpsertStocks(ctx, []Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519"}})
	if err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}
	err = lines.UpsertDailyKLines(ctx, []DailyKLine{
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", TradeDate: serviceTestDate(t, "2025-06-29"), ClosePrice: decimal.NewFromInt(9)},
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", TradeDate: serviceTestDate(t, "2025-06-30"), ClosePrice: decimal.NewFromInt(10)},
		{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", TradeDate: serviceTestDate(t, "2026-06-30"), ClosePrice: decimal.NewFromInt(11)},
	})
	if err != nil {
		t.Fatalf("UpsertDailyKLines() error = %v", err)
	}

	got, err := service.ListDailyKLines(ctx, ListDailyKLinesQuery{Symbol: "600519"})
	if err != nil {
		t.Fatalf("ListDailyKLines() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(lines) = %d, want 2", len(got))
	}
	if got[0].TradeDate.Format(DateLayout) != "2025-06-30" {
		t.Fatalf("first date = %s, want 2025-06-30", got[0].TradeDate.Format(DateLayout))
	}
}

func TestQueryServiceListDailyKLinesReturnsEmptyWhenStockHasNoLines(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	service := NewQueryService(stocks, NewMockKLineRepository(), NewMockTradeCalendarRepository(), NewMockSyncTaskRepository())

	err := stocks.UpsertStocks(ctx, []Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519"}})
	if err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	got, err := service.ListDailyKLines(ctx, ListDailyKLinesQuery{Symbol: "600519"})
	if err != nil {
		t.Fatalf("ListDailyKLines() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("lines = %+v, want empty slice", got)
	}
}

func TestQueryServiceListDailyKLinesRejectsLargeRange(t *testing.T) {
	ctx := context.Background()
	stocks := NewMockStockRepository()
	service := NewQueryService(stocks, NewMockKLineRepository(), NewMockTradeCalendarRepository(), NewMockSyncTaskRepository())

	err := stocks.UpsertStocks(ctx, []Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519"}})
	if err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	_, err = service.ListDailyKLines(ctx, ListDailyKLinesQuery{
		Symbol:    "600519",
		StartDate: serviceTestDate(t, "2020-01-01"),
		EndDate:   serviceTestDate(t, "2026-01-02"),
	})
	if !IsCode(err, CodeDateRangeTooLarge) {
		t.Fatalf("ListDailyKLines() error = %v, want %s", err, CodeDateRangeTooLarge)
	}
}

func TestQueryServiceTradeCalendarQueries(t *testing.T) {
	ctx := context.Background()
	calendars := NewMockTradeCalendarRepository()
	service := NewQueryServiceWithClock(
		NewMockStockRepository(),
		NewMockKLineRepository(),
		calendars,
		NewMockSyncTaskRepository(),
		func() time.Time { return serviceTestDate(t, "2026-06-30") },
	)
	openDay := serviceTestDate(t, "2026-06-29")
	closedDay := serviceTestDate(t, "2026-06-30")

	err := calendars.UpsertTradeCalendars(ctx, []TradeCalendar{
		{Market: MarketCN, Exchange: ExchangeSSE, CalDate: openDay, IsOpen: true},
		{Market: MarketCN, Exchange: ExchangeSSE, CalDate: closedDay, IsOpen: false, PretradeDate: &openDay},
	})
	if err != nil {
		t.Fatalf("UpsertTradeCalendars() error = %v", err)
	}

	list, err := service.ListTradeCalendars(ctx, ListTradeCalendarsQuery{})
	if err != nil {
		t.Fatalf("ListTradeCalendars() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(calendars) = %d, want 2", len(list))
	}

	latest, err := service.LatestOpenDay(ctx, MarketCN, ExchangeSSE)
	if err != nil {
		t.Fatalf("LatestOpenDay() error = %v", err)
	}
	if !latest.Equal(openDay) {
		t.Fatalf("latest = %s, want %s", latest.Format(DateLayout), openDay.Format(DateLayout))
	}

	isOpen, err := service.IsOpenDay(ctx, IsOpenDayQuery{})
	if err != nil {
		t.Fatalf("IsOpenDay() error = %v", err)
	}
	if isOpen.IsOpen {
		t.Fatalf("IsOpen = true, want false")
	}
	if isOpen.PretradeDate == nil || !isOpen.PretradeDate.Equal(openDay) {
		t.Fatalf("PretradeDate = %v, want %s", isOpen.PretradeDate, openDay.Format(DateLayout))
	}
}

func TestQueryServiceEnumMetadata(t *testing.T) {
	service := NewQueryService(NewMockStockRepository(), NewMockKLineRepository(), NewMockTradeCalendarRepository(), NewMockSyncTaskRepository())

	got := service.EnumMetadata()
	assertEnumContains(t, got.Markets, string(MarketCN))
	assertEnumContains(t, got.AssetTypes, string(AssetTypeStock))
	assertEnumContains(t, got.Exchanges, string(ExchangeSSE))
	assertEnumContains(t, got.StockStatuses, string(StockStatusListed))
	assertEnumContains(t, got.DataSources, string(DataSourceMock))
	assertEnumContains(t, got.SyncTaskTypes, string(SyncTaskTypeDailyKLines))
	assertEnumContains(t, got.SyncTaskStatuses, string(SyncTaskStatusPartialSuccess))
}

func serviceTestDate(t *testing.T, value string) time.Time {
	t.Helper()

	date, err := ParseDate(value)
	if err != nil {
		t.Fatalf("ParseDate(%q) error = %v", value, err)
	}
	return date
}

func assertEnumContains(t *testing.T, values []EnumValue, want string) {
	t.Helper()

	for _, value := range values {
		if value.Value == want {
			return
		}
	}
	t.Fatalf("enum values = %+v, want %s", values, want)
}
