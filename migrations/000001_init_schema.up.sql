/*
    Alpha-Trade Database Initial Schema (PostgreSQL)
    ===============================================
    Precision Policy: DECIMAL(36, 18) for all financial fields.
    Security: Passkey (WebAuthn) + Envelope Encryption + Step-up Auth.
*/

-- 1. Users (系统用户表)
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE,
    username VARCHAR(64) NOT NULL UNIQUE,
    display_name VARCHAR(64) NOT NULL,
    password_hash VARCHAR(255),
    role VARCHAR(20) NOT NULL DEFAULT 'VIEWER',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE users IS '系统用户表：管理后台访问人员';
COMMENT ON COLUMN users.id IS '内部自增主键';
COMMENT ON COLUMN users.uuid IS 'WebAuthn User Handle (UUID)';
COMMENT ON COLUMN users.username IS '登录用户名';
COMMENT ON COLUMN users.display_name IS '用户显示名称';
COMMENT ON COLUMN users.password_hash IS '静态密码哈希：仅用于 Break-Glass 紧急恢复 (Argon2id)';
COMMENT ON COLUMN users.role IS '用户角色 [ENUM: ADMIN(超级管理员), OPERATOR(交易员), VIEWER(只读观察员)]';
COMMENT ON COLUMN users.is_active IS '账号激活状态';
COMMENT ON COLUMN users.created_at IS '账号创建时间';
COMMENT ON COLUMN users.updated_at IS '最后更新时间';

-- 2. WebAuthn Credentials (通行证凭证表)
CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    webauthn_id BYTEA NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    attestation_type VARCHAR(32) NOT NULL,
    transport JSONB,
    aaguid UUID,
    sign_count INT DEFAULT 0,
    clone_warning BOOLEAN DEFAULT FALSE,
    device_name VARCHAR(64),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE webauthn_credentials IS 'WebAuthn 凭证表：存储 FIDO2/Passkey 硬件认证数据';
COMMENT ON COLUMN webauthn_credentials.id IS '凭证自增 ID';
COMMENT ON COLUMN webauthn_credentials.user_id IS '关联的用户 ID';
COMMENT ON COLUMN webauthn_credentials.webauthn_id IS '浏览器返回的唯一凭证 ID';
COMMENT ON COLUMN webauthn_credentials.public_key IS 'COSE 编码的认证公钥';
COMMENT ON COLUMN webauthn_credentials.attestation_type IS '认证器声明类型';
COMMENT ON COLUMN webauthn_credentials.transport IS '支持的传输协议 [ENUM: usb, nfc, ble, internal, hybrid]';
COMMENT ON COLUMN webauthn_credentials.aaguid IS '验证器型号唯一标识';
COMMENT ON COLUMN webauthn_credentials.sign_count IS '签名计数器：用于检测凭证克隆风险';
COMMENT ON COLUMN webauthn_credentials.clone_warning IS '凭证是否存在克隆嫌疑标记';
COMMENT ON COLUMN webauthn_credentials.device_name IS '用户定义的硬件设备名称';
COMMENT ON COLUMN webauthn_credentials.created_at IS '凭证注册时间';
COMMENT ON COLUMN webauthn_credentials.last_used_at IS '最后一次认证时间';

-- 3. Permissions (权限资源表)
CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    module VARCHAR(64) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE permissions IS '权限资源表：定义系统原子操作权限';
COMMENT ON COLUMN permissions.id IS '权限 ID';
COMMENT ON COLUMN permissions.code IS '权限唯一编码 (e.g., strategy.start)';
COMMENT ON COLUMN permissions.name IS '权限显示名称';
COMMENT ON COLUMN permissions.module IS '所属业务模块 [ENUM: CORE, STRATEGY, RISK, SYSTEM, MONITOR]';
COMMENT ON COLUMN permissions.description IS '详细描述';
COMMENT ON COLUMN permissions.created_at IS '创建时间';

-- 4. Role-Permissions (角色权限映射表)
CREATE TABLE IF NOT EXISTS role_permissions (
    role VARCHAR(20) NOT NULL,
    permission_code VARCHAR(64) NOT NULL REFERENCES permissions(code) ON DELETE CASCADE,
    PRIMARY KEY (role, permission_code)
);

COMMENT ON TABLE role_permissions IS '角色权限映射表：RBAC 核心配置';
COMMENT ON COLUMN role_permissions.role IS '角色名称 [ENUM: ADMIN, OPERATOR, VIEWER]';
COMMENT ON COLUMN role_permissions.permission_code IS '关联的权限唯一编码';

-- 5. Menus (动态菜单表)
CREATE TABLE IF NOT EXISTS menus (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT REFERENCES menus(id) ON DELETE CASCADE,
    name VARCHAR(64) NOT NULL,
    icon VARCHAR(64),
    path VARCHAR(128),
    permission_code VARCHAR(64) REFERENCES permissions(code) ON DELETE SET NULL,
    sort_order INT DEFAULT 0,
    is_hidden BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE menus IS '动态菜单表：驱动前端侧边栏导航';
COMMENT ON COLUMN menus.id IS '菜单 ID';
COMMENT ON COLUMN menus.parent_id IS '父级菜单 ID (支持多级)';
COMMENT ON COLUMN menus.name IS '菜单显示名称';
COMMENT ON COLUMN menus.icon IS '图标标识符 (Lucide/Iconify)';
COMMENT ON COLUMN menus.path IS '前端页面路由路径';
COMMENT ON COLUMN menus.permission_code IS '访问此菜单所需的权限编码';
COMMENT ON COLUMN menus.sort_order IS '排序权重 (从小到大)';
COMMENT ON COLUMN menus.is_hidden IS '是否在导航栏隐藏该菜单';
COMMENT ON COLUMN menus.created_at IS '创建时间';

-- 6. Exchange Accounts (交易所账户配置表)
CREATE TABLE IF NOT EXISTS exchange_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label VARCHAR(64) NOT NULL,
    exchange VARCHAR(32) NOT NULL,
    api_key VARCHAR(128) NOT NULL,
    encrypted_api_secret TEXT NOT NULL,
    encrypted_passphrase TEXT,
    config JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE exchange_accounts IS '交易所账户表：存储 API 密钥与账户属性';
COMMENT ON COLUMN exchange_accounts.id IS '账户自增 ID';
COMMENT ON COLUMN exchange_accounts.user_id IS '所属系统用户 ID';
COMMENT ON COLUMN exchange_accounts.label IS '账户备注名称 (e.g., Binance_Sub_01)';
COMMENT ON COLUMN exchange_accounts.exchange IS '交易所类型 [ENUM: BINANCE, OKX]';
COMMENT ON COLUMN exchange_accounts.api_key IS '交易所 API Key';
COMMENT ON COLUMN exchange_accounts.encrypted_api_secret IS '加密后的 API Secret (AES-256-GCM)';
COMMENT ON COLUMN exchange_accounts.encrypted_passphrase IS '加密后的 OKX Passphrase (仅 OKX 必填)';
COMMENT ON COLUMN exchange_accounts.config IS '扩展配置 [JSON: permissions, ip_whitelist, is_master]';
COMMENT ON COLUMN exchange_accounts.is_active IS '账户启用状态';
COMMENT ON COLUMN exchange_accounts.created_at IS '记录创建时间';
COMMENT ON COLUMN exchange_accounts.updated_at IS '最后更新时间';

-- 7. Audit Logs (系统审计日志表)
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id VARCHAR(64),
    changes JSONB,
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE audit_logs IS '系统审计日志：追踪所有敏感业务操作';
COMMENT ON COLUMN audit_logs.id IS '日志自增 ID';
COMMENT ON COLUMN audit_logs.user_id IS '操作执行人 ID';
COMMENT ON COLUMN audit_logs.ip_address IS '操作者 IP 地址';
COMMENT ON COLUMN audit_logs.action IS '行为类型 [ENUM: KILL_SWITCH, UPDATE_RISK, API_KEY_ADD, STRATEGY_START]';
COMMENT ON COLUMN audit_logs.target_type IS '操作对象类别 [ENUM: ACCOUNT, RISK_CONFIG, STRATEGY, USER]';
COMMENT ON COLUMN audit_logs.target_id IS '操作对象 ID';
COMMENT ON COLUMN audit_logs.changes IS '变更 Diff 内容 [JSON: old, new]';
COMMENT ON COLUMN audit_logs.is_verified IS '是否通过了 Passkey 二次提级认证';
COMMENT ON COLUMN audit_logs.created_at IS '审计记录时间';

-- 8. Orders (交易订单表)
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_id VARCHAR(64),
    client_oid VARCHAR(64) NOT NULL UNIQUE,
    exchange VARCHAR(32) NOT NULL, 
    symbol VARCHAR(32) NOT NULL,   
    side VARCHAR(10) NOT NULL,     
    type VARCHAR(20) NOT NULL,     
    price DECIMAL(36, 18) DEFAULT 0,    
    quantity DECIMAL(36, 18) NOT NULL,  
    amount DECIMAL(36, 18) DEFAULT 0,    
    status VARCHAR(20) NOT NULL,
    avg_price DECIMAL(36, 18) DEFAULT 0,  
    filled_qty DECIMAL(36, 18) DEFAULT 0, 
    cum_quote DECIMAL(36, 18) DEFAULT 0,  
    strategy_id VARCHAR(32), 
    error_msg TEXT,          
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE orders IS '交易订单表：追踪委托全生命周期状态';
COMMENT ON COLUMN orders.id IS '本地流水 ID';
COMMENT ON COLUMN orders.order_id IS '交易所原始订单 ID';
COMMENT ON COLUMN orders.client_oid IS '本地生成的唯一 ID (幂等键)';
COMMENT ON COLUMN orders.exchange IS '交易所名称';
COMMENT ON COLUMN orders.symbol IS '交易对名称 (e.g., BTCUSDT)';
COMMENT ON COLUMN orders.side IS '交易方向 [ENUM: BUY, SELL]';
COMMENT ON COLUMN orders.type IS '委托类型 [ENUM: LIMIT, MARKET, STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT, LIMIT_MAKER]';
COMMENT ON COLUMN orders.price IS '委托价格';
COMMENT ON COLUMN orders.quantity IS '委托数量 (Base Asset)';
COMMENT ON COLUMN orders.amount IS '委托金额 (Quote Asset，市价买单使用)';
COMMENT ON COLUMN orders.status IS '订单状态 [ENUM: NEW, PARTIALLY_FILLED, FILLED, CANCELED, REJECTED, EXPIRED]';
COMMENT ON COLUMN orders.avg_price IS '成交均价';
COMMENT ON COLUMN orders.filled_qty IS '已成交数量';
COMMENT ON COLUMN orders.cum_quote IS '累计成交金额 (Quote Asset)';
COMMENT ON COLUMN orders.strategy_id IS '所属策略 ID';
COMMENT ON COLUMN orders.error_msg IS '交易所返回的错误描述';
COMMENT ON COLUMN orders.created_at IS '订单创建时间';
COMMENT ON COLUMN orders.updated_at IS '状态最后同步时间';

CREATE INDEX idx_orders_client_oid ON orders (client_oid);
CREATE INDEX idx_orders_status ON orders (status);

-- 9. Executions (成交明细表)
CREATE TABLE IF NOT EXISTS executions (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT,             
    client_oid VARCHAR(64) NOT NULL,
    exec_id VARCHAR(64) NOT NULL,   
    symbol VARCHAR(32) NOT NULL,
    side VARCHAR(10) NOT NULL,
    price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    quote_qty DECIMAL(36, 18) NOT NULL, 
    fee DECIMAL(36, 18) DEFAULT 0,      
    fee_asset VARCHAR(10),              
    traded_at TIMESTAMP WITH TIME ZONE NOT NULL, 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (exec_id, symbol)
);

COMMENT ON TABLE executions IS '成交明细表：记录交易所推送的每一笔 Trade';
COMMENT ON COLUMN executions.id IS '流水 ID';
COMMENT ON COLUMN executions.order_id IS '关联的本地 Orders ID';
COMMENT ON COLUMN executions.client_oid IS '关联的本地 ClientOid';
COMMENT ON COLUMN executions.exec_id IS '交易所成交 ID (Trade ID)';
COMMENT ON COLUMN executions.symbol IS '交易对';
COMMENT ON COLUMN executions.side IS '成交方向 [ENUM: BUY, SELL]';
COMMENT ON COLUMN executions.price IS '成交价格';
COMMENT ON COLUMN executions.quantity IS '成交数量';
COMMENT ON COLUMN executions.quote_qty IS '成交金额 (Quote Asset)';
COMMENT ON COLUMN executions.fee IS '手续费数值';
COMMENT ON COLUMN executions.fee_asset IS '手续费计价币种';
COMMENT ON COLUMN executions.traded_at IS '交易所成交撮合时间';
COMMENT ON COLUMN executions.created_at IS '记录存库时间';

CREATE INDEX idx_executions_client_oid ON executions (client_oid);

-- 10. Risk Records (风控拦截日志表)
CREATE TABLE IF NOT EXISTS risk_records (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    level VARCHAR(20) NOT NULL,
    symbol VARCHAR(32),
    strategy_id VARCHAR(32),
    details JSONB, 
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE risk_records IS '风控拦截日志：记录 RiskManager 的实时决策';
COMMENT ON COLUMN risk_records.id IS '流水 ID';
COMMENT ON COLUMN risk_records.event_type IS '事件类型 [ENUM: POS_LIMIT, LOSS_CIRCUIT, SLIPPAGE, PRICE_PROTECT]';
COMMENT ON COLUMN risk_records.level IS '风险级别 [ENUM: INFO, WARNING, BLOCK, CRITICAL]';
COMMENT ON COLUMN risk_records.symbol IS '涉及标的';
COMMENT ON COLUMN risk_records.strategy_id IS '涉及策略';
COMMENT ON COLUMN risk_records.details IS '触发时的上下文快照 [JSON]';
COMMENT ON COLUMN risk_records.triggered_at IS '触发时间';

-- 11. Asset Snapshots (资产快照表)
CREATE TABLE IF NOT EXISTS asset_snapshots (
    id BIGSERIAL PRIMARY KEY,
    exchange VARCHAR(32) NOT NULL,
    total_balance DECIMAL(36, 18) NOT NULL,     
    available_balance DECIMAL(36, 18) NOT NULL, 
    frozen_balance DECIMAL(36, 18) NOT NULL,    
    snapshot_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE asset_snapshots IS '资产快照表：每日/定时权益对账记录';
COMMENT ON COLUMN asset_snapshots.id IS '流水 ID';
COMMENT ON COLUMN asset_snapshots.exchange IS '交易所类型';
COMMENT ON COLUMN asset_snapshots.total_balance IS '总权益估值 (USDT)';
COMMENT ON COLUMN asset_snapshots.available_balance IS '可用余额 (USDT)';
COMMENT ON COLUMN asset_snapshots.frozen_balance IS '冻结金额 (USDT)';
COMMENT ON COLUMN asset_snapshots.snapshot_time IS '快照生成时间';

-- 12. Strategy Configs (策略动态配置表)
CREATE TABLE IF NOT EXISTS strategy_configs (
    key_name VARCHAR(64) PRIMARY KEY, 
    value_str VARCHAR(255),
    value_num DECIMAL(36, 18),
    description VARCHAR(255),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE strategy_configs IS '策略配置表：存储动态风控阈值与业务参数';
COMMENT ON COLUMN strategy_configs.key_name IS '配置键名 (e.g., risk.max_pos_ratio)';
COMMENT ON COLUMN strategy_configs.value_str IS '字符串类型值';
COMMENT ON COLUMN strategy_configs.value_num IS '数值类型值';
COMMENT ON COLUMN strategy_configs.description IS '配置项含义描述';
COMMENT ON COLUMN strategy_configs.updated_at IS '最后修改时间';

-- 13. Settlements (统一结算单表)
CREATE TABLE IF NOT EXISTS settlements (
    id BIGSERIAL PRIMARY KEY,
    trade_id VARCHAR(64) NOT NULL, 
    symbol VARCHAR(32) NOT NULL,
    market_type VARCHAR(10) NOT NULL, 
    side VARCHAR(10) NOT NULL,
    realized_pnl DECIMAL(36, 18) NOT NULL, 
    commission DECIMAL(36, 18) NOT NULL,   
    funding_fee DECIMAL(36, 18) DEFAULT 0, 
    entry_price DECIMAL(36, 18) NOT NULL,
    exit_price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    roi DECIMAL(12, 6), 
    opened_at TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_seconds BIGINT, 
    metadata JSONB, 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE settlements IS '统一结算单表：实现跨市场的 PnL 最终归因';
COMMENT ON COLUMN settlements.id IS '结算单 ID';
COMMENT ON COLUMN settlements.trade_id IS '逻辑交易 ID (跨订单关联标识)';
COMMENT ON COLUMN settlements.symbol IS '交易对';
COMMENT ON COLUMN settlements.market_type IS '市场类型 [ENUM: SPOT, SWAP, FUTURE]';
COMMENT ON COLUMN settlements.side IS '盈亏方向 [ENUM: LONG, SHORT]';
COMMENT ON COLUMN settlements.realized_pnl IS '已实现净盈亏 (扣费后)';
COMMENT ON COLUMN settlements.commission IS '累计手续费支出';
COMMENT ON COLUMN settlements.funding_fee IS '累计资金费支出 (仅限合约)';
COMMENT ON COLUMN settlements.entry_price IS '平均开仓价';
COMMENT ON COLUMN settlements.exit_price IS '平均平仓价';
COMMENT ON COLUMN settlements.quantity IS '结算头寸数量';
COMMENT ON COLUMN settlements.roi IS '收益率百分比';
COMMENT ON COLUMN settlements.opened_at IS '逻辑开仓时间';
COMMENT ON COLUMN settlements.closed_at IS '逻辑平仓时间';
COMMENT ON COLUMN settlements.duration_seconds IS '持仓总时长 (秒)';
COMMENT ON COLUMN settlements.metadata IS '扩展元数据 [JSON: leverage, liq_price]';
COMMENT ON COLUMN settlements.created_at IS '记录生成时间';

-- ----------------------------
-- 初始数据初始化 (Baseline Data)
-- ----------------------------

INSERT INTO permissions (code, name, module, description) VALUES
('dashboard.view', '查看仪表盘', 'CORE', '查看系统运行概览'),
('monitor.pnl', '盈亏分析', 'MONITOR', '实时查看权益与收益率曲线'),
('monitor.gateway', '网关监控', 'MONITOR', '监控交易所延迟与限流状态'),
('ai.news_stream', '市场情报', 'AI', '查看 AI 情绪得分与新闻流'),
('strategy.manage', '策略管理', 'TRADE', '启动/停止/参数配置交易策略'),
('trade.manual', '手动终端', 'TRADE', '紧急下单、撤单与一键平仓'),
('order.view', '历史订单', 'TRADE', '查询所有成交记录与流水'),
('risk.config', '风控配置', 'RISK', '调整 L1/L2 级熔断阈值'),
('account.manage', '资产管理', 'ASSET', '绑定 API Key 与查看余额'),
('system.users', '账号管理', 'SYSTEM', '管理系统用户与 Passkey'),
('audit.view', '审计查看', 'SYSTEM', '查看全局敏感操作日志')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role, permission_code) 
SELECT 'ADMIN', code FROM permissions ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_code) VALUES
('OPERATOR', 'dashboard.view'), ('OPERATOR', 'monitor.pnl'), ('OPERATOR', 'ai.news_stream'),
('OPERATOR', 'strategy.manage'), ('OPERATOR', 'trade.manual'), ('OPERATOR', 'order.view'),
('VIEWER', 'dashboard.view'), ('VIEWER', 'monitor.pnl'), ('VIEWER', 'order.view')
ON CONFLICT DO NOTHING;

INSERT INTO menus (id, name, icon, path, permission_code, sort_order) VALUES
(1, '仪表盘', 'layout-dashboard', '/dashboard', 'dashboard.view', 1),
(2, '市场监控', 'activity', NULL, NULL, 2),
(3, '执行中心', 'zap', NULL, NULL, 3),
(4, '风控资产', 'shield-check', NULL, NULL, 4),
(5, '系统设置', 'settings', NULL, NULL, 5);

INSERT INTO menus (parent_id, name, icon, path, permission_code, sort_order) VALUES
(2, '实时净值', 'trending-up', '/monitor/pnl', 'monitor.pnl', 1),
(2, 'AI 情绪流', 'brain-circuit', '/monitor/ai', 'ai.news_stream', 2),
(2, '网关状态', 'server', '/monitor/gateway', 'monitor.gateway', 3),
(3, '策略控制', 'play-circle', '/trade/strategies', 'strategy.manage', 1),
(3, '历史订单', 'clipboard-list', '/trade/orders', 'order.view', 2),
(3, '手动终端', 'terminal', '/trade/manual', 'trade.manual', 3),
(4, '风控配置', 'sliders', '/risk/config', 'risk.config', 1),
(4, 'API Key 管理', 'key', '/account/api-keys', 'account.manage', 2),
(5, '用户管理', 'users', '/system/users', 'system.users', 1),
(5, '审计日志', 'file-text', '/system/logs', 'audit.view', 2);
