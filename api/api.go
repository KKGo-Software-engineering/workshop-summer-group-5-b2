package api

import (
	"database/sql"

	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/KKGo-Software-engineering/workshop-summer/api/eslip"
	"github.com/KKGo-Software-engineering/workshop-summer/api/health"
	"github.com/KKGo-Software-engineering/workshop-summer/api/mlog"
	"github.com/KKGo-Software-engineering/workshop-summer/api/spender"
	"github.com/KKGo-Software-engineering/workshop-summer/api/transaction"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	*echo.Echo
}

func New(db *sql.DB, cfg config.Config, logger *zap.Logger) *Server {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(mlog.Middleware(logger))

	v1 := e.Group("/api/v1")

	v1.GET("/slow", health.Slow)
	v1.GET("/health", health.Check(db))
	v1.POST("/upload", eslip.Upload)

	handleE := transaction.New(cfg.FeatureFlag, db)
	v1.GET("/expenses", handleE.GetAll)

	v1.Use(middleware.BasicAuth(AuthCheck))

	{
		h := spender.New(cfg.FeatureFlag, db)
		v1.GET("/spenders", h.GetAll)
		v1.POST("/spenders", h.Create)
		v1.GET("/spenders/:id", h.GetSpenderByID)
		v1.GET("/categories", h.GetAllCategories)
	}
	{
		h := transaction.New(cfg.FeatureFlag, db)
		v1.POST("/transactions", h.Create)
		v1.PUT("/transaction/:id", h.PutTransaction)
		v1.GET("/spenders/:id/transactions", h.GetSpenderTransactions)
		v1.GET("/spenders/:id/transactions/summary", h.GetSpenderTransactionSummary)
		v1.GET("/categorize", h.GetTransactionsGroupedByCategory)
	}

	return &Server{e}
}
