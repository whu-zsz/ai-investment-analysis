package main

import (
	"context"
	"net/http"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/internal/router"
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/llm"
	"stock-analysis-backend/pkg/logger"
	"stock-analysis-backend/pkg/marketdata"
	"stock-analysis-backend/pkg/news"
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
	marketBoardSnapshotRepo := repository.NewMarketBoardSnapshotRepository(db)
	marketBoardConstituentRepo := repository.NewMarketBoardConstituentRepository(db)
	stockQuoteDetailRepo := repository.NewStockQuoteDetailRepository(db)
	stockKlineRepo := repository.NewStockKlineRepository(db)
	stockMetricRepo := repository.NewStockAnalysisMetricRepository(db)
	revokedTokenRepo := repository.NewRevokedTokenRepository(db)

	llmProvider, err := llm.NewProvider(cfg)
	if err != nil {
		log.Fatal("Failed to initialize llm provider", zap.Error(err), zap.String("provider", cfg.LLM.Provider))
	}
	var marketProvider marketdata.Provider
	var marketRankingProvider marketdata.MarketRankingProvider
	if cfg.Market.Enabled {
		marketProvider, err = marketdata.NewProvider(cfg.Market)
		if err != nil {
			log.Fatal("Failed to initialize market provider", zap.Error(err), zap.String("provider", cfg.Market.Provider))
		}
		log.Info("Market data provider initialized",
			zap.String("provider", cfg.Market.Provider),
			zap.String("symbols", cfg.Market.Symbols),
			zap.Int("snapshot_interval_seconds", cfg.Market.SnapshotInterval),
		)
		if cfg.Market.Provider == "mock" {
			log.Warn("Market data provider is mock; analysis and recommendation features will not use real-time market data")
		}
	} else {
		log.Info("Market data integration disabled")
	}
	marketRankingProvider = marketdata.NewRankingProvider(cfg.Market)

	userService := service.NewUserService(userRepo, revokedTokenRepo, cfg.JWT)
	fileParserService := service.NewFileParserService()
	marketStockService := service.NewMarketStockService(cfg.Market, marketProvider, marketRankingProvider, stockQuoteDetailRepo, stockKlineRepo, marketBoardConstituentRepo)
	portfolioService := service.NewPortfolioService(portfolioRepo, transactionRepo, marketStockService)
	uploadService := service.NewUploadService(uploadedFileRepo, transactionRepo, portfolioService, fileParserService, cfg.Upload)
	transactionService := service.NewTransactionService(transactionRepo, portfolioService)
	marketDataService := service.NewMarketDataService(cfg.Market, marketProvider, marketRankingProvider, marketSnapshotRepo, marketBoardSnapshotRepo, marketBoardConstituentRepo, stockKlineRepo)
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
	recommendationService := service.NewRecommendationService(
		userRepo,
		transactionRepo,
		portfolioRepo,
		marketSnapshotRepo,
		marketDataService,
		llmProvider,
	)
	marketSnapshotService := service.NewMarketSnapshotService(marketSnapshotRepo, marketBoardSnapshotRepo, marketBoardConstituentRepo, stockQuoteDetailRepo, marketDataService)
	newsHTTPClient := &http.Client{Timeout: 12 * time.Second}
	newsService := service.NewNewsService(
		news.NewEastmoneyProvider(newsHTTPClient),
		news.NewGoogleNewsProvider(newsHTTPClient),
		news.NewSinaProvider(newsHTTPClient),
	)
	stockChatService := service.NewStockChatService(marketStockService, newsService, llmProvider, log)
	boardChatService := service.NewBoardChatService(marketSnapshotService, newsService, llmProvider, log)
	marketScheduler := service.NewMarketScheduler(time.Duration(cfg.Market.SnapshotInterval)*time.Second, marketDataService, log)

	userHandler := handler.NewUserHandler(userService)
	uploadHandler := handler.NewUploadHandler(uploadService, cfg.Upload)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)
	analysisHandler := handler.NewAnalysisHandler(aiService, recommendationService, stockChatService, boardChatService)
	marketHandler := handler.NewMarketHandler(marketSnapshotService, marketStockService, newsService)

	router := router.SetupRouter(
		userHandler,
		uploadHandler,
		transactionHandler,
		portfolioHandler,
		analysisHandler,
		marketHandler,
		revokedTokenRepo,
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
