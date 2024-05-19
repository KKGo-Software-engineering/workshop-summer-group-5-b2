package spender

import (
	"database/sql"
	"net/http"

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

type Category struct {
	Name string `json:"name"`
}

func New(cfg config.FeatureFlag, db *sql.DB) *handler {
	return &handler{cfg, db}
}

const (
	cStmt   = `INSERT INTO spender (name, email) VALUES ($1, $2) RETURNING id;`
	getStmt = `SELECT id, name, email FROM spender WHERE id = $1;`
	getAllCats  = `SELECT DISTINCT category FROM transaction;`
)

func (h handler) Create(c echo.Context) error {
	if !h.flag.EnableCreateSpender {
		return c.JSON(http.StatusForbidden, "create new spender feature is disabled")
	}

	logger := mlog.L(c)
	ctx := c.Request().Context()
	var sp Spender
	err := c.Bind(&sp)
	if err != nil {
		logger.Error("bad request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var lastInsertId int64
	err = h.db.QueryRowContext(ctx, cStmt, sp.Name, sp.Email).Scan(&lastInsertId)
	if err != nil {
		logger.Error("query row error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	logger.Info("create successfully", zap.Int64("id", lastInsertId))
	sp.ID = lastInsertId
	return c.JSON(http.StatusCreated, sp)
}

func (h handler) GetAll(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	rows, err := h.db.QueryContext(ctx, `SELECT id, name, email FROM spender`)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var sps []Spender
	for rows.Next() {
		var sp Spender
		err := rows.Scan(&sp.ID, &sp.Name, &sp.Email)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		sps = append(sps, sp)
	}

	return c.JSON(http.StatusOK, sps)
}

func (h handler) GetSpenderByID(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()
	spenderID := c.Param("id")

	var sp Spender
	err := h.db.QueryRowContext(ctx, getStmt, spenderID).Scan(&sp.ID, &sp.Name, &sp.Email)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, "spender not found")
	} else if err != nil {
		logger.Error("query row error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, sp)
}

func (h handler) GetAllCategories(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	rows, err := h.db.QueryContext(ctx, getAllCats)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		categories = append(categories, category)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}
