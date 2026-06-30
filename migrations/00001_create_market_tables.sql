-- +goose Up
CREATE TABLE stocks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    market VARCHAR(16) NOT NULL,
    asset_type VARCHAR(32) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    ts_code VARCHAR(32) NOT NULL,
    name VARCHAR(128) NOT NULL,
    exchange VARCHAR(32) NOT NULL,
    board VARCHAR(32) NOT NULL DEFAULT 'UNKNOWN',
    area VARCHAR(64) NOT NULL DEFAULT '',
    industry VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL,
    list_date DATE NULL,
    delist_date DATE NULL,
    data_source VARCHAR(32) NOT NULL,
    synced_at DATETIME(3) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_stock_identity (market, asset_type, symbol),
    KEY idx_stocks_filter (market, asset_type, exchange, status),
    KEY idx_stocks_symbol (symbol),
    KEY idx_stocks_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE daily_k_lines (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    market VARCHAR(16) NOT NULL,
    asset_type VARCHAR(32) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    trade_date DATE NOT NULL,
    open_price DECIMAL(20,4) NOT NULL,
    high_price DECIMAL(20,4) NOT NULL,
    low_price DECIMAL(20,4) NOT NULL,
    close_price DECIMAL(20,4) NOT NULL,
    pre_close DECIMAL(20,4) NOT NULL,
    change_amt DECIMAL(20,4) NOT NULL,
    pct_change DECIMAL(20,4) NOT NULL,
    volume DECIMAL(24,4) NOT NULL,
    amount DECIMAL(24,4) NOT NULL,
    data_source VARCHAR(32) NOT NULL,
    synced_at DATETIME(3) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_daily_kline_identity (market, asset_type, symbol, trade_date),
    KEY idx_daily_kline_query (market, asset_type, symbol, trade_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE trade_calendars (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    market VARCHAR(16) NOT NULL,
    exchange VARCHAR(32) NOT NULL,
    cal_date DATE NOT NULL,
    is_open TINYINT(1) NOT NULL,
    pretrade_date DATE NULL,
    data_source VARCHAR(32) NOT NULL,
    synced_at DATETIME(3) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_trade_calendar_identity (market, exchange, cal_date),
    KEY idx_trade_calendar_open_day (market, exchange, is_open, cal_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE sync_tasks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    uid CHAR(26) NOT NULL,
    task_type VARCHAR(64) NOT NULL,
    market VARCHAR(16) NOT NULL,
    asset_type VARCHAR(32) NOT NULL DEFAULT 'STOCK',
    data_source VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    total_items BIGINT NOT NULL DEFAULT 0,
    processed_items BIGINT NOT NULL DEFAULT 0,
    success_items BIGINT NOT NULL DEFAULT 0,
    failed_items BIGINT NOT NULL DEFAULT 0,
    request_id VARCHAR(128) NOT NULL DEFAULT '',
    started_at DATETIME(3) NULL,
    finished_at DATETIME(3) NULL,
    error_msg TEXT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_sync_tasks_uid (uid),
    KEY idx_sync_tasks_status (status),
    KEY idx_sync_tasks_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE sync_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id BIGINT UNSIGNED NOT NULL,
    task_uid CHAR(26) NOT NULL,
    step VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    market VARCHAR(16) NOT NULL DEFAULT '',
    asset_type VARCHAR(32) NOT NULL DEFAULT '',
    symbol VARCHAR(32) NOT NULL DEFAULT '',
    data_source VARCHAR(32) NOT NULL DEFAULT '',
    message TEXT NULL,
    error_detail TEXT NULL,
    affected_rows BIGINT NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_sync_logs_task_id (task_id),
    KEY idx_sync_logs_task_uid (task_uid),
    KEY idx_sync_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE IF EXISTS sync_logs;
DROP TABLE IF EXISTS sync_tasks;
DROP TABLE IF EXISTS trade_calendars;
DROP TABLE IF EXISTS daily_k_lines;
DROP TABLE IF EXISTS stocks;
