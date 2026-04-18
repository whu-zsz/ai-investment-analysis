package main

import (
	"context"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/internal/router"
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/llm"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	log := logger.InitLogger()
	defer logger.Sync(log)

	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect database", zap.Error(err))
	}
	defer config.CloseDB(db)

	if err := config.AutoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}
	log.Info("Database migration completed successfully")

	userRepo := repository.NewUserRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	analysisTaskRepo := repository.NewAnalysisTaskRepository(db)
	analysisReportRepo := repository.NewAnalysisReportRepository(db)
	analysisReportItemRepo := repository.NewAnalysisReportItemRepository(db)
	uploadedFileRepo := repository.NewUploadedFileRepository(db)
	marketSnapshotRepo := repository.NewMarketSnapshotRepository(db)
	stockMetricRepo := repository.NewStockAnalysisMetricRepository(db)

	llmProvider, err := llm.NewProvider(cfg)
	if err != nil {
		log.Fatal("Failed to initialize llm provider", zap.Error(err), zap.String("provider", cfg.LLM.Provider))
	}
	marketProvider, err := marketdata.NewProvider(cfg.Market)
	if err != nil {
		log.Fatal("Failed to initialize market provider", zap.Error(err))
	}

	userService := service.NewUserService(userRepo, cfg.JWT)
	fileParserService := service.NewFileParserService()
	uploadService := service.NewUploadService(uploadedFileRepo, transactionRepo, fileParserService, cfg.Upload)
	portfolioService := service.NewPortfolioService(portfolioRepo, transactionRepo)
	transactionService := service.NewTransactionService(transactionRepo, portfolioService)
	marketDataService := service.NewMarketDataService(cfg.Market, marketProvider, marketSnapshotRepo)
	stockMetricService := service.NewStockAnalysisMetricService(stockMetricRepo, transactionRepo, marketSnapshotRepo, marketDataService)
	aiService := service.NewAIService(
		analysisTaskRepo,
		analysisReportRepo,
		analysisReportItemRepo,
		transactionRepo,
		stockMetricService,
		llmProvider,
		log,
	)
	marketSnapshotService := service.NewMarketSnapshotService(marketSnapshotRepo)
	marketScheduler := service.NewMarketScheduler(time.Duration(cfg.Market.SnapshotInterval)*time.Second, marketDataService, log)

	userHandler := handler.NewUserHandler(userService)
	uploadHandler := handler.NewUploadHandler(uploadService, cfg.Upload)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)
	analysisHandler := handler.NewAnalysisHandler(aiService)
	marketHandler := handler.NewMarketHandler(marketSnapshotService)

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

	log.Info("Server starting", zap.String("port", cfg.Server.Port))
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
