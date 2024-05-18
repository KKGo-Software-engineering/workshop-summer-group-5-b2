//go:build integration

package transactions

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/KKGo-Software-engineering/workshop-summer/migration"
)

func TestGetSpenderTransactionsIT(t *testing.T) {
	t.Run("retrieve transactions for a specific spender", func(t *testing.T) {
		sqlDB, err := getTestDatabaseFromConfig()
		if err != nil {
			t.Fatal("Failed to get database connection:", err)
		}
		defer sqlDB.Close()

		migration.ApplyMigrations(sqlDB)
		defer migration.RollbackMigrations(sqlDB)

		// Insert sample data for testing
		insertSampleTransaction(sqlDB)

		// Setup handler and server
		h := New(config.FeatureFlag{}, sqlDB)
		e := echo.New()
		defer e.Close()

		e.GET("/api/v1/spenders/:id/transactions", h.GetSpenderTransactions)

		// Make the HTTP request
		spenderID := "123" // Example spender ID
		req := httptest.NewRequest(http.MethodGet, "/api/v1/spenders/"+spenderID+"/transactions", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(spenderID)

		// Serve HTTP and assert
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "HTTP status code should be 200 OK")
		var response SpenderIDTransactionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); assert.NoError(t, err) {
			assert.NotEmpty(t, response.Transactions, "Transactions should not be empty")
			// Additional assertions can be made here depending on expected results
		}
	})
}

// Helper function to insert sample transactions for the spender
func insertSampleTransaction(db *sql.DB) string {
	var id int64 // Use int64 to store the ID returned from the database
	query := `INSERT INTO "transaction" (date, amount, category, transaction_type, note, image_url, spender_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := db.QueryRow(query, time.Now(), 50, "Groceries", "expense", "Weekly shopping", "http://example.com/image.png", 1).Scan(&id)
	if err != nil {
		panic("Failed to insert sample transaction and get ID: " + err.Error())
	}
	return strconv.FormatInt(id, 10) // Convert the ID to a string
}

func getTestDatabaseFromConfig() (*sql.DB, error) {
	cfg := config.Parse("DOCKER")
	db, err := sql.Open("postgres", cfg.PostgresURI())
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestPutTransactionIT(t *testing.T) {
	db, err := getTestDatabaseFromConfig()
	if err != nil {
		t.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	migration.ApplyMigrations(db)
	defer migration.RollbackMigrations(db)

	// Insert sample data for testing
	transactionID := insertSampleTransaction(db)

	e := echo.New()
	h := New(config.FeatureFlag{}, db) // Assuming NewHandler initializes the handler

	// Define PUT endpoint
	e.PUT("/api/v1/transactions/:id", h.PutTransaction)

	t.Run("update transaction successfully", func(t *testing.T) {
		updatedTransaction := PutTransaction{
			Date:            time.Now(),
			Amount:          100,
			Category:        "Utilities",
			TransactionType: "expense",
			Note:            "Electricity bill",
			ImageUrl:        "http://example.com/new-image.png",
			SpenderId:       1,
		}
		reqBody, _ := json.Marshal(updatedTransaction)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/transactions/"+transactionID, bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(transactionID)

		if assert.NoError(t, h.PutTransaction(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); assert.NoError(t, err) {
				assert.Equal(t, "Transaction updated successfully", resp["message"])
			}
		}
	})
}

func insertSampleTransaction(db *sql.DB) string {
	// Insert a sample transaction and return its ID
	result, err := db.Exec(`INSERT INTO "transaction" (date, amount, category, transaction_type, note, image_url, spender_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		time.Now(), 50, "Groceries", "expense", "Weekly shopping", "http://example.com/image.png", 1)
	if err != nil {
		panic("Failed to insert sample transaction: " + err.Error())
	}
	id, err := result.LastInsertId()
	if err != nil {
		panic("Failed to get inserted transaction ID: " + err.Error())
	}
	return strconv.FormatInt(id, 10)
}
