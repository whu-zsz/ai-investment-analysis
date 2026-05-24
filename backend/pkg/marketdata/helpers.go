package marketdata

import (
	"math"
	"strings"
)

func DefaultName(symbol string) string {
	switch strings.ToUpper(symbol) {
	case "000001.SH", "000001.SZ":
		return "上证指数"
	case "399001.SZ":
		return "深证成指"
	case "399006.SZ":
		return "创业板指"
	case "000300.SH", "000300.SZ":
		return "沪深300"
	default:
		return symbol
	}
}

func Round(value float64) float64 {
	return math.Round(value*10000) / 10000
}
