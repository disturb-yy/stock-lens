package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stock-lens/internal/config"
)

func TestConfigPath(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "default path",
			args: []string{"server"},
			want: "configs/config.yaml",
		},
		{
			name: "explicit path",
			args: []string{"server", "-config", "configs/config.test.yaml"},
			want: "configs/config.test.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConfigPath(tt.args)
			if err != nil {
				t.Fatalf("ConfigPath() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ConfigPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewHTTPServer(t *testing.T) {
	cfg := config.Config{
		Server: config.ServerConfig{
			Addr: "127.0.0.1",
			Port: 30078,
		},
		Auth: config.AuthConfig{
			AdminToken: "test-admin-token",
		},
	}

	srv := NewHTTPServer(cfg, nil)
	if srv.Addr != "127.0.0.1:30078" {
		t.Fatalf("server addr = %q, want %q", srv.Addr, "127.0.0.1:30078")
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestNewHTTPServerRegistersMarketRoutes(t *testing.T) {
	cfg := config.Config{
		Server: config.ServerConfig{
			Addr: "127.0.0.1",
			Port: 30078,
		},
		Auth: config.AuthConfig{
			AdminToken: "test-admin-token",
		},
	}

	srv := NewHTTPServer(cfg, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/market/meta/enums", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestNewMarketProvidersUsesConfiguredProvider(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.Config
		wantPrefix string
	}{
		{
			name: "mock",
			cfg: config.Config{
				Market: config.MarketConfig{Provider: "mock"},
			},
			wantPrefix: "*market.mock",
		},
		{
			name: "tushare",
			cfg: config.Config{
				Market:  config.MarketConfig{Provider: "tushare"},
				Tushare: config.TushareConfig{BaseURL: "https://api.tushare.pro", Token: "secret"},
			},
			wantPrefix: "*market.tushare",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instrument, calendar, data := newMarketProviders(tt.cfg)
			got := fmt.Sprintf("%T|%T|%T", instrument, calendar, data)
			if !strings.Contains(got, tt.wantPrefix) {
				t.Fatalf("providers = %s, want containing %s", got, tt.wantPrefix)
			}
		})
	}
}

func TestRunReturnsConfigErrorBeforeStartingServer(t *testing.T) {
	err := Run(context.Background(), []string{"server", "-config", "missing.yaml"})
	if err == nil {
		t.Fatalf("Run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "load config") {
		t.Fatalf("Run() error = %q, want load config error", err.Error())
	}
}

func TestInitializeRuntimeConfiguresDefaultLogger(t *testing.T) {
	original := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(original)
	})

	initializeRuntime(config.Config{
		Log: config.LogConfig{Level: "debug"},
	})

	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		t.Fatalf("default logger does not enable debug records")
	}
}

func TestRunReturnsDatabasePingErrorBeforeStartingServer(t *testing.T) {
	pingErr := errors.New("ping failed")
	served := false

	err := runWithHooks(context.Background(), []string{"server", "-config", "../../configs/config.test.yaml"}, runHooks{
		openDatabase: func(context.Context, config.DatabaseConfig) (databaseHandle, error) {
			return &fakeDatabase{pingErr: pingErr}, nil
		},
		serve: func(context.Context, *http.Server) error {
			served = true
			return nil
		},
	})

	if err == nil {
		t.Fatalf("runWithHooks() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "ping database") {
		t.Fatalf("runWithHooks() error = %q, want ping database error", err.Error())
	}
	if served {
		t.Fatalf("HTTP server was started after database ping failed")
	}
}

func TestRunWiresDatabaseReadinessCheck(t *testing.T) {
	db := &fakeDatabase{}

	err := runWithHooks(context.Background(), []string{"server", "-config", "../../configs/config.test.yaml"}, runHooks{
		openDatabase: func(context.Context, config.DatabaseConfig) (databaseHandle, error) {
			return db, nil
		},
		serve: func(_ context.Context, srv *http.Server) error {
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			return nil
		},
	})

	if err != nil {
		t.Fatalf("runWithHooks() error = %v", err)
	}
	if db.pingCount != 2 {
		t.Fatalf("database ping count = %d, want %d", db.pingCount, 2)
	}
}

func TestRunReturnsReadinessFailureFromDatabasePing(t *testing.T) {
	db := &fakeDatabase{
		readyErr: errors.New("database unavailable"),
	}

	err := runWithHooks(context.Background(), []string{"server", "-config", "../../configs/config.test.yaml"}, runHooks{
		openDatabase: func(context.Context, config.DatabaseConfig) (databaseHandle, error) {
			return db, nil
		},
		serve: func(_ context.Context, srv *http.Server) error {
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
			}
			return nil
		},
	})

	if err != nil {
		t.Fatalf("runWithHooks() error = %v", err)
	}
}

type fakeDatabase struct {
	pingErr   error
	readyErr  error
	pingCount int
}

func (f *fakeDatabase) PingContext(context.Context) error {
	f.pingCount++
	if f.pingCount > 1 && f.readyErr != nil {
		return f.readyErr
	}
	return f.pingErr
}

func (f *fakeDatabase) Close() error {
	return nil
}
