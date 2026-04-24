package repository_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAnalysisTaskTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&model.AnalysisTask{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestAnalysisTaskRepository_Create(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	task := &model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "pending",
		ProgressStage:       "pending",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
	}

	err := repo.Create(task)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if task.ID == 0 {
		t.Error("Create() should set task ID")
	}
}

func TestAnalysisTaskRepository_FindByIDAndUserID(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建测试数据
	task := &model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "pending",
		ProgressStage:       "pending",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
	}
	repo.Create(task)

	// 测试查找
	found, err := repo.FindByIDAndUserID(task.ID, 1)
	if err != nil {
		t.Errorf("FindByIDAndUserID() error = %v", err)
	}

	if found.TaskType != "stock_analysis" {
		t.Errorf("FindByIDAndUserID() TaskType = %v, want stock_analysis", found.TaskType)
	}
}

func TestAnalysisTaskRepository_FindByIDAndUserID_WrongUser(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建测试数据
	task := &model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "pending",
		ProgressStage:       "pending",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
	}
	repo.Create(task)

	// 测试用其他用户ID查找
	_, err := repo.FindByIDAndUserID(task.ID, 999)
	if err == nil {
		t.Error("FindByIDAndUserID() should return error for wrong user")
	}
}

func TestAnalysisTaskRepository_FindByUserID(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建多个任务
	now := time.Now()
	repo.Create(&model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "success",
		ProgressStage:       "completed",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
	})
	repo.Create(&model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "pending",
		ProgressStage:       "pending",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
	})
	repo.Create(&model.AnalysisTask{
		UserID:              2,
		TaskType:            "stock_analysis",
		Status:              "pending",
		ProgressStage:       "pending",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
	})

	// 测试查找用户1的任务
	tasks, total, err := repo.FindByUserID(1, "", 10, 0)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("FindByUserID() returned %d tasks, want 2", len(tasks))
	}
	if total != 2 {
		t.Errorf("FindByUserID() total = %d, want 2", total)
	}
}

func TestAnalysisTaskRepository_FindByUserID_WithStatus(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建多个任务
	now := time.Now()
	repo.Create(&model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "success",
		ProgressStage:       "completed",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
	})
	repo.Create(&model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "pending",
		ProgressStage:       "pending",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
	})

	// 测试按状态筛选
	tasks, total, err := repo.FindByUserID(1, "pending", 10, 0)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("FindByUserID() returned %d tasks, want 1", len(tasks))
	}
	if total != 1 {
		t.Errorf("FindByUserID() total = %d, want 1", total)
	}
}

func TestAnalysisTaskRepository_FindByUserID_Pagination(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建多个任务
	now := time.Now()
	for i := 0; i < 15; i++ {
		repo.Create(&model.AnalysisTask{
			UserID:              1,
			TaskType:            "stock_analysis",
			Status:              "success",
			ProgressStage:       "completed",
			AnalysisPeriodStart: now,
			AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
		})
	}

	// 测试分页
	tasks, total, err := repo.FindByUserID(1, "", 10, 0)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}

	if len(tasks) != 10 {
		t.Errorf("FindByUserID() returned %d tasks, want 10", len(tasks))
	}
	if total != 15 {
		t.Errorf("FindByUserID() total = %d, want 15", total)
	}
}

func TestAnalysisTaskRepository_FindByUserID_Empty(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	tasks, total, err := repo.FindByUserID(999, "", 10, 0)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("FindByUserID() returned %d tasks, want 0", len(tasks))
	}
	if total != 0 {
		t.Errorf("FindByUserID() total = %d, want 0", total)
	}
}

func TestAnalysisTaskRepository_HasRunningTask_True(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建运行中的任务
	now := time.Now()
	repo.Create(&model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "processing",
		ProgressStage:       "generating_stock_reports",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
	})

	// 测试是否有运行中的任务
	hasRunning, err := repo.HasRunningTask(1, "stock_analysis")
	if err != nil {
		t.Errorf("HasRunningTask() error = %v", err)
	}

	if !hasRunning {
		t.Error("HasRunningTask() should return true for running task")
	}
}

func TestAnalysisTaskRepository_HasRunningTask_False(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建已完成的任务
	now := time.Now()
	repo.Create(&model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "success",
		ProgressStage:       "completed",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
	})

	// 测试是否有运行中的任务
	hasRunning, err := repo.HasRunningTask(1, "stock_analysis")
	if err != nil {
		t.Errorf("HasRunningTask() error = %v", err)
	}

	if hasRunning {
		t.Error("HasRunningTask() should return false for completed task")
	}
}

func TestAnalysisTaskRepository_UpdateProgress_Status(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建测试数据
	task := &model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "pending",
		ProgressStage:       "pending",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
	}
	repo.Create(task)

	// 更新状态
	err := repo.UpdateProgress(task.ID, "processing", "collecting_transactions", nil, nil, nil, nil)
	if err != nil {
		t.Errorf("UpdateProgress() error = %v", err)
	}

	// 验证更新
	found, _ := repo.FindByIDAndUserID(task.ID, 1)
	if found.Status != "processing" {
		t.Errorf("UpdateProgress() status = %v, want processing", found.Status)
	}
	if found.ProgressStage != "collecting_transactions" {
		t.Errorf("UpdateProgress() stage = %v, want collecting_transactions", found.ProgressStage)
	}
}

func TestAnalysisTaskRepository_UpdateProgress_WithError(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建测试数据
	task := &model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "processing",
		ProgressStage:       "generating_stock_reports",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
	}
	repo.Create(task)

	// 更新为失败状态
	errMsg := "AI service error"
	finishedAt := time.Now()
	err := repo.UpdateProgress(task.ID, "failed", "completed", &errMsg, nil, nil, &finishedAt)
	if err != nil {
		t.Errorf("UpdateProgress() error = %v", err)
	}

	// 验证更新
	found, _ := repo.FindByIDAndUserID(task.ID, 1)
	if found.Status != "failed" {
		t.Errorf("UpdateProgress() status = %v, want failed", found.Status)
	}
	if found.ErrorMessage == nil || *found.ErrorMessage != errMsg {
		t.Errorf("UpdateProgress() error_message not set correctly")
	}
}

func TestAnalysisTaskRepository_UpdateProgress_WithReportID(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	repo := repository.NewAnalysisTaskRepository(db)

	// 创建测试数据
	task := &model.AnalysisTask{
		UserID:              1,
		TaskType:            "stock_analysis",
		Status:              "processing",
		ProgressStage:       "persisting_report",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
	}
	repo.Create(task)

	// 更新为成功状态并设置报告ID
	reportID := uint64(123)
	finishedAt := time.Now()
	err := repo.UpdateProgress(task.ID, "success", "completed", nil, &reportID, nil, &finishedAt)
	if err != nil {
		t.Errorf("UpdateProgress() error = %v", err)
	}

	// 验证更新
	found, _ := repo.FindByIDAndUserID(task.ID, 1)
	if found.Status != "success" {
		t.Errorf("UpdateProgress() status = %v, want success", found.Status)
	}
	if found.ResultReportID == nil || *found.ResultReportID != reportID {
		t.Errorf("UpdateProgress() result_report_id not set correctly")
	}
}

func TestAnalysisTaskRepository_Interface(t *testing.T) {
	db := setupAnalysisTaskTestDB(t)
	var _ repository.AnalysisTaskRepository = repository.NewAnalysisTaskRepository(db)
}
