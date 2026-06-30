package market

import "errors"

type ErrorCode string

const (
	CodeInvalidSymbol               ErrorCode = "MARKET_INVALID_SYMBOL"
	CodeInvalidMarket               ErrorCode = "MARKET_INVALID_MARKET"
	CodeInvalidAssetType            ErrorCode = "MARKET_INVALID_ASSET_TYPE"
	CodeInvalidExchange             ErrorCode = "MARKET_INVALID_EXCHANGE"
	CodeInvalidStatus               ErrorCode = "MARKET_INVALID_STATUS"
	CodeInvalidPageSize             ErrorCode = "MARKET_INVALID_PAGE_SIZE"
	CodeInvalidDateRange            ErrorCode = "MARKET_INVALID_DATE_RANGE"
	CodeDateRangeTooLarge           ErrorCode = "MARKET_DATE_RANGE_TOO_LARGE"
	CodeInvalidTaskStatus           ErrorCode = "MARKET_INVALID_TASK_STATUS"
	CodeStockNotFound               ErrorCode = "MARKET_STOCK_NOT_FOUND"
	CodeSyncTaskNotFound            ErrorCode = "MARKET_SYNC_TASK_NOT_FOUND"
	CodeTradeCalendarNotFound       ErrorCode = "MARKET_TRADE_CALENDAR_NOT_FOUND"
	CodeTradeCalendarNotInitialized ErrorCode = "MARKET_TRADE_CALENDAR_NOT_INITIALIZED"
	CodeStocksNotInitialized        ErrorCode = "MARKET_STOCKS_NOT_INITIALIZED"
	CodeSyncTaskConflict            ErrorCode = "MARKET_SYNC_TASK_CONFLICT"
	CodeProviderError               ErrorCode = "MARKET_PROVIDER_ERROR"
	CodeStoreError                  ErrorCode = "MARKET_STORE_ERROR"
	CodeSyncTaskFailed              ErrorCode = "MARKET_SYNC_TASK_FAILED"
)

type Error struct {
	Code    ErrorCode
	Message string
}

func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func (e *Error) Error() string {
	return e.Message
}

func IsCode(err error, code ErrorCode) bool {
	var marketErr *Error
	if !errors.As(err, &marketErr) {
		return false
	}
	return marketErr.Code == code
}
