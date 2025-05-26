package ai

import (
	"fmt"
	"testing"
	"time"
)

func TestAIService_Configuration(t *testing.T) {
	config := AIConfig{
		Provider:           "claude",
		APIKey:             "test-key",
		Model:              "claude-3-sonnet-20240229",
		MaxTokens:          1000,
		Temperature:        0.7,
		Timeout:            30 * time.Second,
		MaxRetries:         3,
		RetryDelay:         1 * time.Second,
		RateLimitRequests:  60,
		RateLimitDuration:  1 * time.Minute,
		EnableCaching:      true,
		CacheTTL:           10 * time.Minute,
	}

	service, err := NewAIService(config)
	if err != nil {
		t.Fatalf("Failed to create AI service: %v", err)
	}

	if service.GetProviderName() != "claude" {
		t.Errorf("Expected provider 'claude', got '%s'", service.GetProviderName())
	}

	stats := service.GetStats()
	if stats["provider"] != "claude" {
		t.Errorf("Expected provider 'claude' in stats, got '%s'", stats["provider"])
	}
}

func TestAIService_InvalidProvider(t *testing.T) {
	config := AIConfig{
		Provider: "invalid-provider",
		APIKey:   "test-key",
	}

	_, err := NewAIService(config)
	if err == nil {
		t.Error("Expected error for invalid provider")
	}
}

func TestClaudeProvider_Validation(t *testing.T) {
	// Test missing API key
	config := AIConfig{
		Provider: "claude",
		APIKey:   "",
	}

	err := ValidateClaudeConfig(config)
	if err == nil {
		t.Error("Expected error for missing API key")
	}

	// Test invalid model
	config = AIConfig{
		Provider: "claude",
		APIKey:   "test-key",
		Model:    "invalid-model",
	}

	err = ValidateClaudeConfig(config)
	if err == nil {
		t.Error("Expected error for invalid model")
	}

	// Test valid configuration
	config = AIConfig{
		Provider:    "claude",
		APIKey:      "test-key",
		Model:       "claude-3-sonnet-20240229",
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	err = ValidateClaudeConfig(config)
	if err != nil {
		t.Errorf("Unexpected error for valid config: %v", err)
	}
}

func TestRateLimiter(t *testing.T) {
	// Create rate limiter: 5 requests per second
	rl := NewRateLimiter(5, 1*time.Second)

	// Should allow initial requests
	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Should reject the 6th request
	if rl.Allow() {
		t.Error("6th request should be rejected")
	}

	// Wait for refill and test again
	time.Sleep(250 * time.Millisecond) // Should refill 1 token
	if !rl.Allow() {
		t.Error("Request after refill should be allowed")
	}

	stats := rl.GetStats()
	if stats["max_tokens"] != 5 {
		t.Errorf("Expected max_tokens 5, got %v", stats["max_tokens"])
	}
}

func TestResponseCache(t *testing.T) {
	cache := NewResponseCache(100 * time.Millisecond)

	// Test cache miss
	result := cache.Get("key1")
	if result != "" {
		t.Error("Expected cache miss for non-existent key")
	}

	// Test cache set and hit
	cache.Set("key1", "value1")
	result = cache.Get("key1")
	if result != "value1" {
		t.Errorf("Expected 'value1', got '%s'", result)
	}

	// Test cache expiration
	time.Sleep(150 * time.Millisecond)
	result = cache.Get("key1")
	if result != "" {
		t.Error("Expected cache miss after expiration")
	}

	// Test stats
	cache.Set("key2", "value2")
	cache.Get("key2") // hit
	cache.Get("key3") // miss

	stats := cache.GetStats()
	if stats["hits"].(int64) != 1 {
		t.Errorf("Expected 1 hit, got %v", stats["hits"])
	}
	if stats["misses"].(int64) != 2 {
		t.Errorf("Expected 2 misses, got %v", stats["misses"])
	}
}

func TestHashString(t *testing.T) {
	hash1 := hashString("test string")
	hash2 := hashString("test string")
	hash3 := hashString("different string")

	if hash1 != hash2 {
		t.Error("Same strings should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("Different strings should produce different hashes")
	}

	if len(hash1) != 8 {
		t.Errorf("Hash should be 8 characters long, got %d", len(hash1))
	}
}

func TestIsNonRetryableError(t *testing.T) {
	testCases := []struct {
		error       string
		shouldRetry bool
	}{
		{"authentication failed", false},
		{"unauthorized access", false},
		{"quota exceeded", false},
		{"invalid api key", false},
		{"network timeout", true},
		{"server error", true},
		{"connection refused", true},
	}

	for _, tc := range testCases {
		err := fmt.Errorf(tc.error)
		isNonRetryable := isNonRetryableError(err)
		
		if isNonRetryable == tc.shouldRetry {
			t.Errorf("Error '%s': expected shouldRetry=%t, got isNonRetryable=%t", 
				tc.error, tc.shouldRetry, isNonRetryable)
		}
	}
}

// Mock tests (these would need a test API key to run against real Claude API)
func TestClaudeProvider_MockResponse(t *testing.T) {
	// This is a placeholder test - in real testing you'd either:
	// 1. Use a test API key and make real API calls
	// 2. Mock the Claude client
	// 3. Use integration tests with environment variables

	config := AIConfig{
		Provider:    "claude",
		APIKey:      "test-key",
		Model:       "claude-3-sonnet-20240229",
		MaxTokens:   100,
		Temperature: 0.7,
		Timeout:     5 * time.Second,
	}

	// This will fail without a real API key, which is expected
	_, err := NewClaudeProvider(config)
	if err != nil && err.Error() != "Claude API key is required" {
		// If we have a test key, this test would continue
		t.Skip("Skipping Claude provider test - no API key available")
	}
}

// Benchmark tests
func BenchmarkHashString(b *testing.B) {
	testString := "This is a test string for hashing benchmark"
	
	for i := 0; i < b.N; i++ {
		hashString(testString)
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	rl := NewRateLimiter(1000, 1*time.Second)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

func BenchmarkResponseCache(b *testing.B) {
	cache := NewResponseCache(1 * time.Hour)
	cache.Set("test-key", "test-value")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("test-key")
	}
}
