package repository_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"gorm.io/gorm"
)

// InMemoryUploadedFileRepository 内存上传文件仓储用于测试
type InMemoryUploadedFileRepository struct {
	files  map[uint64]*model.UploadedFile
	nextID uint64
}

func NewInMemoryUploadedFileRepository() *InMemoryUploadedFileRepository {
	return &InMemoryUploadedFileRepository{
		files:  make(map[uint64]*model.UploadedFile),
		nextID: 1,
	}
}

func (r *InMemoryUploadedFileRepository) Create(file *model.UploadedFile) error {
	file.ID = r.nextID
	r.files[r.nextID] = file
	r.nextID++
	return nil
}

func (r *InMemoryUploadedFileRepository) FindByID(id uint64) (*model.UploadedFile, error) {
	file, ok := r.files[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return file, nil
}

func (r *InMemoryUploadedFileRepository) FindByUserID(userID uint64) ([]model.UploadedFile, error) {
	var result []model.UploadedFile
	for _, f := range r.files {
		if f.UserID == userID {
			result = append(result, *f)
		}
	}
	return result, nil
}

func (r *InMemoryUploadedFileRepository) UpdateStatus(id uint64, status string, recordsImported int, errorMsg *string) error {
	file, ok := r.files[id]
	if !ok {
		return nil // GORM Updates 不会对不存在的记录报错
	}
	file.UploadStatus = status
	file.RecordsImported = recordsImported
	now := time.Now()
	file.ProcessedAt = &now
	if errorMsg != nil {
		file.ErrorMessage = errorMsg
	}
	return nil
}

// TestUploadedFileRepository_Create 测试创建上传文件记录
func TestUploadedFileRepository_Create(t *testing.T) {
	repo := NewInMemoryUploadedFileRepository()

	file := &model.UploadedFile{
		UserID:       1,
		FileName:     "test.csv",
		FilePath:     "/uploads/test.csv",
		FileSize:     1024,
		FileType:     "csv",
		UploadStatus: "pending",
		UploadedAt:   time.Now(),
	}

	err := repo.Create(file)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if file.ID == 0 {
		t.Error("Create() should set file ID")
	}
}

// TestUploadedFileRepository_FindByID 测试通过ID查找
func TestUploadedFileRepository_FindByID(t *testing.T) {
	repo := NewInMemoryUploadedFileRepository()

	file := &model.UploadedFile{
		UserID:       1,
		FileName:     "test.csv",
		FilePath:     "/uploads/test.csv",
		FileSize:     1024,
		FileType:     "csv",
		UploadStatus: "pending",
		UploadedAt:   time.Now(),
	}
	repo.Create(file)

	found, err := repo.FindByID(file.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}

	if found.FileName != "test.csv" {
		t.Errorf("FindByID() FileName = %v, want test.csv", found.FileName)
	}
}

// TestUploadedFileRepository_FindByID_NotFound 测试查找不存在的记录
func TestUploadedFileRepository_FindByID_NotFound(t *testing.T) {
	repo := NewInMemoryUploadedFileRepository()

	_, err := repo.FindByID(999)
	if err != gorm.ErrRecordNotFound {
		t.Error("FindByID() should return ErrRecordNotFound for non-existent ID")
	}
}

// TestUploadedFileRepository_FindByUserID 测试通过用户ID查找
func TestUploadedFileRepository_FindByUserID(t *testing.T) {
	repo := NewInMemoryUploadedFileRepository()

	now := time.Now()
	repo.Create(&model.UploadedFile{
		UserID:       1,
		FileName:     "test1.csv",
		FilePath:     "/uploads/test1.csv",
		FileSize:     1024,
		FileType:     "csv",
		UploadStatus: "success",
		UploadedAt:   now,
	})
	repo.Create(&model.UploadedFile{
		UserID:       1,
		FileName:     "test2.xlsx",
		FilePath:     "/uploads/test2.xlsx",
		FileSize:     2048,
		FileType:     "excel",
		UploadStatus: "success",
		UploadedAt:   now.Add(time.Hour),
	})
	repo.Create(&model.UploadedFile{
		UserID:       2,
		FileName:     "test3.csv",
		FilePath:     "/uploads/test3.csv",
		FileSize:     512,
		FileType:     "csv",
		UploadStatus: "pending",
		UploadedAt:   now,
	})

	files, err := repo.FindByUserID(1)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("FindByUserID() returned %d files, want 2", len(files))
	}
}

// TestUploadedFileRepository_FindByUserID_Empty 测试空结果
func TestUploadedFileRepository_FindByUserID_Empty(t *testing.T) {
	repo := NewInMemoryUploadedFileRepository()

	files, err := repo.FindByUserID(999)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("FindByUserID() returned %d files, want 0", len(files))
	}
}

// TestUploadedFileRepository_UpdateStatus_Success 测试更新状态为成功
func TestUploadedFileRepository_UpdateStatus_Success(t *testing.T) {
	repo := NewInMemoryUploadedFileRepository()

	file := &model.UploadedFile{
		UserID:       1,
		FileName:     "test.csv",
		FilePath:     "/uploads/test.csv",
		FileSize:     1024,
		FileType:     "csv",
		UploadStatus: "pending",
		UploadedAt:   time.Now(),
	}
	repo.Create(file)

	err := repo.UpdateStatus(file.ID, "success", 100, nil)
	if err != nil {
		t.Errorf("UpdateStatus() error = %v", err)
	}

	found, _ := repo.FindByID(file.ID)
	if found.UploadStatus != "success" {
		t.Errorf("UpdateStatus() status = %v, want success", found.UploadStatus)
	}
	if found.RecordsImported != 100 {
		t.Errorf("UpdateStatus() records_imported = %v, want 100", found.RecordsImported)
	}
}

// TestUploadedFileRepository_UpdateStatus_Failed 测试更新状态为失败
func TestUploadedFileRepository_UpdateStatus_Failed(t *testing.T) {
	repo := NewInMemoryUploadedFileRepository()

	file := &model.UploadedFile{
		UserID:       1,
		FileName:     "test.csv",
		FilePath:     "/uploads/test.csv",
		FileSize:     1024,
		FileType:     "csv",
		UploadStatus: "processing",
		UploadedAt:   time.Now(),
	}
	repo.Create(file)

	errMsg := "parse error: invalid format"
	err := repo.UpdateStatus(file.ID, "failed", 0, &errMsg)
	if err != nil {
		t.Errorf("UpdateStatus() error = %v", err)
	}

	found, _ := repo.FindByID(file.ID)
	if found.UploadStatus != "failed" {
		t.Errorf("UpdateStatus() status = %v, want failed", found.UploadStatus)
	}
	if found.ErrorMessage == nil || *found.ErrorMessage != errMsg {
		t.Errorf("UpdateStatus() error_message not set correctly")
	}
}

// 确保 InMemoryUploadedFileRepository 实现了 UploadedFileRepository 接口
var _ repository.UploadedFileRepository = (*InMemoryUploadedFileRepository)(nil)

// TestUploadedFileRepository_Interface 测试接口实现
func TestUploadedFileRepository_Interface(t *testing.T) {
	var repo repository.UploadedFileRepository = NewInMemoryUploadedFileRepository()

	file := &model.UploadedFile{
		UserID:       1,
		FileName:     "test.csv",
		FilePath:     "/uploads/test.csv",
		FileSize:     1024,
		FileType:     "csv",
		UploadStatus: "pending",
		UploadedAt:   time.Now(),
	}

	_ = repo.Create(file)
	_, _ = repo.FindByID(file.ID)
	_, _ = repo.FindByUserID(1)
	_ = repo.UpdateStatus(file.ID, "success", 100, nil)
}
