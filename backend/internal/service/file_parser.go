package service

import (
	"encoding/csv"
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

type UploadRowError struct {
	RowNumber int
	Reason    string
}

type FileParseResult struct {
	Transactions    []model.Transaction
	RecordsTotal    int
	RecordsImported int
	RecordsFailed   int
	Errors          []UploadRowError
}

type FileParserService interface {
	ParseCSV(filePath string, userID uint64) (*FileParseResult, error)
	ParseExcel(filePath string, userID uint64) (*FileParseResult, error)
}

type fileParserService struct{}

func NewFileParserService() FileParserService {
	return &fileParserService{}
}

func (s *fileParserService) ParseCSV(filePath string, userID uint64) (*FileParseResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	result := &FileParseResult{}

	_, err = reader.Read()
	if err == io.EOF {
		return result, nil
	}
	if err != nil {
		return nil, err
	}

	rowNumber := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		rowNumber++
		if isBlankCells(record) {
			continue
		}

		result.RecordsTotal++
		transaction, err := s.parseCSVRecord(record, userID)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, UploadRowError{RowNumber: rowNumber, Reason: err.Error()})
			continue
		}

		result.Transactions = append(result.Transactions, *transaction)
		result.RecordsImported++
	}

	return result, nil
}

func (s *fileParserService) parseCSVRecord(record []string, userID uint64) (*model.Transaction, error) {
	if len(record) < 7 {
		return nil, fmt.Errorf("列数不足，至少需要 7 列")
	}

	transactionDate, err := time.Parse("2006-01-02", strings.TrimSpace(record[0]))
	if err != nil {
		return nil, fmt.Errorf("交易日期格式错误，应为 YYYY-MM-DD")
	}

	transactionType := strings.ToLower(strings.TrimSpace(record[1]))
	if transactionType == "" {
		return nil, fmt.Errorf("交易类型不能为空")
	}

	assetType := strings.TrimSpace(record[2])
	if assetType == "" {
		return nil, fmt.Errorf("资产类型不能为空")
	}

	assetCode := strings.TrimSpace(record[3])
	if assetCode == "" {
		return nil, fmt.Errorf("资产代码不能为空")
	}

	assetName := strings.TrimSpace(record[4])
	if assetName == "" {
		return nil, fmt.Errorf("资产名称不能为空")
	}

	quantity, err := decimal.NewFromString(strings.TrimSpace(record[5]))
	if err != nil {
		return nil, fmt.Errorf("数量格式错误")
	}

	pricePerUnit, err := decimal.NewFromString(strings.TrimSpace(record[6]))
	if err != nil {
		return nil, fmt.Errorf("单价格式错误")
	}

	totalAmount := quantity.Mul(pricePerUnit)

	commission := decimal.Zero
	if len(record) > 7 {
		commission, _ = decimal.NewFromString(strings.TrimSpace(record[7]))
	}

	transaction := &model.Transaction{
		UserID:          userID,
		TransactionDate: transactionDate,
		TransactionType: transactionType,
		AssetType:       assetType,
		AssetCode:       assetCode,
		AssetName:       assetName,
		Quantity:        quantity,
		PricePerUnit:    pricePerUnit,
		TotalAmount:     totalAmount,
		Commission:      commission,
	}

	return transaction, nil
}

func (s *fileParserService) ParseExcel(filePath string, userID uint64) (*FileParseResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	result := &FileParseResult{}
	for i, row := range rows {
		if i == 0 || isBlankCells(row) {
			continue
		}

		result.RecordsTotal++
		transaction, err := s.parseExcelRow(row, userID)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, UploadRowError{RowNumber: i + 1, Reason: err.Error()})
			continue
		}

		result.Transactions = append(result.Transactions, *transaction)
		result.RecordsImported++
	}

	return result, nil
}

func (s *fileParserService) parseExcelRow(row []string, userID uint64) (*model.Transaction, error) {
	if len(row) < 7 {
		return nil, fmt.Errorf("列数不足，至少需要 7 列")
	}

	var transactionDate time.Time
	dateStr := strings.TrimSpace(row[0])
	if _, err := strconv.ParseFloat(dateStr, 64); err == nil {
		excelDate, _ := strconv.ParseFloat(dateStr, 64)
		transactionDate, err = excelize.ExcelDateToTime(excelDate, false)
		if err != nil {
			return nil, fmt.Errorf("交易日期格式错误")
		}
	} else {
		var err error
		transactionDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("交易日期格式错误，应为 YYYY-MM-DD")
		}
	}

	transactionType := strings.ToLower(strings.TrimSpace(row[1]))
	if transactionType == "" {
		return nil, fmt.Errorf("交易类型不能为空")
	}

	assetType := strings.TrimSpace(row[2])
	if assetType == "" {
		return nil, fmt.Errorf("资产类型不能为空")
	}

	assetCode := strings.TrimSpace(row[3])
	if assetCode == "" {
		return nil, fmt.Errorf("资产代码不能为空")
	}

	assetName := strings.TrimSpace(row[4])
	if assetName == "" {
		return nil, fmt.Errorf("资产名称不能为空")
	}

	quantity, err := decimal.NewFromString(strings.TrimSpace(row[5]))
	if err != nil {
		return nil, fmt.Errorf("数量格式错误")
	}

	pricePerUnit, err := decimal.NewFromString(strings.TrimSpace(row[6]))
	if err != nil {
		return nil, fmt.Errorf("单价格式错误")
	}

	totalAmount := quantity.Mul(pricePerUnit)

	commission := decimal.Zero
	if len(row) > 7 {
		commission, _ = decimal.NewFromString(strings.TrimSpace(row[7]))
	}

	transaction := &model.Transaction{
		UserID:          userID,
		TransactionDate: transactionDate,
		TransactionType: transactionType,
		AssetType:       assetType,
		AssetCode:       assetCode,
		AssetName:       assetName,
		Quantity:        quantity,
		PricePerUnit:    pricePerUnit,
		TotalAmount:     totalAmount,
		Commission:      commission,
	}

	return transaction, nil
}

func isBlankCells(cells []string) bool {
	for _, cell := range cells {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
