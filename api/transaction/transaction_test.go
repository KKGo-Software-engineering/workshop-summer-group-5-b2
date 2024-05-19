package transaction

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetAllSpender(t *testing.T) {
	t.Run("get all spender succesfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "date", "amount", "category", "transaction_type", "note", "image_url", "spender_id"}).
			AddRow(1, "2024-05-18 08:45:24.119432+00", "0.0", "Food", "expense", "", "", "1")
		mock.ExpectQuery(`SELECT * FROM transaction WHERE transaction_type='expense'`).WillReturnRows(rows)

		h := New(config.FeatureFlag{}, db)
		err := h.GetAll(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `[{"id": 1,
		"date": "2024-05-18 08:45:24.119432+00",
		"amount": 0.0,
		"category":"Food",
		"transaction_type":"expense",
		"note":"",
		"image_url":"",
		"spender_id":1}
		]`, rec.Body.String())
	})

	// t.Run("get all spender failed on database", func(t *testing.T) {
	// 	e := echo.New()
	// 	defer e.Close()

	// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	// 	defer db.Close()

	// 	mock.ExpectQuery(`SELECT id, name, email FROM spender`).WillReturnError(assert.AnError)

	// 	h := New(config.FeatureFlag{}, db)
	// 	err := h.GetAll(c)

	// 	assert.NoError(t, err)
	// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	// })
}

func TestCreate(t *testing.T) {
	t.Run("create transaction succesfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"date":"2024-05-18T15:00:37.557628+07:00","amount":200.99,"category":"refund","transaction_type":"income","spender_id":2}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()
		cStmt := `INSERT INTO transaction ("date", "amount", "category", "transaction_type", "spender_id") VALUES ($1, $2, $3, $4, $5) RETURNING id;`
		row := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery(cStmt).WithArgs("2024-05-18T15:00:37.557628+07:00", 200.99, "refund", "income", 2).WillReturnRows(row)
		cfg := config.FeatureFlag{EnableCreateSpender: true}

		h := New(cfg, db)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"id":1,"date":"2024-05-18T15:00:37.557628+07:00","amount":200.99,"category":"refund","transaction_type":"income","note":"","image_url":"","spender_id":2}`, rec.Body.String())
	})
}

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

func TestPutTransaction(t *testing.T) {
	query := `UPDATE "transaction" SET date=$1, amount=$2, category=$3, transaction_type=$4, spender_id=$5, note=$6, image_url=$7 WHERE id=$8`

	e := echo.New()
	defer e.Close()

	// Correct setup for time.Time for the date
	testDate, _ := time.Parse(time.RFC3339, "2024-05-17T00:00:00Z")

	// Update the test data to send a time.Time object for the date
	updateData := PutTransaction{
		Date:            testDate,
		Amount:          100,
		Category:        "Utilities",
		TransactionType: "Expense",
		SpenderId:       1,
		Note:            "Electricity bill",
		ImageUrl:        "http://example.com/receipt.jpg",
	}
	bodyData, _ := json.Marshal(updateData)
	//e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/transaction/1", bytes.NewReader(bodyData))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	defer db.Close()

	// Setup mock to expect a time.Time object for the date
	mock.ExpectExec(query).WithArgs(
		testDate, // Exact time.Time object
		updateData.Amount, updateData.Category, updateData.TransactionType, updateData.SpenderId, updateData.Note, updateData.ImageUrl, "1",
	).WillReturnResult(sqlmock.NewResult(1, 1))

	h := New(config.FeatureFlag{}, db)
	err := h.PutTransaction(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"message": "Transaction updated successfully"}`, rec.Body.String())
}

func TestGetSpenderTransactionsSummarySuccess(t *testing.T) {
	e := echo.New()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	h := &handler{db: db}

	// Prepare the SQL query regex, allowing for flexible whitespace
	sqlQuery := `SELECT id, date, amount, category, transaction_type, note, image_url, spender_id FROM public.transaction WHERE spender_id=$1`
	sqlQuery = regexp.QuoteMeta(sqlQuery)
	sqlQuery = strings.Replace(sqlQuery, "\\ ", "\\s*", -1) // Allow any amount of whitespace

	// Set up the mock expectation with the modified regex
	mock.ExpectQuery(sqlQuery).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date", "amount", "category", "transaction_type", "note", "image_url", "spender_id"}).
			AddRow(1, time.Now(), 100.00, "Income", "income", "Salary", "http://example.com/img.jpg", 1).
			AddRow(2, time.Now(), 50.00, "Food", "expense", "Groceries", "http://example.com/img2.jpg", 1))

	req := httptest.NewRequest(http.MethodGet, "/spender/1/transactions/summart", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	if assert.NoError(t, h.GetSpenderTransactionSummary(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
