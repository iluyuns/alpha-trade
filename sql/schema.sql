-- 数据库初始化脚本
-- 适用数据库: MySQL 8.0+ / PostgreSQL 14+ (语法尽量通用，特定类型请根据实际调整)
-- ----------------------------
-- 1. 订单表 (Orders)
-- 记录所有本地生成的订单请求及其状态
-- ----------------------------
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    order_id VARCHAR(64) NOT NULL COMMENT '交易所返回的订单ID',
    client_oid VARCHAR(64) NOT NULL UNIQUE COMMENT '本地生成的唯一订单ID (幂等键)',
    exchange VARCHAR(32) NOT NULL COMMENT '交易所 (BINANCE, OKX)',
    symbol VARCHAR(32) NOT NULL COMMENT '交易对 (BTCUSDT)',
    -- 订单详情
    side VARCHAR(10) NOT NULL COMMENT '方向 (BUY, SELL)',
    type VARCHAR(20) NOT NULL COMMENT '类型 (LIMIT, MARKET, STOP_LOSS)',
    price DECIMAL(24, 8) DEFAULT 0 COMMENT '委托价格',
    quantity DECIMAL(24, 8) NOT NULL COMMENT '委托数量',
    amount DECIMAL(24, 8) DEFAULT 0 COMMENT '委托金额(市价买单用)',
    -- 状态与结果
    status VARCHAR(20) NOT NULL COMMENT '状态 (NEW, PARTIALLY_FILLED, FILLED, CANCELED, REJECTED)',
    avg_price DECIMAL(24, 8) DEFAULT 0 COMMENT '平均成交价',
    filled_qty DECIMAL(24, 8) DEFAULT 0 COMMENT '已成交数量',
    cum_quote DECIMAL(24, 8) DEFAULT 0 COMMENT '成交总金额 (Quote Asset)',
    -- 审计信息
    strategy_id VARCHAR(32) COMMENT '触发策略ID',
    error_msg TEXT COMMENT '错误信息(若被拒或失败)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_client_oid (client_oid),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) COMMENT = '交易订单表';
-- ----------------------------
-- 2. 成交明细表 (Executions / Trades)
-- 记录交易所推送的每一笔实际成交
-- ----------------------------
CREATE TABLE IF NOT EXISTS executions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT NOT NULL COMMENT '关联的本地订单ID',
    exec_id VARCHAR(64) NOT NULL COMMENT '交易所成交ID (Trade ID)',
    symbol VARCHAR(32) NOT NULL,
    side VARCHAR(10) NOT NULL,
    price DECIMAL(24, 8) NOT NULL COMMENT '成交价格',
    quantity DECIMAL(24, 8) NOT NULL COMMENT '成交数量',
    quote_qty DECIMAL(24, 8) NOT NULL COMMENT '成交金额',
    fee DECIMAL(24, 8) DEFAULT 0 COMMENT '手续费',
    fee_asset VARCHAR(10) COMMENT '手续费币种 (BNB, USDT)',
    traded_at TIMESTAMP NOT NULL COMMENT '成交时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_exec_id (exec_id, symbol),
    INDEX idx_order_id (order_id),
    INDEX idx_traded_at (traded_at)
) COMMENT = '成交明细表';
-- ----------------------------
-- 3. 风控日志表 (Risk Records)
-- 记录所有被风控拦截的事件或系统熔断记录
-- ----------------------------
CREATE TABLE IF NOT EXISTS risk_records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL COMMENT '事件类型 (POS_LIMIT, LOSS_CIRCUIT, SLIPPAGE)',
    level VARCHAR(20) NOT NULL COMMENT '级别 (WARNING, BLOCK, CRITICAL)',
    symbol VARCHAR(32) COMMENT '相关币种',
    strategy_id VARCHAR(32) COMMENT '来源策略',
    details JSON COMMENT '详细快照 (如当时仓位、价格、拦截原因)',
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_event_type (event_type),
    INDEX idx_triggered_at (triggered_at)
) COMMENT = '风控拦截与告警记录';
-- ----------------------------
-- 4. 账户快照表 (Asset Snapshots)
-- 用于计算回撤和生成日报
-- ----------------------------
CREATE TABLE IF NOT EXISTS asset_snapshots (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    exchange VARCHAR(32) NOT NULL,
    total_balance DECIMAL(24, 8) NOT NULL COMMENT '总资产估值(USDT)',
    available_balance DECIMAL(24, 8) NOT NULL COMMENT '可用余额',
    frozen_balance DECIMAL(24, 8) NOT NULL COMMENT '冻结余额',
    snapshot_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_snapshot_time (snapshot_time)
) COMMENT = '每日资产快照';
-- ----------------------------
-- 5. 策略配置表 (Strategy Configs)
-- 支持动态调整策略参数，无需重启
-- ----------------------------
CREATE TABLE IF NOT EXISTS strategy_configs (
    key_name VARCHAR(64) PRIMARY KEY COMMENT '配置键 (e.g. volatility.threshold)',
    value_str VARCHAR(255) COMMENT '字符串值',
    value_num DECIMAL(24, 8) COMMENT '数值值',
    description VARCHAR(255) COMMENT '说明',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) COMMENT = '策略动态配置';
-- 初始化示例配置
INSERT INTO strategy_configs (key_name, value_num, description)
VALUES ('risk.max_pos_ratio', 0.30, '总仓位上限'),
    ('risk.single_pos_ratio', 0.05, '单笔仓位上限'),
    ('risk.max_loss_circuit', 3, '连续亏损熔断次数');