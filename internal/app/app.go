package app

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	// 注册 database/sql 的 MySQL driver，实际连接仍由 database/sql 管理。
	_ "github.com/go-sql-driver/mysql"

	"stock-lens/domain/market"
	"stock-lens/internal/config"
	"stock-lens/internal/logging"
	"stock-lens/internal/server"
	"stock-lens/pkg/tushare"
)

const defaultConfigPath = "configs/config.yaml"

const shutdownTimeout = 5 * time.Second

type databaseHandle interface {
	PingContext(context.Context) error
	Close() error
}

type gormDatabaseHandle interface {
	databaseHandle
	GormDB() *gorm.DB
}

type appDatabase struct {
	gormDB *gorm.DB
	sqlDB  sqlDatabase
}

type sqlDatabase interface {
	PingContext(context.Context) error
	Close() error
}

type runHooks struct {
	openDatabase func(context.Context, config.DatabaseConfig) (databaseHandle, error)
	serve        func(context.Context, *http.Server) error
}

func ConfigPath(args []string) (string, error) {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	configPath := flags.String("config", defaultConfigPath, "config file path")
	if err := flags.Parse(args[1:]); err != nil {
		return "", fmt.Errorf("parse flags: %w", err)
	}
	return *configPath, nil
}

func NewHTTPServer(cfg config.Config, ready server.ReadinessCheck) *http.Server {
	return &http.Server{
		Addr:    net.JoinHostPort(cfg.Server.Addr, strconv.Itoa(cfg.Server.Port)),
		Handler: server.NewRouter(ready, server.WithRoutes(defaultMarketRouteRegistrar(cfg))),
	}
}

func Run(ctx context.Context, args []string) error {
	return runWithHooks(ctx, args, defaultRunHooks())
}

func runWithHooks(ctx context.Context, args []string, hooks runHooks) error {
	path, err := ConfigPath(args)
	if err != nil {
		return err
	}

	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	initializeRuntime(cfg)

	db, err := hooks.openDatabase(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	marketHandler, err := initializeMarketRuntime(ctx, cfg, db)
	if err != nil {
		return err
	}
	srv := NewHTTPServerWithMarketHandler(cfg, databaseReadiness(db), marketHandler)
	return hooks.serve(ctx, srv)
}

func defaultRunHooks() runHooks {
	return runHooks{
		openDatabase: openDatabase,
		serve:        serve,
	}
}

func openDatabase(_ context.Context, cfg config.DatabaseConfig) (databaseHandle, error) {
	gormDB, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("unwrap database: %w", err)
	}
	return &appDatabase{gormDB: gormDB, sqlDB: sqlDB}, nil
}

func databaseReadiness(db databaseHandle) server.ReadinessCheck {
	return func() error {
		return db.PingContext(context.Background())
	}
}

func initializeRuntime(cfg config.Config) {
	// 全局默认 logger 只在配置校验通过后替换，避免启动失败路径产生半初始化状态。
	slog.SetDefault(logging.NewLogger(cfg.Log))
}

func (db *appDatabase) PingContext(ctx context.Context) error {
	return db.sqlDB.PingContext(ctx)
}

func (db *appDatabase) Close() error {
	return db.sqlDB.Close()
}

func (db *appDatabase) GormDB() *gorm.DB {
	return db.gormDB
}

func NewHTTPServerWithMarketHandler(cfg config.Config, ready server.ReadinessCheck, handler *market.Handler) *http.Server {
	return &http.Server{
		Addr:    net.JoinHostPort(cfg.Server.Addr, strconv.Itoa(cfg.Server.Port)),
		Handler: server.NewRouter(ready, server.WithRoutes(marketRouteRegistrarForHandler(cfg, handler))),
	}
}

func defaultMarketRouteRegistrar(cfg config.Config) server.RouteRegistrar {
	return marketRouteRegistrarForHandler(cfg, newInMemoryMarketHandler(cfg))
}

func marketRouteRegistrarForHandler(cfg config.Config, handler *market.Handler) server.RouteRegistrar {
	return func(engine *gin.Engine) {
		handler.RegisterRoutes(engine.Group("/api/v1/market"), cfg.Auth.AdminToken)
	}
}

func initializeMarketRuntime(ctx context.Context, cfg config.Config, db databaseHandle) (*market.Handler, error) {
	gormHandle, ok := db.(gormDatabaseHandle)
	if !ok || gormHandle.GormDB() == nil {
		return newInMemoryMarketHandler(cfg), nil
	}
	handler, syncService := newMySQLMarketHandler(gormHandle.GormDB(), cfg)
	if err := syncService.RecoverStaleTasks(ctx); err != nil {
		return nil, fmt.Errorf("recover stale sync tasks: %w", err)
	}
	return handler, nil
}

func newInMemoryMarketHandler(cfg config.Config) *market.Handler {
	stocks := market.NewMockStockRepository()
	kLines := market.NewMockKLineRepository()
	tradeCalendars := market.NewMockTradeCalendarRepository()
	syncTasks := market.NewMockSyncTaskRepository()
	instrumentProvider, calendarProvider, marketDataProvider := newMarketProviders(cfg)
	queryService := market.NewQueryService(stocks, kLines, tradeCalendars, syncTasks)
	syncService := market.NewSyncService(
		stocks,
		kLines,
		tradeCalendars,
		syncTasks,
		instrumentProvider,
		calendarProvider,
		marketDataProvider,
	)
	return market.NewHandler(queryService, syncService)
}

func newMySQLMarketHandler(db *gorm.DB, cfg config.Config) (*market.Handler, *market.SyncService) {
	repos := market.NewMySQLRepositories(db)
	instrumentProvider, calendarProvider, marketDataProvider := newMarketProviders(cfg)
	queryService := market.NewQueryService(repos.Stocks(), repos.KLines(), repos.TradeCalendars(), repos.SyncTasks())
	syncService := market.NewSyncService(
		repos.Stocks(),
		repos.KLines(),
		repos.TradeCalendars(),
		repos.SyncTasks(),
		instrumentProvider,
		calendarProvider,
		marketDataProvider,
	)
	return market.NewHandler(queryService, syncService), syncService
}

func newMarketProviders(cfg config.Config) (market.InstrumentProvider, market.CalendarProvider, market.MarketDataProvider) {
	if cfg.Market.Provider == "tushare" {
		client := tushare.NewClient(cfg.Tushare.BaseURL, cfg.Tushare.Token)
		return market.NewTushareInstrumentProvider(client),
			market.NewTushareCalendarProvider(client),
			market.NewTushareMarketDataProvider(client)
	}
	return market.NewMockInstrumentProvider(nil),
		market.NewMockCalendarProvider(nil),
		market.NewMockMarketDataProvider(nil)
}

func serve(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		// Phase 1 只做 HTTP 优雅关闭；运行中的同步任务由下次启动标记为失败。
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return ctx.Err()
	}
}
