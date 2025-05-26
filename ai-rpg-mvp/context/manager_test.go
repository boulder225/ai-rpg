package context

import (
	"testing"
	"time"
)

func TestContextManager_CreateSession(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	// Test creating a new session
	sessionID, err := cm.CreateSession("player123", "TestPlayer")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if sessionID == "" {
		t.Fatal("Expected non-empty session ID")
	}

	// Verify session exists
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		t.Fatalf("Failed to get created context: %v", err)
	}

	if ctx.PlayerID != "player123" {
		t.Errorf("Expected PlayerID 'player123', got '%s'", ctx.PlayerID)
	}

	if ctx.Character.Name != "TestPlayer" {
		t.Errorf("Expected character name 'TestPlayer', got '%s'", ctx.Character.Name)
	}

	if ctx.Character.Health.Current != 20 || ctx.Character.Health.Max != 20 {
		t.Errorf("Expected health 20/20, got %d/%d", ctx.Character.Health.Current, ctx.Character.Health.Max)
	}
}

func TestContextManager_RecordAction(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	// Record an action
	err := cm.RecordAction(sessionID, "/attack goblin", "combat", "goblin", "forest", "Hit for 8 damage", []string{"combat_success"})
	if err != nil {
		t.Fatalf("Failed to record action: %v", err)
	}

	// Give time for event processing
	time.Sleep(100 * time.Millisecond)

	// Verify action was recorded
	actions, err := cm.GetRecentActions(sessionID, 10)
	if err != nil {
		t.Fatalf("Failed to get recent actions: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}

	action := actions[0]
	if action.Command != "/attack goblin" {
		t.Errorf("Expected command '/attack goblin', got '%s'", action.Command)
	}

	if action.Type != "combat" {
		t.Errorf("Expected type 'combat', got '%s'", action.Type)
	}

	if action.Target != "goblin" {
		t.Errorf("Expected target 'goblin', got '%s'", action.Target)
	}
}

func TestContextManager_UpdateLocation(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	// Update location
	err := cm.UpdateLocation(sessionID, "new_forest")
	if err != nil {
		t.Fatalf("Failed to update location: %v", err)
	}

	// Verify location was updated
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		t.Fatalf("Failed to get context: %v", err)
	}

	if ctx.Location.Current != "new_forest" {
		t.Errorf("Expected current location 'new_forest', got '%s'", ctx.Location.Current)
	}

	if ctx.Location.Previous != "starting_village" {
		t.Errorf("Expected previous location 'starting_village', got '%s'", ctx.Location.Previous)
	}

	if len(ctx.Location.LocationHistory) != 1 {
		t.Errorf("Expected 1 location in history, got %d", len(ctx.Location.LocationHistory))
	}
}

func TestContextManager_UpdateNPCRelationship(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	// Update NPC relationship
	err := cm.UpdateNPCRelationship(sessionID, "npc1", "Test NPC", 15, []string{"friendly_greeting", "helpful"})
	if err != nil {
		t.Fatalf("Failed to update NPC relationship: %v", err)
	}

	// Verify NPC relationship was updated
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		t.Fatalf("Failed to get context: %v", err)
	}

	npcRel, exists := ctx.NPCStates["npc1"]
	if !exists {
		t.Fatal("Expected NPC relationship to exist")
	}

	if npcRel.Name != "Test NPC" {
		t.Errorf("Expected NPC name 'Test NPC', got '%s'", npcRel.Name)
	}

	if npcRel.Disposition != 15 {
		t.Errorf("Expected disposition 15, got %d", npcRel.Disposition)
	}

	if npcRel.Mood != "helpful" {
		t.Errorf("Expected mood 'helpful', got '%s'", npcRel.Mood)
	}

	if len(npcRel.KnownFacts) != 2 {
		t.Errorf("Expected 2 known facts, got %d", len(npcRel.KnownFacts))
	}
}

func TestContextManager_UpdateCharacterHealth(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	// Test damage
	err := cm.UpdateCharacterHealth(sessionID, -5)
	if err != nil {
		t.Fatalf("Failed to update character health: %v", err)
	}

	ctx, _ := cm.GetContext(sessionID)
	if ctx.Character.Health.Current != 15 {
		t.Errorf("Expected current health 15, got %d", ctx.Character.Health.Current)
	}

	// Test healing
	err = cm.UpdateCharacterHealth(sessionID, 3)
	if err != nil {
		t.Fatalf("Failed to heal character: %v", err)
	}

	ctx, _ = cm.GetContext(sessionID)
	if ctx.Character.Health.Current != 18 {
		t.Errorf("Expected current health 18, got %d", ctx.Character.Health.Current)
	}

	// Test over-healing (should cap at max)
	err = cm.UpdateCharacterHealth(sessionID, 10)
	if err != nil {
		t.Fatalf("Failed to over-heal character: %v", err)
	}

	ctx, _ = cm.GetContext(sessionID)
	if ctx.Character.Health.Current != 20 {
		t.Errorf("Expected current health capped at 20, got %d", ctx.Character.Health.Current)
	}
}

func TestContextManager_UpdateReputation(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	// Test reputation increase
	err := cm.UpdateReputation(sessionID, 25)
	if err != nil {
		t.Fatalf("Failed to update reputation: %v", err)
	}

	ctx, _ := cm.GetContext(sessionID)
	if ctx.Character.Reputation != 25 {
		t.Errorf("Expected reputation 25, got %d", ctx.Character.Reputation)
	}

	// Test reputation decrease
	err = cm.UpdateReputation(sessionID, -50)
	if err != nil {
		t.Fatalf("Failed to decrease reputation: %v", err)
	}

	ctx, _ = cm.GetContext(sessionID)
	if ctx.Character.Reputation != -25 {
		t.Errorf("Expected reputation -25, got %d", ctx.Character.Reputation)
	}

	// Test reputation cap (should clamp at -100)
	err = cm.UpdateReputation(sessionID, -100)
	if err != nil {
		t.Fatalf("Failed to set very low reputation: %v", err)
	}

	ctx, _ = cm.GetContext(sessionID)
	if ctx.Character.Reputation != -100 {
		t.Errorf("Expected reputation capped at -100, got %d", ctx.Character.Reputation)
	}
}

func TestContextManager_GetContextSummary(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	// Add some context
	cm.UpdateLocation(sessionID, "tavern")
	cm.UpdateReputation(sessionID, 30)
	cm.UpdateNPCRelationship(sessionID, "bartender", "Bob", 20, []string{"regular_customer"})

	// Get summary
	summary, err := cm.GetContextSummary(sessionID)
	if err != nil {
		t.Fatalf("Failed to get context summary: %v", err)
	}

	if summary.CurrentLocation != "tavern" {
		t.Errorf("Expected current location 'tavern', got '%s'", summary.CurrentLocation)
	}

	if summary.PlayerReputation != 30 {
		t.Errorf("Expected reputation 30, got %d", summary.PlayerReputation)
	}

	if summary.PlayerHealth != "20/20" {
		t.Errorf("Expected health '20/20', got '%s'", summary.PlayerHealth)
	}

	if len(summary.ActiveNPCs) != 1 {
		t.Errorf("Expected 1 active NPC, got %d", len(summary.ActiveNPCs))
	}

	if summary.ActiveNPCs[0].Name != "Bob" {
		t.Errorf("Expected NPC name 'Bob', got '%s'", summary.ActiveNPCs[0].Name)
	}
}

func TestContextManager_GenerateAIPrompt(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestHero")

	// Add some game context
	cm.UpdateLocation(sessionID, "dark_forest")
	cm.UpdateReputation(sessionID, 40)
	cm.RecordAction(sessionID, "/examine tree", "explore", "ancient_tree", "dark_forest", "You find strange markings", []string{"exploration_success"})
	
	// Give time for event processing
	time.Sleep(100 * time.Millisecond)

	// Generate AI prompt
	prompt, err := cm.GenerateAIPrompt(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate AI prompt: %v", err)
	}

	if prompt == "" {
		t.Fatal("Expected non-empty AI prompt")
	}

	// Check that prompt contains key information
	expectedStrings := []string{
		"dark_forest",
		"TestHero",
		"40",
		"examine tree",
		"GAME MASTER CONTEXT",
	}

	for _, expected := range expectedStrings {
		if !contains([]string{prompt}, expected) {
			t.Errorf("Expected AI prompt to contain '%s'", expected)
		}
	}
}

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()

	// Test saving and loading context
	ctx := &PlayerContext{
		SessionID: "test123",
		PlayerID:  "player456",
		Character: CharacterState{
			Name: "TestChar",
			Health: HealthStatus{Current: 15, Max: 20},
		},
	}

	// Save context
	err := storage.SaveContext(ctx)
	if err != nil {
		t.Fatalf("Failed to save context: %v", err)
	}

	// Load context
	loadedCtx, err := storage.LoadContext("test123")
	if err != nil {
		t.Fatalf("Failed to load context: %v", err)
	}

	if loadedCtx.SessionID != "test123" {
		t.Errorf("Expected session ID 'test123', got '%s'", loadedCtx.SessionID)
	}

	if loadedCtx.Character.Name != "TestChar" {
		t.Errorf("Expected character name 'TestChar', got '%s'", loadedCtx.Character.Name)
	}

	// Test non-existent context
	_, err = storage.LoadContext("nonexistent")
	if err == nil {
		t.Error("Expected error when loading non-existent context")
	}

	// Test listing sessions
	sessions, err := storage.ListActiveSessions()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	if sessions[0] != "test123" {
		t.Errorf("Expected session 'test123', got '%s'", sessions[0])
	}

	// Test delete
	err = storage.DeleteContext("test123")
	if err != nil {
		t.Fatalf("Failed to delete context: %v", err)
	}

	sessions, _ = storage.ListActiveSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions after delete, got %d", len(sessions))
	}
}

func TestEventProcessing(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	// Record action with consequences
	err := cm.RecordAction(sessionID, "/defeat dragon", "combat", "dragon", "mountain", 
		"Epic victory!", []string{"reputation_increase", "health_damage"})
	if err != nil {
		t.Fatalf("Failed to record action: %v", err)
	}

	// Give time for event processing
	time.Sleep(200 * time.Millisecond)

	// Check that consequences were processed
	ctx, _ := cm.GetContext(sessionID)
	
	// Should have gained reputation
	if ctx.Character.Reputation <= 0 {
		t.Error("Expected reputation increase from combat victory")
	}

	// Should have session stats updated
	if ctx.SessionStats.TotalActions != 1 {
		t.Errorf("Expected 1 total action, got %d", ctx.SessionStats.TotalActions)
	}

	if ctx.SessionStats.CombatActions != 1 {
		t.Errorf("Expected 1 combat action, got %d", ctx.SessionStats.CombatActions)
	}
}

func TestNPCMoodCalculation(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	testCases := []struct {
		disposition int
		expectedMood string
	}{
		{75, "friendly"},
		{30, "helpful"},
		{10, "neutral"},
		{-10, "suspicious"},
		{-30, "unfriendly"},
		{-60, "hostile"},
	}

	for _, tc := range testCases {
		mood := cm.calculateMood(tc.disposition)
		if mood != tc.expectedMood {
			t.Errorf("For disposition %d, expected mood '%s', got '%s'", 
				tc.disposition, tc.expectedMood, mood)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	// Simulate concurrent access
	done := make(chan bool, 10)

	// Multiple goroutines updating context
	for i := 0; i < 10; i++ {
		go func(i int) {
			cm.UpdateReputation(sessionID, 1)
			cm.UpdateCharacterHealth(sessionID, -1)
			cm.RecordAction(sessionID, fmt.Sprintf("/action_%d", i), "test", "target", "location", "outcome", []string{})
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Give time for event processing
	time.Sleep(500 * time.Millisecond)

	// Verify final state is consistent
	ctx, _ := cm.GetContext(sessionID)
	
	if ctx.Character.Reputation != 10 {
		t.Errorf("Expected reputation 10 from concurrent updates, got %d", ctx.Character.Reputation)
	}

	if ctx.Character.Health.Current != 10 {
		t.Errorf("Expected health 10 from concurrent updates, got %d", ctx.Character.Health.Current)
	}
}

func BenchmarkContextManager_GetContext(b *testing.B) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cm.GetContext(sessionID)
		if err != nil {
			b.Fatalf("Failed to get context: %v", err)
		}
	}
}

func BenchmarkContextManager_RecordAction(b *testing.B) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := cm.RecordAction(sessionID, "/test action", "test", "target", "location", "outcome", []string{})
		if err != nil {
			b.Fatalf("Failed to record action: %v", err)
		}
	}
}

func BenchmarkContextManager_GenerateAIPrompt(b *testing.B) {
	storage := NewMemoryStorage()
	cm := NewContextManager(storage)
	defer cm.Shutdown()

	sessionID, _ := cm.CreateSession("player123", "TestPlayer")
	
	// Add some context data
	cm.UpdateLocation(sessionID, "test_location")
	cm.UpdateReputation(sessionID, 25)
	cm.RecordAction(sessionID, "/test", "test", "target", "location", "outcome", []string{})
	time.Sleep(100 * time.Millisecond) // Let event process

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cm.GenerateAIPrompt(sessionID)
		if err != nil {
			b.Fatalf("Failed to generate AI prompt: %v", err)
		}
	}
}
