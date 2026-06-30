package market

import (
	"testing"
	"time"
)

func TestEnumValidation(t *testing.T) {
	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "market CN", got: MarketCN.Valid(), want: true},
		{name: "market US", got: Market("US").Valid(), want: false},
		{name: "asset type stock", got: AssetTypeStock.Valid(), want: true},
		{name: "asset type fund", got: AssetType("FUND").Valid(), want: false},
		{name: "exchange SSE", got: ExchangeSSE.Valid(), want: true},
		{name: "exchange HKEX", got: Exchange("HKEX").Valid(), want: false},
		{name: "stock status listed", got: StockStatusListed.Valid(), want: true},
		{name: "stock status active", got: StockStatus("ACTIVE").Valid(), want: false},
		{name: "data source mock", got: DataSourceMock.Valid(), want: true},
		{name: "data source yahoo", got: DataSource("YAHOO").Valid(), want: false},
		{name: "sync task status pending", got: SyncTaskStatusPending.Valid(), want: true},
		{name: "sync task status cancelled", got: SyncTaskStatus("CANCELLED").Valid(), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("Valid() = %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestValidSymbol(t *testing.T) {
	tests := []struct {
		symbol string
		want   bool
	}{
		{symbol: "600519", want: true},
		{symbol: "000001", want: true},
		{symbol: "600519.SH", want: false},
		{symbol: "abc123", want: false},
		{symbol: "12345", want: false},
		{symbol: "1234567", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			got := ValidSymbol(tt.symbol)
			if got != tt.want {
				t.Fatalf("ValidSymbol(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	got, err := ParseDate("2026-06-30")
	if err != nil {
		t.Fatalf("ParseDate() error = %v", err)
	}
	if got.Format(DateLayout) != "2026-06-30" {
		t.Fatalf("date = %s, want 2026-06-30", got.Format(DateLayout))
	}

	if _, err := ParseDate("2026/06/30"); err == nil {
		t.Fatalf("ParseDate() error = nil, want error")
	}
}

func TestValidDateRange(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	if !ValidDateRange(start, end) {
		t.Fatalf("ValidDateRange() = false, want true")
	}
	if ValidDateRange(end, start) {
		t.Fatalf("ValidDateRange() = true, want false")
	}
	if !ValidDateRange(start, start) {
		t.Fatalf("same-day ValidDateRange() = false, want true")
	}
}
