package model

import (
	"testing"
	"time"
)

func TestOrder_IsFilled(t *testing.T) {
	tests := []struct {
		name   string
		status OrderStatus
		want   bool
	}{
		{"filled order", OrderStatusFilled, true},
		{"pending order", OrderStatusPending, false},
		{"submitted order", OrderStatusSubmitted, false},
		{"partial filled", OrderStatusPartialFilled, false},
		{"cancelled order", OrderStatusCancelled, false},
		{"rejected order", OrderStatusRejected, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			if got := o.IsFilled(); got != tt.want {
				t.Errorf("IsFilled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrder_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status OrderStatus
		want   bool
	}{
		{"pending", OrderStatusPending, true},
		{"submitted", OrderStatusSubmitted, true},
		{"partial filled", OrderStatusPartialFilled, true},
		{"filled", OrderStatusFilled, false},
		{"cancelled", OrderStatusCancelled, false},
		{"rejected", OrderStatusRejected, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			if got := o.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrder_IsClosed(t *testing.T) {
	tests := []struct {
		name   string
		status OrderStatus
		want   bool
	}{
		{"filled", OrderStatusFilled, true},
		{"cancelled", OrderStatusCancelled, true},
		{"rejected", OrderStatusRejected, true},
		{"pending", OrderStatusPending, false},
		{"submitted", OrderStatusSubmitted, false},
		{"partial filled", OrderStatusPartialFilled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			if got := o.IsClosed(); got != tt.want {
				t.Errorf("IsClosed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrder_FilledPercent(t *testing.T) {
	tests := []struct {
		name     string
		quantity string
		filled   string
		want     string
	}{
		{"fully filled", "10", "10", "1"},
		{"half filled", "10", "5", "0.5"},
		{"not filled", "10", "0", "0"},
		{"zero quantity", "0", "0", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{
				Quantity: MustMoney(tt.quantity),
				Filled:   MustMoney(tt.filled),
			}
			got := o.FilledPercent()
			want := MustMoney(tt.want)
			if !got.EQ(want) {
				t.Errorf("FilledPercent() = %s, want %s", got.String(), want.String())
			}
		})
	}
}

func TestOrder_RemainingQty(t *testing.T) {
	tests := []struct {
		name     string
		quantity string
		filled   string
		want     string
	}{
		{"no fill", "10", "0", "10"},
		{"partial fill", "10", "3", "7"},
		{"full fill", "10", "10", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{
				Quantity: MustMoney(tt.quantity),
				Filled:   MustMoney(tt.filled),
			}
			got := o.RemainingQty()
			want := MustMoney(tt.want)
			if !got.EQ(want) {
				t.Errorf("RemainingQty() = %s, want %s", got.String(), want.String())
			}
		})
	}
}

func TestOrderSide_String(t *testing.T) {
	tests := []struct {
		side OrderSide
		want string
	}{
		{OrderSideBuy, "BUY"},
		{OrderSideSell, "SELL"},
		{OrderSide(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.side.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrderType_String(t *testing.T) {
	tests := []struct {
		typ  OrderType
		want string
	}{
		{OrderTypeLimit, "LIMIT"},
		{OrderTypeMarket, "MARKET"},
		{OrderTypeIOC, "IOC"},
		{OrderTypeFOK, "FOK"},
		{OrderType(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrderStatus_String(t *testing.T) {
	tests := []struct {
		status OrderStatus
		want   string
	}{
		{OrderStatusPending, "PENDING"},
		{OrderStatusSubmitted, "SUBMITTED"},
		{OrderStatusPartialFilled, "PARTIAL_FILLED"},
		{OrderStatusFilled, "FILLED"},
		{OrderStatusCancelled, "CANCELLED"},
		{OrderStatusRejected, "REJECTED"},
		{OrderStatus(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarketType_String(t *testing.T) {
	tests := []struct {
		mt   MarketType
		want string
	}{
		{MarketTypeSpot, "SPOT"},
		{MarketTypeFuture, "FUTURE"},
		{MarketType(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mt.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrder_CompleteFlow(t *testing.T) {
	// 模拟完整订单流程
	order := &Order{
		ClientOrderID: "test-order-1",
		Symbol:        "BTCUSDT",
		MarketType:    MarketTypeSpot,
		Side:          OrderSideBuy,
		Type:          OrderTypeLimit,
		Price:         MustMoney("50000"),
		Quantity:      MustMoney("0.1"),
		Filled:        Zero(),
		Status:        OrderStatusPending,
		CreatedAt:     time.Now(),
	}

	// 初始状态
	if !order.IsActive() {
		t.Error("new order should be active")
	}
	if order.IsClosed() {
		t.Error("new order should not be closed")
	}

	// 部分成交
	order.Filled = MustMoney("0.05")
	order.Status = OrderStatusPartialFilled
	if !order.IsActive() {
		t.Error("partial filled order should be active")
	}
	if percent := order.FilledPercent(); !percent.EQ(MustMoney("0.5")) {
		t.Errorf("expected 50%% filled, got %s", percent.String())
	}
	if remaining := order.RemainingQty(); !remaining.EQ(MustMoney("0.05")) {
		t.Errorf("expected remaining 0.05, got %s", remaining.String())
	}

	// 完全成交
	order.Filled = order.Quantity
	order.Status = OrderStatusFilled
	if order.IsActive() {
		t.Error("filled order should not be active")
	}
	if !order.IsClosed() {
		t.Error("filled order should be closed")
	}
	if !order.IsFilled() {
		t.Error("order should be filled")
	}
}
