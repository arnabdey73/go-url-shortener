package storage

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"sync"
	"time"
)

// MemoryStore implements Store using in-memory map
type MemoryStore struct {
	urls  map[string]*URL
	mutex sync.RWMutex
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urls: make(map[string]*URL),
	}
}

// generateID creates a unique shortcode
func generateID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)[:length], nil
}

// Create implements Store.Create
func (s *MemoryStore) Create(original string) (*URL, error) {
	// Validate URL
	parsedURL, err := url.ParseRequestURI(original)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, ErrInvalid
	}

	// Generate short ID (6 characters)
	id, err := generateID(6)
	if err != nil {
		return nil, err
	}

	// Create record
	record := &URL{
		ID:        id,
		Original:  original,
		CreatedAt: time.Now(),
		Hits:      0,
	}

	// Store URL
	s.mutex.Lock()
	s.urls[id] = record
	s.mutex.Unlock()

	return record, nil
}

// Get implements Store.Get
func (s *MemoryStore) Get(id string) (*URL, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	url, exists := s.urls[id]
	if !exists {
		return nil, ErrNotFound
	}
	
	// Increment hit counter
	url.Hits++
	
	return url, nil
}

// GetStats implements Store.GetStats
func (s *MemoryStore) GetStats() ([]*URL, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	result := make([]*URL, 0, len(s.urls))
	for _, url := range s.urls {
		result = append(result, url)
	}
	
	return result, nil
}

// GetTotalCount implements Store.GetTotalCount
func (s *MemoryStore) GetTotalCount() (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return len(s.urls), nil
}

// GetTotalHits implements Store.GetTotalHits
func (s *MemoryStore) GetTotalHits() (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	total := 0
	for _, url := range s.urls {
		total += url.Hits
	}
	
	return total, nil
}

// Close implements Store.Close (no-op for memory store)
func (s *MemoryStore) Close() error {
	return nil
}
