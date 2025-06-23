package handler

import (
	"net/http"
	"go-url-shortener/storage"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// URLHandler manages URL shortening requests
type URLHandler struct {
	store            storage.Store
	redirectCounter  *prometheus.CounterVec
	shortenCounter   prometheus.Counter
	errorCounter     prometheus.Counter
}

// ShortenRequest represents the request to shorten a URL
type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

// NewURLHandler creates a new URL handler
func NewURLHandler(store storage.Store, registry *prometheus.Registry) *URLHandler {
	// Create metrics
	redirectCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_redirects_total",
			Help: "Total number of redirects by URL ID",
		},
		[]string{"url_id"},
	)

	shortenCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortener_shorten_requests_total",
			Help: "Total number of shorten requests",
		},
	)

	errorCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortener_errors_total",
			Help: "Total number of errors",
		},
	)

	// Register metrics
	registry.MustRegister(redirectCounter, shortenCounter, errorCounter)

	return &URLHandler{
		store:           store,
		redirectCounter: redirectCounter,
		shortenCounter:  shortenCounter,
		errorCounter:    errorCounter,
	}
}

// Shorten handles URL shortening requests
func (h *URLHandler) Shorten(c *gin.Context) {
	h.shortenCounter.Inc()

	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorCounter.Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	url, err := h.store.Create(req.URL)
	if err != nil {
		h.errorCounter.Inc()
		if err == storage.ErrInvalid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shortened URL"})
		}
		return
	}

	// Return shortened URL
	c.JSON(http.StatusOK, url)
}

// Redirect handles URL redirection
func (h *URLHandler) Redirect(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.errorCounter.Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL ID is required"})
		return
	}

	url, err := h.store.Get(id)
	if err != nil {
		h.errorCounter.Inc()
		if err == storage.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get URL"})
		}
		return
	}

	// Update metrics
	h.redirectCounter.With(prometheus.Labels{"url_id": id}).Inc()

	// Redirect to original URL
	c.Redirect(http.StatusFound, url.Original)
}

// GetStats returns stats for all URLs
func (h *URLHandler) GetStats(c *gin.Context) {
	urls, err := h.store.GetStats()
	if err != nil {
		h.errorCounter.Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"urls": urls})
}

// GetMetrics returns metrics for Prometheus
func (h *URLHandler) GetMetrics(c *gin.Context) {
	totalCount, err := h.store.GetTotalCount()
	if err != nil {
		h.errorCounter.Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total count"})
		return
	}

	totalHits, err := h.store.GetTotalHits()
	if err != nil {
		h.errorCounter.Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total hits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_urls": totalCount,
		"total_hits": totalHits,
	})
}
