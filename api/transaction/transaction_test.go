package transaction

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		cStmt := `INSERT INTO "transaction" ("date", "amount", "category", "transaction_type", "spender_id") VALUES ($1, $2, $3, $4, $5) RETURNING id;`
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
