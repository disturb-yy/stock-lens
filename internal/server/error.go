package server

import "net/http"

const (
	CodeInvalidArgument = "INVALID_ARGUMENT"
	CodeNotFound        = "NOT_FOUND"
	CodeInternalError   = "INTERNAL_ERROR"

	CodeMarketInvalidSymbol               = "MARKET_INVALID_SYMBOL"
	CodeMarketInvalidMarket               = "MARKET_INVALID_MARKET"
	CodeMarketInvalidAssetType            = "MARKET_INVALID_ASSET_TYPE"
	CodeMarketInvalidExchange             = "MARKET_INVALID_EXCHANGE"
	CodeMarketInvalidStatus               = "MARKET_INVALID_STATUS"
	CodeMarketInvalidPageSize             = "MARKET_INVALID_PAGE_SIZE"
	CodeMarketInvalidDateRange            = "MARKET_INVALID_DATE_RANGE"
	CodeMarketDateRangeTooLarge           = "MARKET_DATE_RANGE_TOO_LARGE"
	CodeMarketInvalidTaskStatus           = "MARKET_INVALID_TASK_STATUS"
	CodeMarketStockNotFound               = "MARKET_STOCK_NOT_FOUND"
	CodeMarketSyncTaskNotFound            = "MARKET_SYNC_TASK_NOT_FOUND"
	CodeMarketTradeCalendarNotFound       = "MARKET_TRADE_CALENDAR_NOT_FOUND"
	CodeMarketTradeCalendarNotInitialized = "MARKET_TRADE_CALENDAR_NOT_INITIALIZED"
	CodeMarketStocksNotInitialized        = "MARKET_STOCKS_NOT_INITIALIZED"
	CodeMarketSyncTaskConflict            = "MARKET_SYNC_TASK_CONFLICT"
	CodeMarketProviderError               = "MARKET_PROVIDER_ERROR"
	CodeMarketStoreError                  = "MARKET_STORE_ERROR"
	CodeMarketSyncTaskFailed              = "MARKET_SYNC_TASK_FAILED"
)

type ErrorSpec struct {
	HTTPStatus int
	Message    string
}

func ErrorSpecForCode(code string) ErrorSpec {
	spec, ok := errorSpecs[code]
	if !ok {
		return errorSpecs[CodeInternalError]
	}
	return spec
}

var errorSpecs = map[string]ErrorSpec{
	CodeInvalidArgument: {HTTPStatus: http.StatusBadRequest, Message: "invalid argument"},
	CodeUnauthorized:    {HTTPStatus: http.StatusUnauthorized, Message: MessageUnauthorized},
	CodeNotFound:        {HTTPStatus: http.StatusNotFound, Message: "not found"},
	CodeInternalError:   {HTTPStatus: http.StatusInternalServerError, Message: "internal error"},

	CodeMarketInvalidSymbol:               {HTTPStatus: http.StatusBadRequest, Message: "invalid stock symbol"},
	CodeMarketInvalidMarket:               {HTTPStatus: http.StatusBadRequest, Message: "invalid market"},
	CodeMarketInvalidAssetType:            {HTTPStatus: http.StatusBadRequest, Message: "invalid asset type"},
	CodeMarketInvalidExchange:             {HTTPStatus: http.StatusBadRequest, Message: "invalid exchange"},
	CodeMarketInvalidStatus:               {HTTPStatus: http.StatusBadRequest, Message: "invalid stock status"},
	CodeMarketInvalidPageSize:             {HTTPStatus: http.StatusBadRequest, Message: "invalid page size"},
	CodeMarketInvalidDateRange:            {HTTPStatus: http.StatusBadRequest, Message: "invalid date range"},
	CodeMarketDateRangeTooLarge:           {HTTPStatus: http.StatusBadRequest, Message: "date range too large"},
	CodeMarketInvalidTaskStatus:           {HTTPStatus: http.StatusBadRequest, Message: "invalid task status"},
	CodeMarketStockNotFound:               {HTTPStatus: http.StatusNotFound, Message: "stock not found"},
	CodeMarketSyncTaskNotFound:            {HTTPStatus: http.StatusNotFound, Message: "sync task not found"},
	CodeMarketTradeCalendarNotFound:       {HTTPStatus: http.StatusNotFound, Message: "trade calendar not found"},
	CodeMarketTradeCalendarNotInitialized: {HTTPStatus: http.StatusConflict, Message: "trade calendar not initialized"},
	CodeMarketStocksNotInitialized:        {HTTPStatus: http.StatusConflict, Message: "stocks not initialized"},
	CodeMarketSyncTaskConflict:            {HTTPStatus: http.StatusConflict, Message: "sync task already running"},
	CodeMarketProviderError:               {HTTPStatus: http.StatusBadGateway, Message: "market provider error"},
	CodeMarketStoreError:                  {HTTPStatus: http.StatusInternalServerError, Message: "market store error"},
	CodeMarketSyncTaskFailed:              {HTTPStatus: http.StatusInternalServerError, Message: "sync task failed"},
}
