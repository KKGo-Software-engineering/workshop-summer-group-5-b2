//go:build integration

package transaction

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/lib/pq"

	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateIT(t *testing.T) {
	t.Run("create transactions successfully", func(t *testing.T) {
		conn := newDatabase(t)
		t.Cleanup(func() {
			conn.Query("DELETE FROM transaction Where amount=$1 AND category=$2 AND date=$3;", 200.99, "refund", "2024-05-18T15:00:37.557628+07:00")
		})

		h := New(config.FeatureFlag{EnableCreateSpender: true}, conn)
		e := echo.New()
		defer e.Close()

		e.POST("/transactions", h.Create)

		payload := `{"date":"2024-05-18T15:00:37.557628+07:00","amount":200.99,"category":"refund","transaction_type":"income","spender_id":2}`
		req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.NotEmpty(t, rec.Body.String())
	})
}

func newDatabase(t *testing.T) *sql.DB {
	t.Helper()
	cfg := config.Parse("DOCKER")
	conn, err := sql.Open("postgres", cfg.PostgresURI())
	if err != nil {
		t.Fatal(err)
	}
	return conn
}
