package health

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/proullon/ramsql/driver"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	db, _ := sql.Open("ramsql", "TestHealth")

	err := Check(db)(c)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSlow(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	err := Slow(c)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
