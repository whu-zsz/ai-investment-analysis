package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"stock-analysis-backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	if strings.EqualFold(strings.TrimSpace(cfg.Host), "localhost") {
		logger.Default.LogMode(logger.Warn).Warn(context.Background(), "DB_HOST=localhost may resolve to ::1 locally; prefer 127.0.0.1 when running the backend on the host")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database %s@%s:%s/%s: %w", cfg.User, cfg.Host, cfg.Port, cfg.DBName, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := waitForDB(sqlDB, cfg); err != nil {
		return nil, err
	}

	return db, nil
}

func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func AutoMigrate(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.User{}) {
		if err := db.AutoMigrate(&model.User{}); err != nil {
			return err
		}
	}
	if !db.Migrator().HasTable(&model.StockQuoteDetail{}) {
		if err := db.AutoMigrate(&model.StockQuoteDetail{}); err != nil {
			return err
		}
	}

	if err := db.AutoMigrate(
		&model.Transaction{},
		&model.Portfolio{},
		&model.AnalysisTask{},
		&model.StockAnalysisMetric{},
		&model.AnalysisReport{},
		&model.AnalysisReportItem{},
		&model.UploadedFile{},
		&model.MarketSnapshot{},
		&model.MarketBoardSnapshot{},
		&model.MarketBoardConstituent{},
		&model.StockKlineBar{},
	); err != nil {
		return err
	}

	if err := ensureAnalysisReportColumns(db); err != nil {
		return err
	}

	return ensureRevokedTokenSchema(db)
}

func waitForDB(sqlDB interface{ PingContext(context.Context) error }, cfg *DatabaseConfig) error {
	const maxAttempts = 10
	const delay = 2 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := sqlDB.PingContext(ctx)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err

		if attempt < maxAttempts {
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("database not ready after %d attempts (%s@%s:%s/%s): %w", maxAttempts, cfg.User, cfg.Host, cfg.Port, cfg.DBName, lastErr)
}

func ensureAnalysisReportColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.AnalysisReport{}) {
		return nil
	}

	if err := ensureColumnType(db, "ai_analysis_reports", "profit_rate", "DECIMAL(10,4) NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnType(db, "ai_analysis_reports", "total_investment", "DECIMAL(15,2) NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnType(db, "ai_analysis_reports", "total_profit", "DECIMAL(15,2) NOT NULL"); err != nil {
		return err
	}

	return nil
}

func ensureColumnType(db *gorm.DB, tableName, columnName, targetType string) error {
	if !db.Migrator().HasColumn(tableName, columnName) {
		return nil
	}

	columnTypes, err := db.Migrator().ColumnTypes(tableName)
	if err != nil {
		return err
	}

	for _, columnType := range columnTypes {
		if !strings.EqualFold(columnType.Name(), columnName) {
			continue
		}

		currentType := strings.ToUpper(columnType.DatabaseTypeName())
		if strings.EqualFold(columnName, "profit_rate") && currentType == "DECIMAL" {
			if precision, scale, ok := columnType.DecimalSize(); ok && precision >= 10 && scale >= 4 {
				return nil
			}
		}
		if (columnName == "total_investment" || columnName == "total_profit") && currentType == "DECIMAL" {
			if precision, scale, ok := columnType.DecimalSize(); ok && precision >= 15 && scale >= 2 {
				return nil
			}
		}

		return db.Exec(fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s", tableName, columnName, targetType)).Error
	}

	return nil
}

func ensureRevokedTokenSchema(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.RevokedToken{}) {
		if err := db.Exec(`
			CREATE TABLE IF NOT EXISTS revoked_tokens (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				user_id BIGINT UNSIGNED NOT NULL,
				jti VARCHAR(64) NOT NULL,
				token_expires_at DATETIME(3) NOT NULL,
				revoked_at DATETIME(3) NOT NULL,
				reason VARCHAR(20) NOT NULL DEFAULT 'logout',
				created_at DATETIME(3) NULL,
				PRIMARY KEY (id),
				KEY idx_revoked_tokens_user_id (user_id),
				KEY idx_revoked_tokens_token_expires_at (token_expires_at)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
		`).Error; err != nil {
			return err
		}
	}

	return ensureRevokedTokenIndexes(db)
}

func ensureRevokedTokenIndexes(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.RevokedToken{}) {
		return nil
	}

	type uniqueIndexStat struct {
		IndexName string
		NonUnique int
	}

	var stats []uniqueIndexStat
	err := db.Raw(`
		SELECT index_name, non_unique
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
		  AND table_name = ?
		  AND column_name = ?
	`, "revoked_tokens", "jti").Scan(&stats).Error
	if err != nil {
		return err
	}

	for _, stat := range stats {
		if stat.NonUnique == 0 {
			return nil
		}
	}

	return db.Exec("CREATE UNIQUE INDEX idx_revoked_tokens_jti ON revoked_tokens (jti)").Error
}
