package service

import (
	"errors"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"time"

	"github.com/shopspring/decimal"
)

type TransactionService interface {
	CreateTransaction(userID uint64, req *request.CreateTransactionRequest) error
	GetTransactions(userID uint64, page, pageSize int) (*response.TransactionListResponse, error)
	GetTransactionByID(userID uint64, id uint64) (*model.Transaction, error)
	DeleteTransaction(userID uint64, id uint64) error
	GetTransactionStats(userID uint64) (*response.TransactionStats, error)
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
	portfolioService PortfolioService
}

func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	portfolioService PortfolioService,
) TransactionService {
	return &transactionService{
		transactionRepo:  transactionRepo,
		portfolioService: portfolioService,
	}
}

func (s *transactionService) CreateTransaction(userID uint64, req *request.CreateTransactionRequest) error {
	// 解析日期
	transactionDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		return errors.New("invalid date format, use YYYY-MM-DD")
	}

	// 解析数量
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return errors.New("invalid quantity")
	}

	// 解析单价
	pricePerUnit, err := decimal.NewFromString(req.PricePerUnit)
	if err != nil {
		return errors.New("invalid price per unit")
	}

	// 解析手续费
	commission := decimal.Zero
	if req.Commission != "" {
		commission, err = decimal.NewFromString(req.Commission)
		if err != nil {
			return errors.New("invalid commission")
		}
	}

	// 计算总金额
	totalAmount := quantity.Mul(pricePerUnit).Add(commission)

	// 创建交易记录
	transaction := &model.Transaction{
		UserID:          userID,
		TransactionDate: transactionDate,
		TransactionType: req.TransactionType,
		AssetType:       req.AssetType,
		AssetCode:       req.AssetCode,
		AssetName:       req.AssetName,
		Quantity:        quantity,
		PricePerUnit:    pricePerUnit,
		TotalAmount:     totalAmount,
		Commission:      commission,
		Notes:           req.Notes,
	}

	// 插入交易记录
	if err := s.transactionRepo.Create(transaction); err != nil {
		return err
	}

	// 更新持仓
	if err := s.portfolioService.UpdatePortfolioFromTransaction(userID, transaction); err != nil {
		return err
	}

	return nil
}

func (s *transactionService) GetTransactions(userID uint64, page, pageSize int) (*response.TransactionListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	transactions, total, err := s.transactionRepo.FindByUserID(userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	var transactionResponses []response.TransactionResponse
	for _, t := range transactions {
		var profit *string
		if t.Profit != nil {
			profitStr := t.Profit.String()
			profit = &profitStr
		}

		transactionResponses = append(transactionResponses, response.TransactionResponse{
			ID:              t.ID,
			TransactionDate: t.TransactionDate.Format("2006-01-02"),
			TransactionType: t.TransactionType,
			AssetType:       t.AssetType,
			AssetCode:       t.AssetCode,
			AssetName:       t.AssetName,
			Quantity:        t.Quantity.String(),
			PricePerUnit:    t.PricePerUnit.String(),
			TotalAmount:     t.TotalAmount.String(),
			Commission:      t.Commission.String(),
			Profit:          profit,
			Notes:           t.Notes,
			CreatedAt:       t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &response.TransactionListResponse{
		Transactions: transactionResponses,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}, nil
}

func (s *transactionService) GetTransactionByID(userID uint64, id uint64) (*model.Transaction, error) {
	transaction, err := s.transactionRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 验证交易记录属于该用户
	if transaction.UserID != userID {
		return nil, errors.New("transaction not found")
	}

	return transaction, nil
}

func (s *transactionService) DeleteTransaction(userID uint64, id uint64) error {
	// 验证交易记录存在且属于该用户
	transaction, err := s.GetTransactionByID(userID, id)
	if err != nil {
		return err
	}

	// 删除交易记录
	if err := s.transactionRepo.Delete(id); err != nil {
		return err
	}

	// 重新计算持仓
	if err := s.portfolioService.RecalculatePortfolio(userID, transaction.AssetCode); err != nil {
		return err
	}

	return nil
}

func (s *transactionService) GetTransactionStats(userID uint64) (*response.TransactionStats, error) {
	return s.transactionRepo.GetTransactionStats(userID)
}
