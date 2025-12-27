package model

import (
	"github.com/shopspring/decimal"
)

// Money 封装 decimal.Decimal 以统一处理金额和数量
// 核心原则: 严禁在领域层使用 float64，必须全链路保持高精度。
type Money = decimal.Decimal

var (
	// Zero 零值常量
	Zero = decimal.NewFromInt(0)
)

// NewMoney 推荐的构造函数，强制使用字符串初始化以避免 float 精度问题
func NewMoney(val string) (Money, error) {
	return decimal.NewFromString(val)
}

// NewMoneyFromInt 从整数构建
func NewMoneyFromInt(val int64) Money {
	return decimal.NewFromInt(val)
}
