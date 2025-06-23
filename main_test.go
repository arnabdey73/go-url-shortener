package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
	
	"go-url-shortener/storage"
)

// TestApplicationStartup verifies that the application starts up successfully
// This is an integration test that starts the actual server
func TestApplicationStartup(t *testing.T) {
	// Start the application in memory mode in a goroutine
	go func() {
		// Use a different port to avoid conflicts with a running server
		args := []string{"cmd", "--port=8081", "--db=memory"}
		// Ignore returned errors as we're killing the server after test
		runServer(args)
	}()

	// Give the server a moment to start
	time.Sleep(500 * time.Millisecond)

	// Create a context with timeout for requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a request to check server health
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8081/metrics", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
}
// TestFullUserFlow tests the complete user flow - shorten URL, get stats, redirect
func TestFullUserFlow(t *testing.T) {
	// Start the application in memory mode in a goroutine
	serverPort := 8082
	go func() {
		// Use a different port to avoid conflicts with other tests
		args := []string{"cmd", "--port=" + string(rune(serverPort+'0')), "--db=memory"}
		// Ignore returned errors as we're killing the server after test
		runServer(args)
	}()

	// Give the server a moment to start
	time.Sleep(500 * time.Millisecond)
	baseURL := "http://localhost:8082"
	
	// Create a context with timeout for requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Step 1: Shorten a URL
	t.Run("ShortenURL", func(t *testing.T) {
		payload := map[string]string{
			"url": "https://example.com/full-flow-test",
		}
		jsonData, _ := json.Marshal(payload)
		
		req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/shorten", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()
		
		// Check response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK, got %v", resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Response body: %s", body)
			return
		}
		
		// Parse response to get shortened URL ID
		var shortenResp storage.URL
		if err := json.NewDecoder(resp.Body).Decode(&shortenResp); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		
		// Step 2: Get stats to verify the URL was created
		t.Run("GetStats", func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/stats", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()
			
			// Check response status
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status OK, got %v", resp.StatusCode)
				return
			}
			
			// Parse response
			var statsResp struct {
				URLs []*storage.URL `json:"urls"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&statsResp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			
			// Verify the URL is in stats
			var found bool
			for _, url := range statsResp.URLs {
				if url.ID == shortenResp.ID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("URL not found in stats")
			}
		})
		
		// Step 3: Redirect to the URL
		t.Run("Redirect", func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/"+shortenResp.ID, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			// Don't follow redirects automatically
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()
			
			// Check response status (should be redirect)
			if resp.StatusCode != http.StatusFound {
				t.Errorf("Expected status Found, got %v", resp.StatusCode)
				return
			}
			
			// Check redirect location
			location := resp.Header.Get("Location")
			if location != "https://example.com/full-flow-test" {
				t.Errorf("Expected redirect to %q, got %q", "https://example.com/full-flow-test", location)
			}
		})
	})
}
}
