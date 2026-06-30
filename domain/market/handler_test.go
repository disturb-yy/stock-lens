package market

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"stock-lens/internal/server"
)

func TestHandlerListStocks(t *testing.T) {
	router, repos := newHandlerTestRouter(t)
	ctx := context.Background()

	err := repos.stocks.UpsertStocks(ctx, []Stock{
		{
			Market:     MarketCN,
			AssetType:  AssetTypeStock,
			Symbol:     "600519",
			Name:       "Kweichow Moutai",
			Exchange:   ExchangeSSE,
			Board:      BoardMain,
			Status:     StockStatusListed,
			DataSource: DataSourceMock,
		},
	})
	if err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}

	rec := performRequest(router, http.MethodGet, "/api/v1/market/stocks?keyword=Moutai&page=1&page_size=20", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body responseBody
	decodeJSON(t, rec, &body)
	if body.Code != server.CodeOK {
		t.Fatalf("code = %s, want %s", body.Code, server.CodeOK)
	}

	data := body.Data.(map[string]any)
	items := data["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	item := items[0].(map[string]any)
	if item["symbol"] != "600519" {
		t.Fatalf("symbol = %v, want 600519", item["symbol"])
	}
}

func TestHandlerGetStockDetailNotFound(t *testing.T) {
	router, _ := newHandlerTestRouter(t)

	rec := performRequest(router, http.MethodGet, "/api/v1/market/stocks/600519", nil, "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}

	var body responseBody
	decodeJSON(t, rec, &body)
	if body.Code != string(CodeStockNotFound) {
		t.Fatalf("code = %s, want %s", body.Code, CodeStockNotFound)
	}
}

func TestHandlerListDailyKLinesFormatsNumericStrings(t *testing.T) {
	router, repos := newHandlerTestRouter(t)
	ctx := context.Background()
	tradeDate := serviceTestDate(t, "2026-06-30")

	if err := repos.stocks.UpsertStocks(ctx, []Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519"}}); err != nil {
		t.Fatalf("UpsertStocks() error = %v", err)
	}
	if err := repos.kLines.UpsertDailyKLines(ctx, []DailyKLine{
		{
			Market:     MarketCN,
			AssetType:  AssetTypeStock,
			Symbol:     "600519",
			TradeDate:  tradeDate,
			OpenPrice:  decimal.NewFromInt(10),
			HighPrice:  decimal.NewFromInt(12),
			LowPrice:   decimal.NewFromInt(9),
			ClosePrice: decimal.NewFromInt(11),
			DataSource: DataSourceMock,
		},
	}); err != nil {
		t.Fatalf("UpsertDailyKLines() error = %v", err)
	}

	rec := performRequest(router, http.MethodGet, "/api/v1/market/stocks/600519/daily-k-lines?start_date=2026-06-30&end_date=2026-06-30", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body responseBody
	decodeJSON(t, rec, &body)
	items := body.Data.([]any)
	if len(items) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(items))
	}
	item := items[0].(map[string]any)
	if item["open"] != "10.0000" {
		t.Fatalf("open = %v, want 10.0000", item["open"])
	}
}

func TestHandlerSyncStocksRequiresAdminToken(t *testing.T) {
	router, _ := newHandlerTestRouter(t)

	rec := performRequest(router, http.MethodPost, "/api/v1/market/sync/stocks", []byte(`{}`), "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestHandlerSyncStocksCreatesTask(t *testing.T) {
	router, _ := newHandlerTestRouter(t)

	rec := performRequest(router, http.MethodPost, "/api/v1/market/sync/stocks", []byte(`{"market":"CN","asset_type":"STOCK"}`), "Bearer admin-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body responseBody
	decodeJSON(t, rec, &body)
	data := body.Data.(map[string]any)
	if data["task_uid"] == "" {
		t.Fatalf("task_uid is empty")
	}
	if data["status"] != string(SyncTaskStatusPending) {
		t.Fatalf("status = %v, want %s", data["status"], SyncTaskStatusPending)
	}
}

func TestHandlerRegisterRoutesThroughServerRouter(t *testing.T) {
	handler, _ := newHandlerTestComponents(t)
	router := server.NewRouter(nil, server.WithRoutes(func(engine *gin.Engine) {
		handler.RegisterRoutes(engine.Group("/api/v1/market"), "admin-token")
	}))

	rec := performRequest(router, http.MethodGet, "/api/v1/market/meta/enums", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

type handlerTestRepos struct {
	stocks         StockRepository
	kLines         KLineRepository
	tradeCalendars TradeCalendarRepository
	syncTasks      SyncTaskRepository
}

type responseBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
}

func newHandlerTestRouter(t *testing.T) (*gin.Engine, handlerTestRepos) {
	t.Helper()

	handler, repos := newHandlerTestComponents(t)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(server.RequestIDMiddleware())
	handler.RegisterRoutes(router.Group("/api/v1/market"), "admin-token")
	return router, repos
}

func newHandlerTestComponents(t *testing.T) (*Handler, handlerTestRepos) {
	t.Helper()

	repos := handlerTestRepos{
		stocks:         NewMockStockRepository(),
		kLines:         NewMockKLineRepository(),
		tradeCalendars: NewMockTradeCalendarRepository(),
		syncTasks:      NewMockSyncTaskRepository(),
	}
	query := NewQueryService(repos.stocks, repos.kLines, repos.tradeCalendars, repos.syncTasks)
	syncSvc := NewSyncServiceWithRunner(
		repos.stocks,
		repos.kLines,
		repos.tradeCalendars,
		repos.syncTasks,
		NewMockInstrumentProvider([]Stock{{Market: MarketCN, AssetType: AssetTypeStock, Symbol: "600519", DataSource: DataSourceMock}}),
		NewMockCalendarProvider(nil),
		NewMockMarketDataProvider(nil),
		runSyncForTest,
	)
	return NewHandler(query, syncSvc), repos
}

func performRequest(router http.Handler, method string, path string, body []byte, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, out any) {
	t.Helper()

	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
}
