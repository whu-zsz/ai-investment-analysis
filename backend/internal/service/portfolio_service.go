package service

import (
	"errors"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"

	"github.com/shopspring/decimal"
)

type PortfolioService interface {
	GetPortfolios(userID uint64) ([]model.Portfolio, error)
	UpdatePortfolioFromTransaction(userID uint64, transaction *model.Transaction) error
	RecalculatePortfolio(userID uint64, assetCode string) error
}

type portfolioService struct {
	portfolioRepo   repository.PortfolioRepository
	transactionRepo repository.TransactionRepository
}

func NewPortfolioService(portfolioRepo repository.PortfolioRepository, transactionRepo repository.TransactionRepository) PortfolioService {
	return &portfolioService{
		portfolioRepo:   portfolioRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *portfolioService) GetPortfolios(userID uint64) ([]model.Portfolio, error) {
	return s.portfolioRepo.FindByUserID(userID)
}

func (s *portfolioService) UpdatePortfolioFromTransaction(userID uint64, transaction *model.Transaction) error {
	// 查找该资产的持仓记录
	portfolio, err := s.portfolioRepo.FindByUserAndAsset(userID, transaction.AssetCode)

	if err != nil {
		// 持仓不存在，创建新持仓
		if transaction.TransactionType == "buy" {
			portfolio = &model.Portfolio{
				UserID:            userID,
				AssetCode:         transaction.AssetCode,
				AssetName:         transaction.AssetName,
				AssetType:         transaction.AssetType,
				TotalQuantity:     transaction.Quantity,
				AvailableQuantity: transaction.Quantity,
				AverageCost:       transaction.PricePerUnit,
			}
			return s.portfolioRepo.Create(portfolio)
		}
		return errors.New("cannot sell asset that you don't own")
	}

	// 根据交易类型更新持仓
	switch transaction.TransactionType {
	case "buy":
		// 买入：计算新的平均成本
		oldTotal := portfolio.TotalQuantity.Mul(portfolio.AverageCost)
		newTotal := transaction.Quantity.Mul(transaction.PricePerUnit)
		totalQuantity := portfolio.TotalQuantity.Add(transaction.Quantity)

		portfolio.AverageCost = oldTotal.Add(newTotal).Div(totalQuantity)
		portfolio.TotalQuantity = totalQuantity
		portfolio.AvailableQuantity = portfolio.AvailableQuantity.Add(transaction.Quantity)

		return s.portfolioRepo.Update(portfolio)

	case "sell":
		// 卖出：验证持仓数量
		if portfolio.AvailableQuantity.LessThan(transaction.Quantity) {
			return errors.New("insufficient available quantity")
		}

		portfolio.AvailableQuantity = portfolio.AvailableQuantity.Sub(transaction.Quantity)
		portfolio.TotalQuantity = portfolio.TotalQuantity.Sub(transaction.Quantity)

		// 如果持仓数量为0，删除记录
		if portfolio.TotalQuantity.IsZero() {
			return s.portfolioRepo.Delete(portfolio.ID)
		}

		return s.portfolioRepo.Update(portfolio)

	case "dividend":
		// 分红：不影响持仓数量
		return nil

	default:
		return errors.New("invalid transaction type")
	}
}

func (s *portfolioService) RecalculatePortfolio(userID uint64, assetCode string) error {
	// 获取该资产的所有交易记录
	transactions, err := s.transactionRepo.FindByAssetCode(userID, assetCode)
	if err != nil {
		return err
	}

	// 获取当前持仓（如果存在）
	portfolio, _ := s.portfolioRepo.FindByUserAndAsset(userID, assetCode)
	if len(transactions) == 0 {
		if portfolio != nil {
			return s.portfolioRepo.Delete(portfolio.ID)
		}
		return nil
	}

	// 重新计算持仓
	var totalQuantity decimal.Decimal
	var totalCost decimal.Decimal
	var availableQuantity decimal.Decimal
	assetName := transactions[0].AssetName
	assetType := transactions[0].AssetType
	var avgCost decimal.Decimal
	if portfolio != nil {
		avgCost = portfolio.AverageCost
	}

	for _, t := range transactions {
		if t.TransactionType == "buy" {
			totalCost = totalCost.Add(t.Quantity.Mul(t.PricePerUnit))
			totalQuantity = totalQuantity.Add(t.Quantity)
			availableQuantity = availableQuantity.Add(t.Quantity)
		} else if t.TransactionType == "sell" {
			totalQuantity = totalQuantity.Sub(t.Quantity)
			availableQuantity = availableQuantity.Sub(t.Quantity)
			if !avgCost.IsZero() {
				totalCost = totalCost.Sub(t.Quantity.Mul(avgCost))
			}
		}
	}

	if totalQuantity.IsZero() {
		// 删除持仓记录
		if portfolio != nil {
			return s.portfolioRepo.Delete(portfolio.ID)
		}
		return nil
	}

	// 计算平均成本
	averageCost := totalCost.Div(totalQuantity)

	// 更新或创建持仓
	if portfolio == nil {
		// 创建新持仓
		portfolio = &model.Portfolio{
			UserID:            userID,
			AssetCode:         assetCode,
			AssetName:         assetName,
			AssetType:         assetType,
			TotalQuantity:     totalQuantity,
			AvailableQuantity: availableQuantity,
			AverageCost:       averageCost,
		}
		return s.portfolioRepo.Create(portfolio)
	}

	// 更新持仓
	portfolio.AssetName = assetName
	portfolio.AssetType = assetType
	portfolio.TotalQuantity = totalQuantity
	portfolio.AvailableQuantity = availableQuantity
	portfolio.AverageCost = averageCost
	return s.portfolioRepo.Update(portfolio)
}
