package order

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// PostgresRepo PostgreSQL 订单仓储（实现 port.OrderRepo 接口）
type PostgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo 创建 PostgreSQL 订单仓储
func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}

// SaveOrder 保存订单（幂等）
func (r *PostgresRepo) SaveOrder(ctx context.Context, order *model.Order) error {
	query := `
		INSERT INTO orders (
			client_oid, order_id, exchange, symbol, side, type,
			price, quantity, filled_qty, status, 
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		ON CONFLICT (client_oid) DO UPDATE SET
			order_id = EXCLUDED.order_id,
			status = EXCLUDED.status,
			filled_qty = EXCLUDED.filled_qty,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		order.ClientOrderID,
		order.ExchangeID,
		getExchangeName(order.Symbol), // 简化处理，从 Symbol 推断
		order.Symbol,
		order.Side.String(),
		order.Type.String(),
		order.Price.String(),
		order.Quantity.String(),
		order.Filled.String(),
		order.Status.String(),
		order.CreatedAt,
		order.UpdatedAt,
	)

	return err
}

// GetOrder 根据 ClientOrderID 获取订单
func (r *PostgresRepo) GetOrder(ctx context.Context, clientOrderID string) (*model.Order, error) {
	query := `
		SELECT 
			client_oid, order_id, symbol, side, type,
			price, quantity, filled_qty, status,
			created_at, updated_at
		FROM orders
		WHERE client_oid = $1
	`

	var (
		clientOid, exchangeID, symbol, side, orderType, status string
		price, quantity, filled                                 string
		createdAt, updatedAt                                    time.Time
	)

	err := r.db.QueryRowContext(ctx, query, clientOrderID).Scan(
		&clientOid, &exchangeID, &symbol, &side, &orderType,
		&price, &quantity, &filled, &status,
		&createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found: %s", clientOrderID)
	}
	if err != nil {
		return nil, fmt.Errorf("get order failed: %w", err)
	}

	return &model.Order{
		ClientOrderID: clientOid,
		ExchangeID:    exchangeID,
		Symbol:        symbol,
		Side:          parseOrderSide(side),
		Type:          parseOrderType(orderType),
		Price:         model.MustMoney(price),
		Quantity:      model.MustMoney(quantity),
		Filled:        model.MustMoney(filled),
		Status:        parseOrderStatus(status),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

// GetOrderByExchangeID 根据交易所订单ID获取订单
func (r *PostgresRepo) GetOrderByExchangeID(ctx context.Context, exchangeID string) (*model.Order, error) {
	query := `
		SELECT 
			client_oid, order_id, symbol, side, type,
			price, quantity, filled_qty, status,
			created_at, updated_at
		FROM orders
		WHERE order_id = $1
	`

	var (
		clientOid, exchID, symbol, side, orderType, status string
		price, quantity, filled                            string
		createdAt, updatedAt                               time.Time
	)

	err := r.db.QueryRowContext(ctx, query, exchangeID).Scan(
		&clientOid, &exchID, &symbol, &side, &orderType,
		&price, &quantity, &filled, &status,
		&createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found: %s", exchangeID)
	}
	if err != nil {
		return nil, fmt.Errorf("get order by exchange id failed: %w", err)
	}

	return &model.Order{
		ClientOrderID: clientOid,
		ExchangeID:    exchID,
		Symbol:        symbol,
		Side:          parseOrderSide(side),
		Type:          parseOrderType(orderType),
		Price:         model.MustMoney(price),
		Quantity:      model.MustMoney(quantity),
		Filled:        model.MustMoney(filled),
		Status:        parseOrderStatus(status),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

// UpdateOrderStatus 原子更新订单状态
func (r *PostgresRepo) UpdateOrderStatus(ctx context.Context, clientOrderID string, status model.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $2, updated_at = $3
		WHERE client_oid = $1
	`

	result, err := r.db.ExecContext(ctx, query, clientOrderID, status.String(), time.Now())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("order not found: %s", clientOrderID)
	}

	return nil
}

// UpdateFilled 更新成交数量
func (r *PostgresRepo) UpdateFilled(ctx context.Context, clientOrderID string, filled model.Money) error {
	query := `
		UPDATE orders
		SET filled_qty = $2, updated_at = $3
		WHERE client_oid = $1
	`

	result, err := r.db.ExecContext(ctx, query, clientOrderID, filled.String(), time.Now())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("order not found: %s", clientOrderID)
	}

	return nil
}

// ListActiveOrders 列出所有活跃订单
func (r *PostgresRepo) ListActiveOrders(ctx context.Context) ([]*model.Order, error) {
	query := `
		SELECT 
			client_oid, order_id, symbol, side, type,
			price, quantity, filled_qty, status,
			created_at, updated_at
		FROM orders
		WHERE status IN ('PENDING', 'SUBMITTED', 'PARTIAL_FILLED')
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list active orders failed: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var (
			clientOid, exchangeID, symbol, side, orderType, status string
			price, quantity, filled                                 string
			createdAt, updatedAt                                    time.Time
		)

		err := rows.Scan(
			&clientOid, &exchangeID, &symbol, &side, &orderType,
			&price, &quantity, &filled, &status,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order failed: %w", err)
		}

		orders = append(orders, &model.Order{
			ClientOrderID: clientOid,
			ExchangeID:    exchangeID,
			Symbol:        symbol,
			Side:          parseOrderSide(side),
			Type:          parseOrderType(orderType),
			Price:         model.MustMoney(price),
			Quantity:      model.MustMoney(quantity),
			Filled:        model.MustMoney(filled),
			Status:        parseOrderStatus(status),
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		})
	}

	return orders, rows.Err()
}

// ListOrdersBySymbol 列出指定标的的订单
func (r *PostgresRepo) ListOrdersBySymbol(ctx context.Context, symbol string, limit int) ([]*model.Order, error) {
	query := `
		SELECT 
			client_oid, order_id, symbol, side, type,
			price, quantity, filled_qty, status,
			created_at, updated_at
		FROM orders
		WHERE symbol = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("list orders by symbol failed: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var (
			clientOid, exchangeID, sym, side, orderType, status string
			price, quantity, filled                             string
			createdAt, updatedAt                                time.Time
		)

		err := rows.Scan(
			&clientOid, &exchangeID, &sym, &side, &orderType,
			&price, &quantity, &filled, &status,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order failed: %w", err)
		}

		orders = append(orders, &model.Order{
			ClientOrderID: clientOid,
			ExchangeID:    exchangeID,
			Symbol:        sym,
			Side:          parseOrderSide(side),
			Type:          parseOrderType(orderType),
			Price:         model.MustMoney(price),
			Quantity:      model.MustMoney(quantity),
			Filled:        model.MustMoney(filled),
			Status:        parseOrderStatus(status),
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		})
	}

	return orders, rows.Err()
}

// 辅助函数：解析订单方向
func parseOrderSide(s string) model.OrderSide {
	switch s {
	case "BUY":
		return model.OrderSideBuy
	case "SELL":
		return model.OrderSideSell
	default:
		return 0
	}
}

// 辅助函数：解析订单类型
func parseOrderType(t string) model.OrderType {
	switch t {
	case "LIMIT":
		return model.OrderTypeLimit
	case "MARKET":
		return model.OrderTypeMarket
	case "IOC":
		return model.OrderTypeIOC
	case "FOK":
		return model.OrderTypeFOK
	default:
		return 0
	}
}

// 辅助函数：解析订单状态
func parseOrderStatus(s string) model.OrderStatus {
	switch s {
	case "PENDING":
		return model.OrderStatusPending
	case "SUBMITTED":
		return model.OrderStatusSubmitted
	case "PARTIAL_FILLED":
		return model.OrderStatusPartialFilled
	case "FILLED":
		return model.OrderStatusFilled
	case "CANCELLED":
		return model.OrderStatusCancelled
	case "REJECTED":
		return model.OrderStatusRejected
	default:
		return 0
	}
}

// 辅助函数：从 Symbol 推断交易所名称（简化实现）
func getExchangeName(symbol string) string {
	// TODO: 根据实际业务逻辑实现
	return "binance"
}
