package model

import (
	"testing"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func TestPortfolio_TableName(t *testing.T) {
	p := Portfolio{}
	if p.TableName() != "portfolios" {
		t.Errorf("TableName() = %v, want %v", p.TableName(), "portfolios")
	}
}

func TestPortfolio_BeforeSave_CalculatesMarketValue(t *testing.T) {
	// 市值 = 持仓数量 * 当前价格
	quantity := decimal.NewFromInt(100)
	currentPrice := decimal.NewFromInt(50)
	avgCost := decimal.NewFromInt(40)

	portfolio := &Portfolio{
		TotalQuantity:     quantity,
		AverageCost:       avgCost,
		CurrentPrice:      &currentPrice,
	}

	portfolio.BeforeSave(nil)

	expectedMarketValue := decimal.NewFromInt(5000) // 100 * 50
	if !portfolio.MarketValue.Equal(expectedMarketValue) {
		t.Errorf("MarketValue = %v, want %v", portfolio.MarketValue, expectedMarketValue)
	}
}

func TestPortfolio_BeforeSave_CalculatesProfitLoss(t *testing.T) {
	// 盈亏 = (当前价格 - 成本价) * 持仓数量
	quantity := decimal.NewFromInt(100)
	currentPrice := decimal.NewFromInt(50)
	avgCost := decimal.NewFromInt(40)

	portfolio := &Portfolio{
		TotalQuantity:     quantity,
		AverageCost:       avgCost,
		CurrentPrice:      &currentPrice,
	}

	portfolio.BeforeSave(nil)

	expectedProfitLoss := decimal.NewFromInt(1000) // (50 - 40) * 100
	if !portfolio.ProfitLoss.Equal(expectedProfitLoss) {
		t.Errorf("ProfitLoss = %v, want %v", portfolio.ProfitLoss, expectedProfitLoss)
	}
}

func TestPortfolio_BeforeSave_CalculatesProfitLossPercent(t *testing.T) {
	// 盈亏百分比 = (当前价格 - 成本价) / 成本价 * 100
	quantity := decimal.NewFromInt(100)
	currentPrice := decimal.NewFromInt(50)
	avgCost := decimal.NewFromInt(40)

	portfolio := &Portfolio{
		TotalQuantity:     quantity,
		AverageCost:       avgCost,
		CurrentPrice:      &currentPrice,
	}

	portfolio.BeforeSave(nil)

	expectedPercent := decimal.NewFromInt(25) // (50 - 40) / 40 * 100 = 25%
	if !portfolio.ProfitLossPercent.Equal(expectedPercent) {
		t.Errorf("ProfitLossPercent = %v, want %v", portfolio.ProfitLossPercent, expectedPercent)
	}
}

func TestPortfolio_BeforeSave_ZeroAverageCost(t *testing.T) {
	// 当成本为零时，盈亏百分比不应除以零
	quantity := decimal.NewFromInt(100)
	currentPrice := decimal.NewFromInt(50)
	avgCost := decimal.Zero

	portfolio := &Portfolio{
		TotalQuantity:     quantity,
		AverageCost:       avgCost,
		CurrentPrice:      &currentPrice,
	}

	portfolio.BeforeSave(nil)

	// 成本为零时，盈亏百分比应保持为零（避免除以零）
	if !portfolio.ProfitLossPercent.IsZero() {
		t.Errorf("ProfitLossPercent should be zero when AverageCost is zero, got %v", portfolio.ProfitLossPercent)
	}
}

func TestPortfolio_BeforeSave_NilCurrentPrice(t *testing.T) {
	// 当前价格为空时，不应计算市值和盈亏
	quantity := decimal.NewFromInt(100)
	avgCost := decimal.NewFromInt(40)

	portfolio := &Portfolio{
		TotalQuantity:     quantity,
		AverageCost:       avgCost,
		CurrentPrice:      nil,
	}

	portfolio.BeforeSave(nil)

	// CurrentPrice 为 nil 时，相关字段应保持默认值
	if !portfolio.MarketValue.IsZero() {
		t.Errorf("MarketValue should be zero when CurrentPrice is nil, got %v", portfolio.MarketValue)
	}
	if !portfolio.ProfitLoss.IsZero() {
		t.Errorf("ProfitLoss should be zero when CurrentPrice is nil, got %v", portfolio.ProfitLoss)
	}
	if !portfolio.ProfitLossPercent.IsZero() {
		t.Errorf("ProfitLossPercent should be zero when CurrentPrice is nil, got %v", portfolio.ProfitLossPercent)
	}
}

func TestPortfolio_BeforeSave_NegativeProfitLoss(t *testing.T) {
	// 亏损情况：当前价格低于成本价
	quantity := decimal.NewFromInt(100)
	currentPrice := decimal.NewFromInt(30) // 当前价格低于成本
	avgCost := decimal.NewFromInt(40)

	portfolio := &Portfolio{
		TotalQuantity:     quantity,
		AverageCost:       avgCost,
		CurrentPrice:      &currentPrice,
	}

	portfolio.BeforeSave(nil)

	// 亏损 = (30 - 40) * 100 = -1000
	expectedProfitLoss := decimal.NewFromInt(-1000)
	if !portfolio.ProfitLoss.Equal(expectedProfitLoss) {
		t.Errorf("ProfitLoss = %v, want %v", portfolio.ProfitLoss, expectedProfitLoss)
	}

	// 亏损百分比 = (30 - 40) / 40 * 100 = -25%
	expectedPercent := decimal.NewFromInt(-25)
	if !portfolio.ProfitLossPercent.Equal(expectedPercent) {
		t.Errorf("ProfitLossPercent = %v, want %v", portfolio.ProfitLossPercent, expectedPercent)
	}
}

func TestPortfolio_BeforeSave_DecimalPrecision(t *testing.T) {
	// 测试小数精度计算
	quantity := decimal.NewFromFloat(123.45)
	currentPrice := decimal.NewFromFloat(10.55)
	avgCost := decimal.NewFromFloat(8.33)

	portfolio := &Portfolio{
		TotalQuantity:     quantity,
		AverageCost:       avgCost,
		CurrentPrice:      &currentPrice,
	}

	portfolio.BeforeSave(nil)

	// 市值 = 123.45 * 10.55 = 1302.3975
	expectedMarketValue := decimal.NewFromFloat(1302.3975)
	if !portfolio.MarketValue.Equal(expectedMarketValue) {
		t.Errorf("MarketValue = %v, want %v", portfolio.MarketValue, expectedMarketValue)
	}

	// 盈亏 = (10.55 - 8.33) * 123.45 = 2.22 * 123.45 = 274.059
	// 注意: decimal 库计算 10.55 - 8.33 = 2.22, 然后 2.22 * 123.45 = 274.059
	expectedProfitLoss := decimal.NewFromFloat(274.059)
	if !portfolio.ProfitLoss.Equal(expectedProfitLoss) {
		t.Errorf("ProfitLoss = %v, want %v", portfolio.ProfitLoss, expectedProfitLoss)
	}
}

func TestPortfolio_BeforeSave_ReturnsNil(t *testing.T) {
	// BeforeSave 应该始终返回 nil
	portfolio := &Portfolio{}
	err := portfolio.BeforeSave(nil)
	if err != nil {
		t.Errorf("BeforeSave() should return nil, got %v", err)
	}
}

func TestPortfolio_BeforeSave_WithGormDB(t *testing.T) {
	// 测试传入 gorm.DB 参数（虽然不使用）
	quantity := decimal.NewFromInt(100)
	currentPrice := decimal.NewFromInt(50)
	avgCost := decimal.NewFromInt(40)

	portfolio := &Portfolio{
		TotalQuantity:     quantity,
		AverageCost:       avgCost,
		CurrentPrice:      &currentPrice,
	}

	// BeforeSave 应该能接受 nil 作为 *gorm.DB 参数
	err := portfolio.BeforeSave(nil)
	if err != nil {
		t.Errorf("BeforeSave(nil) should not error, got %v", err)
	}

	// 即使传入非 nil 的 DB，也应该正常工作（当前实现不使用它）
	var db *gorm.DB = nil // 保持 nil，因为我们没有真实数据库连接
	portfolio2 := &Portfolio{
		TotalQuantity:     quantity,
		AverageCost:       avgCost,
		CurrentPrice:      &currentPrice,
	}
	err = portfolio2.BeforeSave(db)
	if err != nil {
		t.Errorf("BeforeSave(db) should not error, got %v", err)
	}
}
