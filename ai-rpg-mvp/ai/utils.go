package ai

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, duration time.Duration) *RateLimiter {
	refillRate := duration / time.Duration(maxRequests)
	
	return &RateLimiter{
		tokens:     maxRequests,
		maxTokens:  maxRequests,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed based on the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	
	// Calculate how many tokens to add based on elapsed time
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)
	
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	// Check if we have tokens available
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	return map[string]interface{}{
		"available_tokens": rl.tokens,
		"max_tokens":       rl.maxTokens,
		"refill_rate_ms":   rl.refillRate.Milliseconds(),
	}
}

// ResponseCache implements a simple in-memory cache with TTL
type ResponseCache struct {
	cache   map[string]cacheEntry
	ttl     time.Duration
	mutex   sync.RWMutex
	hits    int64
	misses  int64
}

type cacheEntry struct {
	value     string
	timestamp time.Time
}

// NewResponseCache creates a new response cache
func NewResponseCache(ttl time.Duration) *ResponseCache {
	cache := &ResponseCache{
		cache: make(map[string]cacheEntry),
		ttl:   ttl,
	}

	// Start background cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from the cache
func (rc *ResponseCache) Get(key string) string {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	entry, exists := rc.cache[key]
	if !exists {
		rc.misses++
		return ""
	}

	// Check if entry has expired
	if time.Since(entry.timestamp) > rc.ttl {
		rc.misses++
		return ""
	}

	rc.hits++
	return entry.value
}

// Set stores a value in the cache
func (rc *ResponseCache) Set(key, value string) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	rc.cache[key] = cacheEntry{
		value:     value,
		timestamp: time.Now(),
	}
}

// GetStats returns cache statistics
func (rc *ResponseCache) GetStats() map[string]interface{} {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	total := rc.hits + rc.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(rc.hits) / float64(total)
	}

	return map[string]interface{}{
		"hits":      rc.hits,
		"misses":    rc.misses,
		"hit_rate":  hitRate,
		"size":      len(rc.cache),
		"ttl_hours": rc.ttl.Hours(),
	}
}

// cleanup removes expired entries from the cache
func (rc *ResponseCache) cleanup() {
	ticker := time.NewTicker(rc.ttl / 2) // Clean up twice per TTL period
	defer ticker.Stop()

	for range ticker.C {
		rc.mutex.Lock()
		now := time.Now()
		for key, entry := range rc.cache {
			if now.Sub(entry.timestamp) > rc.ttl {
				delete(rc.cache, key)
			}
		}
		rc.mutex.Unlock()
	}
}

// Clear removes all entries from the cache
func (rc *ResponseCache) Clear() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	rc.cache = make(map[string]cacheEntry)
	rc.hits = 0
	rc.misses = 0
}
