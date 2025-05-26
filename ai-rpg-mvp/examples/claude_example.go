package main

import (
	"fmt"
	"log"
	"time"

	"ai-rpg-mvp/ai"
	"ai-rpg-mvp/config"
	"ai-rpg-mvp/context"
)

func main() {
	fmt.Println("🎮 AI RPG with Claude Integration - Basic Usage Example")
	fmt.Println("====================================================")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize context manager with in-memory storage
	storage := context.NewMemoryStorage()
	contextMgr := context.NewContextManager(storage)
	defer contextMgr.Shutdown()

	// Initialize AI service with Claude
	aiConfig := ai.AIConfig{
		Provider:           cfg.AI.Provider,
		APIKey:             cfg.AI.APIKey,
		Model:              cfg.AI.Model,
		MaxTokens:          cfg.AI.MaxTokens,
		Temperature:        cfg.AI.Temperature,
		Timeout:            cfg.AI.Timeout,
		MaxRetries:         cfg.AI.MaxRetries,
		RetryDelay:         cfg.AI.RetryDelay,
		RateLimitRequests:  cfg.AI.RateLimitRequests,
		RateLimitDuration:  cfg.AI.RateLimitDuration,
		EnableCaching:      cfg.AI.EnableCaching,
		CacheTTL:           cfg.AI.CacheTTL,
	}

	aiService, err := ai.NewAIService(aiConfig)
	if err != nil {
		log.Fatalf("Failed to initialize AI service: %v", err)
	}

	fmt.Printf("✅ Initialized AI service with %s provider\n\n", aiService.GetProviderName())

	// Create a new player session
	sessionID, err := contextMgr.CreateSession("player123", "Aragorn the Ranger")
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	fmt.Printf("🆔 Created session: %s\n\n", sessionID)

	// Run example RPG session with AI responses
	runAIRPGSession(contextMgr, aiService, sessionID)

	// Show final AI statistics
	showAIStatistics(aiService)
}

func runAIRPGSession(contextMgr *context.ContextManager, aiService *ai.AIService, sessionID string) {
	fmt.Println("🎭 Starting AI-Powered RPG Session")
	fmt.Println("==================================")

	// Define a series of player actions to demonstrate AI responses
	playerActions := []struct {
		command     string
		description string
	}{
		{"/look around", "Player examines the environment"},
		{"/talk to the tavern keeper", "Player initiates social interaction"},
		{"/attack the goblin scout", "Player engages in combat"},
		{"/examine the mysterious chest", "Player investigates an object"},
		{"/move to the enchanted forest", "Player travels to a new location"},
	}

	for i, action := range playerActions {
		fmt.Printf("\n--- Turn %d: %s ---\n", i+1, action.description)
		fmt.Printf("🎮 Player: %s\n", action.command)

		// Process the action through the context system
		if err := processPlayerAction(contextMgr, aiService, sessionID, action.command); err != nil {
			log.Printf("❌ Error processing action: %v", err)
			continue
		}

		// Add a small delay to simulate real gameplay
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n🏁 RPG Session Complete!")
}

func processPlayerAction(contextMgr *context.ContextManager, aiService *ai.AIService, sessionID, command string) error {
	// Update location for movement commands
	if command == "/move to the enchanted forest" {
		if err := contextMgr.UpdateLocation(sessionID, "enchanted_forest"); err != nil {
			return fmt.Errorf("failed to update location: %w", err)
		}
	}

	// Generate AI prompt based on current context
	prompt, err := contextMgr.GenerateAIPrompt(sessionID)
	if err != nil {
		return fmt.Errorf("failed to generate AI prompt: %w", err)
	}

	// Enhance prompt with player action
	fullPrompt := fmt.Sprintf("%s\n\nPlayer Action: %s\n\nAs the Game Master, respond to this player action with an engaging, contextual response.", prompt, command)

	// Get AI response
	aiResponse, err := aiService.GenerateGMResponse(fullPrompt)
	if err != nil {
		return fmt.Errorf("failed to get AI response: %w", err)
	}

	fmt.Printf("🤖 AI GM: %s\n", aiResponse)

	// Determine consequences based on action type
	var actionType, target string
	var consequences []string

	switch {
	case command == "/look around":
		actionType = "examine"
		target = "environment"
		consequences = []string{"exploration_success"}

	case command == "/talk to the tavern keeper":
		actionType = "social"
		target = "tavern_keeper"
		consequences = []string{"social_success", "npc_noticed"}
		
		// Update NPC relationship
		contextMgr.UpdateNPCRelationship(sessionID, "tavern_keeper", "Marcus the Tavern Keeper", 5, 
			[]string{"friendly_conversation", "helpful_information"})

	case command == "/attack the goblin scout":
		actionType = "combat"
		target = "goblin_scout"
		consequences = []string{"combat_success", "reputation_increase"}
		
		// Apply combat consequences
		contextMgr.UpdateReputation(sessionID, 10)
		contextMgr.UpdateCharacterHealth(sessionID, -3)

	case command == "/examine the mysterious chest":
		actionType = "examine"
		target = "chest"
		consequences = []string{"item_gained", "exploration_success"}

	case command == "/move to the enchanted forest":
		actionType = "move"
		target = "enchanted_forest"
		consequences = []string{"location_change", "exploration_success"}

	default:
		actionType = "unknown"
		target = "unknown"
		consequences = []string{}
	}

	// Record the action with AI-generated outcome
	ctx, _ := contextMgr.GetContext(sessionID)
	if err := contextMgr.RecordAction(sessionID, command, actionType, target, ctx.Location.Current, aiResponse, consequences); err != nil {
		return fmt.Errorf("failed to record action: %w", err)
	}

	return nil
}

func showAIStatistics(aiService *ai.AIService) {
	fmt.Println("\n📊 AI Service Statistics")
	fmt.Println("========================")

	stats := aiService.GetStats()
	
	fmt.Printf("Provider: %s\n", stats["provider"])
	fmt.Printf("Model: %s\n", stats["model"])

	if rateLimiter, ok := stats["rate_limiter"].(map[string]interface{}); ok {
		fmt.Printf("Rate Limiter - Available Tokens: %v\n", rateLimiter["available_tokens"])
		fmt.Printf("Rate Limiter - Max Tokens: %v\n", rateLimiter["max_tokens"])
	}

	if cache, ok := stats["cache"].(map[string]interface{}); ok {
		fmt.Printf("Cache - Hits: %v\n", cache["hits"])
		fmt.Printf("Cache - Misses: %v\n", cache["misses"])
		fmt.Printf("Cache - Hit Rate: %.2f%%\n", cache["hit_rate"].(float64)*100)
		fmt.Printf("Cache - Size: %v entries\n", cache["size"])
	}
}

func demonstrateAdvancedFeatures(contextMgr *context.ContextManager, aiService *ai.AIService, sessionID string) {
	fmt.Println("\n🚀 Advanced AI Features Demo")
	fmt.Println("============================")

	// Generate NPC dialogue
	fmt.Println("\n--- NPC Dialogue Generation ---")
	npcDialogue, err := aiService.GenerateNPCDialogue(
		"Eldara the Wise", 
		"Ancient elven mage, mysterious and knowledgeable, speaks in riddles", 
		"The player asks about the ancient prophecy and the location of the Crystal of Eternity",
	)
	if err != nil {
		log.Printf("Failed to generate NPC dialogue: %v", err)
	} else {
		fmt.Printf("🧙‍♀️ Eldara: %s\n", npcDialogue)
	}

	// Generate scene description
	fmt.Println("\n--- Scene Description Generation ---")
	sceneDescription, err := aiService.GenerateSceneDescription(
		"Ancient Elven Library", 
		"Player has just discovered a hidden chamber filled with ancient tomes and magical artifacts", 
		"Mysterious and awe-inspiring",
	)
	if err != nil {
		log.Printf("Failed to generate scene description: %v", err)
	} else {
		fmt.Printf("🏛️ Scene: %s\n", sceneDescription)
	}

	fmt.Println("\n✨ Demo complete!")
}
