package service

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestModelDecimalZero 测试获取零值
func TestModelDecimalZero(t *testing.T) {
	result := modelDecimalZero()
	if !result.IsZero() {
		t.Error("modelDecimalZero() should return zero")
	}
}

// TestModelDecimalFromInt 测试从整数创建
func TestModelDecimalFromInt(t *testing.T) {
	tests := []struct {
		input    int
		expected int64
	}{
		{0, 0},
		{1, 1},
		{100, 100},
		{-1, -1},
		{-100, -100},
	}

	for _, tt := range tests {
		result := modelDecimalFromInt(tt.input)
		if !result.Equal(decimal.NewFromInt(tt.expected)) {
			t.Errorf("modelDecimalFromInt(%d) = %v, want %d", tt.input, result, tt.expected)
		}
	}
}

// TestModelDecimalFromInt_LargeValue 测试大整数
func TestModelDecimalFromInt_LargeValue(t *testing.T) {
	largeValue := 1000000000
	result := modelDecimalFromInt(largeValue)
	if !result.Equal(decimal.NewFromInt(int64(largeValue))) {
		t.Errorf("modelDecimalFromInt(%d) = %v, want %d", largeValue, result, largeValue)
	}
}

// TestModelDecimalFromInt_Arithmetic 测试算术运算
func TestModelDecimalFromInt_Arithmetic(t *testing.T) {
	a := modelDecimalFromInt(100)
	b := modelDecimalFromInt(50)

	// 加法
	sum := a.Add(b)
	if !sum.Equal(decimal.NewFromInt(150)) {
		t.Errorf("100 + 50 = %v, want 150", sum)
	}

	// 减法
	diff := a.Sub(b)
	if !diff.Equal(decimal.NewFromInt(50)) {
		t.Errorf("100 - 50 = %v, want 50", diff)
	}

	// 乘法
	product := a.Mul(b)
	if !product.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("100 * 50 = %v, want 5000", product)
	}

	// 除法
	quotient := a.Div(b)
	if !quotient.Equal(decimal.NewFromInt(2)) {
		t.Errorf("100 / 50 = %v, want 2", quotient)
	}
}

// TestModelDecimalZero_Comparisons 测试比较操作
func TestModelDecimalZero_Comparisons(t *testing.T) {
	zero := modelDecimalZero()
	pos := decimal.NewFromInt(1)
	neg := decimal.NewFromInt(-1)

	if !zero.IsZero() {
		t.Error("Zero should be zero")
	}

	if !zero.LessThan(pos) {
		t.Error("Zero should be less than positive")
	}

	if !zero.GreaterThan(neg) {
		t.Error("Zero should be greater than negative")
	}
}
