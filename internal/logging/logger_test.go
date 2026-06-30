package logging

import (
	"context"
	"log/slog"
	"testing"

	"stock-lens/internal/config"
)

func TestNewLoggerAppliesConfiguredLevel(t *testing.T) {
	tests := []struct {
		name         string
		level        string
		enabledLevel slog.Level
		wantEnabled  bool
	}{
		{
			name:         "debug enables debug records",
			level:        "debug",
			enabledLevel: slog.LevelDebug,
			wantEnabled:  true,
		},
		{
			name:         "info disables debug records",
			level:        "info",
			enabledLevel: slog.LevelDebug,
			wantEnabled:  false,
		},
		{
			name:         "warn disables info records",
			level:        "warn",
			enabledLevel: slog.LevelInfo,
			wantEnabled:  false,
		},
		{
			name:         "error enables error records",
			level:        "error",
			enabledLevel: slog.LevelError,
			wantEnabled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(config.LogConfig{Level: tt.level})

			got := logger.Enabled(context.Background(), tt.enabledLevel)
			if got != tt.wantEnabled {
				t.Fatalf("Enabled(%v) = %v, want %v", tt.enabledLevel, got, tt.wantEnabled)
			}
		})
	}
}
