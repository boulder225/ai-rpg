package ai

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// AIProvider defines the interface for AI services
type AIProvider interface {
	GenerateGMResponse(prompt string) (string, error)
	GenerateNPCDialogue(npcName, personality, prompt string) (string, error)
	GenerateSceneDescription(location, context, mood string) (string, error)
	GetProviderName() string
}

// AIService manages AI providers and handles requests
type AIService struct {
	provider    AIProvider
	rateLimiter *RateLimiter
	cache       *ResponseCache
	config      AIConfig
}

// AIConfig holds configuration for AI service
type AIConfig struct {
	Provider          string
	APIKey            string
	Model             string
	MaxTokens         int
	Temperature       float64
	Timeout           time.Duration
	MaxRetries        int
	RetryDelay        time.Duration
	EnableCaching     bool
	CacheTTL          time.Duration
	RateLimitRequests int
	RateLimitDuration time.Duration
}

// NewAIService creates a new AI service with the specified provider
func NewAIService(config AIConfig) (*AIService, error) {
	var provider AIProvider
	var err error

	switch strings.ToLower(config.Provider) {
	case "claude", "anthropic":
		provider, err = NewClaudeProvider(config)
	case "openai":
		provider, err = NewOpenAIProvider(config)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", config.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	service := &AIService{
		provider: provider,
		config:   config,
	}

	// Initialize rate limiter
	if config.RateLimitRequests > 0 {
		service.rateLimiter = NewRateLimiter(config.RateLimitRequests, config.RateLimitDuration)
	}

	// Initialize cache
	if config.EnableCaching {
		service.cache = NewResponseCache(config.CacheTTL)
	}

	return service, nil
}

// GenerateGMResponse generates a Game Master response
func (s *AIService) GenerateGMResponse(prompt string) (string, error) {
	cacheKey := fmt.Sprintf("gm:%s", hashString(prompt))

	// Check cache first
	if s.cache != nil {
		if cached := s.cache.Get(cacheKey); cached != "" {
			return cached, nil
		}
	}

	// Check rate limit
	if s.rateLimiter != nil {
		if !s.rateLimiter.Allow() {
			return "", fmt.Errorf("rate limit exceeded")
		}
	}

	// Generate response with retries
	response, err := s.generateWithRetry(func() (string, error) {
		return s.provider.GenerateGMResponse(prompt)
	})

	if err != nil {
		return "", err
	}

	// Cache response
	if s.cache != nil {
		s.cache.Set(cacheKey, response)
	}

	return response, nil
}

// GenerateNPCDialogue generates NPC dialogue
func (s *AIService) GenerateNPCDialogue(npcName, personality, prompt string) (string, error) {
	cacheKey := fmt.Sprintf("npc:%s:%s", npcName, hashString(prompt))

	// Check cache first
	if s.cache != nil {
		if cached := s.cache.Get(cacheKey); cached != "" {
			return cached, nil
		}
	}

	// Check rate limit
	if s.rateLimiter != nil {
		if !s.rateLimiter.Allow() {
			return "", fmt.Errorf("rate limit exceeded")
		}
	}

	// Generate response with retries
	response, err := s.generateWithRetry(func() (string, error) {
		return s.provider.GenerateNPCDialogue(npcName, personality, prompt)
	})

	if err != nil {
		return "", err
	}

	// Cache response
	if s.cache != nil {
		s.cache.Set(cacheKey, response)
	}

	return response, nil
}

// GenerateSceneDescription generates scene descriptions
func (s *AIService) GenerateSceneDescription(location, contextInfo, mood string) (string, error) {
	cacheKey := fmt.Sprintf("scene:%s:%s:%s", location, mood, hashString(contextInfo))

	// Check cache first
	if s.cache != nil {
		if cached := s.cache.Get(cacheKey); cached != "" {
			return cached, nil
		}
	}

	// Check rate limit
	if s.rateLimiter != nil {
		if !s.rateLimiter.Allow() {
			return "", fmt.Errorf("rate limit exceeded")
		}
	}

	// Generate response with retries
	response, err := s.generateWithRetry(func() (string, error) {
		return s.provider.GenerateSceneDescription(location, contextInfo, mood)
	})

	if err != nil {
		return "", err
	}

	// Cache response
	if s.cache != nil {
		s.cache.Set(cacheKey, response)
	}

	return response, nil
}

// generateWithRetry executes a function with retry logic
func (s *AIService) generateWithRetry(fn func() (string, error)) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(s.config.RetryDelay * time.Duration(attempt))
			log.Printf("AI request retry attempt %d/%d", attempt, s.config.MaxRetries)
		}

		response, err := fn()
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on certain errors (rate limit, invalid key, etc.)
		if isNonRetryableError(err) {
			break
		}
	}

	return "", fmt.Errorf("AI request failed after %d attempts: %w", s.config.MaxRetries+1, lastErr)
}

// GetProviderName returns the name of the current AI provider
func (s *AIService) GetProviderName() string {
	return s.provider.GetProviderName()
}

// GetStats returns service statistics
func (s *AIService) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"provider": s.GetProviderName(),
		"model":    s.config.Model,
	}

	if s.rateLimiter != nil {
		stats["rate_limiter"] = s.rateLimiter.GetStats()
	}

	if s.cache != nil {
		stats["cache"] = s.cache.GetStats()
	}

	return stats
}

// isNonRetryableError checks if an error should not be retried
func isNonRetryableError(err error) bool {
	errStr := strings.ToLower(err.Error())

	// Don't retry on authentication, permission, or quota errors
	nonRetryablePatterns := []string{
		"authentication",
		"unauthorized",
		"forbidden",
		"quota exceeded",
		"invalid api key",
		"billing",
		"payment",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// hashString creates a simple hash of a string for caching
func hashString(s string) string {
	// Simple hash function for cache keys
	h := uint32(0)
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return fmt.Sprintf("%08x", h)
}
