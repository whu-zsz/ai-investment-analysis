package service

import "github.com/shopspring/decimal"

func modelDecimalZero() decimal.Decimal {
	return decimal.Zero
}

func modelDecimalFromInt(value int) decimal.Decimal {
	return decimal.NewFromInt(int64(value))
}
