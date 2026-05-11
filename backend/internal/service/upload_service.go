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

const maxUploadRowErrors = 10

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
	fileExt := strings.ToLower(filepath.Ext(originalFileName))
	if fileExt != ".csv" && fileExt != ".xlsx" && fileExt != ".xls" {
		return nil, errors.New("unsupported file type, only CSV and Excel files are allowed")
	}

	if fileSize > s.uploadCfg.MaxUploadSize {
		return nil, fmt.Errorf("file size exceeds maximum limit of %d bytes", s.uploadCfg.MaxUploadSize)
	}

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

	var parseResult *FileParseResult
	var parseErr error

	if fileType == "csv" {
		parseResult, parseErr = s.fileParserService.ParseCSV(filePath, userID)
	} else {
		parseResult, parseErr = s.fileParserService.ParseExcel(filePath, userID)
	}

	if parseErr != nil {
		errorMsg := parseErr.Error()
		s.uploadedFileRepo.UpdateStatus(uploadedFile.ID, "failed", 0, &errorMsg)
		return nil, parseErr
	}

	if parseResult == nil {
		parseResult = &FileParseResult{}
	}

	if parseResult.RecordsImported == 0 {
		message := "文件中没有可导入的有效记录"
		if parseResult.RecordsFailed > 0 {
			message = fmt.Sprintf("导入失败：共解析 %d 条，全部失败", parseResult.RecordsTotal)
		}
		errorMsg := message
		s.uploadedFileRepo.UpdateStatus(uploadedFile.ID, "failed", 0, &errorMsg)
		return nil, errors.New(message)
	}

	if err := s.transactionRepo.BatchCreate(parseResult.Transactions); err != nil {
		errorMsg := err.Error()
		s.uploadedFileRepo.UpdateStatus(uploadedFile.ID, "failed", 0, &errorMsg)
		return nil, err
	}

	uploadStatus := "success"
	message := fmt.Sprintf("成功导入 %d 条记录", parseResult.RecordsImported)
	if parseResult.RecordsFailed > 0 {
		uploadStatus = "partial_success"
		message = fmt.Sprintf("成功导入 %d 条，失败 %d 条", parseResult.RecordsImported, parseResult.RecordsFailed)
	}

	s.uploadedFileRepo.UpdateStatus(uploadedFile.ID, uploadStatus, parseResult.RecordsImported, nil)

	return &response.UploadResponse{
		FileID:          int64(uploadedFile.ID),
		FileName:        originalFileName,
		UploadStatus:    uploadStatus,
		RecordsTotal:    parseResult.RecordsTotal,
		RecordsImported: parseResult.RecordsImported,
		RecordsFailed:   parseResult.RecordsFailed,
		Errors:          toUploadRowErrors(parseResult.Errors),
		Message:         message,
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

func toUploadRowErrors(errors []UploadRowError) []response.UploadRowError {
	if len(errors) == 0 {
		return nil
	}
	if len(errors) > maxUploadRowErrors {
		errors = errors[:maxUploadRowErrors]
	}

	result := make([]response.UploadRowError, 0, len(errors))
	for _, item := range errors {
		result = append(result, response.UploadRowError{
			RowNumber: item.RowNumber,
			Reason:    item.Reason,
		})
	}
	return result
}
