-- PostgreSQL init script for Alpha-Trade (partitioned candles + Events)
-- 首次启动 PostgreSQL 时执行：创建分区 candles 表与 major_events 表。
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
VALUES
FROM ('2025-01-01 00:00:00+00') TO ('2026-01-01 00:00:00+00');
CREATE TABLE IF NOT EXISTS market_candles_2026 PARTITION OF market_candles FOR
VALUES
FROM ('2026-01-01 00:00:00+00') TO ('2027-01-01 00:00:00+00');
CREATE TABLE IF NOT EXISTS market_candles_default PARTITION OF market_candles DEFAULT;
-- 索引用于过滤 symbol/interval + 时间段
CREATE INDEX IF NOT EXISTS idx_market_candles_symbol_interval_time ON market_candles (symbol, interval, open_time DESC);
CREATE INDEX IF NOT EXISTS idx_market_candles_exchange_symbol_interval_time ON market_candles (exchange, symbol, interval, open_time DESC);
-- `create_market_candles_partition` 用于按年提前创建分区，避免插入落入 DEFAULT。
-- 辅助函数：按年份创建市场分区
-- 使用方式示例：SELECT create_market_candles_partition(2027);
CREATE OR REPLACE FUNCTION create_market_candles_partition(target_year INT) RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE start_ts TIMESTAMPTZ := make_timestamptz(target_year, 1, 1, 0, 0, 0, 'UTC');
end_ts TIMESTAMPTZ := start_ts + INTERVAL '1 year';
part_name TEXT := format('market_candles_%s', target_year);
sql TEXT;
BEGIN sql := format(
  $fmt$ CREATE TABLE IF NOT EXISTS %I PARTITION OF market_candles FOR
  VALUES
  FROM (%L) TO (%L);
$fmt$,
part_name,
start_ts,
end_ts
);
EXECUTE sql;
END;
$$;
COMMENT ON TABLE market_candles IS '按开盘时间分区的K线数据，包含多交易所、多周期';
COMMENT ON COLUMN market_candles.exchange IS '行情来源交易所，例如 BINANCE、OKX';
COMMENT ON COLUMN market_candles.symbol IS '交易对，例如 BTC-USDT';
COMMENT ON COLUMN market_candles.interval IS '周期（15m、1H、4H、1D）';
COMMENT ON COLUMN market_candles.open_time IS '该K线的 UTC 开盘时间';
COMMENT ON COLUMN market_candles.open IS '开盘价';
COMMENT ON COLUMN market_candles.high IS '最高价';
COMMENT ON COLUMN market_candles.low IS '最低价';
COMMENT ON COLUMN market_candles.close IS '收盘价';
COMMENT ON COLUMN market_candles.volume IS '成交量（可为空）';
COMMENT ON COLUMN market_candles.raw IS '原始 OKX 响应 JSON，可调试用';
COMMENT ON COLUMN market_candles.created_at IS '写入记录的时间戳';
COMMENT ON FUNCTION create_market_candles_partition(INT) IS '按指定年份创建对应 open_time 的分区（示例：SELECT create_market_candles_partition(2027);）';
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
  -- 创建时间
  btc_delta_1d DOUBLE PRECISION NULL,
  eth_delta_1d DOUBLE PRECISION NULL,
  btc_delta_1h DOUBLE PRECISION NULL,
  eth_delta_1h DOUBLE PRECISION NULL,
  btc_price_usd DOUBLE PRECISION NULL,
  eth_price_usd DOUBLE PRECISION NULL
);
CREATE INDEX IF NOT EXISTS idx_major_events_date ON major_events(event_date);
COMMENT ON COLUMN major_events.btc_delta_1d IS '事件前后比特币过去24小时涨跌幅（百分比）';
COMMENT ON COLUMN major_events.eth_delta_1d IS '事件前后以太坊过去24小时涨跌幅（百分比）';
COMMENT ON COLUMN major_events.btc_delta_1h IS '事件前后比特币过去1小时涨跌幅（百分比）';
COMMENT ON COLUMN major_events.eth_delta_1h IS '事件前后以太坊过去1小时涨跌幅（百分比）';
COMMENT ON COLUMN major_events.btc_price_usd IS '事件发生时的比特币价格（USD）';
COMMENT ON COLUMN major_events.eth_price_usd IS '事件发生时的以太坊价格（USD）';
COMMENT ON TABLE major_events IS '重大历史事件日志，用于事件驱动回测/AI 补充知识';
COMMENT ON COLUMN major_events.name IS '事件名称';
COMMENT ON COLUMN major_events.event_date IS '事件发生日期';
COMMENT ON COLUMN major_events.severity IS '严重程度评分（1-10）';
COMMENT ON COLUMN major_events.category IS '事件分类';
COMMENT ON COLUMN major_events.tags IS '关键词列表';
COMMENT ON COLUMN major_events.description IS '英文描述';
COMMENT ON COLUMN major_events.description_cn IS '中文描述';
COMMENT ON COLUMN major_events.impact_pattern IS '典型影响模式';
COMMENT ON COLUMN major_events.short_term_impact IS '短期影响百分比（正负）';
COMMENT ON COLUMN major_events.source IS '来源/链接JSON';
COMMENT ON COLUMN major_events.created_at IS '记录创建时间';