package risk

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// PostgresRepo PostgreSQL 风控仓储（实现 port.RiskRepo 接口）
type PostgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo 创建 PostgreSQL 风控仓储
func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}

// LoadState 加载风控状态
func (r *PostgresRepo) LoadState(ctx context.Context, accountID, symbol string) (*model.RiskState, error) {
	query := `
		SELECT 
			initial_equity, current_equity, peak_equity, daily_pnl,
			consecutive_losses, circuit_breaker_open, circuit_breaker_until,
			last_reset_date, state_data
		FROM risk_states
		WHERE account_id = $1 AND symbol = $2
	`

	var (
		initialEquity, currentEquity, peakEquity, dailyPnl string
		consecutiveLosses                                  int
		circuitBreakerOpen                                 bool
		circuitBreakerUntil                                sql.NullTime
		lastResetDate                                      sql.NullTime
		stateData                                          sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, accountID, symbol).Scan(
		&initialEquity, &currentEquity, &peakEquity, &dailyPnl,
		&consecutiveLosses, &circuitBreakerOpen, &circuitBreakerUntil,
		&lastResetDate, &stateData,
	)

	if err == sql.ErrNoRows {
		// 记录不存在，返回新状态
		return model.NewRiskState(accountID, model.Zero()), nil
	}
	if err != nil {
		return nil, fmt.Errorf("load risk state failed: %w", err)
	}

	// 构建状态对象：列数据为主（UpdateEquity/RecordTrade 仅更新列）
	state := &model.RiskState{
		AccountID:          accountID,
		Symbol:             symbol,
		InitialEquity:      model.MustMoney(initialEquity),
		CurrentEquity:      model.MustMoney(currentEquity),
		PeakEquity:         model.MustMoney(peakEquity),
		DailyPnL:           model.MustMoney(dailyPnl),
		ConsecutiveLosses:  consecutiveLosses,
		CircuitBreakerOpen: circuitBreakerOpen,
	}

	if circuitBreakerUntil.Valid {
		state.CircuitBreakerUntil = circuitBreakerUntil.Time.Unix()
	}
	if lastResetDate.Valid {
		state.LastResetDate = lastResetDate.Time.Format("2006-01-02")
	}

	// 从 JSON 补充列中未存储的字段（PositionMap, MDD 等）
	if stateData.Valid && stateData.String != "" {
		var extra model.RiskState
		if err := json.Unmarshal([]byte(stateData.String), &extra); err == nil {
			state.PositionMap = extra.PositionMap
			state.TotalExposure = extra.TotalExposure
			state.MDD = extra.MDD
			state.MDDPercent = extra.MDDPercent
			state.DailyTradeCount = extra.DailyTradeCount
			state.DailyResetTime = extra.DailyResetTime
			state.LastLossTime = extra.LastLossTime
			state.UpdatedAt = extra.UpdatedAt
		}
	}

	return state, nil
}

// SaveState 保存风控状态
func (r *PostgresRepo) SaveState(ctx context.Context, state *model.RiskState) error {
	// 序列化状态为 JSON
	stateData, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state data failed: %w", err)
	}

	// 使用 UPSERT (ON CONFLICT)
	query := `
		INSERT INTO risk_states (
			account_id, symbol, initial_equity, current_equity, peak_equity,
			daily_pnl, consecutive_losses, circuit_breaker_open,
			circuit_breaker_until, last_reset_date, state_data, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		ON CONFLICT (account_id, symbol) DO UPDATE SET
			initial_equity = EXCLUDED.initial_equity,
			current_equity = EXCLUDED.current_equity,
			peak_equity = EXCLUDED.peak_equity,
			daily_pnl = EXCLUDED.daily_pnl,
			consecutive_losses = EXCLUDED.consecutive_losses,
			circuit_breaker_open = EXCLUDED.circuit_breaker_open,
			circuit_breaker_until = EXCLUDED.circuit_breaker_until,
			last_reset_date = EXCLUDED.last_reset_date,
			state_data = EXCLUDED.state_data,
			updated_at = EXCLUDED.updated_at
	`

	var circuitBreakerUntil interface{}
	if state.CircuitBreakerUntil > 0 {
		circuitBreakerUntil = time.Unix(state.CircuitBreakerUntil, 0)
	}

	var lastResetDate interface{}
	if state.LastResetDate != "" {
		t, _ := time.Parse("2006-01-02", state.LastResetDate)
		lastResetDate = t
	}

	_, err = r.db.ExecContext(ctx, query,
		state.AccountID,
		state.Symbol,
		state.InitialEquity.String(),
		state.CurrentEquity.String(),
		state.PeakEquity.String(),
		state.DailyPnL.String(),
		state.ConsecutiveLosses,
		state.CircuitBreakerOpen,
		circuitBreakerUntil,
		lastResetDate,
		stateData,
		time.Now(),
	)

	return err
}

// UpdateEquity 原子更新净值，同时刷新 state_data JSON
func (r *PostgresRepo) UpdateEquity(ctx context.Context, accountID string, newEquity model.Money) error {
	query := `
		UPDATE risk_states
		SET 
			current_equity = $2,
			peak_equity = CASE 
				WHEN $2::DECIMAL > peak_equity THEN $2::DECIMAL 
				ELSE peak_equity 
			END,
			updated_at = $3
		WHERE account_id = $1 AND symbol = ''
	`

	_, err := r.db.ExecContext(ctx, query, accountID, newEquity.String(), time.Now())
	if err != nil {
		return err
	}

	return r.refreshStateData(ctx, accountID)
}

// RecordTrade 记录交易，同时刷新 state_data JSON
func (r *PostgresRepo) RecordTrade(ctx context.Context, accountID string, pnl model.Money) error {
	query := `
		UPDATE risk_states
		SET 
			daily_pnl = daily_pnl + $2::DECIMAL,
			consecutive_losses = CASE 
				WHEN $2::DECIMAL < 0 THEN consecutive_losses + 1
				WHEN $2::DECIMAL > 0 THEN 0
				ELSE consecutive_losses
			END,
			updated_at = $3
		WHERE account_id = $1 AND symbol = ''
	`

	_, err := r.db.ExecContext(ctx, query, accountID, pnl.String(), time.Now())
	if err != nil {
		return err
	}

	return r.refreshStateData(ctx, accountID)
}

// OpenCircuitBreaker 打开熔断器
func (r *PostgresRepo) OpenCircuitBreaker(ctx context.Context, accountID string, duration int64) error {
	until := time.Now().Add(time.Duration(duration) * time.Second)
	query := `
		UPDATE risk_states
		SET 
			circuit_breaker_open = true,
			circuit_breaker_until = $2,
			updated_at = $3
		WHERE account_id = $1 AND symbol = ''
	`

	_, err := r.db.ExecContext(ctx, query, accountID, until, time.Now())
	return err
}

// CloseCircuitBreaker 关闭熔断器
func (r *PostgresRepo) CloseCircuitBreaker(ctx context.Context, accountID string) error {
	query := `
		UPDATE risk_states
		SET 
			circuit_breaker_open = false,
			circuit_breaker_until = NULL,
			updated_at = $2
		WHERE account_id = $1 AND symbol = ''
	`

	_, err := r.db.ExecContext(ctx, query, accountID, time.Now())
	return err
}

// refreshStateData 重新序列化列数据到 state_data JSON（保持一致性）
func (r *PostgresRepo) refreshStateData(ctx context.Context, accountID string) error {
	state, err := r.LoadState(ctx, accountID, "")
	if err != nil {
		return nil // best-effort
	}
	data, err := json.Marshal(state)
	if err != nil {
		return nil
	}
	_, _ = r.db.ExecContext(ctx,
		`UPDATE risk_states SET state_data = $2 WHERE account_id = $1 AND symbol = ''`,
		accountID, data)
	return nil
}

// IsCircuitBreakerOpen 检查熔断器状态
func (r *PostgresRepo) IsCircuitBreakerOpen(ctx context.Context, accountID string) (bool, error) {
	query := `
		SELECT circuit_breaker_open, circuit_breaker_until
		FROM risk_states
		WHERE account_id = $1 AND symbol = ''
	`

	var (
		open  bool
		until sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, accountID).Scan(&open, &until)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if !open {
		return false, nil
	}

	// 检查是否过期
	if until.Valid && time.Now().After(until.Time) {
		// 自动关闭过期的熔断器
		_ = r.CloseCircuitBreaker(ctx, accountID)
		return false, nil
	}

	return true, nil
}
