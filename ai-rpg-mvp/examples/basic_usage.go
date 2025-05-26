package main

import (
	"fmt"
	"log"
	"time"

	"ai-rpg-mvp/context"
)

func main() {
	// Initialize context manager with in-memory storage
	storage := context.NewMemoryStorage()
	contextMgr := context.NewContextManager(storage)
	defer contextMgr.Shutdown()

	// Create a new player session
	sessionID, err := contextMgr.CreateSession("player123", "Aragorn")
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	fmt.Printf("Created session: %s\n", sessionID)

	// Simulate a complete RPG session
	runExampleSession(contextMgr, sessionID)

	// Generate AI prompt at the end
	demonstrateAIIntegration(contextMgr, sessionID)
}

func runExampleSession(contextMgr *context.ContextManager, sessionID string) {
	fmt.Println("\n=== Simulating RPG Session ===")

	// Player starts in a village
	contextMgr.UpdateLocation(sessionID, "thornwick_village")
	fmt.Println("Player entered Thornwick Village")

	// Player examines the area
	contextMgr.RecordAction(
		sessionID,
		"/look around",
		"examine",
		"village_square",
		"thornwick_village",
		"You see a bustling village square with a tavern, blacksmith, and merchant stalls",
		[]string{"exploration_success"},
	)

	// Player talks to tavern keeper
	contextMgr.RecordAction(
		sessionID,
		"/talk tavern_keeper",
		"social",
		"tavern_keeper",
		"thornwick_village",
		"Marcus the tavern keeper greets you warmly",
		[]string{"social_success", "npc_noticed"},
	)

	// Update NPC relationship
	contextMgr.UpdateNPCRelationship(
		sessionID,
		"tavern_keeper_marcus",
		"Marcus the Tavern Keeper",
		10, // +10 disposition
		[]string{"met_in_village", "friendly_greeting"},
	)

	// Player gets into combat
	contextMgr.RecordAction(
		sessionID,
		"/attack goblin",
		"combat",
		"goblin_scout",
		"thornwick_village",
		"You successfully strike the goblin for 8 damage",
		[]string{"combat_success", "health_damage", "reputation_increase"},
	)

	// Apply combat consequences
	contextMgr.UpdateCharacterHealth(sessionID, -3) // Player takes some damage
	contextMgr.UpdateReputation(sessionID, 5)       // Gains reputation for defending village

	// Player moves to forest
	contextMgr.UpdateLocation(sessionID, "thornwick_forest")
	fmt.Println("Player moved to Thornwick Forest")

	// Player finds treasure
	contextMgr.RecordAction(
		sessionID,
		"/search chest",
		"examine",
		"treasure_chest",
		"thornwick_forest",
		"You find a magical sword in the ancient chest",
		[]string{"item_gained", "exploration_success"},
	)

	// Player returns to village
	contextMgr.UpdateLocation(sessionID, "thornwick_village")

	// Player talks to blacksmith about the sword
	contextMgr.RecordAction(
		sessionID,
		"/talk blacksmith",
		"social",
		"village_blacksmith",
		"thornwick_village",
		"The blacksmith examines your new sword with interest",
		[]string{"social_success", "npc_noticed"},
	)

	// Update blacksmith relationship
	contextMgr.UpdateNPCRelationship(
		sessionID,
		"blacksmith_elena",
		"Elena the Blacksmith",
		5, // +5 disposition
		[]string{"brought_magical_sword", "potential_customer"},
	)

	fmt.Println("Session simulation complete!")
}

func demonstrateAIIntegration(contextMgr *context.ContextManager, sessionID string) {
	fmt.Println("\n=== AI Integration Demo ===")

	// Generate context summary
	summary, err := contextMgr.GetContextSummary(sessionID)
	if err != nil {
		log.Printf("Failed to get context summary: %v", err)
		return
	}

	fmt.Printf("Context Summary:\n")
	fmt.Printf("  Current Location: %s\n", summary.CurrentLocation)
	fmt.Printf("  Previous Location: %s\n", summary.PreviousLocation)
	fmt.Printf("  Player Health: %s\n", summary.PlayerHealth)
	fmt.Printf("  Player Reputation: %d\n", summary.PlayerReputation)
	fmt.Printf("  Session Duration: %.1f minutes\n", summary.SessionDuration)
	fmt.Printf("  Player Mood: %s\n", summary.PlayerMood)

	fmt.Printf("\nRecent Actions:\n")
	for _, action := range summary.RecentActions {
		fmt.Printf("  - %s\n", action)
	}

	fmt.Printf("\nActive NPCs:\n")
	for _, npc := range summary.ActiveNPCs {
		fmt.Printf("  - %s (%s): %s mood, %s relationship\n", 
			npc.Name, npc.ID, npc.Mood, npc.Relationship)
		if len(npc.KnownFacts) > 0 {
			fmt.Printf("    Facts: %v\n", npc.KnownFacts)
		}
	}

	// Generate AI prompt
	fmt.Println("\n=== Generated AI Prompt ===")
	prompt, err := contextMgr.GenerateAIPrompt(sessionID)
	if err != nil {
		log.Printf("Failed to generate AI prompt: %v", err)
		return
	}

	fmt.Println(prompt)

	// Show recent actions in detail
	fmt.Println("\n=== Recent Actions Detail ===")
	recentActions, err := contextMgr.GetRecentActions(sessionID, 5)
	if err != nil {
		log.Printf("Failed to get recent actions: %v", err)
		return
	}

	for _, action := range recentActions {
		fmt.Printf("Action: %s\n", action.Command)
		fmt.Printf("  Type: %s\n", action.Type)
		fmt.Printf("  Target: %s\n", action.Target)
		fmt.Printf("  Location: %s\n", action.Location)
		fmt.Printf("  Outcome: %s\n", action.Outcome)
		fmt.Printf("  Consequences: %v\n", action.Consequences)
		fmt.Printf("  Time: %s\n", action.Timestamp.Format(time.RFC3339))
		fmt.Println("---")
	}

	// Demonstrate context metrics
	fmt.Println("\n=== Context Manager Metrics ===")
	metrics := contextMgr.GetContextMetrics()
	for key, value := range metrics {
		fmt.Printf("  %s: %v\n", key, value)
	}
}
