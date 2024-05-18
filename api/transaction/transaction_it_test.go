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
	"github.com/KKGo-Software-engineering/workshop-summer/migration"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateIT(t *testing.T) {
	t.Run("create transactions successfully", func(t *testing.T) {
		sql, err := getTestDatabaseFromConfig()
		if err != nil {
			t.Error(err)
		}
		migration.ApplyMigrations(sql)
		defer migration.RollbackMigrations(sql)

		h := New(config.FeatureFlag{EnableCreateSpender: true}, sql)
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

func getTestDatabaseFromConfig() (*sql.DB, error) {
	cfg := config.Parse("DOCKER")
	sql, err := sql.Open("postgres", cfg.PostgresURI())
	if err != nil {
		return nil, err
	}
	return sql, nil
}
