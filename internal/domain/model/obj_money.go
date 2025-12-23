package model

import (
	"github.com/shopspring/decimal"
)

// Money 封装 decimal.Decimal 以统一处理金额和数量
type Money = decimal.Decimal

var (
	Zero = decimal.NewFromInt(0)
)

// ToFloat64 辅助方法，用于技术指标计算
func ToFloat64(d Money) float64 {
	f, _ := d.Float64()
	return f
}

// FromFloat64 辅助方法
func FromFloat64(f float64) Money {
	return decimal.NewFromFloat(f)
}

