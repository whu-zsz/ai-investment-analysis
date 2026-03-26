package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"strings"
	"time"
)

type UploadService interface {
	ProcessUploadedFile(userID uint64, filePath string, originalFileName string, fileSize int64, fileType string) (*response.UploadResponse, error)
	GetUploadHistory(userID uint64) ([]response.UploadHistoryResponse, error)
}

type uploadService struct {
	uploadedFileRepo  repository.UploadedFileRepository
	transactionRepo   repository.TransactionRepository
	fileParserService FileParserService
	uploadCfg         config.UploadConfig
}

func NewUploadService(
	uploadedFileRepo repository.UploadedFileRepository,
	transactionRepo repository.TransactionRepository,
	fileParserService FileParserService,
	uploadCfg config.UploadConfig,
) UploadService {
	return &uploadService{
		uploadedFileRepo:  uploadedFileRepo,
		transactionRepo:   transactionRepo,
		fileParserService: fileParserService,
		uploadCfg:         uploadCfg,
	}
}

func (s *uploadService) ProcessUploadedFile(userID uint64, filePath string, originalFileName string, fileSize int64, fileType string) (*response.UploadResponse, error) {
	// 1. 验证文件类型
	fileExt := strings.ToLower(filepath.Ext(originalFileName))
	if fileExt != ".csv" && fileExt != ".xlsx" && fileExt != ".xls" {
		return nil, errors.New("unsupported file type, only CSV and Excel files are allowed")
	}

	// 2. 验证文件大小
	if fileSize > s.uploadCfg.MaxUploadSize {
		return nil, fmt.Errorf("file size exceeds maximum limit of %d bytes", s.uploadCfg.MaxUploadSize)
	}

	// 3. 创建uploaded_files记录
	uploadedFile := &model.UploadedFile{
		UserID:       userID,
		FileName:     originalFileName,
		FilePath:     filePath,
		FileSize:     fileSize,
		FileType:     fileType,
		UploadStatus: "processing",
		UploadedAt:   time.Now(),
	}

	if err := s.uploadedFileRepo.Create(uploadedFile); err != nil {
		return nil, err
	}

	// 4. 解析文件
	var transactions []model.Transaction
	var parseErr error

	if fileType == "csv" {
		transactions, parseErr = s.fileParserService.ParseCSV(filePath, userID)
	} else {
		transactions, parseErr = s.fileParserService.ParseExcel(filePath, userID)
	}

	if parseErr != nil {
		errorMsg := parseErr.Error()
		s.uploadedFileRepo.UpdateStatus(uploadedFile.ID, "failed", 0, &errorMsg)
		return nil, parseErr
	}

	// 5. 批量插入交易记录
	if len(transactions) > 0 {
		if err := s.transactionRepo.BatchCreate(transactions); err != nil {
			errorMsg := err.Error()
			s.uploadedFileRepo.UpdateStatus(uploadedFile.ID, "failed", 0, &errorMsg)
			return nil, err
		}
	}

	// 6. 更新uploaded_files状态
	s.uploadedFileRepo.UpdateStatus(uploadedFile.ID, "success", len(transactions), nil)

	// 7. 返回结果
	return &response.UploadResponse{
		FileID:          int64(uploadedFile.ID),
		FileName:        originalFileName,
		RecordsImported: len(transactions),
		Message:         fmt.Sprintf("Successfully imported %d transactions", len(transactions)),
	}, nil
}

func (s *uploadService) GetUploadHistory(userID uint64) ([]response.UploadHistoryResponse, error) {
	files, err := s.uploadedFileRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	var history []response.UploadHistoryResponse
	for _, file := range files {
		var processedAt string
		if file.ProcessedAt != nil {
			processedAt = file.ProcessedAt.Format("2006-01-02 15:04:05")
		}

		history = append(history, response.UploadHistoryResponse{
			ID:              file.ID,
			FileName:        file.FileName,
			FileSize:        file.FileSize,
			FileType:        file.FileType,
			UploadStatus:    file.UploadStatus,
			RecordsImported: file.RecordsImported,
			UploadedAt:      file.UploadedAt.Format("2006-01-02 15:04:05"),
			ProcessedAt:     processedAt,
		})
	}

	return history, nil
}
