-- PostgreSQL init script for Alpha-Trade (partitioned candles + Events + Core Trading)
-- 首次启动 PostgreSQL 时执行：创建分区 candles 表、major_events 表及核心交易表。

-- ==========================================
-- Part 1: Market Data (Candles & Events)
-- ==========================================

CREATE TABLE IF NOT EXISTS market_candles (
  exchange TEXT NOT NULL,
  symbol TEXT NOT NULL,
  interval TEXT NOT NULL,
  open_time TIMESTAMPTZ NOT NULL,
  open DOUBLE PRECISION NOT NULL,
  high DOUBLE PRECISION NOT NULL,
  low DOUBLE PRECISION NOT NULL,
  close DOUBLE PRECISION NOT NULL,
  volume DOUBLE PRECISION NULL,
  raw JSONB NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (exchange, symbol, interval, open_time)
) PARTITION BY RANGE (open_time);

CREATE TABLE IF NOT EXISTS market_candles_2025 PARTITION OF market_candles FOR
VALUES FROM ('2025-01-01 00:00:00+00') TO ('2026-01-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS market_candles_2026 PARTITION OF market_candles FOR
VALUES FROM ('2026-01-01 00:00:00+00') TO ('2027-01-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS market_candles_default PARTITION OF market_candles DEFAULT;

-- 索引用于过滤 symbol/interval + 时间段
CREATE INDEX IF NOT EXISTS idx_market_candles_symbol_interval_time ON market_candles (symbol, interval, open_time DESC);
CREATE INDEX IF NOT EXISTS idx_market_candles_exchange_symbol_interval_time ON market_candles (exchange, symbol, interval, open_time DESC);

-- 辅助函数：按年份创建市场分区
CREATE OR REPLACE FUNCTION create_market_candles_partition(target_year INT) RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE 
  start_ts TIMESTAMPTZ := make_timestamptz(target_year, 1, 1, 0, 0, 0, 'UTC');
  end_ts TIMESTAMPTZ := start_ts + INTERVAL '1 year';
  part_name TEXT := format('market_candles_%s', target_year);
  sql TEXT;
BEGIN 
  sql := format(
    $fmt$ CREATE TABLE IF NOT EXISTS %I PARTITION OF market_candles FOR VALUES FROM (%L) TO (%L); $fmt$,
    part_name, start_ts, end_ts
  );
  EXECUTE sql;
END;
$$;

CREATE TABLE IF NOT EXISTS major_events (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  event_date DATE NOT NULL,
  severity INT NOT NULL,
  category TEXT NOT NULL,
  tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  description TEXT NOT NULL,
  description_cn TEXT NOT NULL,
  impact_pattern TEXT NOT NULL,
  short_term_impact DOUBLE PRECISION NOT NULL,
  source JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  btc_delta_1d DOUBLE PRECISION NULL,
  eth_delta_1d DOUBLE PRECISION NULL,
  btc_delta_1h DOUBLE PRECISION NULL,
  eth_delta_1h DOUBLE PRECISION NULL,
  btc_price_usd DOUBLE PRECISION NULL,
  eth_price_usd DOUBLE PRECISION NULL
);
CREATE INDEX IF NOT EXISTS idx_major_events_date ON major_events(event_date);

-- ==========================================
-- Part 2: Core Trading System (High Precision)
-- ==========================================

-- 1. 订单表 (Orders)
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_id VARCHAR(64), -- 允许 NULL，支持先落库(PENDING)
    client_oid VARCHAR(64) NOT NULL UNIQUE, -- 幂等键
    exchange VARCHAR(32) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    side VARCHAR(10) NOT NULL,
    type VARCHAR(20) NOT NULL,
    
    -- 高精度财务字段 NUMERIC(36, 18)
    price NUMERIC(36, 18) DEFAULT 0,
    quantity NUMERIC(36, 18) NOT NULL,
    amount NUMERIC(36, 18) DEFAULT 0,
    
    status VARCHAR(20) NOT NULL,
    avg_price NUMERIC(36, 18) DEFAULT 0,
    filled_qty NUMERIC(36, 18) DEFAULT 0,
    cum_quote NUMERIC(36, 18) DEFAULT 0,
    
    strategy_id VARCHAR(32),
    error_msg TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_orders_client_oid ON orders(client_oid);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- 2. 成交明细表 (Executions)
CREATE TABLE IF NOT EXISTS executions (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT,
    client_oid VARCHAR(64) NOT NULL,
    exec_id VARCHAR(64) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    side VARCHAR(10) NOT NULL,
    
    price NUMERIC(36, 18) NOT NULL,
    quantity NUMERIC(36, 18) NOT NULL,
    quote_qty NUMERIC(36, 18) NOT NULL,
    fee NUMERIC(36, 18) DEFAULT 0,
    fee_asset VARCHAR(10),
    
    traded_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (exec_id, symbol)
);
CREATE INDEX IF NOT EXISTS idx_executions_client_oid ON executions(client_oid);
CREATE INDEX IF NOT EXISTS idx_executions_traded_at ON executions(traded_at);

-- 3. 风控日志表 (Risk Records)
CREATE TABLE IF NOT EXISTS risk_records (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    level VARCHAR(20) NOT NULL,
    symbol VARCHAR(32),
    strategy_id VARCHAR(32),
    details JSONB,
    triggered_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_risk_records_event_type ON risk_records(event_type);
CREATE INDEX IF NOT EXISTS idx_risk_records_triggered_at ON risk_records(triggered_at);

-- 4. 账户快照表 (Asset Snapshots)
CREATE TABLE IF NOT EXISTS asset_snapshots (
    id BIGSERIAL PRIMARY KEY,
    exchange VARCHAR(32) NOT NULL,
    total_balance NUMERIC(36, 18) NOT NULL,
    available_balance NUMERIC(36, 18) NOT NULL,
    frozen_balance NUMERIC(36, 18) NOT NULL,
    snapshot_time TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_asset_snapshots_time ON asset_snapshots(snapshot_time);

-- 5. 策略配置表 (Strategy Configs)
CREATE TABLE IF NOT EXISTS strategy_configs (
    key_name VARCHAR(64) PRIMARY KEY,
    value_str VARCHAR(255),
    value_num NUMERIC(36, 18),
    description VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT now()
);

INSERT INTO strategy_configs (key_name, value_num, description)
VALUES 
    ('risk.max_pos_ratio', 0.30, '总仓位上限'),
    ('risk.single_pos_ratio', 0.05, '单笔仓位上限'),
    ('risk.max_loss_circuit', 3, '连续亏损熔断次数')
ON CONFLICT (key_name) DO NOTHING;

-- 6. 统一结算单表 (Settlements)
CREATE TABLE IF NOT EXISTS settlements (
    id BIGSERIAL PRIMARY KEY,
    trade_id VARCHAR(64) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    market_type VARCHAR(10) NOT NULL,
    side VARCHAR(10) NOT NULL,
    
    realized_pnl NUMERIC(36, 18) NOT NULL,
    commission NUMERIC(36, 18) NOT NULL,
    funding_fee NUMERIC(36, 18) DEFAULT 0,
    
    entry_price NUMERIC(36, 18) NOT NULL,
    exit_price NUMERIC(36, 18) NOT NULL,
    quantity NUMERIC(36, 18) NOT NULL,
    roi NUMERIC(12, 6),
    
    opened_at TIMESTAMPTZ NOT NULL,
    closed_at TIMESTAMPTZ NOT NULL,
    duration_seconds BIGINT,
    
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_settlements_symbol_market ON settlements(symbol, market_type);
CREATE INDEX IF NOT EXISTS idx_settlements_closed_at ON settlements(closed_at);
