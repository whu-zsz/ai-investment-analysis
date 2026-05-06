//go:build integration
// +build integration

package repository_test

import (
	"fmt"
	"testing"

	"stock-analysis-backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDBConfig 测试数据库配置
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// DefaultTestDBConfig 默认测试数据库配置
func DefaultTestDBConfig() TestDBConfig {
	return TestDBConfig{
		Host:     "localhost",
		Port:     "3306",
		User:     "root",
		Password: "soyorin114",
		DBName:   "stock_analysis_test",
	}
}

// SetupTestDB 创建测试数据库连接
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	cfg := DefaultTestDBConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&model.User{},
		&model.Transaction{},
		&model.Portfolio{},
		&model.UploadedFile{},
		&model.AnalysisTask{},
		&model.AnalysisReport{},
		&model.AnalysisReportItem{},
		&model.MarketSnapshot{},
		&model.StockAnalysisMetric{},
	)
	if err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	return db
}

// CleanupDB 清理测试数据
func CleanupDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	// 按照外键依赖顺序删除
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	db.Exec("TRUNCATE TABLE stock_analysis_metrics")
	db.Exec("TRUNCATE TABLE market_snapshots")
	db.Exec("TRUNCATE TABLE ai_analysis_report_items")
	db.Exec("TRUNCATE TABLE ai_analysis_reports")
	db.Exec("TRUNCATE TABLE ai_analysis_tasks")
	db.Exec("TRUNCATE TABLE uploaded_files")
	db.Exec("TRUNCATE TABLE portfolios")
	db.Exec("TRUNCATE TABLE transactions")
	db.Exec("TRUNCATE TABLE users")
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
}
