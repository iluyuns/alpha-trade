package model

import (
	"testing"
	"time"
)

func TestCandle_IsBullish(t *testing.T) {
	tests := []struct {
		name  string
		open  string
		close string
		want  bool
	}{
		{"bullish candle", "100", "110", true},
		{"bearish candle", "110", "100", false},
		{"doji (equal)", "100", "100", true}, // close >= open
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Candle{
				Open:  MustMoney(tt.open),
				Close: MustMoney(tt.close),
			}
			if got := c.IsBullish(); got != tt.want {
				t.Errorf("IsBullish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCandle_IsBearish(t *testing.T) {
	tests := []struct {
		name  string
		open  string
		close string
		want  bool
	}{
		{"bearish candle", "110", "100", true},
		{"bullish candle", "100", "110", false},
		{"doji (equal)", "100", "100", false}, // close >= open
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Candle{
				Open:  MustMoney(tt.open),
				Close: MustMoney(tt.close),
			}
			if got := c.IsBearish(); got != tt.want {
				t.Errorf("IsBearish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCandle_Body(t *testing.T) {
	tests := []struct {
		name  string
		open  string
		close string
		want  string
	}{
		{"bullish body", "100", "110", "10"},
		{"bearish body", "110", "100", "10"},
		{"doji", "100", "100", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Candle{
				Open:  MustMoney(tt.open),
				Close: MustMoney(tt.close),
			}
			got := c.Body()
			want := MustMoney(tt.want)
			if !got.EQ(want) {
				t.Errorf("Body() = %s, want %s", got.String(), want.String())
			}
		})
	}
}

func TestCandle_Range(t *testing.T) {
	tests := []struct {
		name string
		high string
		low  string
		want string
	}{
		{"normal range", "120", "80", "40"},
		{"small range", "101", "100", "1"},
		{"zero range", "100", "100", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Candle{
				High: MustMoney(tt.high),
				Low:  MustMoney(tt.low),
			}
			got := c.Range()
			want := MustMoney(tt.want)
			if !got.EQ(want) {
				t.Errorf("Range() = %s, want %s", got.String(), want.String())
			}
		})
	}
}

func TestCandle_UpperShadow(t *testing.T) {
	tests := []struct {
		name  string
		open  string
		high  string
		close string
		want  string
	}{
		{"bullish upper shadow", "100", "120", "110", "10"}, // high - close
		{"bearish upper shadow", "110", "120", "100", "10"}, // high - open
		{"no upper shadow (bullish)", "100", "110", "110", "0"},
		{"no upper shadow (bearish)", "110", "110", "100", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Candle{
				Open:  MustMoney(tt.open),
				High:  MustMoney(tt.high),
				Close: MustMoney(tt.close),
			}
			got := c.UpperShadow()
			want := MustMoney(tt.want)
			if !got.EQ(want) {
				t.Errorf("UpperShadow() = %s, want %s", got.String(), want.String())
			}
		})
	}
}

func TestCandle_LowerShadow(t *testing.T) {
	tests := []struct {
		name  string
		open  string
		low   string
		close string
		want  string
	}{
		{"bullish lower shadow", "100", "90", "110", "10"}, // open - low
		{"bearish lower shadow", "110", "90", "100", "10"}, // close - low
		{"no lower shadow (bullish)", "100", "100", "110", "0"},
		{"no lower shadow (bearish)", "110", "100", "100", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Candle{
				Open:  MustMoney(tt.open),
				Low:   MustMoney(tt.low),
				Close: MustMoney(tt.close),
			}
			got := c.LowerShadow()
			want := MustMoney(tt.want)
			if !got.EQ(want) {
				t.Errorf("LowerShadow() = %s, want %s", got.String(), want.String())
			}
		})
	}
}

func TestCandle_CompleteAnalysis(t *testing.T) {
	// 完整的看涨K线：开100，高130，低95，收120
	bullish := &Candle{
		Symbol:    "BTCUSDT",
		Interval:  "1m",
		Open:      MustMoney("100"),
		High:      MustMoney("130"),
		Low:       MustMoney("95"),
		Close:     MustMoney("120"),
		Volume:    MustMoney("1000"),
		OpenTime:  time.Now(),
		CloseTime: time.Now().Add(time.Minute),
	}

	if !bullish.IsBullish() {
		t.Error("expected bullish candle")
	}
	if bullish.IsBearish() {
		t.Error("should not be bearish")
	}

	// 实体：120 - 100 = 20
	if body := bullish.Body(); !body.EQ(MustMoney("20")) {
		t.Errorf("expected body 20, got %s", body.String())
	}

	// 振幅：130 - 95 = 35
	if r := bullish.Range(); !r.EQ(MustMoney("35")) {
		t.Errorf("expected range 35, got %s", r.String())
	}

	// 上影线：130 - 120 = 10 (看涨时：high - close)
	if upper := bullish.UpperShadow(); !upper.EQ(MustMoney("10")) {
		t.Errorf("expected upper shadow 10, got %s", upper.String())
	}

	// 下影线：100 - 95 = 5 (看涨时：open - low)
	if lower := bullish.LowerShadow(); !lower.EQ(MustMoney("5")) {
		t.Errorf("expected lower shadow 5, got %s", lower.String())
	}
}

func TestCandle_BearishAnalysis(t *testing.T) {
	// 完整的看跌K线：开120，高130，低95，收100
	bearish := &Candle{
		Symbol:   "BTCUSDT",
		Interval: "1m",
		Open:     MustMoney("120"),
		High:     MustMoney("130"),
		Low:      MustMoney("95"),
		Close:    MustMoney("100"),
	}

	if bearish.IsBullish() {
		t.Error("should not be bullish")
	}
	if !bearish.IsBearish() {
		t.Error("expected bearish candle")
	}

	// 实体：|100 - 120| = 20
	if body := bearish.Body(); !body.EQ(MustMoney("20")) {
		t.Errorf("expected body 20, got %s", body.String())
	}

	// 上影线：130 - 120 = 10 (看跌时：high - open)
	if upper := bearish.UpperShadow(); !upper.EQ(MustMoney("10")) {
		t.Errorf("expected upper shadow 10, got %s", upper.String())
	}

	// 下影线：100 - 95 = 5 (看跌时：close - low)
	if lower := bearish.LowerShadow(); !lower.EQ(MustMoney("5")) {
		t.Errorf("expected lower shadow 5, got %s", lower.String())
	}
}

func TestTick_Basic(t *testing.T) {
	tick := &Tick{
		Symbol:    "BTCUSDT",
		Price:     MustMoney("50000"),
		Volume:    MustMoney("0.5"),
		EventTime: time.Now(),
		RecvTime:  time.Now(),
	}

	if tick.Symbol != "BTCUSDT" {
		t.Errorf("expected symbol BTCUSDT, got %s", tick.Symbol)
	}
	if !tick.Price.EQ(MustMoney("50000")) {
		t.Errorf("expected price 50000, got %s", tick.Price.String())
	}
}
