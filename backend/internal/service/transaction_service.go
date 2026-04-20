package service

import (
	"errors"
	"sort"
	"stock-analysis-backend/internal/dto/request"
	dtoResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type TransactionService interface {
	CreateTransaction(userID uint64, req *request.CreateTransactionRequest) error
	GetTransactions(userID uint64, page, pageSize int) (*dtoResponse.TransactionListResponse, error)
	GetTransactionByID(userID uint64, id uint64) (*model.Transaction, error)
	UpdateTransaction(userID uint64, id uint64, req *request.UpdateTransactionRequest) (*model.Transaction, error)
	DeleteTransaction(userID uint64, id uint64) error
	GetTransactionStats(userID uint64) (*dtoResponse.TransactionStats, error)
}

type transactionService struct {
	transactionRepo  repository.TransactionRepository
	portfolioService PortfolioService
}

type parsedTransactionInput struct {
	TransactionDate time.Time
	TransactionType string
	AssetType       string
	AssetCode       string
	AssetName       string
	Quantity        decimal.Decimal
	PricePerUnit    decimal.Decimal
	TotalAmount     decimal.Decimal
	Commission      decimal.Decimal
	Notes           *string
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
	input, err := parseCreateTransactionRequest(req)
	if err != nil {
		return err
	}

	transaction := &model.Transaction{
		UserID: userID,
	}
	applyParsedTransactionInput(transaction, input)

	if err := s.transactionRepo.Create(transaction); err != nil {
		return err
	}

	if err := s.portfolioService.UpdatePortfolioFromTransaction(userID, transaction); err != nil {
		return err
	}

	return nil
}

func (s *transactionService) GetTransactions(userID uint64, page, pageSize int) (*dtoResponse.TransactionListResponse, error) {
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

	transactionResponses := make([]dtoResponse.TransactionResponse, 0, len(transactions))
	for i := range transactions {
		transactionResponses = append(transactionResponses, dtoResponse.NewTransactionResponse(&transactions[i]))
	}

	return &dtoResponse.TransactionListResponse{
		Transactions: transactionResponses,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}, nil
}

func (s *transactionService) GetTransactionByID(userID uint64, id uint64) (*model.Transaction, error) {
	transaction, err := s.transactionRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	if transaction.UserID != userID {
		return nil, ErrTransactionNotFound
	}

	return transaction, nil
}

func (s *transactionService) UpdateTransaction(userID uint64, id uint64, req *request.UpdateTransactionRequest) (*model.Transaction, error) {
	existing, err := s.GetTransactionByID(userID, id)
	if err != nil {
		return nil, err
	}

	input, err := parseUpdateTransactionRequest(req)
	if err != nil {
		return nil, err
	}

	updated := *existing
	oldAssetCode := existing.AssetCode
	applyParsedTransactionInput(&updated, input)

	if err := s.validateUpdatedTransaction(userID, existing, &updated); err != nil {
		return nil, err
	}

	if err := s.transactionRepo.Update(&updated); err != nil {
		return nil, err
	}

	if oldAssetCode == updated.AssetCode {
		if err := s.portfolioService.RecalculatePortfolio(userID, updated.AssetCode); err != nil {
			return nil, err
		}
		return &updated, nil
	}

	if err := s.portfolioService.RecalculatePortfolio(userID, oldAssetCode); err != nil {
		return nil, err
	}
	if err := s.portfolioService.RecalculatePortfolio(userID, updated.AssetCode); err != nil {
		return nil, err
	}

	return &updated, nil
}

func (s *transactionService) DeleteTransaction(userID uint64, id uint64) error {
	transaction, err := s.GetTransactionByID(userID, id)
	if err != nil {
		return err
	}

	if err := s.transactionRepo.Delete(id); err != nil {
		return err
	}

	if err := s.portfolioService.RecalculatePortfolio(userID, transaction.AssetCode); err != nil {
		return err
	}

	return nil
}

func (s *transactionService) GetTransactionStats(userID uint64) (*dtoResponse.TransactionStats, error) {
	return s.transactionRepo.GetTransactionStats(userID)
}

func parseCreateTransactionRequest(req *request.CreateTransactionRequest) (*parsedTransactionInput, error) {
	return parseTransactionInput(
		req.TransactionDate,
		req.TransactionType,
		req.AssetType,
		req.AssetCode,
		req.AssetName,
		req.Quantity,
		req.PricePerUnit,
		req.Commission,
		req.Notes,
	)
}

func parseUpdateTransactionRequest(req *request.UpdateTransactionRequest) (*parsedTransactionInput, error) {
	return parseTransactionInput(
		req.TransactionDate,
		req.TransactionType,
		req.AssetType,
		req.AssetCode,
		req.AssetName,
		req.Quantity,
		req.PricePerUnit,
		req.Commission,
		req.Notes,
	)
}

func parseTransactionInput(
	transactionDateRaw string,
	transactionType string,
	assetType string,
	assetCode string,
	assetName string,
	quantityRaw string,
	pricePerUnitRaw string,
	commissionRaw string,
	notes *string,
) (*parsedTransactionInput, error) {
	transactionDate, err := time.Parse("2006-01-02", transactionDateRaw)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	quantity, err := decimal.NewFromString(quantityRaw)
	if err != nil {
		return nil, errors.New("invalid quantity")
	}

	pricePerUnit, err := decimal.NewFromString(pricePerUnitRaw)
	if err != nil {
		return nil, errors.New("invalid price per unit")
	}

	commission := decimal.Zero
	if commissionRaw != "" {
		commission, err = decimal.NewFromString(commissionRaw)
		if err != nil {
			return nil, errors.New("invalid commission")
		}
	}

	return &parsedTransactionInput{
		TransactionDate: transactionDate,
		TransactionType: transactionType,
		AssetType:       assetType,
		AssetCode:       assetCode,
		AssetName:       assetName,
		Quantity:        quantity,
		PricePerUnit:    pricePerUnit,
		TotalAmount:     quantity.Mul(pricePerUnit).Add(commission),
		Commission:      commission,
		Notes:           notes,
	}, nil
}

func applyParsedTransactionInput(transaction *model.Transaction, input *parsedTransactionInput) {
	transaction.TransactionDate = input.TransactionDate
	transaction.TransactionType = input.TransactionType
	transaction.AssetType = input.AssetType
	transaction.AssetCode = input.AssetCode
	transaction.AssetName = input.AssetName
	transaction.Quantity = input.Quantity
	transaction.PricePerUnit = input.PricePerUnit
	transaction.TotalAmount = input.TotalAmount
	transaction.Commission = input.Commission
	transaction.Notes = input.Notes
}

func (s *transactionService) validateUpdatedTransaction(userID uint64, existing *model.Transaction, updated *model.Transaction) error {
	if existing.AssetCode == updated.AssetCode {
		return s.validateAssetTransactions(userID, updated.AssetCode, existing.ID, updated)
	}

	if err := s.validateAssetTransactions(userID, existing.AssetCode, existing.ID, nil); err != nil {
		return err
	}

	return s.validateAssetTransactions(userID, updated.AssetCode, 0, updated)
}

func (s *transactionService) validateAssetTransactions(userID uint64, assetCode string, excludeID uint64, include *model.Transaction) error {
	transactions, err := s.transactionRepo.FindByAssetCode(userID, assetCode)
	if err != nil {
		return err
	}

	candidates := make([]model.Transaction, 0, len(transactions)+1)
	for _, transaction := range transactions {
		if excludeID != 0 && transaction.ID == excludeID {
			continue
		}
		candidates = append(candidates, transaction)
	}
	if include != nil {
		candidates = append(candidates, *include)
	}

	return validateTransactionSequence(candidates)
}

func validateTransactionSequence(transactions []model.Transaction) error {
	sort.SliceStable(transactions, func(i, j int) bool {
		if transactions[i].TransactionDate.Equal(transactions[j].TransactionDate) {
			if transactions[i].CreatedAt.Equal(transactions[j].CreatedAt) {
				return transactions[i].ID < transactions[j].ID
			}
			return transactions[i].CreatedAt.Before(transactions[j].CreatedAt)
		}
		return transactions[i].TransactionDate.Before(transactions[j].TransactionDate)
	})

	availableQuantity := decimal.Zero
	for _, transaction := range transactions {
		switch transaction.TransactionType {
		case "buy":
			availableQuantity = availableQuantity.Add(transaction.Quantity)
		case "sell":
			if availableQuantity.LessThan(transaction.Quantity) {
				return errors.New("insufficient available quantity")
			}
			availableQuantity = availableQuantity.Sub(transaction.Quantity)
		case "dividend":
			continue
		default:
			return errors.New("invalid transaction type")
		}
	}

	return nil
}
