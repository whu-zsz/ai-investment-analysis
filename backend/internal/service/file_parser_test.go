package service_test

import (
	"os"
	"path/filepath"
	"stock-analysis-backend/internal/service"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tmpFile.Name()
}

func writeTempExcel(t *testing.T, rows [][]interface{}) string {
	t.Helper()
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			t.Fatalf("Failed to build cell name: %v", err)
		}
		if err := file.SetSheetRow(sheetName, cell, &row); err != nil {
			t.Fatalf("Failed to write sheet row: %v", err)
		}
	}
	tmpPath := filepath.Join(t.TempDir(), "test.xlsx")
	if err := file.SaveAs(tmpPath); err != nil {
		t.Fatalf("Failed to save excel file: %v", err)
	}
	return tmpPath
}

func TestFileParserService_ParseCSV(t *testing.T) {
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费
2024-06-15,buy,stock,600519,贵州茅台,100,1800.00,50.00
2024-07-01,buy,stock,000858,五粮液,200,150.00,30.00
2024-08-15,sell,stock,600519,贵州茅台,50,1900.00,25.00
`
	filePath := writeTempCSV(t, content)
	defer os.Remove(filePath)

	parser := service.NewFileParserService()
	result, err := parser.ParseCSV(filePath, 1)
	if err != nil {
		t.Fatalf("ParseCSV() error = %v", err)
	}

	if result.RecordsTotal != 3 {
		t.Errorf("ParseCSV() RecordsTotal = %d, want 3", result.RecordsTotal)
	}
	if result.RecordsImported != 3 {
		t.Errorf("ParseCSV() RecordsImported = %d, want 3", result.RecordsImported)
	}
	if result.RecordsFailed != 0 {
		t.Errorf("ParseCSV() RecordsFailed = %d, want 0", result.RecordsFailed)
	}
	if len(result.Transactions) != 3 {
		t.Fatalf("ParseCSV() returned %d transactions, want 3", len(result.Transactions))
	}
	if result.Transactions[0].AssetCode != "600519" {
		t.Errorf("ParseCSV() AssetCode = %v, want 600519", result.Transactions[0].AssetCode)
	}
	if result.Transactions[0].TransactionType != "buy" {
		t.Errorf("ParseCSV() TransactionType = %v, want buy", result.Transactions[0].TransactionType)
	}
}

func TestFileParserService_ParseCSV_FileNotFound(t *testing.T) {
	parser := service.NewFileParserService()
	_, err := parser.ParseCSV("/non/existent/file.csv", 1)
	if err == nil {
		t.Error("ParseCSV() should return error for non-existent file")
	}
}

func TestFileParserService_ParseCSV_EmptyFile(t *testing.T) {
	content := "交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费\n"
	filePath := writeTempCSV(t, content)
	defer os.Remove(filePath)

	parser := service.NewFileParserService()
	result, err := parser.ParseCSV(filePath, 1)
	if err != nil {
		t.Fatalf("ParseCSV() error = %v", err)
	}

	if result.RecordsTotal != 0 || result.RecordsImported != 0 || result.RecordsFailed != 0 {
		t.Errorf("ParseCSV() unexpected stats: %+v", result)
	}
}

func TestFileParserService_ParseCSV_InvalidRows(t *testing.T) {
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费
invalid-date,buy,stock,600519,贵州茅台,100,1800.00,50.00
2024-07-01,buy,stock,000858,五粮液,200,150.00,30.00
2024-08-01,buy,stock,000001,平安银行,invalid,12.50,2.00
`
	filePath := writeTempCSV(t, content)
	defer os.Remove(filePath)

	parser := service.NewFileParserService()
	result, err := parser.ParseCSV(filePath, 1)
	if err != nil {
		t.Fatalf("ParseCSV() error = %v", err)
	}

	if result.RecordsTotal != 3 {
		t.Errorf("ParseCSV() RecordsTotal = %d, want 3", result.RecordsTotal)
	}
	if result.RecordsImported != 1 {
		t.Errorf("ParseCSV() RecordsImported = %d, want 1", result.RecordsImported)
	}
	if result.RecordsFailed != 2 {
		t.Errorf("ParseCSV() RecordsFailed = %d, want 2", result.RecordsFailed)
	}
	if len(result.Errors) != 2 {
		t.Fatalf("ParseCSV() errors = %d, want 2", len(result.Errors))
	}
	if result.Errors[0].RowNumber != 2 {
		t.Errorf("first error row = %d, want 2", result.Errors[0].RowNumber)
	}
	if !strings.Contains(result.Errors[0].Reason, "交易日期") {
		t.Errorf("first error reason = %q, want date error", result.Errors[0].Reason)
	}
	if result.Errors[1].RowNumber != 4 {
		t.Errorf("second error row = %d, want 4", result.Errors[1].RowNumber)
	}
	if !strings.Contains(result.Errors[1].Reason, "数量") {
		t.Errorf("second error reason = %q, want quantity error", result.Errors[1].Reason)
	}
}

func TestFileParserService_ParseCSV_TotalAmount(t *testing.T) {
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费
2024-06-15,buy,stock,600519,贵州茅台,100,1800.00,50.00
`
	filePath := writeTempCSV(t, content)
	defer os.Remove(filePath)

	parser := service.NewFileParserService()
	result, err := parser.ParseCSV(filePath, 1)
	if err != nil {
		t.Fatalf("ParseCSV() error = %v", err)
	}
	if len(result.Transactions) != 1 {
		t.Fatalf("ParseCSV() returned %d transactions, want 1", len(result.Transactions))
	}
	if result.Transactions[0].TotalAmount.String() != "180000" {
		t.Errorf("ParseCSV() TotalAmount = %v, want 180000", result.Transactions[0].TotalAmount.String())
	}
}

func TestFileParserService_ParseCSV_WithoutCommission(t *testing.T) {
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价
2024-06-15,buy,stock,600519,贵州茅台,100,1800.00
`
	filePath := writeTempCSV(t, content)
	defer os.Remove(filePath)

	parser := service.NewFileParserService()
	result, err := parser.ParseCSV(filePath, 1)
	if err != nil {
		t.Fatalf("ParseCSV() error = %v", err)
	}
	if len(result.Transactions) != 1 {
		t.Fatalf("ParseCSV() returned %d transactions, want 1", len(result.Transactions))
	}
	if !result.Transactions[0].Commission.IsZero() {
		t.Errorf("ParseCSV() Commission should be zero when not provided")
	}
}

func TestFileParserService_ParseExcel_PartialSuccess(t *testing.T) {
	filePath := writeTempExcel(t, [][]interface{}{
		{"交易日期", "交易类型", "资产类型", "资产代码", "资产名称", "数量", "单价", "手续费"},
		{"2024-06-15", "buy", "stock", "600519", "贵州茅台", "100", "1800.00", "50.00"},
		{"bad-date", "buy", "stock", "000858", "五粮液", "200", "150.00", "30.00"},
		{"2024-08-01", "sell", "stock", "000001", "平安银行", "10", "12.50", "1.00"},
	})

	parser := service.NewFileParserService()
	result, err := parser.ParseExcel(filePath, 1)
	if err != nil {
		t.Fatalf("ParseExcel() error = %v", err)
	}

	if result.RecordsTotal != 3 {
		t.Errorf("ParseExcel() RecordsTotal = %d, want 3", result.RecordsTotal)
	}
	if result.RecordsImported != 2 {
		t.Errorf("ParseExcel() RecordsImported = %d, want 2", result.RecordsImported)
	}
	if result.RecordsFailed != 1 {
		t.Errorf("ParseExcel() RecordsFailed = %d, want 1", result.RecordsFailed)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("ParseExcel() errors = %d, want 1", len(result.Errors))
	}
	if result.Errors[0].RowNumber != 3 {
		t.Errorf("ParseExcel() error row = %d, want 3", result.Errors[0].RowNumber)
	}
}

func TestFileParserService_Interface(t *testing.T) {
	var _ service.FileParserService = service.NewFileParserService()
}
