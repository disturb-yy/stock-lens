package market

import "time"

const (
	pctChangeUnit = "percent"
	volumeUnit    = "lot"
	amountUnit    = "thousand_cny"
)

type pageResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type listResponse[T any] struct {
	Items      []T          `json:"items"`
	Pagination pageResponse `json:"pagination"`
}

type stockResponse struct {
	Symbol     string  `json:"symbol"`
	Name       string  `json:"name"`
	Market     string  `json:"market"`
	AssetType  string  `json:"asset_type"`
	Exchange   string  `json:"exchange"`
	Board      string  `json:"board"`
	Area       string  `json:"area"`
	Industry   string  `json:"industry"`
	Status     string  `json:"status"`
	ListDate   *string `json:"list_date"`
	DelistDate *string `json:"delist_date"`
	DataSource string  `json:"data_source"`
	SyncedAt   string  `json:"synced_at"`
}

type stockDetailResponse struct {
	stockResponse
	LatestDailyKLine *dailyKLineResponse `json:"latest_daily_k_line"`
}

type dailyKLineResponse struct {
	Symbol        string `json:"symbol"`
	Market        string `json:"market"`
	AssetType     string `json:"asset_type"`
	TradeDate     string `json:"trade_date"`
	Open          string `json:"open"`
	High          string `json:"high"`
	Low           string `json:"low"`
	Close         string `json:"close"`
	PreClose      string `json:"pre_close"`
	Change        string `json:"change"`
	PctChange     string `json:"pct_change"`
	PctChangeUnit string `json:"pct_change_unit"`
	Volume        string `json:"volume"`
	VolumeUnit    string `json:"volume_unit"`
	Amount        string `json:"amount"`
	AmountUnit    string `json:"amount_unit"`
	DataSource    string `json:"data_source"`
	SyncedAt      string `json:"synced_at"`
}

type tradeCalendarResponse struct {
	Market       string  `json:"market"`
	Exchange     string  `json:"exchange"`
	CalDate      string  `json:"cal_date"`
	IsOpen       bool    `json:"is_open"`
	PretradeDate *string `json:"pretrade_date"`
	DataSource   string  `json:"data_source"`
	SyncedAt     string  `json:"synced_at"`
}

type latestOpenDayResponse struct {
	Market   string `json:"market"`
	Exchange string `json:"exchange"`
	Date     string `json:"date"`
}

type isOpenDayResponse struct {
	Market       string  `json:"market"`
	Exchange     string  `json:"exchange"`
	Date         string  `json:"date"`
	IsOpen       bool    `json:"is_open"`
	PretradeDate *string `json:"pretrade_date"`
}

type enumMetadataResponse struct {
	Markets          []string `json:"markets"`
	AssetTypes       []string `json:"asset_types"`
	Exchanges        []string `json:"exchanges"`
	Boards           []string `json:"boards"`
	StockStatuses    []string `json:"stock_statuses"`
	DataSources      []string `json:"data_sources"`
	SyncTaskTypes    []string `json:"sync_task_types"`
	SyncTaskStatuses []string `json:"sync_task_statuses"`
}

type syncTaskCreationResponse struct {
	TaskUID string `json:"task_uid"`
	Status  string `json:"status"`
}

type syncTaskResponse struct {
	TaskUID        string  `json:"task_uid"`
	TaskType       string  `json:"task_type"`
	Status         string  `json:"status"`
	Market         string  `json:"market"`
	AssetType      string  `json:"asset_type"`
	DataSource     string  `json:"data_source"`
	TotalItems     int64   `json:"total_items"`
	ProcessedItems int64   `json:"processed_items"`
	SuccessItems   int64   `json:"success_items"`
	FailedItems    int64   `json:"failed_items"`
	RequestID      string  `json:"request_id"`
	StartedAt      *string `json:"started_at"`
	FinishedAt     *string `json:"finished_at"`
	ErrorMsg       string  `json:"error_msg"`
	CreatedAt      string  `json:"created_at"`
}

type syncLogResponse struct {
	TaskUID      string `json:"task_uid"`
	Step         string `json:"step"`
	Status       string `json:"status"`
	Market       string `json:"market"`
	AssetType    string `json:"asset_type"`
	Symbol       string `json:"symbol"`
	DataSource   string `json:"data_source"`
	Message      string `json:"message"`
	ErrorDetail  string `json:"error_detail"`
	AffectedRows int64  `json:"affected_rows"`
	CreatedAt    string `json:"created_at"`
}

func toStockResponse(stock Stock) stockResponse {
	return stockResponse{
		Symbol:     stock.Symbol,
		Name:       stock.Name,
		Market:     string(stock.Market),
		AssetType:  string(stock.AssetType),
		Exchange:   string(stock.Exchange),
		Board:      string(stock.Board),
		Area:       stock.Area,
		Industry:   stock.Industry,
		Status:     string(stock.Status),
		ListDate:   formatOptionalDate(stock.ListDate),
		DelistDate: formatOptionalDate(stock.DelistDate),
		DataSource: string(stock.DataSource),
		SyncedAt:   formatDateTime(stock.SyncedAt),
	}
}

func toStockDetailResponse(detail StockDetailResult) stockDetailResponse {
	resp := stockDetailResponse{stockResponse: toStockResponse(detail.Stock)}
	if detail.LatestDailyKLine != nil {
		line := toDailyKLineResponse(*detail.LatestDailyKLine)
		resp.LatestDailyKLine = &line
	}
	return resp
}

func toDailyKLineResponse(line DailyKLine) dailyKLineResponse {
	return dailyKLineResponse{
		Symbol:        line.Symbol,
		Market:        string(line.Market),
		AssetType:     string(line.AssetType),
		TradeDate:     formatDate(line.TradeDate),
		Open:          line.OpenPrice.StringFixed(4),
		High:          line.HighPrice.StringFixed(4),
		Low:           line.LowPrice.StringFixed(4),
		Close:         line.ClosePrice.StringFixed(4),
		PreClose:      line.PreClose.StringFixed(4),
		Change:        line.ChangeAmt.StringFixed(4),
		PctChange:     line.PctChange.StringFixed(4),
		PctChangeUnit: pctChangeUnit,
		Volume:        line.Volume.StringFixed(4),
		VolumeUnit:    volumeUnit,
		Amount:        line.Amount.StringFixed(4),
		AmountUnit:    amountUnit,
		DataSource:    string(line.DataSource),
		SyncedAt:      formatDateTime(line.SyncedAt),
	}
}

func toTradeCalendarResponse(calendar TradeCalendar) tradeCalendarResponse {
	return tradeCalendarResponse{
		Market:       string(calendar.Market),
		Exchange:     string(calendar.Exchange),
		CalDate:      formatDate(calendar.CalDate),
		IsOpen:       calendar.IsOpen,
		PretradeDate: formatOptionalDate(calendar.PretradeDate),
		DataSource:   string(calendar.DataSource),
		SyncedAt:     formatDateTime(calendar.SyncedAt),
	}
}

func toSyncTaskResponse(task SyncTask) syncTaskResponse {
	return syncTaskResponse{
		TaskUID:        task.UID,
		TaskType:       string(task.TaskType),
		Status:         string(task.Status),
		Market:         string(task.Market),
		AssetType:      string(task.AssetType),
		DataSource:     string(task.DataSource),
		TotalItems:     task.TotalItems,
		ProcessedItems: task.ProcessedItems,
		SuccessItems:   task.SuccessItems,
		FailedItems:    task.FailedItems,
		RequestID:      task.RequestID,
		StartedAt:      formatOptionalDateTime(task.StartedAt),
		FinishedAt:     formatOptionalDateTime(task.FinishedAt),
		ErrorMsg:       task.ErrorMsg,
		CreatedAt:      formatDateTime(task.CreatedAt),
	}
}

func toSyncLogResponse(log SyncLog) syncLogResponse {
	return syncLogResponse{
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
		CreatedAt:    formatDateTime(log.CreatedAt),
	}
}

func toPageResponse(page PageResult) pageResponse {
	return pageResponse{
		Page:       page.Page,
		PageSize:   page.PageSize,
		Total:      page.Total,
		TotalPages: page.TotalPages,
	}
}

func toEnumMetadataResponse(metadata EnumMetadataResult) enumMetadataResponse {
	return enumMetadataResponse{
		Markets:          enumStrings(metadata.Markets),
		AssetTypes:       enumStrings(metadata.AssetTypes),
		Exchanges:        enumStrings(metadata.Exchanges),
		Boards:           enumStrings(metadata.Boards),
		StockStatuses:    enumStrings(metadata.StockStatuses),
		DataSources:      enumStrings(metadata.DataSources),
		SyncTaskTypes:    enumStrings(metadata.SyncTaskTypes),
		SyncTaskStatuses: enumStrings(metadata.SyncTaskStatuses),
	}
}

func enumStrings(values []EnumValue) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, value.Value)
	}
	return items
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(DateLayout)
}

func formatOptionalDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatDate(*value)
	return &formatted
}

func formatDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339Nano)
}

func formatOptionalDateTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatDateTime(*value)
	return &formatted
}
