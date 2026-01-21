package model

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Money 封装金额/数量，基于 decimal.Decimal 保证精度
type Money struct {
	v decimal.Decimal
}

// NewMoney 创建 Money 实例
func NewMoney(value string) (Money, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return Money{}, fmt.Errorf("invalid money value: %w", err)
	}
	return Money{v: d}, nil
}

// MustMoney 创建 Money 实例，panic on error
func MustMoney(value string) Money {
	m, err := NewMoney(value)
	if err != nil {
		panic(err)
	}
	return m
}

// NewMoneyFromFloat 从 float64 创建（仅用于策略层信号计算，不进入会计）
func NewMoneyFromFloat(f float64) Money {
	return Money{v: decimal.NewFromFloat(f)}
}

// NewMoneyFromInt 从 int64 创建
func NewMoneyFromInt(i int64) Money {
	return Money{v: decimal.NewFromInt(i)}
}

// Zero 零值
func Zero() Money {
	return Money{v: decimal.Zero}
}

// Add 加法
func (m Money) Add(other Money) Money {
	return Money{v: m.v.Add(other.v)}
}

// Sub 减法
func (m Money) Sub(other Money) Money {
	return Money{v: m.v.Sub(other.v)}
}

// Mul 乘法
func (m Money) Mul(other Money) Money {
	return Money{v: m.v.Mul(other.v)}
}

// Div 除法
func (m Money) Div(other Money) Money {
	return Money{v: m.v.Div(other.v)}
}

// LT 小于
func (m Money) LT(other Money) bool {
	return m.v.LessThan(other.v)
}

// LE 小于等于
func (m Money) LE(other Money) bool {
	return m.v.LessThanOrEqual(other.v)
}

// GT 大于
func (m Money) GT(other Money) bool {
	return m.v.GreaterThan(other.v)
}

// GE 大于等于
func (m Money) GE(other Money) bool {
	return m.v.GreaterThanOrEqual(other.v)
}

// EQ 等于
func (m Money) EQ(other Money) bool {
	return m.v.Equal(other.v)
}

// IsZero 是否为零
func (m Money) IsZero() bool {
	return m.v.IsZero()
}

// IsPositive 是否为正
func (m Money) IsPositive() bool {
	return m.v.IsPositive()
}

// IsNegative 是否为负
func (m Money) IsNegative() bool {
	return m.v.IsNegative()
}

// Abs 绝对值
func (m Money) Abs() Money {
	return Money{v: m.v.Abs()}
}

// Neg 取反
func (m Money) Neg() Money {
	return Money{v: m.v.Neg()}
}

// String 格式化为字符串
func (m Money) String() string {
	return m.v.String()
}

// Float64 转换为 float64（精度损失，仅用于展示/策略信号）
func (m Money) Float64() float64 {
	f, _ := m.v.Float64()
	return f
}

// Decimal 获取底层 decimal.Decimal（内部使用）
func (m Money) Decimal() decimal.Decimal {
	return m.v
}
