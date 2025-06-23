package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go-url-shortener/storage"
)

// mockErrorStore is a Store implementation that returns errors
type mockErrorStore struct{}

func (s *mockErrorStore) Create(url string) (*storage.URL, error) {
	return nil, storage.ErrInvalid
}

func (s *mockErrorStore) Get(id string) (*storage.URL, error) {
	return nil, storage.ErrNotFound
}

func (s *mockErrorStore) GetStats() ([]*storage.URL, error) {
	return nil, errors.New("database error")
}

func (s *mockErrorStore) GetTotalCount() (int, error) {
	return 0, errors.New("database error")
}

func (s *mockErrorStore) GetTotalHits() (int, error) {
	return 0, errors.New("database error")
}

func (s *mockErrorStore) Close() error {
	return nil
}

// setupErrorTestEnvironment sets up a test environment with a mock error store
func setupErrorTestEnvironment() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create mock error store
	store := &mockErrorStore{}
	
	// Create registry
	registry := prometheus.NewRegistry()
	
	// Create handler
	handler := NewURLHandler(store, registry)
	
	// Set up routes
	router.POST("/api/shorten", handler.Shorten)
	router.GET("/api/stats", handler.GetStats)
	router.GET("/:id", handler.Redirect)
	router.GET("/metrics", handler.GetMetrics)
	
	return router
}

// TestErrorHandling tests error handling in handlers
func TestErrorHandling(t *testing.T) {
	router := setupErrorTestEnvironment()
	
	t.Run("Handle Create Error", func(t *testing.T) {
		// Create request with valid URL, but store will return error
		reqBody := `{"url":"https://example.com/test"}`
		req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code (should be Bad Request due to ErrInvalid)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", w.Code)
		}
		
		// Verify error response
		var resp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		
		if resp.Error == "" {
			t.Error("Expected error message in response")
		}
	})
	
	t.Run("Handle Get Error", func(t *testing.T) {
		// Create request for non-existent URL
		req, _ := http.NewRequest("GET", "/abc123", nil)
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code (should be Not Found)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", w.Code)
		}
	})
	
	t.Run("Handle GetStats Error", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code (should be Internal Server Error)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status Internal Server Error, got %v", w.Code)
		}
	})
	
	t.Run("Handle GetMetrics Error", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("GET", "/metrics", nil)
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code (should be Internal Server Error)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status Internal Server Error, got %v", w.Code)
		}
	})
}
