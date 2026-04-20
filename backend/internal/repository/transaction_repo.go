package repository

import (
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(transaction *model.Transaction) error
	BatchCreate(transactions []model.Transaction) error
	FindByID(id uint64) (*model.Transaction, error)
	FindByUserID(userID uint64, limit, offset int) ([]model.Transaction, int64, error)
	FindByAssetCode(userID uint64, assetCode string) ([]model.Transaction, error)
	FindByDateRange(userID uint64, startDate, endDate string) ([]model.Transaction, error)
	Update(transaction *model.Transaction) error
	Delete(id uint64) error
	GetTransactionStats(userID uint64) (*response.TransactionStats, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(transaction *model.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) BatchCreate(transactions []model.Transaction) error {
	return r.db.CreateInBatches(transactions, 100).Error
}

func (r *transactionRepository) FindByID(id uint64) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.db.First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) FindByUserID(userID uint64, limit, offset int) ([]model.Transaction, int64, error) {
	var transactions []model.Transaction
	var total int64

	db := r.db.Model(&model.Transaction{}).Where("user_id = ?", userID)
	db.Count(&total)

	err := db.Order("transaction_date DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, total, err
}

func (r *transactionRepository) FindByAssetCode(userID uint64, assetCode string) ([]model.Transaction, error) {
	var transactions []model.Transaction
	err := r.db.Where("user_id = ? AND asset_code = ?", userID, assetCode).
		Order("transaction_date DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) FindByDateRange(userID uint64, startDate, endDate string) ([]model.Transaction, error) {
	var transactions []model.Transaction
	err := r.db.Where("user_id = ? AND transaction_date BETWEEN ? AND ?", userID, startDate, endDate).
		Order("transaction_date DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) Update(transaction *model.Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *transactionRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Transaction{}, id).Error
}

func (r *transactionRepository) GetTransactionStats(userID uint64) (*response.TransactionStats, error) {
	var stats response.TransactionStats

	// 总交易次数
	r.db.Model(&model.Transaction{}).Where("user_id = ?", userID).Count(&stats.TotalTransactions)

	// 买入次数
	r.db.Model(&model.Transaction{}).Where("user_id = ? AND transaction_type = ?", userID, "buy").Count(&stats.BuyCount)

	// 卖出次数
	r.db.Model(&model.Transaction{}).Where("user_id = ? AND transaction_type = ?", userID, "sell").Count(&stats.SellCount)

	// 总投资额
	var totalInvestment decimal.Decimal
	r.db.Model(&model.Transaction{}).
		Where("user_id = ? AND transaction_type = ?", userID, "buy").
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalInvestment)
	stats.TotalInvestment = totalInvestment.String()

	// 总盈亏
	var totalProfit decimal.Decimal
	r.db.Model(&model.Transaction{}).
		Where("user_id = ?", userID).
		Where("profit IS NOT NULL").
		Select("COALESCE(SUM(profit), 0)").
		Scan(&totalProfit)
	stats.TotalProfit = totalProfit.String()

	return &stats, nil
}
