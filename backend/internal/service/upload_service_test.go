package service_test

import (
	"errors"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"gorm.io/gorm"
)

// MockUploadedFileRepository 模拟上传文件仓储
type MockUploadedFileRepository struct {
	Files  map[uint64]*model.UploadedFile
	NextID uint64
}

func NewMockUploadedFileRepository() *MockUploadedFileRepository {
	return &MockUploadedFileRepository{
		Files:  make(map[uint64]*model.UploadedFile),
		NextID: 1,
	}
}

func (r *MockUploadedFileRepository) Create(file *model.UploadedFile) error {
	file.ID = r.NextID
	r.Files[r.NextID] = file
	r.NextID++
	return nil
}

func (r *MockUploadedFileRepository) FindByUserID(userID uint64) ([]model.UploadedFile, error) {
	var result []model.UploadedFile
	for _, f := range r.Files {
		if f.UserID == userID {
			result = append(result, *f)
		}
	}
	return result, nil
}

func (r *MockUploadedFileRepository) FindByID(id uint64) (*model.UploadedFile, error) {
	file, ok := r.Files[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return file, nil
}

func (r *MockUploadedFileRepository) UpdateStatus(id uint64, status string, recordsImported int, errorMsg *string) error {
	file, ok := r.Files[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	file.UploadStatus = status
	file.RecordsImported = recordsImported
	if errorMsg != nil {
		e := *errorMsg
		file.ErrorMessage = &e
	}
	return nil
}

// MockTransactionRepositoryForUpload 模拟交易仓储
type MockTransactionRepositoryForUpload struct {
	Transactions []model.Transaction
	Err          error
}

func (r *MockTransactionRepositoryForUpload) Create(transaction *model.Transaction) error {
	return nil
}

func (r *MockTransactionRepositoryForUpload) BatchCreate(transactions []model.Transaction) error {
	if r.Err != nil {
		return r.Err
	}
	r.Transactions = append(r.Transactions, transactions...)
	return nil
}

func (r *MockTransactionRepositoryForUpload) FindByID(id uint64) (*model.Transaction, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *MockTransactionRepositoryForUpload) FindByUserID(userID uint64, limit, offset int) ([]model.Transaction, int64, error) {
	return []model.Transaction{}, 0, nil
}

func (r *MockTransactionRepositoryForUpload) FindByAssetCode(userID uint64, assetCode string) ([]model.Transaction, error) {
	return []model.Transaction{}, nil
}

func (r *MockTransactionRepositoryForUpload) FindByDateRange(userID uint64, startDate, endDate string) ([]model.Transaction, error) {
	return []model.Transaction{}, nil
}

func (r *MockTransactionRepositoryForUpload) Update(transaction *model.Transaction) error {
	return nil
}

func (r *MockTransactionRepositoryForUpload) Delete(id uint64) error {
	return nil
}

func (r *MockTransactionRepositoryForUpload) GetTransactionStats(userID uint64) (*response.TransactionStats, error) {
	return &response.TransactionStats{}, nil
}

// MockFileParserService 模拟文件解析服务
type MockFileParserService struct {
	Result     *service.FileParseResult
	ParseError error
}

func (m *MockFileParserService) ParseCSV(filePath string, userID uint64) (*service.FileParseResult, error) {
	if m.ParseError != nil {
		return nil, m.ParseError
	}
	return m.Result, nil
}

func (m *MockFileParserService) ParseExcel(filePath string, userID uint64) (*service.FileParseResult, error) {
	if m.ParseError != nil {
		return nil, m.ParseError
	}
	return m.Result, nil
}

func TestUploadService_ProcessUploadedFile_CSV(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	txRepo := &MockTransactionRepositoryForUpload{}
	parser := &MockFileParserService{
		Result: &service.FileParseResult{
			Transactions:    []model.Transaction{{UserID: 1, AssetCode: "600519", AssetName: "贵州茅台"}},
			RecordsTotal:    1,
			RecordsImported: 1,
			RecordsFailed:   0,
		},
	}
	uploadCfg := config.UploadConfig{MaxUploadSize: 10485760}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	result, err := uploadService.ProcessUploadedFile(1, "/tmp/test.csv", "test.csv", 1024, "csv")
	if err != nil {
		t.Fatalf("ProcessUploadedFile() error = %v", err)
	}

	if result.UploadStatus != "success" {
		t.Errorf("Expected upload status success, got %s", result.UploadStatus)
	}
	if result.RecordsTotal != 1 || result.RecordsImported != 1 || result.RecordsFailed != 0 {
		t.Errorf("Unexpected stats: %+v", result)
	}
	if result.FileName != "test.csv" {
		t.Errorf("Expected filename test.csv, got %s", result.FileName)
	}
}

func TestUploadService_ProcessUploadedFile_Excel(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	txRepo := &MockTransactionRepositoryForUpload{}
	parser := &MockFileParserService{
		Result: &service.FileParseResult{
			Transactions: []model.Transaction{
				{UserID: 1, AssetCode: "600519", AssetName: "贵州茅台"},
				{UserID: 1, AssetCode: "000858", AssetName: "五粮液"},
			},
			RecordsTotal:    2,
			RecordsImported: 2,
			RecordsFailed:   0,
		},
	}
	uploadCfg := config.UploadConfig{MaxUploadSize: 10485760}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	result, err := uploadService.ProcessUploadedFile(1, "/tmp/test.xlsx", "test.xlsx", 2048, "excel")
	if err != nil {
		t.Fatalf("ProcessUploadedFile() error = %v", err)
	}

	if result.RecordsImported != 2 {
		t.Errorf("Expected 2 records imported, got %d", result.RecordsImported)
	}
}

func TestUploadService_ProcessUploadedFile_UnsupportedType(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	txRepo := &MockTransactionRepositoryForUpload{}
	parser := &MockFileParserService{}
	uploadCfg := config.UploadConfig{MaxUploadSize: 10485760}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	_, err := uploadService.ProcessUploadedFile(1, "/tmp/test.txt", "test.txt", 1024, "text")
	if err == nil {
		t.Error("Expected error for unsupported file type")
	}
}

func TestUploadService_ProcessUploadedFile_FileTooLarge(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	txRepo := &MockTransactionRepositoryForUpload{}
	parser := &MockFileParserService{}
	uploadCfg := config.UploadConfig{MaxUploadSize: 1000}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	_, err := uploadService.ProcessUploadedFile(1, "/tmp/test.csv", "test.csv", 2048, "csv")
	if err == nil {
		t.Error("Expected error for file too large")
	}
}

func TestUploadService_ProcessUploadedFile_ParseError(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	txRepo := &MockTransactionRepositoryForUpload{}
	parser := &MockFileParserService{ParseError: errors.New("parse error")}
	uploadCfg := config.UploadConfig{MaxUploadSize: 10485760}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	_, err := uploadService.ProcessUploadedFile(1, "/tmp/test.csv", "test.csv", 1024, "csv")
	if err == nil {
		t.Error("Expected error for parse error")
	}

	file, _ := fileRepo.FindByID(1)
	if file.UploadStatus != "failed" {
		t.Errorf("Expected status 'failed', got %s", file.UploadStatus)
	}
}

func TestUploadService_ProcessUploadedFile_PartialSuccess(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	txRepo := &MockTransactionRepositoryForUpload{}
	parser := &MockFileParserService{
		Result: &service.FileParseResult{
			Transactions: []model.Transaction{{UserID: 1, AssetCode: "600519", AssetName: "贵州茅台"}},
			RecordsTotal: 3,
			RecordsImported: 1,
			RecordsFailed: 2,
			Errors: []service.UploadRowError{
				{RowNumber: 2, Reason: "交易日期格式错误，应为 YYYY-MM-DD"},
				{RowNumber: 4, Reason: "数量格式错误"},
			},
		},
	}
	uploadCfg := config.UploadConfig{MaxUploadSize: 10485760}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	result, err := uploadService.ProcessUploadedFile(1, "/tmp/test.csv", "test.csv", 1024, "csv")
	if err != nil {
		t.Fatalf("ProcessUploadedFile() error = %v", err)
	}

	if result.UploadStatus != "partial_success" {
		t.Errorf("Expected partial_success, got %s", result.UploadStatus)
	}
	if result.RecordsFailed != 2 || len(result.Errors) != 2 {
		t.Errorf("Unexpected partial result: %+v", result)
	}
	file, _ := fileRepo.FindByID(1)
	if file.UploadStatus != "partial_success" {
		t.Errorf("Expected stored status 'partial_success', got %s", file.UploadStatus)
	}
}

func TestUploadService_ProcessUploadedFile_AllRowsFailed(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	txRepo := &MockTransactionRepositoryForUpload{}
	parser := &MockFileParserService{
		Result: &service.FileParseResult{
			RecordsTotal:    2,
			RecordsImported: 0,
			RecordsFailed:   2,
			Errors: []service.UploadRowError{{RowNumber: 2, Reason: "交易日期格式错误，应为 YYYY-MM-DD"}},
		},
	}
	uploadCfg := config.UploadConfig{MaxUploadSize: 10485760}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	_, err := uploadService.ProcessUploadedFile(1, "/tmp/test.csv", "test.csv", 1024, "csv")
	if err == nil {
		t.Fatal("Expected error when all rows fail")
	}
	file, _ := fileRepo.FindByID(1)
	if file.UploadStatus != "failed" {
		t.Errorf("Expected status failed, got %s", file.UploadStatus)
	}
}

func TestUploadService_ProcessUploadedFile_BatchCreateError(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	txRepo := &MockTransactionRepositoryForUpload{Err: errors.New("database error")}
	parser := &MockFileParserService{
		Result: &service.FileParseResult{
			Transactions:    []model.Transaction{{UserID: 1}},
			RecordsTotal:    1,
			RecordsImported: 1,
			RecordsFailed:   0,
		},
	}
	uploadCfg := config.UploadConfig{MaxUploadSize: 10485760}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	_, err := uploadService.ProcessUploadedFile(1, "/tmp/test.csv", "test.csv", 1024, "csv")
	if err == nil {
		t.Error("Expected error for batch create error")
	}
}

func TestUploadService_GetUploadHistory(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	now := time.Now()
	processedAt := now.Add(time.Minute)
	fileRepo.Create(&model.UploadedFile{
		UserID:          1,
		FileName:        "test.csv",
		FileSize:        1024,
		FileType:        "csv",
		UploadStatus:    "success",
		RecordsImported: 10,
		UploadedAt:      now,
		ProcessedAt:     &processedAt,
	})

	txRepo := &MockTransactionRepositoryForUpload{}
	parser := &MockFileParserService{}
	uploadCfg := config.UploadConfig{}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	history, err := uploadService.GetUploadHistory(1)
	if err != nil {
		t.Fatalf("GetUploadHistory() error = %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 history record, got %d", len(history))
	}
	if history[0].FileName != "test.csv" {
		t.Errorf("Expected filename test.csv, got %s", history[0].FileName)
	}
}

func TestUploadService_GetUploadHistory_Empty(t *testing.T) {
	fileRepo := NewMockUploadedFileRepository()
	txRepo := &MockTransactionRepositoryForUpload{}
	parser := &MockFileParserService{}
	uploadCfg := config.UploadConfig{}

	uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

	history, err := uploadService.GetUploadHistory(1)
	if err != nil {
		t.Fatalf("GetUploadHistory() error = %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected 0 history records, got %d", len(history))
	}
}

func TestUploadService_ProcessUploadedFile_DifferentExtensions(t *testing.T) {
	tests := []struct {
		filename    string
		fileType    string
		expectError bool
	}{
		{"test.csv", "csv", false},
		{"test.CSV", "csv", false},
		{"test.xlsx", "excel", false},
		{"test.XLSX", "excel", false},
		{"test.xls", "excel", false},
		{"test.XLS", "excel", false},
		{"test.txt", "text", true},
		{"test.pdf", "pdf", true},
	}

	for _, tt := range tests {
		fileRepo := NewMockUploadedFileRepository()
		txRepo := &MockTransactionRepositoryForUpload{}
		parser := &MockFileParserService{
			Result: &service.FileParseResult{
				Transactions:    []model.Transaction{{UserID: 1}},
				RecordsTotal:    1,
				RecordsImported: 1,
				RecordsFailed:   0,
			},
		}
		uploadCfg := config.UploadConfig{MaxUploadSize: 10485760}

		uploadService := service.NewUploadService(fileRepo, txRepo, parser, uploadCfg)

		_, err := uploadService.ProcessUploadedFile(1, "/tmp/"+tt.filename, tt.filename, 1024, tt.fileType)
		if tt.expectError && err == nil {
			t.Errorf("Expected error for %s", tt.filename)
		}
		if !tt.expectError && err != nil {
			t.Errorf("Unexpected error for %s: %v", tt.filename, err)
		}
	}
}
