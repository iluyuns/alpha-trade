package risk

import (
	"context"
	"fmt"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// CheckPositionLimit 仓位限制规则
// 检查项：
// 1. 单标的仓位占比 <= MaxSinglePositionPercent
// 2. 总敞口占比 <= MaxTotalExposurePercent
// 3. 现金储备 >= MinCashReservePercent
// 4. 合约杠杆 <= MaxLeverage
// 5. 大额单强制降档（ProtectPrice 机制）
func (m *Manager) CheckPositionLimit(ctx context.Context, req *OrderContext, state *model.RiskState) DecisionDetail {
	// 计算订单名义价值
	orderNotional := calculateNotional(req)

	// 1. 单标的仓位限制
	if m.config.MaxSinglePositionPercent > 0 && state.CurrentEquity.IsPositive() {
		existingPosition := state.PositionMap[req.Symbol]
		newPosition := existingPosition.Add(orderNotional)
		positionPercent := newPosition.Div(state.CurrentEquity).Float64()

		if positionPercent > m.config.MaxSinglePositionPercent {
			// 计算建议数量
			maxAllowed := state.CurrentEquity.Mul(model.NewMoneyFromFloat(m.config.MaxSinglePositionPercent))
			availableNotional := maxAllowed.Sub(existingPosition)

			if availableNotional.LE(model.Zero()) {
				return NewBlock(
					fmt.Sprintf("single position limit exceeded for %s: %.2f%% > %.2f%%",
						req.Symbol, positionPercent*100, m.config.MaxSinglePositionPercent*100),
					"PositionLimit:SinglePosition",
				)
			}

			// 建议降档
			price := req.Price
			if price.IsZero() {
				price = req.CurrentPrice
			}
			suggestedQty := availableNotional.Div(price)
			return NewReduce(
				fmt.Sprintf("single position limit: reduce to %.2f%%", m.config.MaxSinglePositionPercent*100),
				"PositionLimit:SinglePosition",
				suggestedQty.String(),
				1, // 强制1x杠杆
			)
		}
	}

	// 2. 总敞口限制
	if m.config.MaxTotalExposurePercent > 0 && state.CurrentEquity.IsPositive() {
		newTotalExposure := state.TotalExposure.Add(orderNotional)
		exposurePercent := newTotalExposure.Div(state.CurrentEquity).Float64()

		if exposurePercent > m.config.MaxTotalExposurePercent {
			return NewBlock(
				fmt.Sprintf("total exposure limit exceeded: %.2f%% > %.2f%%",
					exposurePercent*100, m.config.MaxTotalExposurePercent*100),
				"PositionLimit:TotalExposure",
			)
		}
	}

	// 3. 现金储备检查
	if m.config.MinCashReservePercent > 0 && state.CurrentEquity.IsPositive() {
		requiredCash := state.CurrentEquity.Mul(model.NewMoneyFromFloat(m.config.MinCashReservePercent))
		availableCash := state.CurrentEquity.Sub(state.TotalExposure).Sub(orderNotional)

		if availableCash.LT(requiredCash) {
			return NewBlock(
				fmt.Sprintf("insufficient cash reserve: required %.2f%%, available %.2f%%",
					m.config.MinCashReservePercent*100,
					availableCash.Div(state.CurrentEquity).Float64()*100),
				"PositionLimit:CashReserve",
			)
		}
	}

	// 4. 合约杠杆限制
	if req.MarketType == model.MarketTypeFuture {
		if m.config.MaxLeverage > 0 && req.Leverage > m.config.MaxLeverage {
			return NewReduce(
				fmt.Sprintf("leverage %dx exceeds max %dx", req.Leverage, m.config.MaxLeverage),
				"PositionLimit:Leverage",
				req.Quantity.String(),
				m.config.MaxLeverage,
			)
		}

		// 5. 大额单强制1x
		if m.config.ForceLeverageOne && state.CurrentEquity.IsPositive() {
			orderSizePercent := orderNotional.Div(state.CurrentEquity).Float64()
			if orderSizePercent > m.config.LargeOrderThreshold && req.Leverage > 1 {
				return NewReduce(
					fmt.Sprintf("large order (%.2f%% of equity) requires 1x leverage", orderSizePercent*100),
					"PositionLimit:LargeOrder",
					req.Quantity.String(),
					1,
				)
			}
		}
	}

	return NewAllow()
}

// checkPositionLimit 内部调用（manager.go 中的短路链）
func (m *Manager) checkPositionLimit(ctx context.Context, req *OrderContext, state *model.RiskState) DecisionDetail {
	return m.CheckPositionLimit(ctx, req, state)
}

// calculateNotional 计算订单名义价值
func calculateNotional(req *OrderContext) model.Money {
	price := req.Price
	if price.IsZero() {
		// 市价单使用当前价格
		price = req.CurrentPrice
	}

	notional := price.Mul(req.Quantity)

	// 合约杠杆调整
	if req.MarketType == model.MarketTypeFuture && req.Leverage > 1 {
		notional = notional.Div(model.NewMoneyFromInt(int64(req.Leverage)))
	}

	return notional
}
