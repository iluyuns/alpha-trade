package model

import "time"

// Tick 行情快照（逐笔/最新价）
type Tick struct {
	Symbol    string
	Price     Money
	Volume    Money     // 当前成交量
	EventTime time.Time // 事件发生时间（回测用此时间）
	RecvTime  time.Time // 系统接收时间
}

// Candle K线数据
type Candle struct {
	Symbol   string
	Interval string // "1m", "5m", "1h", "1d"

	Open   Money
	High   Money
	Low    Money
	Close  Money
	Volume Money

	OpenTime  time.Time // K线开盘时间（事件时间）
	CloseTime time.Time // K线收盘时间
	RecvTime  time.Time // 系统接收时间
}

// IsBullish 是否阳线
func (c *Candle) IsBullish() bool {
	return c.Close.GE(c.Open)
}

// IsBearish 是否阴线
func (c *Candle) IsBearish() bool {
	return c.Close.LT(c.Open)
}

// Body K线实体大小
func (c *Candle) Body() Money {
	return c.Close.Sub(c.Open).Abs()
}

// Range K线振幅
func (c *Candle) Range() Money {
	return c.High.Sub(c.Low)
}

// UpperShadow 上影线
func (c *Candle) UpperShadow() Money {
	if c.IsBullish() {
		return c.High.Sub(c.Close)
	}
	return c.High.Sub(c.Open)
}

// LowerShadow 下影线
func (c *Candle) LowerShadow() Money {
	if c.IsBullish() {
		return c.Open.Sub(c.Low)
	}
	return c.Close.Sub(c.Low)
}
