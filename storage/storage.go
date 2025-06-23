package storage

import (
	"errors"
	"time"
)

// URL represents a shortened URL record
type URL struct {
	ID        string    `json:"id"`
	Original  string    `json:"original"`
	CreatedAt time.Time `json:"created_at"`
	Hits      int       `json:"hits"`
}

// Store defines the interface for URL storage
type Store interface {
	// Create stores a new shortened URL
	Create(url string) (*URL, error)
	
	// Get retrieves a URL by its ID and increments hit counter
	Get(id string) (*URL, error)
	
	// GetStats retrieves all URLs stats
	GetStats() ([]*URL, error)
	
	// GetTotalCount returns the total number of shortened URLs
	GetTotalCount() (int, error)
	
	// GetTotalHits returns the total number of hits across all URLs
	GetTotalHits() (int, error)
	
	// Close cleans up any resources
	Close() error
}

// Common errors
var (
	ErrNotFound = errors.New("url not found")
	ErrInvalid  = errors.New("invalid url")
)
