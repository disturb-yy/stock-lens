package market

import "context"

type mockInstrumentProvider struct {
	stocks []Stock
	err    error
}

type mockCalendarProvider struct {
	calendars []TradeCalendar
	err       error
}

type mockMarketDataProvider struct {
	lines []DailyKLine
	err   error
}

func NewMockInstrumentProvider(stocks []Stock) InstrumentProvider {
	return &mockInstrumentProvider{
		stocks: append([]Stock(nil), stocks...),
	}
}

func NewFailingMockInstrumentProvider(err error) InstrumentProvider {
	return &mockInstrumentProvider{err: err}
}

func NewMockCalendarProvider(calendars []TradeCalendar) CalendarProvider {
	return &mockCalendarProvider{
		calendars: append([]TradeCalendar(nil), calendars...),
	}
}

func NewFailingMockCalendarProvider(err error) CalendarProvider {
	return &mockCalendarProvider{err: err}
}

func NewMockMarketDataProvider(lines []DailyKLine) MarketDataProvider {
	return &mockMarketDataProvider{
		lines: append([]DailyKLine(nil), lines...),
	}
}

func NewFailingMockMarketDataProvider(err error) MarketDataProvider {
	return &mockMarketDataProvider{err: err}
}

func (p *mockInstrumentProvider) DataSource() DataSource {
	return DataSourceMock
}

func (p *mockCalendarProvider) DataSource() DataSource {
	return DataSourceMock
}

func (p *mockMarketDataProvider) DataSource() DataSource {
	return DataSourceMock
}

func (p *mockInstrumentProvider) FetchStocks(_ context.Context, req FetchStocksRequest) ([]Stock, error) {
	if p.err != nil {
		return nil, p.err
	}

	items := make([]Stock, 0)
	for _, stock := range p.stocks {
		if req.Market != "" && stock.Market != req.Market {
			continue
		}
		stock.DataSource = DataSourceMock
		items = append(items, stock)
	}
	return items, nil
}

func (p *mockCalendarProvider) FetchTradeCalendars(_ context.Context, req FetchTradeCalendarsRequest) ([]TradeCalendar, error) {
	if p.err != nil {
		return nil, p.err
	}

	items := make([]TradeCalendar, 0)
	for _, calendar := range p.calendars {
		if !matchProviderTradeCalendar(calendar, req) {
			continue
		}
		calendar.DataSource = DataSourceMock
		items = append(items, calendar)
	}
	return items, nil
}

func (p *mockMarketDataProvider) FetchDailyKLines(_ context.Context, req FetchDailyKLinesRequest) ([]DailyKLine, error) {
	if p.err != nil {
		return nil, p.err
	}

	items := make([]DailyKLine, 0)
	for _, line := range p.lines {
		if !matchProviderDailyKLine(line, req) {
			continue
		}
		line.DataSource = DataSourceMock
		items = append(items, line)
	}
	return items, nil
}

func matchProviderTradeCalendar(calendar TradeCalendar, req FetchTradeCalendarsRequest) bool {
	if req.Market != "" && calendar.Market != req.Market {
		return false
	}
	if req.Exchange != "" && calendar.Exchange != req.Exchange {
		return false
	}
	return withinDateRange(calendar.CalDate, req.StartDate, req.EndDate)
}

func matchProviderDailyKLine(line DailyKLine, req FetchDailyKLinesRequest) bool {
	if req.Market != "" && line.Market != req.Market {
		return false
	}
	if req.AssetType != "" && line.AssetType != req.AssetType {
		return false
	}
	if req.Symbol != "" && line.Symbol != req.Symbol {
		return false
	}
	return withinDateRange(line.TradeDate, req.StartDate, req.EndDate)
}
