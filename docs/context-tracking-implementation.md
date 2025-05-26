---
title: "Basic Context Tracking Implementation"
date: "2025-05-25"
version: "1.0"
---

# Basic Context Tracking Implementation

## Overview

Context tracking enables the AI GM to maintain awareness of player state, world events, and session history to generate contextual responses. This system forms the memory foundation for autonomous AI agents.

---

## Core Data Structures

### Player Context Schema

```go
type PlayerContext struct {
    // Identity & Session
    PlayerID    string    `json:"player_id"`
    SessionID   string    `json:"session_id"`
    StartTime   time.Time `json:"start_time"`
    LastUpdate  time.Time `json:"last_update"`
    
    // Character State
    Character   CharacterState `json:"character"`
    
    // Location & Movement
    Location    LocationState  `json:"location"`
    
    // Interaction History
    Actions     []ActionEvent  `json:"actions"`
    
    // Relationships
    NPCStates   map[string]NPCRelationship `json:"npc_states"`
    
    // Session Metrics
    SessionStats SessionMetrics `json:"session_stats"`
}

type CharacterState struct {
    Name        string            `json:"name"`
    Health      HealthStatus      `json:"health"`
    Equipment   []EquipmentItem   `json:"equipment"`
    Inventory   []InventoryItem   `json:"inventory"`
    Reputation  int              `json:"reputation"` // -100 to 100
    Attributes  map[string]int   `json:"attributes"` // strength, charisma, etc.
}

type LocationState struct {
    Current     string    `json:"current"`
    Previous    string    `json:"previous"`
    VisitCount  int       `json:"visit_count"`
    FirstVisit  time.Time `json:"first_visit"`
    TimeInLocation int    `json:"time_in_location"` // minutes
}

type ActionEvent struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time             `json:"timestamp"`
    Type        string                `json:"type"` // "move", "talk", "attack", "examine"
    Command     string                `json:"command"`
    Target      string                `json:"target,omitempty"`
    Location    string                `json:"location"`
    Outcome     string                `json:"outcome"`
    Consequences []string             `json:"consequences"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type NPCRelationship struct {
    NPCID       string    `json:"npc_id"`
    Disposition int       `json:"disposition"` // -100 to 100
    FirstMet    time.Time `json:"first_met"`
    LastInteraction time.Time `json:"last_interaction"`
    InteractionCount int  `json:"interaction_count"`
    KnownFacts  []string  `json:"known_facts"`
    Mood        string    `json:"mood"` // "friendly", "hostile", "neutral"
}
```

---

## Context Manager Implementation

### Core Context Manager

```go
type ContextManager struct {
    storage     ContextStorage
    cache       *sync.Map // session_id -> PlayerContext
    eventQueue  chan ContextEvent
    
    // Configuration
    maxActions      int           // Keep last N actions
    cacheTimeout    time.Duration // How long to keep in memory
    persistInterval time.Duration // How often to save to storage
}

type ContextEvent struct {
    SessionID string
    Event     ActionEvent
}

func NewContextManager(storage ContextStorage) *ContextManager {
    cm := &ContextManager{
        storage:         storage,
        cache:          &sync.Map{},
        eventQueue:     make(chan ContextEvent, 1000),
        maxActions:     50,
        cacheTimeout:   30 * time.Minute,
        persistInterval: 5 * time.Minute,
    }
    
    // Start background processors
    go cm.processEvents()
    go cm.persistentSaver()
    
    return cm
}
```

### Context Operations

```go
// Get current context for a session
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

// Update context with new action
func (cm *ContextManager) RecordAction(sessionID, command, actionType, target, location string, outcome string, consequences []string) error {
    action := ActionEvent{
        ID:           generateActionID(),
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
    cm.eventQueue <- ContextEvent{
        SessionID: sessionID,
        Event:     action,
    }
    
    return nil
}

// Update location context
func (cm *ContextManager) UpdateLocation(sessionID, newLocation string) error {
    ctx, err := cm.GetContext(sessionID)
    if err != nil {
        return err
    }
    
    // Update location state
    if ctx.Location.Current != newLocation {
        ctx.Location.Previous = ctx.Location.Current
        ctx.Location.Current = newLocation
        ctx.Location.TimeInLocation = 0
        
        // Increment visit count for this location
        ctx.Location.VisitCount++
        if ctx.Location.FirstVisit.IsZero() {
            ctx.Location.FirstVisit = time.Now()
        }
    }
    
    ctx.LastUpdate = time.Now()
    cm.cache.Store(sessionID, ctx)
    
    return nil
}

// Update NPC relationship
func (cm *ContextManager) UpdateNPCRelationship(sessionID, npcID string, dispositionChange int, facts []string) error {
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
            Disposition: 0,
            FirstMet:    time.Now(),
            KnownFacts:  []string{},
            Mood:        "neutral",
        }
    }
    
    // Update relationship
    npcRel.Disposition += dispositionChange
    npcRel.LastInteraction = time.Now()
    npcRel.InteractionCount++
    
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
```

---

## Context Processing Engine

### Event Processing

```go
func (cm *ContextManager) processEvents() {
    for event := range cm.eventQueue {
        cm.processContextEvent(event)
    }
}

func (cm *ContextManager) processContextEvent(event ContextEvent) {
    ctx, err := cm.GetContext(event.SessionID)
    if err != nil {
        log.Printf("Error getting context for session %s: %v", event.SessionID, err)
        return
    }
    
    // Add action to history
    ctx.Actions = append(ctx.Actions, event.Event)
    
    // Trim action history if too long
    if len(ctx.Actions) > cm.maxActions {
        ctx.Actions = ctx.Actions[len(ctx.Actions)-cm.maxActions:]
    }
    
    // Process action consequences
    cm.processActionConsequences(ctx, event.Event)
    
    // Update session stats
    ctx.SessionStats.TotalActions++
    if event.Event.Type == "combat" {
        ctx.SessionStats.CombatActions++
    }
    
    ctx.LastUpdate = time.Now()
    cm.cache.Store(event.SessionID, ctx)
}

func (cm *ContextManager) processActionConsequences(ctx *PlayerContext, action ActionEvent) {
    for _, consequence := range action.Consequences {
        switch consequence {
        case "reputation_increase":
            ctx.Character.Reputation += 5
        case "reputation_decrease":
            ctx.Character.Reputation -= 10
        case "health_damage":
            if damage, ok := action.Metadata["damage"].(int); ok {
                ctx.Character.Health.Current -= damage
                if ctx.Character.Health.Current < 0 {
                    ctx.Character.Health.Current = 0
                }
            }
        case "npc_noticed":
            // Mark that NPCs in the area have noticed the player
            if npcID, ok := action.Metadata["npc_id"].(string); ok {
                cm.UpdateNPCRelationship(ctx.SessionID, npcID, 0, []string{
                    fmt.Sprintf("saw_player_with_%s", action.Target),
                })
            }
        }
    }
}
```

### Context Query Interface

```go
// Get recent actions for AI context
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

// Get context summary for AI prompt
func (cm *ContextManager) GetContextSummary(sessionID string) (*ContextSummary, error) {
    ctx, err := cm.GetContext(sessionID)
    if err != nil {
        return nil, err
    }
    
    summary := &ContextSummary{
        CurrentLocation:    ctx.Location.Current,
        PreviousLocation:   ctx.Location.Previous,
        PlayerHealth:       fmt.Sprintf("%d/%d", ctx.Character.Health.Current, ctx.Character.Health.Max),
        PlayerReputation:   ctx.Character.Reputation,
        RecentActions:      cm.getActionSummary(ctx.Actions, 5),
        ActiveNPCs:         cm.getRelevantNPCs(ctx),
        SessionDuration:    time.Since(ctx.StartTime).Minutes(),
    }
    
    return summary, nil
}

type ContextSummary struct {
    CurrentLocation    string            `json:"current_location"`
    PreviousLocation   string            `json:"previous_location"`
    PlayerHealth       string            `json:"player_health"`
    PlayerReputation   int               `json:"player_reputation"`
    RecentActions      []string          `json:"recent_actions"`
    ActiveNPCs         []NPCContextInfo  `json:"active_npcs"`
    SessionDuration    float64           `json:"session_duration_minutes"`
}

type NPCContextInfo struct {
    ID           string   `json:"id"`
    Disposition  int      `json:"disposition"`
    Mood         string   `json:"mood"`
    KnownFacts   []string `json:"known_facts"`
    LastSeen     string   `json:"last_seen"`
}
```

---

## Storage Implementation

### In-Memory Storage (Development)

```go
type MemoryContextStorage struct {
    contexts map[string]*PlayerContext
    mutex    sync.RWMutex
}

func NewMemoryStorage() *MemoryContextStorage {
    return &MemoryContextStorage{
        contexts: make(map[string]*PlayerContext),
    }
}

func (s *MemoryContextStorage) LoadContext(sessionID string) (*PlayerContext, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    ctx, exists := s.contexts[sessionID]
    if !exists {
        return nil, fmt.Errorf("context not found for session %s", sessionID)
    }
    
    return ctx, nil
}

func (s *MemoryContextStorage) SaveContext(ctx *PlayerContext) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    s.contexts[ctx.SessionID] = ctx
    return nil
}
```

### Database Storage (Production)

```go
type PostgreSQLContextStorage struct {
    db *sql.DB
}

func (s *PostgreSQLContextStorage) SaveContext(ctx *PlayerContext) error {
    query := `
        INSERT INTO player_contexts (session_id, context_data, last_update)
        VALUES ($1, $2, $3)
        ON CONFLICT (session_id)
        DO UPDATE SET context_data = $2, last_update = $3
    `
    
    contextJSON, err := json.Marshal(ctx)
    if err != nil {
        return err
    }
    
    _, err = s.db.Exec(query, ctx.SessionID, contextJSON, ctx.LastUpdate)
    return err
}

func (s *PostgreSQLContextStorage) LoadContext(sessionID string) (*PlayerContext, error) {
    query := "SELECT context_data FROM player_contexts WHERE session_id = $1"
    
    var contextJSON []byte
    err := s.db.QueryRow(query, sessionID).Scan(&contextJSON)
    if err != nil {
        return nil, err
    }
    
    var ctx PlayerContext
    err = json.Unmarshal(contextJSON, &ctx)
    if err != nil {
        return nil, err
    }
    
    return &ctx, nil
}
```

---

## Integration with AI GM

### Context to Prompt Conversion

```go
func (cm *ContextManager) GenerateAIPrompt(sessionID string) (string, error) {
    summary, err := cm.GetContextSummary(sessionID)
    if err != nil {
        return "", err
    }
    
    recentActions, err := cm.GetRecentActions(sessionID, 3)
    if err != nil {
        return "", err
    }
    
    prompt := fmt.Sprintf(`
CURRENT GAME STATE:
- Location: %s (previously: %s)
- Player Health: %s
- Player Reputation: %d
- Session Duration: %.1f minutes

RECENT PLAYER ACTIONS:
%s

ACTIVE NPCS:
%s

CONTEXT: The player is currently in %s. Respond as the Game Master with appropriate 
environmental description, NPC reactions, or event narration based on the current situation.
`,
        summary.CurrentLocation,
        summary.PreviousLocation,
        summary.PlayerHealth,
        summary.PlayerReputation,
        summary.SessionDuration,
        cm.formatRecentActions(recentActions),
        cm.formatActiveNPCs(summary.ActiveNPCs),
        summary.CurrentLocation,
    )
    
    return prompt, nil
}
```

---

## Usage Examples

### Basic Usage

```go
// Initialize context manager
storage := NewMemoryStorage()
contextMgr := NewContextManager(storage)

// Record player action
err := contextMgr.RecordAction(
    "session123",
    "/attack goblin",
    "combat",
    "goblin",
    "forest_clearing",
    "hit for 8 damage",
    []string{"combat_success", "goblin_wounded"},
)

// Update location
err = contextMgr.UpdateLocation("session123", "village_square")

// Update NPC relationship
err = contextMgr.UpdateNPCRelationship(
    "session123",
    "tavern_keeper",
    5, // +5 disposition
    []string{"player_helped_village", "trustworthy_stranger"},
)

// Generate AI prompt with context
prompt, err := contextMgr.GenerateAIPrompt("session123")
```

### Integration with Game Loop

```go
func handlePlayerCommand(sessionID, command string) (string, error) {
    // Parse command
    action := parsePlayerCommand(command)
    
    // Get current context
    ctx, err := contextMgr.GetContext(sessionID)
    if err != nil {
        return "", err
    }
    
    // Generate AI prompt with context
    prompt, err := contextMgr.GenerateAIPrompt(sessionID)
    if err != nil {
        return "", err
    }
    
    // Get AI response
    aiResponse, err := callAIGM(prompt)
    if err != nil {
        return "", err
    }
    
    // Record the action and outcome
    err = contextMgr.RecordAction(
        sessionID,
        command,
        action.Type,
        action.Target,
        ctx.Location.Current,
        aiResponse,
        action.Consequences,
    )
    
    return aiResponse, nil
}
```

---

## Performance Considerations

### Optimization Strategies

1. **Memory Management**: Limit action history to last 50 events
2. **Caching**: Keep active sessions in memory, persist to database periodically
3. **Lazy Loading**: Load context only when needed
4. **Batch Processing**: Process multiple context updates together
5. **Context Compression**: Summarize old actions instead of storing full details

### Monitoring Metrics

- Context retrieval time (target: <50ms)
- Memory usage per session (target: <1MB)
- Cache hit rate (target: >90%)
- Persistence frequency (every 5 minutes)
- Context accuracy (manual validation)

This implementation provides the foundation for contextual AI interactions while maintaining performance and scalability for your AI RPG platform.