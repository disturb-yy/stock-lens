package market

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"stock-lens/pkg/tushare"
)

const tushareDateLayout = "20060102"

type tushareInstrumentProvider struct {
	client *tushare.Client
	now    func() time.Time
}

type tushareCalendarProvider struct {
	client *tushare.Client
	now    func() time.Time
}

type tushareMarketDataProvider struct {
	client *tushare.Client
	now    func() time.Time
}

func NewTushareInstrumentProvider(client *tushare.Client) InstrumentProvider {
	return &tushareInstrumentProvider{client: client, now: time.Now}
}

func NewTushareCalendarProvider(client *tushare.Client) CalendarProvider {
	return &tushareCalendarProvider{client: client, now: time.Now}
}

func NewTushareMarketDataProvider(client *tushare.Client) MarketDataProvider {
	return &tushareMarketDataProvider{client: client, now: time.Now}
}

func (p *tushareInstrumentProvider) DataSource() DataSource {
	return DataSourceTushare
}

func (p *tushareCalendarProvider) DataSource() DataSource {
	return DataSourceTushare
}

func (p *tushareMarketDataProvider) DataSource() DataSource {
	return DataSourceTushare
}

func (p *tushareInstrumentProvider) FetchStocks(ctx context.Context, req FetchStocksRequest) ([]Stock, error) {
	if req.Market != MarketCN {
		return nil, NewError(CodeInvalidMarket, "invalid market")
	}
	// Tushare stock_basic 需要按上市状态分别拉取，空状态只返回默认上市集合。
	statuses := []string{"L", "D", "P"}
	items := make([]Stock, 0)
	for _, status := range statuses {
		resp, err := p.client.Call(ctx, "stock_basic", map[string]any{
			"exchange":    "",
			"list_status": status,
		}, tushare.StockBasicFields())
		if err != nil {
			return nil, wrapProviderError(err)
		}

		for _, row := range tushare.Rows(resp) {
			stock, err := mapTushareStock(row, p.now())
			if err != nil {
				// 实盘返回可能混入非 A 股 ts_code，股票主数据同步跳过坏行而不是整体失败。
				continue
			}
			items = append(items, stock)
		}
	}
	if len(items) == 0 {
		return nil, NewError(CodeProviderError, "empty stock list")
	}
	return items, nil
}

func (p *tushareCalendarProvider) FetchTradeCalendars(ctx context.Context, req FetchTradeCalendarsRequest) ([]TradeCalendar, error) {
	if req.Market != MarketCN {
		return nil, NewError(CodeInvalidMarket, "invalid market")
	}
	resp, err := p.client.Call(ctx, "trade_cal", map[string]any{
		"exchange":   tushareExchangeParam(req.Exchange),
		"start_date": formatTushareDate(req.StartDate),
		"end_date":   formatTushareDate(req.EndDate),
	}, tushare.TradeCalFields())
	if err != nil {
		return nil, wrapProviderError(err)
	}

	rows := tushare.Rows(resp)
	if len(rows) == 0 {
		return nil, NewError(CodeProviderError, "empty trade calendar")
	}
	items := make([]TradeCalendar, 0, len(rows))
	for _, row := range rows {
		calendar, err := mapTushareTradeCalendar(row, req.Market, p.now())
		if err != nil {
			return nil, wrapProviderError(err)
		}
		items = append(items, calendar)
	}
	return items, nil
}

func (p *tushareMarketDataProvider) FetchDailyKLines(ctx context.Context, req FetchDailyKLinesRequest) ([]DailyKLine, error) {
	if req.Market != MarketCN {
		return nil, NewError(CodeInvalidMarket, "invalid market")
	}
	resp, err := p.client.Call(ctx, "daily", map[string]any{
		"ts_code":    tsCodeFromSymbol(req.Symbol),
		"start_date": formatTushareDate(req.StartDate),
		"end_date":   formatTushareDate(req.EndDate),
	}, tushare.DailyFields())
	if err != nil {
		return nil, wrapProviderError(err)
	}

	rows := tushare.Rows(resp)
	items := make([]DailyKLine, 0, len(rows))
	for _, row := range rows {
		line, err := mapTushareDailyKLine(row, req.Market, req.AssetType, p.now())
		if err != nil {
			return nil, wrapProviderError(err)
		}
		items = append(items, line)
	}
	return items, nil
}

func mapTushareStock(row map[string]string, syncedAt time.Time) (Stock, error) {
	symbol, exchange, err := symbolExchangeFromTSCode(row["ts_code"])
	if err != nil {
		return Stock{}, err
	}
	status, err := mapTushareStockStatus(row["list_status"])
	if err != nil {
		return Stock{}, err
	}
	listDate, err := parseOptionalTushareDate(row["list_date"])
	if err != nil {
		return Stock{}, err
	}
	delistDate, err := parseOptionalTushareDate(row["delist_date"])
	if err != nil {
		return Stock{}, err
	}

	return Stock{
		Market:     MarketCN,
		AssetType:  AssetTypeStock,
		Symbol:     symbol,
		TSCode:     row["ts_code"],
		Name:       row["name"],
		Exchange:   exchange,
		Board:      inferBoard(symbol, exchange),
		Area:       row["area"],
		Industry:   row["industry"],
		Status:     status,
		ListDate:   listDate,
		DelistDate: delistDate,
		DataSource: DataSourceTushare,
		SyncedAt:   syncedAt,
	}, nil
}

func mapTushareTradeCalendar(row map[string]string, market Market, syncedAt time.Time) (TradeCalendar, error) {
	calDate, err := parseRequiredTushareDate(row["cal_date"])
	if err != nil {
		return TradeCalendar{}, err
	}
	pretradeDate, err := parseOptionalTushareDate(row["pretrade_date"])
	if err != nil {
		return TradeCalendar{}, err
	}
	exchange, err := mapTushareExchange(row["exchange"])
	if err != nil {
		return TradeCalendar{}, err
	}

	return TradeCalendar{
		Market:       market,
		Exchange:     exchange,
		CalDate:      calDate,
		IsOpen:       row["is_open"] == "1",
		PretradeDate: pretradeDate,
		DataSource:   DataSourceTushare,
		SyncedAt:     syncedAt,
	}, nil
}

func mapTushareDailyKLine(row map[string]string, market Market, assetType AssetType, syncedAt time.Time) (DailyKLine, error) {
	symbol, _, err := symbolExchangeFromTSCode(row["ts_code"])
	if err != nil {
		return DailyKLine{}, err
	}
	tradeDate, err := parseRequiredTushareDate(row["trade_date"])
	if err != nil {
		return DailyKLine{}, err
	}

	openPrice, err := parseDecimalField(row, "open")
	if err != nil {
		return DailyKLine{}, err
	}
	highPrice, err := parseDecimalField(row, "high")
	if err != nil {
		return DailyKLine{}, err
	}
	lowPrice, err := parseDecimalField(row, "low")
	if err != nil {
		return DailyKLine{}, err
	}
	closePrice, err := parseDecimalField(row, "close")
	if err != nil {
		return DailyKLine{}, err
	}
	preClose, err := parseDecimalField(row, "pre_close")
	if err != nil {
		return DailyKLine{}, err
	}
	changeAmt, err := parseDecimalField(row, "change")
	if err != nil {
		return DailyKLine{}, err
	}
	pctChange, err := parseDecimalField(row, "pct_chg")
	if err != nil {
		return DailyKLine{}, err
	}
	volume, err := parseDecimalField(row, "vol")
	if err != nil {
		return DailyKLine{}, err
	}
	amount, err := parseDecimalField(row, "amount")
	if err != nil {
		return DailyKLine{}, err
	}

	return DailyKLine{
		Market:     market,
		AssetType:  assetType,
		Symbol:     symbol,
		TradeDate:  tradeDate,
		OpenPrice:  openPrice,
		HighPrice:  highPrice,
		LowPrice:   lowPrice,
		ClosePrice: closePrice,
		PreClose:   preClose,
		ChangeAmt:  changeAmt,
		PctChange:  pctChange,
		Volume:     volume,
		Amount:     amount,
		DataSource: DataSourceTushare,
		SyncedAt:   syncedAt,
	}, nil
}

func symbolExchangeFromTSCode(tsCode string) (string, Exchange, error) {
	// Tushare 标识按 provider 身份保存，领域内查询仍统一使用 6 位 symbol。
	tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
	parts := strings.Split(tsCode, ".")
	if len(parts) != 2 || !ValidSymbol(parts[0]) {
		return "", "", fmt.Errorf("invalid tushare ts_code")
	}
	exchange, err := mapTushareExchange(parts[1])
	if err != nil {
		return "", "", err
	}
	return parts[0], exchange, nil
}

func mapTushareExchange(value string) (Exchange, error) {
	switch value {
	case "SH", "SSE":
		return ExchangeSSE, nil
	case "SZ", "SZSE":
		return ExchangeSZSE, nil
	case "BJ", "BSE":
		return ExchangeBSE, nil
	default:
		return "", fmt.Errorf("invalid tushare exchange")
	}
}

func tushareExchangeParam(exchange Exchange) string {
	switch exchange {
	case ExchangeSSE:
		return "SSE"
	case ExchangeSZSE:
		return "SZSE"
	case ExchangeBSE:
		return "BSE"
	default:
		return ""
	}
}

func mapTushareStockStatus(value string) (StockStatus, error) {
	switch value {
	case "L":
		return StockStatusListed, nil
	case "D":
		return StockStatusDelisted, nil
	case "P":
		return StockStatusPaused, nil
	default:
		return "", fmt.Errorf("invalid tushare stock status")
	}
}

func inferBoard(symbol string, exchange Exchange) Board {
	if exchange == ExchangeBSE || strings.HasPrefix(symbol, "8") || strings.HasPrefix(symbol, "4") {
		return BoardBSE
	}
	if strings.HasPrefix(symbol, "688") {
		return BoardSTAR
	}
	if strings.HasPrefix(symbol, "300") {
		return BoardGEM
	}
	if symbol == "" {
		return BoardUnknown
	}
	return BoardMain
}

func parseRequiredTushareDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("missing tushare date")
	}
	date, err := time.Parse(tushareDateLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse tushare date: %w", err)
	}
	return date, nil
}

func parseOptionalTushareDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	date, err := parseRequiredTushareDate(value)
	if err != nil {
		return nil, err
	}
	return &date, nil
}

func parseDecimalField(row map[string]string, field string) (decimal.Decimal, error) {
	value := strings.TrimSpace(row[field])
	if value == "" {
		return decimal.Decimal{}, fmt.Errorf("missing decimal field %s", field)
	}
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("parse decimal field %s: %w", field, err)
	}
	return parsed, nil
}

func formatTushareDate(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format(tushareDateLayout)
}

func tsCodeFromSymbol(symbol string) string {
	if symbol == "" {
		return ""
	}
	if strings.HasPrefix(symbol, "6") {
		return symbol + ".SH"
	}
	if strings.HasPrefix(symbol, "8") || strings.HasPrefix(symbol, "4") {
		return symbol + ".BJ"
	}
	return symbol + ".SZ"
}

func wrapProviderError(err error) error {
	if IsCode(err, CodeProviderError) {
		return err
	}
	return NewError(CodeProviderError, err.Error())
}
