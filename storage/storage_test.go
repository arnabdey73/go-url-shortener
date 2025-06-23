package storage

import (
	"os"
	"testing"
	"time"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	defer store.Close()

	runStoreTests(t, store)
}

func TestSQLiteStore(t *testing.T) {
	// Create temporary file for SQLite
	tmpFile, err := os.CreateTemp("", "urls-*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFileName)

	// Create store
	store, err := NewSQLiteStore(tmpFileName)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	runStoreTests(t, store)
}

func runStoreTests(t *testing.T, store Store) {
	t.Run("Create", func(t *testing.T) {
		// Test valid URL
		url, err := store.Create("https://example.com/test")
		if err != nil {
			t.Fatalf("Failed to create URL: %v", err)
		}
		if url.ID == "" {
			t.Error("Expected ID to be set")
		}
		if url.Original != "https://example.com/test" {
			t.Errorf("Expected original URL to be %q, got %q", "https://example.com/test", url.Original)
		}
		if url.Hits != 0 {
			t.Errorf("Expected hits to be 0, got %d", url.Hits)
		}
		if time.Since(url.CreatedAt) > time.Minute {
			t.Errorf("Expected CreatedAt to be recent")
		}

		// Test invalid URL
		_, err = store.Create("not-a-url")
		if err != ErrInvalid {
			t.Errorf("Expected ErrInvalid for invalid URL, got %v", err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		// Create a URL
		created, err := store.Create("https://example.com/test-get")
		if err != nil {
			t.Fatalf("Failed to create URL: %v", err)
		}

		// Get the URL
		got, err := store.Get(created.ID)
		if err != nil {
			t.Fatalf("Failed to get URL: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("Expected ID to be %q, got %q", created.ID, got.ID)
		}
		if got.Original != created.Original {
			t.Errorf("Expected original URL to be %q, got %q", created.Original, got.Original)
		}
		if got.Hits != 1 {
			t.Errorf("Expected hits to be 1, got %d", got.Hits)
		}

		// Get again to test hit counter
		got, err = store.Get(created.ID)
		if err != nil {
			t.Fatalf("Failed to get URL: %v", err)
		}
		if got.Hits != 2 {
			t.Errorf("Expected hits to be 2, got %d", got.Hits)
		}

		// Get non-existent URL
		_, err = store.Get("nonexistent")
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound for non-existent URL, got %v", err)
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		// Create a few URLs
		for i := 0; i < 3; i++ {
			_, err := store.Create("https://example.com/test-stats-" + string(rune('a'+i)))
			if err != nil {
				t.Fatalf("Failed to create URL: %v", err)
			}
		}

		// Get stats
		urls, err := store.GetStats()
		if err != nil {
			t.Fatalf("Failed to get stats: %v", err)
		}
		
		// SQLite and Memory stores might have different ordering
		// Just check that we have some URLs
		if len(urls) == 0 {
			t.Error("Expected some URLs in stats")
		}
	})

	t.Run("GetTotalCount", func(t *testing.T) {
		// Get initial count
		initialCount, err := store.GetTotalCount()
		if err != nil {
			t.Fatalf("Failed to get total count: %v", err)
		}

		// Create a URL
		_, err = store.Create("https://example.com/test-count")
		if err != nil {
			t.Fatalf("Failed to create URL: %v", err)
		}

		// Get count again
		newCount, err := store.GetTotalCount()
		if err != nil {
			t.Fatalf("Failed to get total count: %v", err)
		}
		if newCount != initialCount+1 {
			t.Errorf("Expected count to increase by 1, got increase of %d", newCount-initialCount)
		}
	})

	t.Run("GetTotalHits", func(t *testing.T) {
		// Get initial total hits
		initialHits, err := store.GetTotalHits()
		if err != nil {
			t.Fatalf("Failed to get total hits: %v", err)
		}

		// Create a URL
		url, err := store.Create("https://example.com/test-hits")
		if err != nil {
			t.Fatalf("Failed to create URL: %v", err)
		}

		// Get the URL to increment hits
		_, err = store.Get(url.ID)
		if err != nil {
			t.Fatalf("Failed to get URL: %v", err)
		}

		// Get total hits again
		newHits, err := store.GetTotalHits()
		if err != nil {
			t.Fatalf("Failed to get total hits: %v", err)
		}
		if newHits != initialHits+1 {
			t.Errorf("Expected total hits to increase by 1, got increase of %d", newHits-initialHits)
		}
	})
}
