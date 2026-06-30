package market

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"stock-lens/pkg/tushare"
)

func TestTushareInstrumentProviderMapsStocks(t *testing.T) {
	client := newTushareStockBasicTestClient(t, map[string][][]string{
		"L": {{"600519.SH", "600519", "Kweichow Moutai", "Guizhou", "Baijiu", "20010827", "", "L"}},
		"D": {},
		"P": {},
	})

	got, err := NewTushareInstrumentProvider(client).FetchStocks(context.Background(), FetchStocksRequest{Market: MarketCN})
	if err != nil {
		t.Fatalf("FetchStocks() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(stocks) = %d, want 1", len(got))
	}
	stock := got[0]
	if stock.Symbol != "600519" || stock.Exchange != ExchangeSSE || stock.Board != BoardMain || stock.Status != StockStatusListed {
		t.Fatalf("stock = %+v, want mapped stock", stock)
	}
	if stock.DataSource != DataSourceTushare {
		t.Fatalf("DataSource = %s, want %s", stock.DataSource, DataSourceTushare)
	}
	if stock.ListDate == nil || stock.ListDate.Format(DateLayout) != "2001-08-27" {
		t.Fatalf("ListDate = %v, want 2001-08-27", stock.ListDate)
	}
}

func TestTushareProviderInfersBoards(t *testing.T) {
	tests := []struct {
		symbol   string
		exchange Exchange
		want     Board
	}{
		{symbol: "688001", exchange: ExchangeSSE, want: BoardSTAR},
		{symbol: "300001", exchange: ExchangeSZSE, want: BoardGEM},
		{symbol: "830001", exchange: ExchangeBSE, want: BoardBSE},
		{symbol: "600519", exchange: ExchangeSSE, want: BoardMain},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			if got := inferBoard(tt.symbol, tt.exchange); got != tt.want {
				t.Fatalf("inferBoard() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestTushareCalendarProviderMapsTradeCalendars(t *testing.T) {
	client := newTushareTestClient(t, tushare.Response{
		Code: 0,
		Data: tushare.Data{
			Fields: tushare.TradeCalFields(),
			Items:  [][]string{{"SSE", "20260630", "1", "20260629"}},
		},
	})

	got, err := NewTushareCalendarProvider(client).FetchTradeCalendars(context.Background(), FetchTradeCalendarsRequest{
		Market:    MarketCN,
		Exchange:  ExchangeSSE,
		StartDate: serviceTestDate(t, "2026-06-30"),
		EndDate:   serviceTestDate(t, "2026-06-30"),
	})
	if err != nil {
		t.Fatalf("FetchTradeCalendars() error = %v", err)
	}
	if len(got) != 1 || !got[0].IsOpen || got[0].PretradeDate == nil {
		t.Fatalf("calendars = %+v, want mapped open calendar", got)
	}
	if got[0].CalDate.Format(DateLayout) != "2026-06-30" {
		t.Fatalf("CalDate = %s, want 2026-06-30", got[0].CalDate.Format(DateLayout))
	}
}

func TestTushareMarketDataProviderMapsDailyKLines(t *testing.T) {
	client := newTushareTestClient(t, tushare.Response{
		Code: 0,
		Data: tushare.Data{
			Fields: tushare.DailyFields(),
			Items: [][]string{{
				"600519.SH", "20260630", "10.1", "11.2", "9.3", "10.8", "10.0", "0.8", "8.0", "123.4", "567.8",
			}},
		},
	})

	got, err := NewTushareMarketDataProvider(client).FetchDailyKLines(context.Background(), FetchDailyKLinesRequest{
		Market:    MarketCN,
		AssetType: AssetTypeStock,
		Symbol:    "600519",
		StartDate: serviceTestDate(t, "2026-06-30"),
		EndDate:   serviceTestDate(t, "2026-06-30"),
	})
	if err != nil {
		t.Fatalf("FetchDailyKLines() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(got))
	}
	line := got[0]
	if line.Symbol != "600519" || line.TradeDate.Format(DateLayout) != "2026-06-30" {
		t.Fatalf("line = %+v, want mapped symbol/date", line)
	}
	if !line.ClosePrice.Equal(decimalFromString(t, "10.8")) || line.DataSource != DataSourceTushare {
		t.Fatalf("line = %+v, want mapped close/data_source", line)
	}
}

func TestTushareInstrumentProviderSkipsInvalidStockBasicTSCode(t *testing.T) {
	client := newTushareStockBasicTestClient(t, map[string][][]string{
		"L": {
			{"bad", "", "", "", "", "", "", "L"},
			{" 600519.sh ", "600519", "Kweichow Moutai", "Guizhou", "Baijiu", "20010827", "", "L"},
		},
		"D": {},
		"P": {},
	})

	got, err := NewTushareInstrumentProvider(client).FetchStocks(context.Background(), FetchStocksRequest{Market: MarketCN})
	if err != nil {
		t.Fatalf("FetchStocks() error = %v", err)
	}
	if len(got) != 1 || got[0].Symbol != "600519" || got[0].Exchange != ExchangeSSE {
		t.Fatalf("stocks = %+v, want one valid mapped stock", got)
	}
}

func TestTushareInstrumentProviderFailsWhenAllStockBasicTSCodeInvalid(t *testing.T) {
	client := newTushareTestClient(t, tushare.Response{
		Code: 0,
		Data: tushare.Data{
			Fields: tushare.StockBasicFields(),
			Items:  [][]string{{"bad", "", "", "", "", "", "", "L"}},
		},
	})

	_, err := NewTushareInstrumentProvider(client).FetchStocks(context.Background(), FetchStocksRequest{Market: MarketCN})
	if !IsCode(err, CodeProviderError) {
		t.Fatalf("FetchStocks() error = %v, want %s", err, CodeProviderError)
	}
}

func TestTushareClientErrorDoesNotExposeTokenThroughProvider(t *testing.T) {
	client := newTushareTestClient(t, tushare.Response{Code: 1001, Msg: "bad token"})

	_, err := NewTushareMarketDataProvider(client).FetchDailyKLines(context.Background(), FetchDailyKLinesRequest{
		Market:    MarketCN,
		AssetType: AssetTypeStock,
		Symbol:    "600519",
	})
	if err == nil {
		t.Fatalf("FetchDailyKLines() error = nil, want error")
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("provider error exposes token: %v", err)
	}
}

func newTushareTestClient(t *testing.T, response tushare.Response) *tushare.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return tushare.NewClient(server.URL, "secret-token")
}

func newTushareStockBasicTestClient(t *testing.T, itemsByStatus map[string][][]string) *tushare.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req tushare.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		status, _ := req.Params["list_status"].(string)
		if err := json.NewEncoder(w).Encode(tushare.Response{
			Code: 0,
			Data: tushare.Data{
				Fields: tushare.StockBasicFields(),
				Items:  itemsByStatus[status],
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return tushare.NewClient(server.URL, "secret-token")
}

func decimalFromString(t *testing.T, value string) decimal.Decimal {
	t.Helper()

	parsed, err := decimal.NewFromString(value)
	if err != nil {
		t.Fatalf("parse decimal: %v", err)
	}
	return parsed
}
