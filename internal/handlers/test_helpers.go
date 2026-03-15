package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"food-order-tracking/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

// TestMain is the entry point for all handler tests.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// withMockDB sets database.DB to a fresh sqlmock, calls f, then restores the original.
func withMockDB(t *testing.T, f func(m sqlmock.Sqlmock)) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	f(mock)
}

// runDashboardRequest fires a GET /api/dashboard request and returns the recorder.
func runDashboardRequest(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	GetDashboardStats(c)
	return w
}
