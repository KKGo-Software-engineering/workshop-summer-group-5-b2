package spender

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateSpender(t *testing.T) {

	t.Run("create spender succesfully when feature toggle is enable", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name": "HongJot", "email": "hong@jot.ok"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		row := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery(cStmt).WithArgs("HongJot", "hong@jot.ok").WillReturnRows(row)
		cfg := config.FeatureFlag{EnableCreateSpender: true}

		h := New(cfg, db)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"id": 1, "name": "HongJot", "email": "hong@jot.ok"}`, rec.Body.String())
	})

	t.Run("create spender failed when feature toggle is disable", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name": "HongJot", "email": "hong@jot.ok"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		cfg := config.FeatureFlag{EnableCreateSpender: false}

		h := New(cfg, nil)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("create spender failed when bad request body", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{ bad request body }`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		cfg := config.FeatureFlag{EnableCreateSpender: true}

		h := New(cfg, nil)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid character")
	})

	t.Run("create spender failed on database (feature toggle is enable) ", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name": "HongJot", "email": "hong@jot.ok"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectQuery(cStmt).WithArgs("HongJot", "hong@jot.ok").WillReturnError(assert.AnError)
		cfg := config.FeatureFlag{EnableCreateSpender: true}

		h := New(cfg, db)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetSpenderByID(t *testing.T) {
	e := echo.New()
	defer e.Close()

	t.Run("get spender successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/spenders/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		row := sqlmock.NewRows([]string{"id", "name", "email"}).AddRow(1, "HongJot", "hong@jot.ok")
		mock.ExpectQuery(getStmt).WithArgs("1").WillReturnRows(row)
		cfg := config.FeatureFlag{}

		h := New(cfg, db)
		err := h.GetSpenderByID(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"id": 1, "name": "HongJot", "email": "hong@jot.ok"}`, rec.Body.String())
	})

	t.Run("get spender not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/spenders/2", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders/:id")
		c.SetParamNames("id")
		c.SetParamValues("2")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectQuery(getStmt).WithArgs("2").WillReturnError(sql.ErrNoRows)
		cfg := config.FeatureFlag{}

		h := New(cfg, db)
		err := h.GetSpenderByID(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "\"spender not found\"", strings.TrimSpace(rec.Body.String()))
	})

	t.Run("get spender database error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/spenders/3", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders/:id")
		c.SetParamNames("id")
		c.SetParamValues("3")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectQuery(getStmt).WithArgs("3").WillReturnError(assert.AnError)
		cfg := config.FeatureFlag{}

		h := New(cfg, db)
		err := h.GetSpenderByID(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetAllCategories(t *testing.T) {
	e := echo.New()
	defer e.Close()

	t.Run("get all categories successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/categories")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		rows := sqlmock.NewRows([]string{"category"}).
			AddRow("Food").
			AddRow("Travel").
			AddRow("Shopping")
		mock.ExpectQuery(getAllCats).WillReturnRows(rows)
		cfg := config.FeatureFlag{}

		h := New(cfg, db)
		err := h.GetAllCategories(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"categories": ["Food", "Travel", "Shopping"]}`, rec.Body.String())
	})

	t.Run("get all categories database error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/categories")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectQuery(getAllCats).WillReturnError(assert.AnError)
		cfg := config.FeatureFlag{}

		h := New(cfg, db)
		err := h.GetAllCategories(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("get all categories scan error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/categories")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		rows := sqlmock.NewRows([]string{"category"}).
			AddRow("Food").
			AddRow(nil) // This will cause a scan error
		mock.ExpectQuery(getAllCats).WillReturnRows(rows)
		cfg := config.FeatureFlag{}

		h := New(cfg, db)
		err := h.GetAllCategories(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetAll(t *testing.T) {
	e := echo.New()
	defer e.Close()

	t.Run("get all successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/spenders", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "name", "email"}).
			AddRow(1, "HongJot", "hong@jot.ok")
		mock.ExpectQuery(`SELECT id, name, email FROM spender`).WillReturnRows(rows)
		h := New(config.FeatureFlag{}, db)
		h.GetAll(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"spenders": [{"id": 1, "name": "HongJot", "email": "hong@jot.ok"}]}`, rec.Body.String())
	})
}
