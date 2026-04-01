package router

import (
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/middleware"

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
	jwtSecret string,
) *gin.Engine {
	router := gin.Default()

	// CORS中间件
	router.Use(middleware.CORS())

	// Swagger文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// 公开接口（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// 需要认证的接口
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			// 用户相关
			user := protected.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PUT("/profile", userHandler.UpdateProfile)
			}

			// 文件上传
			upload := protected.Group("/upload")
			{
				upload.POST("", uploadHandler.UploadFile)
				upload.GET("/history", uploadHandler.GetUploadHistory)
			}

			// 交易记录
			transactions := protected.Group("/transactions")
			{
				transactions.POST("", transactionHandler.CreateTransaction)
				transactions.GET("", transactionHandler.GetTransactions)
				transactions.GET("/stats", transactionHandler.GetTransactionStats)
				transactions.DELETE("/:id", transactionHandler.DeleteTransaction)
			}

			// 持仓管理
			portfolios := protected.Group("/portfolios")
			{
				portfolios.GET("", portfolioHandler.GetPortfolios)
			}

			// Dashboard 市场快照
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/market-snapshot", marketHandler.GetDashboardSnapshot)
			}

			// 市场快照
			market := protected.Group("/market")
			{
				market.GET("/snapshots/latest", marketHandler.GetLatestSnapshots)
				market.GET("/snapshots/history", marketHandler.GetSnapshotHistory)
			}

			// AI分析
			analysis := protected.Group("/analysis")
			{
				analysis.POST("/summary", analysisHandler.GenerateSummary)
				analysis.GET("/reports", analysisHandler.GetReports)
			}
		}
	}

	return router
}
