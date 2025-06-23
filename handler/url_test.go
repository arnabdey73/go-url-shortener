package handler

import (
	"bytes"
	"encoding/json"
	"go-url-shortener/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Setup test environment
func setupTestEnvironment() (*gin.Engine, *URLHandler, storage.Store) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create in-memory store
	store := storage.NewMemoryStore()
	
	// Create registry
	registry := prometheus.NewRegistry()
	
	// Create handler
	handler := NewURLHandler(store, registry)
	
	// Set up routes
	router.POST("/api/shorten", handler.Shorten)
	router.GET("/api/stats", handler.GetStats)
	router.GET("/:id", handler.Redirect)
	
	return router, handler, store
}

func TestShorten(t *testing.T) {
	router, _, store := setupTestEnvironment()
	
	t.Run("Valid URL", func(t *testing.T) {
		// Create request
		reqBody := `{"url":"https://example.com/test"}`
		req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code
		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
		
		// Parse response
		var resp storage.URL
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		
		// Verify response fields
		if resp.ID == "" {
			t.Error("Expected ID to be set")
		}
		if resp.Original != "https://example.com/test" {
			t.Errorf("Expected original URL to be %q, got %q", "https://example.com/test", resp.Original)
		}
	})
	
	t.Run("Invalid URL", func(t *testing.T) {
		// Create request with invalid URL
		reqBody := `{"url":"not-a-url"}`
		req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code (should be Bad Request)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", w.Code)
		}
	})
	
	t.Run("Missing URL", func(t *testing.T) {
		// Create request with missing URL
		reqBody := `{}`
		req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code (should be Bad Request)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", w.Code)
		}
	})
	
	// Clean up
	store.Close()
}

func TestRedirect(t *testing.T) {
	router, _, store := setupTestEnvironment()
	
	// Create a URL for testing redirects
	url, err := store.Create("https://example.com/test-redirect")
	if err != nil {
		t.Fatalf("Failed to create URL: %v", err)
	}
	
	t.Run("Valid Redirect", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("GET", "/"+url.ID, nil)
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code (should be Found/302)
		if w.Code != http.StatusFound {
			t.Errorf("Expected status Found, got %v", w.Code)
		}
		
		// Check redirect location
		if location := w.Header().Get("Location"); location != "https://example.com/test-redirect" {
			t.Errorf("Expected redirect location %q, got %q", "https://example.com/test-redirect", location)
		}
	})
	
	t.Run("Non-existent ID", func(t *testing.T) {
		// Create request with non-existent ID
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code (should be Not Found)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", w.Code)
		}
	})
	
	// Clean up
	store.Close()
}

func TestGetStats(t *testing.T) {
	router, _, store := setupTestEnvironment()
	
	// Create some URLs for testing
	url1, _ := store.Create("https://example.com/test-stats-1")
	url2, _ := store.Create("https://example.com/test-stats-2")
	
	// Increment hits for one URL
	_, _ = store.Get(url1.ID)
	
	t.Run("Get Stats", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		
		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert status code
		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
		
		// Parse response
		var resp struct {
			URLs []*storage.URL `json:"urls"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		
		// Verify response
		if len(resp.URLs) != 2 {
			t.Errorf("Expected 2 URLs, got %d", len(resp.URLs))
		}
		
		// Find url1 in the response
		var found bool
		for _, u := range resp.URLs {
			if u.ID == url1.ID {
				found = true
				if u.Hits != 1 {
					t.Errorf("Expected hits to be 1 for first URL, got %d", u.Hits)
				}
				break
			}
		}
		if !found {
			t.Error("First URL not found in stats")
		}
		
		// Find url2 in the response
		found = false
		for _, u := range resp.URLs {
			if u.ID == url2.ID {
				found = true
				if u.Hits != 0 {
					t.Errorf("Expected hits to be 0 for second URL, got %d", u.Hits)
				}
				break
			}
		}
		if !found {
			t.Error("Second URL not found in stats")
		}
	})
	
	// Clean up
	store.Close()
}
