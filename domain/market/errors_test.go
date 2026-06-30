package market

import (
	"errors"
	"testing"
)

func TestMarketErrorClassification(t *testing.T) {
	err := NewError(CodeInvalidSymbol, "invalid stock symbol")

	var marketErr *Error
	if !errors.As(err, &marketErr) {
		t.Fatalf("errors.As() = false, want true")
	}
	if marketErr.Code != CodeInvalidSymbol {
		t.Fatalf("Code = %s, want %s", marketErr.Code, CodeInvalidSymbol)
	}
	if marketErr.Error() != "invalid stock symbol" {
		t.Fatalf("Error() = %q", marketErr.Error())
	}
	if !IsCode(err, CodeInvalidSymbol) {
		t.Fatalf("IsCode() = false, want true")
	}
	if IsCode(err, CodeStockNotFound) {
		t.Fatalf("IsCode(stock not found) = true, want false")
	}
}

func TestIsCodeRejectsNonMarketError(t *testing.T) {
	if IsCode(errors.New("plain error"), CodeInvalidSymbol) {
		t.Fatalf("IsCode() = true, want false")
	}
}
