package transaction

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/KKGo-Software-engineering/workshop-summer/api/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Transaction struct {
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

func (h handler) GetAll(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	rows, err := h.db.QueryContext(ctx, `SELECT * FROM public.transaction WHERE transaction_type='expense'`)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var exs []Transaction
	for rows.Next() {
		var ex Transaction
		err := rows.Scan(&ex.ID, &ex.Date, &ex.Amount, &ex.Category, &ex.TransactionType, &ex.Note, &ex.ImageURL, &ex.SpenderId)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		exs = append(exs, ex)
	}

	return c.JSON(http.StatusOK, exs)
}

func (h handler) Create(c echo.Context) error {
	msg := "bad request body"
	logger := mlog.L(c)
	ctx := c.Request().Context()
	var req Transaction
	if err := c.Bind(&req); err != nil {
		logger.Error(msg, zap.Error(err))
		return c.JSON(http.StatusBadRequest, msg)
	}
	var lastInsertId int64
	err := h.db.QueryRowContext(ctx, `INSERT INTO public.transaction ("date", "amount", "category", "transaction_type", "spender_id") VALUES ($1, $2, $3, $4, $5) RETURNING id;`, req.Date, req.Amount, req.Category, req.TransactionType, req.SpenderId).Scan(&lastInsertId)
	if err != nil {
		fmt.Println("query row error", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	req.ID = lastInsertId
	return c.JSON(http.StatusCreated, req)
}

type Spender struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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
	msg := "bad request body"
	logger := mlog.L(c)
	ctx := c.Request().Context()

	transactionID := c.Param("id") // Get the ID from the URL path
	var req PutTransaction
	if err := c.Bind(&req); err != nil {
		logger.Error(msg, zap.Error(err))
		return c.JSON(http.StatusBadRequest, msg)
	}

	query := `UPDATE public.transaction SET date=$1, amount=$2, category=$3, transaction_type=$4, spender_id=$5, note=$6, image_url=$7 WHERE id=$8`
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
type SpenderIDTransactionResponseSummary struct {
	Summary Summary `json:"summary"`
}

func (h *handler) GetSpenderTransactionSummary(c echo.Context) error {
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
		err := rows.Scan(&t.ID, &t.Date, &t.Amount, &t.Category, &t.TransactionType, &t.Note, &t.ImageURL, &t.SpenderId)
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
	response := SpenderIDTransactionResponseSummary{
		Summary: Summary{
			TotalIncome:    totalIncome,
			TotalExpenses:  totalExpenses,
			CurrentBalance: currentBalance,
		},
	}

	return c.JSON(http.StatusOK, response)
}

type CategoryTransactions struct {
	Category     string        `json:"category"`
	Transactions []Transaction `json:"transactions"`
}

func (h *handler) GetTransactionsGroupedByCategory(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	sqlQuery := `SELECT id, date, amount, category, transaction_type, note, image_url, spender_id FROM public.transaction`

	rows, err := h.db.QueryContext(ctx, sqlQuery)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	transactionsByCategory := make(map[string][]Transaction)

	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.ID, &t.Date, &t.Amount, &t.Category, &t.TransactionType, &t.Note, &t.ImageURL, &t.SpenderId)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		transactionsByCategory[t.Category] = append(transactionsByCategory[t.Category], t)
	}

	return c.JSON(http.StatusOK, transactionsByCategory)
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
		err := rows.Scan(&t.ID, &t.Date, &t.Amount, &t.Category, &t.TransactionType, &t.Note, &t.ImageURL, &t.SpenderId)
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
