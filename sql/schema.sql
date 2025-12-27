-- 数据库初始化脚本
-- 适用数据库: MySQL 8.0+ / PostgreSQL 14+
-- 精度说明: 核心财务字段统一升级为 DECIMAL(36, 18) 以支持 Meme 币及高精度费率计算
-- ----------------------------

-- ----------------------------
-- 1. 订单表 (Orders)
-- 记录所有本地生成的订单请求及其状态
-- ----------------------------
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    -- 核心优化: 允许 NULL，支持先落库(PENDING)再发送 HTTP 请求，防止 WS 竞态
    order_id VARCHAR(64) COMMENT '交易所返回的订单ID (可为空)',
    client_oid VARCHAR(64) NOT NULL UNIQUE COMMENT '本地生成的唯一订单ID (幂等键)',
    exchange VARCHAR(32) NOT NULL COMMENT '交易所 (BINANCE, OKX)',
    symbol VARCHAR(32) NOT NULL COMMENT '交易对 (BTCUSDT)',
    
    -- 订单详情
    side VARCHAR(10) NOT NULL COMMENT '方向 (BUY, SELL)',
    type VARCHAR(20) NOT NULL COMMENT '类型 (LIMIT, MARKET, STOP_LOSS)',
    price DECIMAL(36, 18) DEFAULT 0 COMMENT '委托价格',
    quantity DECIMAL(36, 18) NOT NULL COMMENT '委托数量',
    amount DECIMAL(36, 18) DEFAULT 0 COMMENT '委托金额(市价买单用)',
    
    -- 状态与结果
    status VARCHAR(20) NOT NULL COMMENT '状态 (NEW, PARTIALLY_FILLED, FILLED, CANCELED, REJECTED)',
    avg_price DECIMAL(36, 18) DEFAULT 0 COMMENT '平均成交价',
    filled_qty DECIMAL(36, 18) DEFAULT 0 COMMENT '已成交数量',
    cum_quote DECIMAL(36, 18) DEFAULT 0 COMMENT '成交总金额 (Quote Asset)',
    
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
    -- 核心优化: 增加 client_oid 冗余，确保在 orders.id 未生成时也能关联
    order_id BIGINT COMMENT '关联的本地订单ID (Orders.id)',
    client_oid VARCHAR(64) NOT NULL COMMENT '关联的本地ClientOid',
    
    exec_id VARCHAR(64) NOT NULL COMMENT '交易所成交ID (Trade ID)',
    symbol VARCHAR(32) NOT NULL,
    side VARCHAR(10) NOT NULL,
    price DECIMAL(36, 18) NOT NULL COMMENT '成交价格',
    quantity DECIMAL(36, 18) NOT NULL COMMENT '成交数量',
    quote_qty DECIMAL(36, 18) NOT NULL COMMENT '成交金额',
    fee DECIMAL(36, 18) DEFAULT 0 COMMENT '手续费',
    fee_asset VARCHAR(10) COMMENT '手续费币种 (BNB, USDT)',
    
    traded_at TIMESTAMP NOT NULL COMMENT '成交时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_exec_id (exec_id, symbol),
    INDEX idx_client_oid (client_oid),
    INDEX idx_traded_at (traded_at)
) COMMENT = '成交明细表';

-- ----------------------------
-- 3. 风控日志表 (Risk Records)
-- ----------------------------
CREATE TABLE IF NOT EXISTS risk_records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL COMMENT '事件类型 (POS_LIMIT, LOSS_CIRCUIT, SLIPPAGE)',
    level VARCHAR(20) NOT NULL COMMENT '级别 (WARNING, BLOCK, CRITICAL)',
    symbol VARCHAR(32) COMMENT '相关币种',
    strategy_id VARCHAR(32) COMMENT '来源策略',
    details JSON COMMENT '详细快照',
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_event_type (event_type),
    INDEX idx_triggered_at (triggered_at)
) COMMENT = '风控拦截与告警记录';

-- ----------------------------
-- 4. 账户快照表 (Asset Snapshots)
-- ----------------------------
CREATE TABLE IF NOT EXISTS asset_snapshots (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    exchange VARCHAR(32) NOT NULL,
    total_balance DECIMAL(36, 18) NOT NULL COMMENT '总资产估值(USDT)',
    available_balance DECIMAL(36, 18) NOT NULL,
    frozen_balance DECIMAL(36, 18) NOT NULL,
    snapshot_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_snapshot_time (snapshot_time)
) COMMENT = '每日资产快照';

-- ----------------------------
-- 5. 策略配置表 (Strategy Configs)
-- ----------------------------
CREATE TABLE IF NOT EXISTS strategy_configs (
    key_name VARCHAR(64) PRIMARY KEY,
    value_str VARCHAR(255),
    value_num DECIMAL(36, 18),
    description VARCHAR(255),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) COMMENT = '策略动态配置';

INSERT INTO strategy_configs (key_name, value_num, description)
VALUES ('risk.max_pos_ratio', 0.30, '总仓位上限'),
    ('risk.single_pos_ratio', 0.05, '单笔仓位上限'),
    ('risk.max_loss_circuit', 3, '连续亏损熔断次数');

-- ----------------------------
-- 6. 统一结算单表 (Settlements)
-- ----------------------------
CREATE TABLE IF NOT EXISTS settlements (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    trade_id VARCHAR(64) NOT NULL COMMENT '逻辑交易ID',
    symbol VARCHAR(32) NOT NULL,
    market_type VARCHAR(10) NOT NULL,
    side VARCHAR(10) NOT NULL,
    
    -- 核心财务 (高精度)
    realized_pnl DECIMAL(36, 18) NOT NULL COMMENT '已实现盈亏 (净值)',
    commission DECIMAL(36, 18) NOT NULL COMMENT '累计手续费支出(折算后)',
    funding_fee DECIMAL(36, 18) DEFAULT 0 COMMENT '累计资金费(折算后)',
    
    entry_price DECIMAL(36, 18) NOT NULL,
    exit_price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    roi DECIMAL(12, 6) COMMENT '收益率',
    
    opened_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP NOT NULL,
    duration_seconds BIGINT,
    
    metadata JSON COMMENT '扩展元数据: {raw_fee_asset: "BNB", raw_fee_qty: "0.005"}',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_symbol_market (symbol, market_type),
    INDEX idx_closed_at (closed_at)
) COMMENT = '统一结算单表';

-- ----------------------------
-- 7. 用户表 (Users) - 适配 WebAuthn
-- ----------------------------
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid CHAR(36) NOT NULL UNIQUE COMMENT 'WebAuthn User Handle',
    username VARCHAR(64) NOT NULL UNIQUE,
    display_name VARCHAR(64) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'VIEWER' COMMENT 'ADMIN, OPERATOR, VIEWER',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username)
) COMMENT = '系统用户表';

-- ----------------------------
-- 8. WebAuthn 凭证表 (WebAuthn Credentials)
-- ----------------------------
CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    webauthn_id VARBINARY(1024) NOT NULL COMMENT 'Credential ID',
    public_key VARBINARY(4096) NOT NULL COMMENT 'COSE Encoded Public Key',
    attestation_type VARCHAR(32) NOT NULL,
    transport JSON COMMENT 'Transport types: usb, nfc, ble, internal',
    aaguid BINARY(16) COMMENT 'Authenticator Attestation GUID',
    sign_count INT UNSIGNED DEFAULT 0 COMMENT 'Signature Counter',
    clone_warning BOOLEAN DEFAULT FALSE,
    device_name VARCHAR(64) COMMENT 'User Defined Device Name',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_webauthn_id (webauthn_id)
) COMMENT = 'Passkeys/WebAuthn 凭证表';
