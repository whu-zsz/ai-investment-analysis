package router

import (
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/middleware"
	"stock-analysis-backend/internal/repository"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	userHandler *handler.UserHandler,
	uploadHandler *handler.UploadHandler,
	transactionHandler *handler.TransactionHandler,
	portfolioHandler *handler.PortfolioHandler,
	analysisHandler *handler.AnalysisHandler,
	marketHandler *handler.MarketHandler,
	revokedTokenRepo repository.RevokedTokenRepository,
	jwtSecret string,
) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORS())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret, revokedTokenRepo))
		{
			user := protected.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PUT("/profile", userHandler.UpdateProfile)
			}

			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", userHandler.Logout)
			}

			upload := protected.Group("/upload")
			{
				upload.POST("", uploadHandler.UploadFile)
				upload.GET("/history", uploadHandler.GetUploadHistory)
			}

			transactions := protected.Group("/transactions")
			{
				transactions.POST("", transactionHandler.CreateTransaction)
				transactions.GET("", transactionHandler.GetTransactions)
				transactions.GET("/stats", transactionHandler.GetTransactionStats)
				transactions.GET("/:id", transactionHandler.GetTransaction)
				transactions.PUT("/:id", transactionHandler.UpdateTransaction)
				transactions.DELETE("/:id", transactionHandler.DeleteTransaction)
			}

			portfolios := protected.Group("/portfolios")
			{
				portfolios.GET("", portfolioHandler.GetPortfolios)
			}

			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/market-snapshot", marketHandler.GetDashboardSnapshot)
				dashboard.GET("/market-breadth", marketHandler.GetDashboardMarketBreadth)
			}

			market := protected.Group("/market")
			{
				market.GET("/snapshots/latest", marketHandler.GetLatestSnapshots)
				market.GET("/snapshots/history", marketHandler.GetSnapshotHistory)
				market.GET("/stocks/search", marketHandler.SearchStocks)
				market.GET("/boards/:boardType/:code", marketHandler.GetBoardDetail)
				market.GET("/boards/:boardType/:code/news", marketHandler.GetBoardNews)
				market.GET("/stocks/:symbol/detail", marketHandler.GetStockDetail)
				market.GET("/stocks/:symbol/profile", marketHandler.GetStockProfile)
				market.GET("/stocks/:symbol/news", marketHandler.GetStockNews)
				market.GET("/stocks/:symbol/kline", marketHandler.GetStockKlines)
			}

			analysis := protected.Group("/analysis")
			{
				analysis.POST("/tasks", analysisHandler.CreateTask)
				analysis.GET("/tasks", analysisHandler.GetTasks)
				analysis.GET("/tasks/:id", analysisHandler.GetTask)
				analysis.POST("/summary", analysisHandler.GenerateSummary)
				analysis.GET("/reports", analysisHandler.GetReports)
				analysis.GET("/reports/:id", analysisHandler.GetReportDetail)
				analysis.GET("/reports/:id/pdf", analysisHandler.ExportReportPDF)
				analysis.POST("/stock-chat", analysisHandler.StockChat)
				analysis.POST("/stock-chat/stream", analysisHandler.StockChatStream)
				analysis.POST("/board-chat", analysisHandler.BoardChat)
				analysis.POST("/board-chat/stream", analysisHandler.BoardChatStream)
				analysis.GET("/chat-contexts/:id", analysisHandler.GetChatContext)
				analysis.POST("/recommendation-chat", analysisHandler.RecommendationChat)
				analysis.POST("/recommendation-chat/stream", analysisHandler.RecommendationChatStream)
				analysis.GET("/recommendation-chat/contexts/:id", analysisHandler.GetRecommendationChatContext)
				analysis.GET("/recommendation-chat/reports/:id/context", analysisHandler.GetRecommendationChatContextByReport)
				analysis.GET("/candidates", analysisHandler.GetCandidates)
				analysis.GET("/recommendations", analysisHandler.GetRecommendations)
			}
		}
	}

	return router
}
