-- Risk States Table (风控状态表)
CREATE TABLE IF NOT EXISTS risk_states (
    id BIGSERIAL PRIMARY KEY,
    account_id VARCHAR(64) NOT NULL,
    symbol VARCHAR(32) NOT NULL DEFAULT '',
    
    -- 权益字段
    initial_equity DECIMAL(36, 18) NOT NULL,
    current_equity DECIMAL(36, 18) NOT NULL,
    peak_equity DECIMAL(36, 18) NOT NULL,
    
    -- 当日统计
    daily_pnl DECIMAL(36, 18) NOT NULL DEFAULT 0,
    consecutive_losses INT NOT NULL DEFAULT 0,
    
    -- 熔断器状态
    circuit_breaker_open BOOLEAN NOT NULL DEFAULT FALSE,
    circuit_breaker_until TIMESTAMP WITH TIME ZONE,
    
    -- 每日重置
    last_reset_date DATE,
    
    -- 完整状态快照 (JSON)
    state_data JSONB,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 唯一约束：一个账户+标的只能有一条状态记录
    UNIQUE(account_id, symbol)
);

COMMENT ON TABLE risk_states IS '风控状态表：存储账户实时风控状态';
COMMENT ON COLUMN risk_states.id IS '自增主键';
COMMENT ON COLUMN risk_states.account_id IS '账户ID（交易所账户或回测账户）';
COMMENT ON COLUMN risk_states.symbol IS '标的（空字符串表示账户全局状态）';
COMMENT ON COLUMN risk_states.initial_equity IS '初始净值';
COMMENT ON COLUMN risk_states.current_equity IS '当前净值';
COMMENT ON COLUMN risk_states.peak_equity IS '峰值净值（用于计算 MDD）';
COMMENT ON COLUMN risk_states.daily_pnl IS '当日盈亏';
COMMENT ON COLUMN risk_states.consecutive_losses IS '连续亏损次数';
COMMENT ON COLUMN risk_states.circuit_breaker_open IS '熔断器是否打开';
COMMENT ON COLUMN risk_states.circuit_breaker_until IS '熔断器关闭时间';
COMMENT ON COLUMN risk_states.last_reset_date IS '最后一次每日重置日期';
COMMENT ON COLUMN risk_states.state_data IS '完整状态 JSON 快照';
COMMENT ON COLUMN risk_states.created_at IS '创建时间';
COMMENT ON COLUMN risk_states.updated_at IS '更新时间';

-- 索引
CREATE INDEX idx_risk_states_account ON risk_states(account_id);
CREATE INDEX idx_risk_states_breaker ON risk_states(circuit_breaker_open, circuit_breaker_until)
    WHERE circuit_breaker_open = TRUE;
