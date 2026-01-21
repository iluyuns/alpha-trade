package model

import (
	"testing"
)

func TestMoney_Arithmetic(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want string
		op   string
	}{
		{"Add positive", "10.5", "5.25", "15.75", "add"},
		{"Add negative", "10", "-5", "5", "add"},
		{"Sub positive", "10.5", "5.25", "5.25", "sub"},
		{"Sub negative", "10", "15", "-5", "sub"},
		{"Mul integers", "10", "5", "50", "mul"},
		{"Mul decimals", "2.5", "4", "10", "mul"},
		{"Div exact", "10", "2", "5", "div"},
		{"Div with decimal", "10", "3", "3.3333333333333333", "div"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MustMoney(tt.a)
			b := MustMoney(tt.b)
			var got Money

			switch tt.op {
			case "add":
				got = a.Add(b)
			case "sub":
				got = a.Sub(b)
			case "mul":
				got = a.Mul(b)
			case "div":
				got = a.Div(b)
			}

			if got.String() != tt.want {
				t.Errorf("got %s, want %s", got.String(), tt.want)
			}
		})
	}
}

func TestMoney_Comparison(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		lt   bool
		le   bool
		gt   bool
		ge   bool
		eq   bool
	}{
		{"equal", "10", "10", false, true, false, true, true},
		{"less than", "5", "10", true, true, false, false, false},
		{"greater than", "10", "5", false, false, true, true, false},
		{"negative equal", "-5", "-5", false, true, false, true, true},
		{"negative less", "-10", "-5", true, true, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MustMoney(tt.a)
			b := MustMoney(tt.b)

			if a.LT(b) != tt.lt {
				t.Errorf("LT: got %v, want %v", a.LT(b), tt.lt)
			}
			if a.LE(b) != tt.le {
				t.Errorf("LE: got %v, want %v", a.LE(b), tt.le)
			}
			if a.GT(b) != tt.gt {
				t.Errorf("GT: got %v, want %v", a.GT(b), tt.gt)
			}
			if a.GE(b) != tt.ge {
				t.Errorf("GE: got %v, want %v", a.GE(b), tt.ge)
			}
			if a.EQ(b) != tt.eq {
				t.Errorf("EQ: got %v, want %v", a.EQ(b), tt.eq)
			}
		})
	}
}

func TestMoney_ZeroAndSign(t *testing.T) {
	zero := Zero()
	pos := MustMoney("10.5")
	neg := MustMoney("-10.5")

	if !zero.IsZero() {
		t.Error("Zero() should be zero")
	}
	if !pos.IsPositive() {
		t.Error("10.5 should be positive")
	}
	if !neg.IsNegative() {
		t.Error("-10.5 should be negative")
	}

	if neg.Abs().String() != "10.5" {
		t.Errorf("Abs(-10.5) = %s, want 10.5", neg.Abs().String())
	}

	if pos.Neg().String() != "-10.5" {
		t.Errorf("Neg(10.5) = %s, want -10.5", pos.Neg().String())
	}
}

func TestNewMoney_InvalidInput(t *testing.T) {
	_, err := NewMoney("invalid")
	if err == nil {
		t.Error("expected error for invalid input")
	}

	_, err = NewMoney("")
	if err == nil {
		t.Error("expected error for empty string")
	}
}
