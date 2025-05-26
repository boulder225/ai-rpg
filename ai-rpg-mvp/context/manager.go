package context

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ContextManager manages player context and game state
type ContextManager struct {
	storage         ContextStorage
	cache          *sync.Map // session_id -> *PlayerContext
	eventQueue     chan ContextEvent
	shutdownCh     chan struct{}
	wg             sync.WaitGroup

	// Configuration
	maxActions      int           // Keep last N actions
	cacheTimeout    time.Duration // How long to keep in memory
	persistInterval time.Duration // How often to save to storage
}

// NewContextManager creates a new context manager instance
func NewContextManager(storage ContextStorage) *ContextManager {
	cm := &ContextManager{
		storage:         storage,
		cache:          &sync.Map{},
		eventQueue:     make(chan ContextEvent, 1000),
		shutdownCh:     make(chan struct{}),
		maxActions:     50,
		cacheTimeout:   30 * time.Minute,
		persistInterval: 5 * time.Minute,
	}

	// Start background processors
	cm.wg.Add(2)
	go cm.processEvents()
	go cm.persistentSaver()

	return cm
}

// Shutdown gracefully shuts down the context manager
func (cm *ContextManager) Shutdown() {
	close(cm.shutdownCh)
	cm.wg.Wait()
	
	// Save all cached contexts before shutdown
	cm.cache.Range(func(key, value interface{}) bool {
		ctx := value.(*PlayerContext)
		if err := cm.storage.SaveContext(ctx); err != nil {
			log.Printf("Error saving context during shutdown: %v", err)
		}
		return true
	})
}

// GetContext retrieves context for a session
func (cm *ContextManager) GetContext(sessionID string) (*PlayerContext, error) {
	// Check cache first
	if cached, ok := cm.cache.Load(sessionID); ok {
		return cached.(*PlayerContext), nil
	}

	// Load from storage
	ctx, err := cm.storage.LoadContext(sessionID)
	if err != nil {
		// Create new context if not found
		ctx = cm.createNewContext(sessionID)
	}

	// Cache for future use
	cm.cache.Store(sessionID, ctx)
	return ctx, nil
}

// CreateSession creates a new player session
func (cm *ContextManager) CreateSession(playerID, playerName string) (string, error) {
	sessionID := uuid.New().String()
	
	ctx := &PlayerContext{
		PlayerID:   playerID,
		SessionID:  sessionID,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Character: CharacterState{
			Name: playerName,
			Health: HealthStatus{
				Current: 20,
				Max:     20,
			},
			Reputation: 0,
			Equipment:  []EquipmentItem{},
			Inventory:  []InventoryItem{},
			Attributes: map[string]int{
				"strength":     10,
				"dexterity":    10,
				"intelligence": 10,
				"charisma":     10,
			},
			Metadata: make(map[string]interface{}),
		},
		Location: LocationState{
			Current:         "starting_village",
			Previous:        "",
			VisitCount:      1,
			FirstVisit:      time.Now(),
			TimeInLocation:  0,
			LocationHistory: []LocationVisit{},
		},
		Actions:    []ActionEvent{},
		NPCStates:  make(map[string]NPCRelationship),
		SessionStats: SessionMetrics{
			TotalActions:     0,
			CombatActions:    0,
			SocialActions:    0,
			ExploreActions:   0,
			SessionTime:      0,
			LocationsVisited: 1,
			NPCsInteracted:   0,
		},
	}

	// Cache and save
	cm.cache.Store(sessionID, ctx)
	if err := cm.storage.SaveContext(ctx); err != nil {
		return "", fmt.Errorf("failed to save new context: %w", err)
	}

	return sessionID, nil
}

// RecordAction records a player action with context
func (cm *ContextManager) RecordAction(sessionID, command, actionType, target, location, outcome string, consequences []string) error {
	action := ActionEvent{
		ID:           uuid.New().String(),
		Timestamp:    time.Now(),
		Type:         actionType,
		Command:      command,
		Target:       target,
		Location:     location,
		Outcome:      outcome,
		Consequences: consequences,
		Metadata:     make(map[string]interface{}),
	}

	// Queue for processing
	select {
	case cm.eventQueue <- ContextEvent{
		SessionID: sessionID,
		Event:     action,
		Timestamp: time.Now(),
	}:
		return nil
	default:
		return fmt.Errorf("event queue full")
	}
}

// UpdateLocation updates player location
func (cm *ContextManager) UpdateLocation(sessionID, newLocation string) error {
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return err
	}

	// Update location state
	if ctx.Location.Current != newLocation {
		// Record exit from previous location
		if len(ctx.Location.LocationHistory) > 0 && ctx.Location.LocationHistory[len(ctx.Location.LocationHistory)-1].ExitTime.IsZero() {
			lastVisit := &ctx.Location.LocationHistory[len(ctx.Location.LocationHistory)-1]
			lastVisit.ExitTime = time.Now()
			lastVisit.Duration = int(time.Since(lastVisit.EntryTime).Minutes())
		}

		// Update current location
		ctx.Location.Previous = ctx.Location.Current
		ctx.Location.Current = newLocation
		ctx.Location.TimeInLocation = 0

		// Add to location history
		ctx.Location.LocationHistory = append(ctx.Location.LocationHistory, LocationVisit{
			Location:  newLocation,
			EntryTime: time.Now(),
		})

		// Increment stats
		ctx.SessionStats.LocationsVisited++
		if ctx.Location.FirstVisit.IsZero() {
			ctx.Location.FirstVisit = time.Now()
		}
	}

	ctx.LastUpdate = time.Now()
	cm.cache.Store(sessionID, ctx)

	return nil
}

// UpdateNPCRelationship updates relationship with an NPC
func (cm *ContextManager) UpdateNPCRelationship(sessionID, npcID, npcName string, dispositionChange int, facts []string) error {
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return err
	}

	if ctx.NPCStates == nil {
		ctx.NPCStates = make(map[string]NPCRelationship)
	}

	npcRel, exists := ctx.NPCStates[npcID]
	if !exists {
		npcRel = NPCRelationship{
			NPCID:       npcID,
			Name:        npcName,
			Disposition: 0,
			FirstMet:    time.Now(),
			KnownFacts:  []string{},
			Mood:        "neutral",
			Location:    ctx.Location.Current,
			Notes:       []string{},
		}
		ctx.SessionStats.NPCsInteracted++
	}

	// Update relationship
	npcRel.Disposition += dispositionChange
	
	// Clamp disposition to valid range
	if npcRel.Disposition > 100 {
		npcRel.Disposition = 100
	} else if npcRel.Disposition < -100 {
		npcRel.Disposition = -100
	}

	npcRel.LastInteraction = time.Now()
	npcRel.InteractionCount++
	npcRel.Location = ctx.Location.Current

	// Add new facts
	for _, fact := range facts {
		if !contains(npcRel.KnownFacts, fact) {
			npcRel.KnownFacts = append(npcRel.KnownFacts, fact)
		}
	}

	// Update mood based on disposition
	npcRel.Mood = cm.calculateMood(npcRel.Disposition)

	ctx.NPCStates[npcID] = npcRel
	ctx.LastUpdate = time.Now()
	cm.cache.Store(sessionID, ctx)

	return nil
}

// UpdateCharacterHealth updates player health
func (cm *ContextManager) UpdateCharacterHealth(sessionID string, healthChange int) error {
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return err
	}

	ctx.Character.Health.Current += healthChange
	
	// Clamp health
	if ctx.Character.Health.Current > ctx.Character.Health.Max {
		ctx.Character.Health.Current = ctx.Character.Health.Max
	} else if ctx.Character.Health.Current < 0 {
		ctx.Character.Health.Current = 0
	}

	ctx.LastUpdate = time.Now()
	cm.cache.Store(sessionID, ctx)

	return nil
}

// UpdateReputation updates player reputation
func (cm *ContextManager) UpdateReputation(sessionID string, reputationChange int) error {
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return err
	}

	ctx.Character.Reputation += reputationChange
	
	// Clamp reputation
	if ctx.Character.Reputation > 100 {
		ctx.Character.Reputation = 100
	} else if ctx.Character.Reputation < -100 {
		ctx.Character.Reputation = -100
	}

	ctx.LastUpdate = time.Now()
	cm.cache.Store(sessionID, ctx)

	return nil
}

// GetRecentActions gets recent actions for AI context
func (cm *ContextManager) GetRecentActions(sessionID string, count int) ([]ActionEvent, error) {
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return nil, err
	}

	actions := ctx.Actions
	if len(actions) > count {
		actions = actions[len(actions)-count:]
	}

	return actions, nil
}

// createNewContext creates a new player context
func (cm *ContextManager) createNewContext(sessionID string) *PlayerContext {
	return &PlayerContext{
		SessionID:  sessionID,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Character: CharacterState{
			Health: HealthStatus{
				Current: 20,
				Max:     20,
			},
			Reputation: 0,
			Equipment:  []EquipmentItem{},
			Inventory:  []InventoryItem{},
			Attributes: make(map[string]int),
			Metadata:   make(map[string]interface{}),
		},
		Location: LocationState{
			Current:         "unknown",
			VisitCount:      0,
			LocationHistory: []LocationVisit{},
		},
		Actions:      []ActionEvent{},
		NPCStates:    make(map[string]NPCRelationship),
		SessionStats: SessionMetrics{},
	}
}

// calculateMood determines NPC mood based on disposition
func (cm *ContextManager) calculateMood(disposition int) string {
	switch {
	case disposition >= 50:
		return "friendly"
	case disposition >= 25:
		return "helpful"
	case disposition >= 0:
		return "neutral"
	case disposition >= -25:
		return "suspicious"
	case disposition >= -50:
		return "unfriendly"
	default:
		return "hostile"
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
