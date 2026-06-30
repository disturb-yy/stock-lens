package market

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"stock-lens/internal/server"
)

type Handler struct {
	query *QueryService
	sync  *SyncService
}

func NewHandler(query *QueryService, sync *SyncService) *Handler {
	return &Handler{
		query: query,
		sync:  sync,
	}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup, adminToken string) {
	group.GET("/stocks", h.listStocks)
	group.GET("/stocks/:symbol", h.getStockDetail)
	group.GET("/stocks/:symbol/daily-k-lines", h.listDailyKLines)
	group.GET("/stocks/:symbol/latest-daily-k-line", h.latestDailyKLine)
	group.GET("/trade-calendars", h.listTradeCalendars)
	group.GET("/trade-calendars/latest-open-day", h.latestOpenDay)
	group.GET("/trade-calendars/is-open", h.isOpenDay)
	group.GET("/meta/enums", h.enumMetadata)
	group.GET("/sync/tasks", h.listSyncTasks)
	group.GET("/sync/tasks/:task_uid", h.getSyncTask)
	group.GET("/sync/tasks/:task_uid/logs", h.listSyncLogs)

	protected := group.Group("/sync", server.AdminTokenMiddleware(adminToken))
	protected.POST("/stocks", h.syncStocks)
	protected.POST("/trade-calendars", h.syncTradeCalendars)
	protected.POST("/daily-k-lines", h.syncDailyKLines)
}

func (h *Handler) listStocks(c *gin.Context) {
	query, ok := parseListStocksQuery(c)
	if !ok {
		return
	}
	result, err := h.query.ListStocks(c.Request.Context(), query)
	if err != nil {
		h.writeError(c, err)
		return
	}

	items := make([]stockResponse, 0, len(result.Items))
	for _, stock := range result.Items {
		items = append(items, toStockResponse(stock))
	}
	h.writeOK(c, listResponse[stockResponse]{Items: items, Pagination: toPageResponse(result.Pagination)})
}

func (h *Handler) getStockDetail(c *gin.Context) {
	detail, err := h.query.GetStockDetail(c.Request.Context(), FindStockQuery{
		Market:    Market(c.Query("market")),
		AssetType: AssetType(c.Query("asset_type")),
		Symbol:    c.Param("symbol"),
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	h.writeOK(c, toStockDetailResponse(detail))
}

func (h *Handler) listDailyKLines(c *gin.Context) {
	query, ok := parseListDailyKLinesQuery(c)
	if !ok {
		return
	}
	items, err := h.query.ListDailyKLines(c.Request.Context(), query)
	if err != nil {
		h.writeError(c, err)
		return
	}

	resp := make([]dailyKLineResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toDailyKLineResponse(item))
	}
	h.writeOK(c, resp)
}

func (h *Handler) latestDailyKLine(c *gin.Context) {
	line, err := h.query.LatestDailyKLine(c.Request.Context(), FindStockQuery{
		Market:    Market(c.Query("market")),
		AssetType: AssetType(c.Query("asset_type")),
		Symbol:    c.Param("symbol"),
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	if line == nil {
		h.writeOK(c, nil)
		return
	}
	h.writeOK(c, toDailyKLineResponse(*line))
}

func (h *Handler) listTradeCalendars(c *gin.Context) {
	query, ok := parseListTradeCalendarsQuery(c)
	if !ok {
		return
	}
	items, err := h.query.ListTradeCalendars(c.Request.Context(), query)
	if err != nil {
		h.writeError(c, err)
		return
	}

	resp := make([]tradeCalendarResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toTradeCalendarResponse(item))
	}
	h.writeOK(c, resp)
}

func (h *Handler) latestOpenDay(c *gin.Context) {
	market := Market(c.Query("market"))
	exchange := Exchange(c.Query("exchange"))
	day, err := h.query.LatestOpenDay(c.Request.Context(), market, exchange)
	if err != nil {
		h.writeError(c, err)
		return
	}
	h.writeOK(c, latestOpenDayResponse{
		Market:   string(defaultMarket(market)),
		Exchange: string(defaultExchange(exchange)),
		Date:     formatDate(day),
	})
}

func (h *Handler) isOpenDay(c *gin.Context) {
	query, ok := parseIsOpenDayQuery(c)
	if !ok {
		return
	}
	result, err := h.query.IsOpenDay(c.Request.Context(), query)
	if err != nil {
		h.writeError(c, err)
		return
	}
	h.writeOK(c, isOpenDayResponse{
		Market:       string(result.Market),
		Exchange:     string(result.Exchange),
		Date:         formatDate(result.Date),
		IsOpen:       result.IsOpen,
		PretradeDate: formatOptionalDate(result.PretradeDate),
	})
}

func (h *Handler) enumMetadata(c *gin.Context) {
	h.writeOK(c, toEnumMetadataResponse(h.query.EnumMetadata()))
}

func (h *Handler) listSyncTasks(c *gin.Context) {
	query, ok := parseListSyncTasksQuery(c)
	if !ok {
		return
	}
	result, err := h.query.ListSyncTasks(c.Request.Context(), query)
	if err != nil {
		h.writeError(c, err)
		return
	}

	items := make([]syncTaskResponse, 0, len(result.Items))
	for _, task := range result.Items {
		items = append(items, toSyncTaskResponse(task))
	}
	h.writeOK(c, listResponse[syncTaskResponse]{Items: items, Pagination: toPageResponse(result.Pagination)})
}

func (h *Handler) getSyncTask(c *gin.Context) {
	task, err := h.query.GetSyncTask(c.Request.Context(), c.Param("task_uid"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	h.writeOK(c, toSyncTaskResponse(task))
}

func (h *Handler) listSyncLogs(c *gin.Context) {
	query, ok := parseListSyncLogsQuery(c)
	if !ok {
		return
	}
	result, err := h.query.ListSyncLogs(c.Request.Context(), query)
	if err != nil {
		h.writeError(c, err)
		return
	}

	items := make([]syncLogResponse, 0, len(result.Items))
	for _, log := range result.Items {
		items = append(items, toSyncLogResponse(log))
	}
	h.writeOK(c, listResponse[syncLogResponse]{Items: items, Pagination: toPageResponse(result.Pagination)})
}

func (h *Handler) syncStocks(c *gin.Context) {
	var req syncStocksRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.sync.SyncStocks(c.Request.Context(), SyncStocksInput{
		Market:    Market(req.Market),
		AssetType: AssetType(req.AssetType),
		RequestID: server.RequestID(c),
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	h.writeOK(c, toSyncTaskCreationResponse(result))
}

func (h *Handler) syncTradeCalendars(c *gin.Context) {
	var req syncTradeCalendarsRequest
	if !bindJSON(c, &req) {
		return
	}
	input, ok := req.toInput(c, server.RequestID(c))
	if !ok {
		return
	}
	result, err := h.sync.SyncTradeCalendars(c.Request.Context(), input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	h.writeOK(c, toSyncTaskCreationResponse(result))
}

func (h *Handler) syncDailyKLines(c *gin.Context) {
	var req syncDailyKLinesRequest
	if !bindJSON(c, &req) {
		return
	}
	input, ok := req.toInput(c, server.RequestID(c))
	if !ok {
		return
	}
	result, err := h.sync.SyncDailyKLines(c.Request.Context(), input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	h.writeOK(c, toSyncTaskCreationResponse(result))
}

func (h *Handler) writeOK(c *gin.Context, data any) {
	server.WriteSuccess(c, http.StatusOK, server.RequestID(c), data)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	var marketErr *Error
	if errors.As(err, &marketErr) {
		server.WriteErrorCode(c, string(marketErr.Code), server.RequestID(c), nil)
		return
	}
	server.WriteErrorCode(c, server.CodeInternalError, server.RequestID(c), nil)
}

type syncStocksRequest struct {
	Market    string `json:"market"`
	AssetType string `json:"asset_type"`
}

type syncTradeCalendarsRequest struct {
	Market    string `json:"market"`
	Exchange  string `json:"exchange"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type syncDailyKLinesRequest struct {
	Market    string `json:"market"`
	AssetType string `json:"asset_type"`
	Symbol    string `json:"symbol"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (r syncTradeCalendarsRequest) toInput(c *gin.Context, requestID string) (SyncTradeCalendarsInput, bool) {
	start, ok := parseRequiredDate(c, r.StartDate)
	if !ok {
		return SyncTradeCalendarsInput{}, false
	}
	end, ok := parseRequiredDate(c, r.EndDate)
	if !ok {
		return SyncTradeCalendarsInput{}, false
	}
	return SyncTradeCalendarsInput{
		Market:    Market(r.Market),
		Exchange:  Exchange(r.Exchange),
		StartDate: start,
		EndDate:   end,
		RequestID: requestID,
	}, true
}

func (r syncDailyKLinesRequest) toInput(c *gin.Context, requestID string) (SyncDailyKLinesInput, bool) {
	start, ok := parseOptionalDate(c, r.StartDate)
	if !ok {
		return SyncDailyKLinesInput{}, false
	}
	end, ok := parseOptionalDate(c, r.EndDate)
	if !ok {
		return SyncDailyKLinesInput{}, false
	}
	return SyncDailyKLinesInput{
		Market:    Market(r.Market),
		AssetType: AssetType(r.AssetType),
		Symbol:    r.Symbol,
		StartDate: start,
		EndDate:   end,
		RequestID: requestID,
	}, true
}

func parseListStocksQuery(c *gin.Context) (ListStocksQuery, bool) {
	page, pageSize, ok := parsePage(c)
	if !ok {
		return ListStocksQuery{}, false
	}
	return ListStocksQuery{
		PageQuery: PageQuery{Page: page, PageSize: pageSize},
		Market:    Market(c.Query("market")),
		AssetType: AssetType(c.Query("asset_type")),
		Keyword:   c.Query("keyword"),
		Exchange:  Exchange(c.Query("exchange")),
		Status:    StockStatus(c.Query("status")),
	}, true
}

func parseListDailyKLinesQuery(c *gin.Context) (ListDailyKLinesQuery, bool) {
	start, ok := parseOptionalDate(c, c.Query("start_date"))
	if !ok {
		return ListDailyKLinesQuery{}, false
	}
	end, ok := parseOptionalDate(c, c.Query("end_date"))
	if !ok {
		return ListDailyKLinesQuery{}, false
	}
	return ListDailyKLinesQuery{
		Market:    Market(c.Query("market")),
		AssetType: AssetType(c.Query("asset_type")),
		Symbol:    c.Param("symbol"),
		StartDate: start,
		EndDate:   end,
	}, true
}

func parseListTradeCalendarsQuery(c *gin.Context) (ListTradeCalendarsQuery, bool) {
	start, ok := parseOptionalDate(c, c.Query("start_date"))
	if !ok {
		return ListTradeCalendarsQuery{}, false
	}
	end, ok := parseOptionalDate(c, c.Query("end_date"))
	if !ok {
		return ListTradeCalendarsQuery{}, false
	}
	return ListTradeCalendarsQuery{
		Market:    Market(c.Query("market")),
		Exchange:  Exchange(c.Query("exchange")),
		StartDate: start,
		EndDate:   end,
	}, true
}

func parseIsOpenDayQuery(c *gin.Context) (IsOpenDayQuery, bool) {
	date, ok := parseOptionalDate(c, c.Query("date"))
	if !ok {
		return IsOpenDayQuery{}, false
	}
	return IsOpenDayQuery{
		Market:   Market(c.Query("market")),
		Exchange: Exchange(c.Query("exchange")),
		Date:     date,
	}, true
}

func parseListSyncTasksQuery(c *gin.Context) (ListSyncTasksQuery, bool) {
	page, pageSize, ok := parsePage(c)
	if !ok {
		return ListSyncTasksQuery{}, false
	}
	return ListSyncTasksQuery{
		PageQuery: PageQuery{Page: page, PageSize: pageSize},
		Status:    SyncTaskStatus(c.Query("status")),
	}, true
}

func parseListSyncLogsQuery(c *gin.Context) (ListSyncLogsQuery, bool) {
	page, pageSize, ok := parsePage(c)
	if !ok {
		return ListSyncLogsQuery{}, false
	}
	return ListSyncLogsQuery{
		PageQuery: PageQuery{Page: page, PageSize: pageSize},
		TaskUID:   c.Param("task_uid"),
		Status:    SyncLogStatus(c.Query("status")),
		Symbol:    c.Query("symbol"),
	}, true
}

func parsePage(c *gin.Context) (int, int, bool) {
	page, ok := parseOptionalInt(c, "page")
	if !ok {
		return 0, 0, false
	}
	pageSize, ok := parseOptionalInt(c, "page_size")
	if !ok {
		return 0, 0, false
	}
	return page, pageSize, true
}

func parseOptionalInt(c *gin.Context, key string) (int, bool) {
	value := c.Query(key)
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		writeInvalidArgument(c)
		return 0, false
	}
	return parsed, true
}

func parseOptionalDate(c *gin.Context, value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, true
	}
	date, err := ParseDate(value)
	if err != nil {
		writeInvalidArgument(c)
		return time.Time{}, false
	}
	return date, true
}

func parseRequiredDate(c *gin.Context, value string) (time.Time, bool) {
	if value == "" {
		writeInvalidArgument(c)
		return time.Time{}, false
	}
	return parseOptionalDate(c, value)
}

func bindJSON(c *gin.Context, out any) bool {
	if err := c.ShouldBindJSON(out); err != nil {
		writeInvalidArgument(c)
		return false
	}
	return true
}

func writeInvalidArgument(c *gin.Context) {
	server.WriteErrorCode(c, server.CodeInvalidArgument, server.RequestID(c), nil)
}

func toSyncTaskCreationResponse(result SyncTaskCreationResult) syncTaskCreationResponse {
	return syncTaskCreationResponse{
		TaskUID: result.TaskUID,
		Status:  string(result.Status),
	}
}
