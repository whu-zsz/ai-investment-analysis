package service

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"stock-analysis-backend/internal/model"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type FileParserService interface {
	ParseCSV(filePath string, userID uint64) ([]model.Transaction, error)
	ParseExcel(filePath string, userID uint64) ([]model.Transaction, error)
}

type fileParserService struct{}

func NewFileParserService() FileParserService {
	return &fileParserService{}
}

func (s *fileParserService) ParseCSV(filePath string, userID uint64) ([]model.Transaction, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var transactions []model.Transaction

	// 跳过标题行
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		transaction, err := s.parseCSVRecord(record, userID)
		if err != nil {
			continue // 跳过无效记录
		}
		transactions = append(transactions, *transaction)
	}

	return transactions, nil
}

func (s *fileParserService) parseCSVRecord(record []string, userID uint64) (*model.Transaction, error) {
	if len(record) < 7 {
		return nil, errors.New("invalid record format")
	}

	// 解析日期
	transactionDate, err := time.Parse("2006-01-02", strings.TrimSpace(record[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// 解析数量
	quantity, err := decimal.NewFromString(strings.TrimSpace(record[4]))
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	// 解析单价
	pricePerUnit, err := decimal.NewFromString(strings.TrimSpace(record[5]))
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	// 计算总金额
	totalAmount := quantity.Mul(pricePerUnit)

	// 解析手续费
	commission := decimal.Zero
	if len(record) > 6 {
		commission, _ = decimal.NewFromString(strings.TrimSpace(record[6]))
	}

	transaction := &model.Transaction{
		UserID:          userID,
		TransactionDate: transactionDate,
		TransactionType: strings.ToLower(strings.TrimSpace(record[1])),
		AssetType:       strings.TrimSpace(record[2]),
		AssetCode:       strings.TrimSpace(record[3]),
		AssetName:       strings.TrimSpace(record[3]),
		Quantity:        quantity,
		PricePerUnit:    pricePerUnit,
		TotalAmount:     totalAmount,
		Commission:      commission,
	}

	return transaction, nil
}

func (s *fileParserService) ParseExcel(filePath string, userID uint64) ([]model.Transaction, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	var transactions []model.Transaction

	// 跳过标题行
	for i, row := range rows {
		if i == 0 {
			continue
		}

		transaction, err := s.parseExcelRow(row, userID)
		if err != nil {
			continue // 跳过无效记录
		}
		transactions = append(transactions, *transaction)
	}

	return transactions, nil
}

func (s *fileParserService) parseExcelRow(row []string, userID uint64) (*model.Transaction, error) {
	if len(row) < 7 {
		return nil, errors.New("invalid row format")
	}

	// 解析日期（Excel日期格式可能是数字或字符串）
	var transactionDate time.Time
	var err error

	dateStr := strings.TrimSpace(row[0])
	if _, err := strconv.ParseFloat(dateStr, 64); err == nil {
		// 如果是Excel日期数字格式
		excelDate, _ := strconv.ParseFloat(dateStr, 64)
		transactionDate, err = excelize.ExcelDateToTime(excelDate, false)
		if err != nil {
			return nil, fmt.Errorf("invalid excel date: %w", err)
		}
	} else {
		// 字符串格式
		transactionDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	// 解析数量
	quantity, err := decimal.NewFromString(strings.TrimSpace(row[4]))
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	// 解析单价
	pricePerUnit, err := decimal.NewFromString(strings.TrimSpace(row[5]))
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	// 计算总金额
	totalAmount := quantity.Mul(pricePerUnit)

	// 解析手续费
	commission := decimal.Zero
	if len(row) > 6 {
		commission, _ = decimal.NewFromString(strings.TrimSpace(row[6]))
	}

	transaction := &model.Transaction{
		UserID:          userID,
		TransactionDate: transactionDate,
		TransactionType: strings.ToLower(strings.TrimSpace(row[1])),
		AssetType:       strings.TrimSpace(row[2]),
		AssetCode:       strings.TrimSpace(row[3]),
		AssetName:       strings.TrimSpace(row[3]),
		Quantity:        quantity,
		PricePerUnit:    pricePerUnit,
		TotalAmount:     totalAmount,
		Commission:      commission,
	}

	return transaction, nil
}
