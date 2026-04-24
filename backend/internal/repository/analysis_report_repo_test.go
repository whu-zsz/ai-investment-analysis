package repository_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAnalysisReportTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&model.AnalysisReport{}, &model.AnalysisReportItem{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestAnalysisReportRepository_Create(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	report := &model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Test Report",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "This is a test summary",
		TotalInvestment:     decimal.NewFromInt(100000),
		TotalProfit:         decimal.NewFromInt(10000),
		ProfitRate:          decimal.NewFromFloat(10.0),
	}

	err := repo.Create(report)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if report.ID == 0 {
		t.Error("Create() should set report ID")
	}
}

func TestAnalysisReportRepository_CreateWithItems(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	report := &model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Test Report",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "This is a test summary",
		TotalInvestment:     decimal.NewFromInt(100000),
		TotalProfit:         decimal.NewFromInt(10000),
		ProfitRate:          decimal.NewFromFloat(10.0),
	}

	items := []model.AnalysisReportItem{
		{
			UserID:            1,
			Symbol:            "600519.SH",
			AssetName:         "贵州茅台",
			TradeCount:        1,
			BuyCount:          1,
			SellCount:         0,
			BuyAmount:         decimal.NewFromInt(180000),
			SellAmount:        decimal.Zero,
			NetQuantity:       decimal.NewFromInt(100),
			RealizedProfit:    decimal.Zero,
			EndingPositionQty: decimal.NewFromInt(100),
			EndingAvgCost:     decimal.NewFromInt(1800),
			LatestPrice:       decimal.NewFromInt(1900),
			LatestMarketValue: decimal.NewFromInt(190000),
			UnrealizedProfit:  decimal.NewFromInt(10000),
			TotalProfit:       decimal.NewFromInt(10000),
			RiskLevel:         "medium",
			AnalysisText:      "Test analysis",
			Recommendation:    "hold",
		},
	}

	err := repo.CreateWithItems(report, items)
	if err != nil {
		t.Errorf("CreateWithItems() error = %v", err)
	}

	if report.ID == 0 {
		t.Error("CreateWithItems() should set report ID")
	}

	// 验证 items 的 report_id 被设置
	if items[0].ReportID != report.ID {
		t.Errorf("CreateWithItems() items[0].ReportID = %v, want %v", items[0].ReportID, report.ID)
	}
}

func TestAnalysisReportRepository_FindByID(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	// 创建测试数据
	report := &model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Test Report",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "This is a test summary",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	}
	repo.Create(report)

	// 测试查找
	found, err := repo.FindByID(report.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}

	if found.ReportTitle != "Test Report" {
		t.Errorf("FindByID() ReportTitle = %v, want Test Report", found.ReportTitle)
	}
}

func TestAnalysisReportRepository_FindByID_NotFound(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	_, err := repo.FindByID(999)
	if err == nil {
		t.Error("FindByID() should return error for non-existent ID")
	}
}

func TestAnalysisReportRepository_FindByIDAndUserID(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	// 创建测试数据
	report := &model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Test Report",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "This is a test summary",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	}
	repo.Create(report)

	// 测试查找
	found, err := repo.FindByIDAndUserID(report.ID, 1)
	if err != nil {
		t.Errorf("FindByIDAndUserID() error = %v", err)
	}

	if found.UserID != 1 {
		t.Errorf("FindByIDAndUserID() UserID = %v, want 1", found.UserID)
	}
}

func TestAnalysisReportRepository_FindByIDAndUserID_WrongUser(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	// 创建测试数据
	report := &model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Test Report",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "This is a test summary",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	}
	repo.Create(report)

	// 测试用其他用户ID查找
	_, err := repo.FindByIDAndUserID(report.ID, 999)
	if err == nil {
		t.Error("FindByIDAndUserID() should return error for wrong user")
	}
}

func TestAnalysisReportRepository_FindByTaskID(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	// 创建测试数据
	taskID := uint64(123)
	report := &model.AnalysisReport{
		TaskID:              &taskID,
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Test Report",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "This is a test summary",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	}
	repo.Create(report)

	// 测试查找
	found, err := repo.FindByTaskID(taskID)
	if err != nil {
		t.Errorf("FindByTaskID() error = %v", err)
	}

	if found.ReportTitle != "Test Report" {
		t.Errorf("FindByTaskID() ReportTitle = %v, want Test Report", found.ReportTitle)
	}
}

func TestAnalysisReportRepository_FindByUserID(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	// 创建多个报告
	now := time.Now()
	repo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Report 1",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "Summary 1",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	})
	repo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "risk",
		ReportTitle:         "Report 2",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
		RiskLevel:           "high",
		SummaryText:         "Summary 2",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	})
	repo.Create(&model.AnalysisReport{
		UserID:              2,
		ReportType:          "summary",
		ReportTitle:         "Report 3",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
		RiskLevel:           "low",
		SummaryText:         "Summary 3",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	})

	// 测试查找用户1的报告
	reports, err := repo.FindByUserID(1, "", 10)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}

	if len(reports) != 2 {
		t.Errorf("FindByUserID() returned %d reports, want 2", len(reports))
	}
}

func TestAnalysisReportRepository_FindByUserID_WithType(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	// 创建多个报告
	now := time.Now()
	repo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Report 1",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "Summary 1",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	})
	repo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "risk",
		ReportTitle:         "Report 2",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
		RiskLevel:           "high",
		SummaryText:         "Summary 2",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	})

	// 测试按类型筛选
	reports, err := repo.FindByUserID(1, "risk", 10)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}

	if len(reports) != 1 {
		t.Errorf("FindByUserID() returned %d reports, want 1", len(reports))
	}
}

func TestAnalysisReportRepository_FindLatestByUser(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	// 创建多个报告
	now := time.Now()
	repo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Old Report",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "Old Summary",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	})
	time.Sleep(time.Millisecond * 100) // 确保时间不同
	repo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "New Report",
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now.AddDate(0, 1, 0),
		RiskLevel:           "low",
		SummaryText:         "New Summary",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	})

	// 测试获取最新报告
	report, err := repo.FindLatestByUser(1, "summary")
	if err != nil {
		t.Errorf("FindLatestByUser() error = %v", err)
	}

	if report.ReportTitle != "New Report" {
		t.Errorf("FindLatestByUser() ReportTitle = %v, want New Report", report.ReportTitle)
	}
}

func TestAnalysisReportRepository_Delete(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	repo := repository.NewAnalysisReportRepository(db)

	// 创建测试数据
	report := &model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "Test Report",
		AnalysisPeriodStart: time.Now(),
		AnalysisPeriodEnd:   time.Now().AddDate(0, 1, 0),
		RiskLevel:           "medium",
		SummaryText:         "This is a test summary",
		TotalInvestment:     decimal.Zero,
		TotalProfit:         decimal.Zero,
		ProfitRate:          decimal.Zero,
	}
	repo.Create(report)

	// 删除
	err := repo.Delete(report.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// 验证删除
	_, err = repo.FindByID(report.ID)
	if err == nil {
		t.Error("Delete() should remove report")
	}
}

func TestAnalysisReportRepository_Interface(t *testing.T) {
	db := setupAnalysisReportTestDB(t)
	var _ repository.AnalysisReportRepository = repository.NewAnalysisReportRepository(db)
}
