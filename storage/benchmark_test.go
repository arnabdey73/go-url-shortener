package storage

import (
	"os"
	"testing"
)

// BenchmarkMemoryStoreCreate benchmarks the Create operation in MemoryStore
func BenchmarkMemoryStoreCreate(b *testing.B) {
	store := NewMemoryStore()
	defer store.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create URL with unique name to avoid collisions
		url := "https://example.com/benchmark-" + string(rune(i%26+97))
		_, err := store.Create(url)
		if err != nil {
			b.Fatalf("Error creating URL: %v", err)
		}
	}
}

// BenchmarkMemoryStoreGet benchmarks the Get operation in MemoryStore
func BenchmarkMemoryStoreGet(b *testing.B) {
	store := NewMemoryStore()
	defer store.Close()
	
	// Create a URL to retrieve during benchmark
	url, err := store.Create("https://example.com/benchmark-get")
	if err != nil {
		b.Fatalf("Error creating URL: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.Get(url.ID)
		if err != nil {
			b.Fatalf("Error getting URL: %v", err)
		}
	}
}

// BenchmarkMemoryStoreGetStats benchmarks the GetStats operation in MemoryStore
func BenchmarkMemoryStoreGetStats(b *testing.B) {
	store := NewMemoryStore()
	defer store.Close()
	
	// Create some URLs for stats
	for i := 0; i < 100; i++ {
		url := "https://example.com/benchmark-stats-" + string(rune(i%26+97))
		_, err := store.Create(url)
		if err != nil {
			b.Fatalf("Error creating URL: %v", err)
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.GetStats()
		if err != nil {
			b.Fatalf("Error getting stats: %v", err)
		}
	}
}

// BenchmarkSQLiteStoreCreate benchmarks the Create operation in SQLiteStore
func BenchmarkSQLiteStoreCreate(b *testing.B) {
	// Create temporary file for SQLite
	tmpFile, err := os.CreateTemp("", "bench-urls-*.db")
	if err != nil {
		b.Fatalf("Failed to create temporary file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFileName)
	
	// Create store
	store, err := NewSQLiteStore(tmpFileName)
	if err != nil {
		b.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create URL with unique name to avoid collisions
		url := "https://example.com/benchmark-" + string(rune(i%26+97))
		_, err := store.Create(url)
		if err != nil {
			b.Fatalf("Error creating URL: %v", err)
		}
	}
}

// BenchmarkSQLiteStoreGet benchmarks the Get operation in SQLiteStore
func BenchmarkSQLiteStoreGet(b *testing.B) {
	// Create temporary file for SQLite
	tmpFile, err := os.CreateTemp("", "bench-urls-get-*.db")
	if err != nil {
		b.Fatalf("Failed to create temporary file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFileName)
	
	// Create store
	store, err := NewSQLiteStore(tmpFileName)
	if err != nil {
		b.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()
	
	// Create a URL to retrieve during benchmark
	url, err := store.Create("https://example.com/benchmark-get")
	if err != nil {
		b.Fatalf("Error creating URL: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.Get(url.ID)
		if err != nil {
			b.Fatalf("Error getting URL: %v", err)
		}
	}
}
