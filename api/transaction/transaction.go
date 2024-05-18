package transaction

import (
	"database/sql"
	"net/http"

	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/KKGo-Software-engineering/workshop-summer/api/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Expense struct {
	ID              int64   `json:"id"`
	Date            string  `json:"date"`
	Amount          float64 `json:"amount"`
	Category        string  `json:"category"`
	TransactionType string  `json:"transaction_type"`
	Note            string  `json:"note"`
	ImageURL        string  `json:"image_url"`
	SpenderId       int64   `json:"spender_id"`
}

type handler struct {
	flag config.FeatureFlag
	db   *sql.DB
}

func New(cfg config.FeatureFlag, db *sql.DB) *handler {
	return &handler{cfg, db}
}

const (
	cStmt = `INSERT INTO spender (name, email) VALUES ($1, $2) RETURNING id;`
)

func (h handler) GetAll(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	rows, err := h.db.QueryContext(ctx, `SELECT * FROM transaction WHERE transaction_type='expense'`)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var exs []Expense
	for rows.Next() {
		var ex Expense
		err := rows.Scan(&ex.ID, &ex.Date, &ex.Amount, &ex.Category, &ex.TransactionType, &ex.Note, &ex.ImageURL, &ex.SpenderId)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		exs = append(exs, ex)
	}

	return c.JSON(http.StatusOK, exs)
}
