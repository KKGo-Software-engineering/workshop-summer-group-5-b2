package transactions

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/kkgo-software-engineering/workshop/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Spender struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type handler struct {
	flag config.FeatureFlag
	db   *sql.DB
}

func New(cfg config.FeatureFlag, db *sql.DB) *handler {
	return &handler{cfg, db}
}

type requestIncome struct {
	Date      time.Time `json:"date"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	SpenderID int64     `json:"spender_id"`
}

type PutTransaction struct {
	Date            time.Time `json:"date"`
	Amount          int       `json:"amount"`
	Category        string    `json:"category"`
	TransactionType string    `json:"transaction_type"`
	Note            string    `json:"note"`
	ImageUrl        string    `json:"image_url"`
	SpenderId       int       `json:"spender_id"`
}

func (h handler) PutTransaction(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	transactionID := c.Param("id") // Get the ID from the URL path
	var req PutTransaction
	if err := c.Bind(&req); err != nil {
		logger.Error("bad request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, "bad request body")
	}

	query := `UPDATE "transaction" SET date=$1, amount=$2, category=$3, transaction_type=$4, spender_id=$5, note=$6, image_url=$7 WHERE id=$8`
	_, err := h.db.ExecContext(ctx, query, req.Date, req.Amount, req.Category, req.TransactionType, req.SpenderId, req.Note, req.ImageUrl, transactionID)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// Confirm the update was successful
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Transaction updated successfully",
	})
}

// Define the struct based on your previous structure
type Transaction struct {
	Id              int       `json:"id"`
	Date            time.Time `json:"date"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	TransactionType string    `json:"transaction_type"`
	Note            string    `json:"note"`
	ImageUrl        string    `json:"image_url"`
	SpenderId       int       `json:"spender_id"`
}

type Summary struct {
	TotalIncome    float64 `json:"total_income"`
	TotalExpenses  float64 `json:"total_expenses"`
	CurrentBalance float64 `json:"current_balance"`
}

type Pagination struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	PerPage     int `json:"per_page"`
}

type SpenderIDTransactionResponse struct {
	Transactions []Transaction `json:"transactions"`
	Summary      Summary       `json:"summary"`
	Pagination   Pagination    `json:"pagination"`
}

func (h *handler) GetSpenderTransactions(c echo.Context) error {
	sqlQuery := `SELECT id, date, amount, category, transaction_type, note, image_url, spender_id FROM public.transaction WHERE spender_id=$1`

	spenderID := c.Param("id") // Get the spender ID from the URL parameter

	// Placeholder for the database connection
	db := h.db // Assuming there is a db field in the handler struct for the database connection

	// Querying the database for transactions related to the spender
	rows, err := db.Query(sqlQuery, spenderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		return err
	}

	defer rows.Close()

	var transactions []Transaction
	var totalIncome, totalExpenses float64

	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.Id, &t.Date, &t.Amount, &t.Category, &t.TransactionType, &t.Note, &t.ImageUrl, &t.SpenderId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		transactions = append(transactions, t)

		// Calculate totals for summary
		if t.TransactionType == "income" {
			totalIncome += t.Amount
		} else {
			totalExpenses += t.Amount
		}
	}

	// Calculate the current balance
	currentBalance := totalIncome - totalExpenses

	// Construct the response
	response := SpenderIDTransactionResponse{
		Transactions: transactions,
		Summary: Summary{
			TotalIncome:    totalIncome,
			TotalExpenses:  totalExpenses,
			CurrentBalance: currentBalance,
		},
		Pagination: Pagination{
			CurrentPage: 1,
			TotalPages:  1,
			PerPage:     len(transactions),
		},
	}

	return c.JSON(http.StatusOK, response)
}
