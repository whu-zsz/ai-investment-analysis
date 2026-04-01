package main

import (
	"context"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/internal/router"
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/deepseek"
	"stock-analysis-backend/pkg/logger"
	"stock-analysis-backend/pkg/marketdata"
	"time"

	"go.uber.org/zap"
)

// @title Stock Analysis API
// @version 1.0
// @description 基于AI大模型的投资记录分析与预测系统后端API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@stock-analysis.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 2. 初始化日志
	log := logger.InitLogger()
	defer logger.Sync(log)

	// 3. 初始化数据库
	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect database", zap.Error(err))
	}
	defer config.CloseDB(db)

	// 4. 自动迁移
	if err := config.AutoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}

	log.Info("Database migration completed successfully")

	// 5. 初始化Repository
	userRepo := repository.NewUserRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	analysisReportRepo := repository.NewAnalysisReportRepository(db)
	uploadedFileRepo := repository.NewUploadedFileRepository(db)
	marketSnapshotRepo := repository.NewMarketSnapshotRepository(db)

	// 6. 初始化客户端
	deepseekClient := deepseek.NewClient(cfg.Deepseek.APIKey, cfg.Deepseek.APIURL)
	marketProvider, err := marketdata.NewProvider(cfg.Market)
	if err != nil {
		log.Fatal("Failed to initialize market provider", zap.Error(err))
	}

	// 7. 初始化Service
	userService := service.NewUserService(userRepo, cfg.JWT)
	fileParserService := service.NewFileParserService()
	uploadService := service.NewUploadService(uploadedFileRepo, transactionRepo, fileParserService, cfg.Upload)
	portfolioService := service.NewPortfolioService(portfolioRepo, transactionRepo)
	transactionService := service.NewTransactionService(transactionRepo, portfolioService)
	aiService := service.NewAIService(analysisReportRepo, transactionRepo, deepseekClient)
	marketDataService := service.NewMarketDataService(cfg.Market, marketProvider, marketSnapshotRepo)
	marketSnapshotService := service.NewMarketSnapshotService(marketSnapshotRepo)
	marketScheduler := service.NewMarketScheduler(time.Duration(cfg.Market.SnapshotInterval)*time.Second, marketDataService, log)

	// 8. 初始化Handler
	userHandler := handler.NewUserHandler(userService)
	uploadHandler := handler.NewUploadHandler(uploadService, cfg.Upload)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)
	analysisHandler := handler.NewAnalysisHandler(aiService)
	marketHandler := handler.NewMarketHandler(marketSnapshotService)

	// 9. 设置路由
	router := router.SetupRouter(
		userHandler,
		uploadHandler,
		transactionHandler,
		portfolioHandler,
		analysisHandler,
		marketHandler,
		cfg.JWT.Secret,
	)

	if cfg.Market.Enabled {
		marketScheduler.Start(context.Background())
	}

	// 10. 启动服务器
	log.Info("Server starting", zap.String("port", cfg.Server.Port))
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
