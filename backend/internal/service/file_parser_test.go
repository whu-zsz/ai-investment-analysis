package service_test

import (
	"os"
	"stock-analysis-backend/internal/service"
	"testing"
)

func TestFileParserService_ParseCSV(t *testing.T) {
	// 创建测试 CSV 文件
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费
2024-06-15,buy,stock,600519,贵州茅台,100,1800.00,50.00
2024-07-01,buy,stock,000858,五粮液,200,150.00,30.00
2024-08-15,sell,stock,600519,贵州茅台,50,1900.00,25.00
`
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// 测试解析
	parser := service.NewFileParserService()
	transactions, err := parser.ParseCSV(tmpFile.Name(), 1)
	if err != nil {
		t.Errorf("ParseCSV() error = %v", err)
	}

	if len(transactions) != 3 {
		t.Errorf("ParseCSV() returned %d transactions, want 3", len(transactions))
	}

	// 验证第一条记录
	if transactions[0].AssetCode != "600519" {
		t.Errorf("ParseCSV() AssetCode = %v, want 600519", transactions[0].AssetCode)
	}
	if transactions[0].TransactionType != "buy" {
		t.Errorf("ParseCSV() TransactionType = %v, want buy", transactions[0].TransactionType)
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
	// 创建空 CSV 文件（只有标题行）
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费
`
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	parser := service.NewFileParserService()
	transactions, err := parser.ParseCSV(tmpFile.Name(), 1)
	if err != nil {
		t.Errorf("ParseCSV() error = %v", err)
	}

	if len(transactions) != 0 {
		t.Errorf("ParseCSV() returned %d transactions, want 0", len(transactions))
	}
}

func TestFileParserService_ParseCSV_InvalidDate(t *testing.T) {
	// 创建包含无效日期的 CSV 文件
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费
invalid-date,buy,stock,600519,贵州茅台,100,1800.00,50.00
2024-07-01,buy,stock,000858,五粮液,200,150.00,30.00
`
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	parser := service.NewFileParserService()
	transactions, err := parser.ParseCSV(tmpFile.Name(), 1)
	if err != nil {
		t.Errorf("ParseCSV() error = %v", err)
	}

	// 无效日期的记录应该被跳过
	if len(transactions) != 1 {
		t.Errorf("ParseCSV() returned %d transactions, want 1 (invalid record skipped)", len(transactions))
	}
}

func TestFileParserService_ParseCSV_InvalidQuantity(t *testing.T) {
	// 创建包含无效数量的 CSV 文件
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费
2024-06-15,buy,stock,600519,贵州茅台,invalid,1800.00,50.00
`
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	parser := service.NewFileParserService()
	transactions, err := parser.ParseCSV(tmpFile.Name(), 1)
	if err != nil {
		t.Errorf("ParseCSV() error = %v", err)
	}

	// 无效数量的记录应该被跳过
	if len(transactions) != 0 {
		t.Errorf("ParseCSV() returned %d transactions, want 0 (invalid record skipped)", len(transactions))
	}
}

func TestFileParserService_ParseCSV_InvalidPrice(t *testing.T) {
	// 创建包含无效单价的 CSV 文件
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费
2024-06-15,buy,stock,600519,贵州茅台,100,invalid,50.00
`
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	parser := service.NewFileParserService()
	transactions, err := parser.ParseCSV(tmpFile.Name(), 1)
	if err != nil {
		t.Errorf("ParseCSV() error = %v", err)
	}

	// 无效单价的记录应该被跳过
	if len(transactions) != 0 {
		t.Errorf("ParseCSV() returned %d transactions, want 0 (invalid record skipped)", len(transactions))
	}
}

func TestFileParserService_ParseCSV_TotalAmount(t *testing.T) {
	// 验证总金额计算
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价,手续费
2024-06-15,buy,stock,600519,贵州茅台,100,1800.00,50.00
`
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	parser := service.NewFileParserService()
	transactions, err := parser.ParseCSV(tmpFile.Name(), 1)
	if err != nil {
		t.Errorf("ParseCSV() error = %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("ParseCSV() returned %d transactions, want 1", len(transactions))
	}

	// 验证总金额 = 数量 * 单价 = 100 * 1800 = 180000
	expectedTotal := "180000"
	if transactions[0].TotalAmount.String() != expectedTotal {
		t.Errorf("ParseCSV() TotalAmount = %v, want %v", transactions[0].TotalAmount.String(), expectedTotal)
	}
}

func TestFileParserService_ParseCSV_WithoutCommission(t *testing.T) {
	// 测试没有手续费列的情况
	content := `交易日期,交易类型,资产类型,资产代码,资产名称,数量,单价
2024-06-15,buy,stock,600519,贵州茅台,100,1800.00
`
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	parser := service.NewFileParserService()
	transactions, err := parser.ParseCSV(tmpFile.Name(), 1)
	if err != nil {
		t.Errorf("ParseCSV() error = %v", err)
	}

	if len(transactions) != 1 {
		t.Errorf("ParseCSV() returned %d transactions, want 1", len(transactions))
	}

	// 手续费应该为 0
	if !transactions[0].Commission.IsZero() {
		t.Errorf("ParseCSV() Commission should be zero when not provided")
	}
}

func TestFileParserService_Interface(t *testing.T) {
	var _ service.FileParserService = service.NewFileParserService()
}
