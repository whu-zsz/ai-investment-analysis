package repository

import (
	"time"

	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockAnalysisMetricRepository interface {
	Upsert(metric *model.StockAnalysisMetric) error
	BatchUpsert(metrics []model.StockAnalysisMetric) error
	FindByUserPeriod(userID uint64, start, end time.Time, symbols []string) ([]model.StockAnalysisMetric, error)
	FindByUserSymbolPeriod(userID uint64, symbol string, start, end time.Time) (*model.StockAnalysisMetric, error)
}

type stockAnalysisMetricRepository struct {
	db *gorm.DB
}

func NewStockAnalysisMetricRepository(db *gorm.DB) StockAnalysisMetricRepository {
	return &stockAnalysisMetricRepository{db: db}
}

func (r *stockAnalysisMetricRepository) Upsert(metric *model.StockAnalysisMetric) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "symbol"}, {Name: "period_start"}, {Name: "period_end"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"task_id",
			"asset_name",
			"market",
			"trade_count",
			"buy_count",
			"sell_count",
			"buy_quantity",
			"sell_quantity",
			"buy_amount",
			"sell_amount",
			"net_quantity",
			"realized_profit",
			"realized_profit_rate",
			"ending_position_qty",
			"ending_avg_cost",
			"latest_price",
			"latest_market_value",
			"unrealized_profit",
			"unrealized_profit_rate",
			"total_profit",
			"total_profit_rate",
			"period_start_price",
			"period_end_price",
			"period_price_change_pct",
			"period_high_price",
			"period_low_price",
			"market_data_status",
			"source_type",
			"computed_at",
			"updated_at",
		}),
	}).Create(metric).Error
}

func (r *stockAnalysisMetricRepository) BatchUpsert(metrics []model.StockAnalysisMetric) error {
	if len(metrics) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "symbol"}, {Name: "period_start"}, {Name: "period_end"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"task_id",
			"asset_name",
			"market",
			"trade_count",
			"buy_count",
			"sell_count",
			"buy_quantity",
			"sell_quantity",
			"buy_amount",
			"sell_amount",
			"net_quantity",
			"realized_profit",
			"realized_profit_rate",
			"ending_position_qty",
			"ending_avg_cost",
			"latest_price",
			"latest_market_value",
			"unrealized_profit",
			"unrealized_profit_rate",
			"total_profit",
			"total_profit_rate",
			"period_start_price",
			"period_end_price",
			"period_price_change_pct",
			"period_high_price",
			"period_low_price",
			"market_data_status",
			"source_type",
			"computed_at",
			"updated_at",
		}),
	}).CreateInBatches(metrics, 100).Error
}

func (r *stockAnalysisMetricRepository) FindByUserPeriod(userID uint64, start, end time.Time, symbols []string) ([]model.StockAnalysisMetric, error) {
	var metrics []model.StockAnalysisMetric
	query := r.db.Where("user_id = ? AND period_start = ? AND period_end = ?", userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if len(symbols) > 0 {
		query = query.Where("symbol IN ?", symbols)
	}
	err := query.Order("symbol ASC").Find(&metrics).Error
	return metrics, err
}

func (r *stockAnalysisMetricRepository) FindByUserSymbolPeriod(userID uint64, symbol string, start, end time.Time) (*model.StockAnalysisMetric, error) {
	var metric model.StockAnalysisMetric
	err := r.db.Where("user_id = ? AND symbol = ? AND period_start = ? AND period_end = ?", userID, symbol, start.Format("2006-01-02"), end.Format("2006-01-02")).First(&metric).Error
	if err != nil {
		return nil, err
	}
	return &metric, nil
}
