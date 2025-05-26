# AI RPG Context Tracking System

A comprehensive context tracking system for autonomous AI agents in RPG games. This system provides the memory and state management foundation for AI Game Masters, NPCs, and monsters to maintain awareness of player actions, relationships, and world state.

## Project Structure

```
ai-rpg-mvp/
├── go.mod                          # Go module configuration
├── context/                        # Core context tracking package
│   ├── types.go                   # Data structures and interfaces
│   ├── manager.go                 # Main context manager
│   ├── events.go                  # Event processing and background tasks
│   ├── storage.go                 # Storage implementations (Memory + PostgreSQL)
│   └── ai_integration.go          # AI prompt generation and integration
└── examples/                      # Usage examples and demos
    ├── basic_usage.go             # Simple command-line example
    └── web_server.go              # Complete web server with API
```

## Key Features

### 🧠 Intelligent Context Tracking
- **Player State Management**: Health, reputation, equipment, location tracking
- **Action History**: Detailed logging of player actions with outcomes and consequences
- **NPC Relationships**: Dynamic relationship tracking with disposition, mood, and memory
- **Location Awareness**: Movement history and location-based context

### ⚡ Performance Optimized
- **In-Memory Caching**: Fast access to active session contexts
- **Background Processing**: Asynchronous event processing without blocking gameplay
- **Automatic Persistence**: Periodic saving to prevent data loss
- **Memory Management**: Configurable limits to prevent memory bloat

### 🤖 AI Integration Ready
- **Structured Prompts**: Generate contextual AI prompts with full game state
- **Personality Analysis**: Determine player mood, play style, and preferences
- **Relationship Context**: Provide NPC disposition and interaction history
- **World Consistency**: Maintain coherent world state across AI interactions

### 🎮 Game-Ready Architecture
- **Concurrent Sessions**: Support multiple simultaneous players
- **Event Consequences**: Automatic processing of action outcomes
- **Flexible Storage**: Pluggable storage (Memory for dev, PostgreSQL for prod)
- **RESTful API**: Ready-to-use web API for game integration

## Quick Start

### 1. Basic Usage

```go
// Initialize context manager
storage := context.NewMemoryStorage()
contextMgr := context.NewContextManager(storage)
defer contextMgr.Shutdown()

// Create player session
sessionID, err := contextMgr.CreateSession("player123", "Aragorn")

// Record player actions
contextMgr.RecordAction(sessionID, "/attack goblin", "combat", "goblin", 
    "forest", "You defeat the goblin!", []string{"combat_success", "reputation_increase"})

// Update NPC relationships
contextMgr.UpdateNPCRelationship(sessionID, "tavern_keeper", "Marcus", 10, 
    []string{"friendly_conversation"})

// Generate AI prompt
prompt, err := contextMgr.GenerateAIPrompt(sessionID)
```

### 2. Web Server Example

```bash
# Run the example web server
go run examples/web_server.go

# Visit http://localhost:8080 for interactive demo
```

### 3. Database Setup (Production)

```go
// Use PostgreSQL for production
storage, err := context.NewPostgreSQLStorage("postgres://user:pass@localhost/rpgdb?sslmode=disable")
contextMgr := context.NewContextManager(storage)
```

## Core Components

### Context Manager
The central orchestrator that manages all player contexts, processes events, and maintains game state consistency.

```go
type ContextManager struct {
    storage         ContextStorage
    cache          *sync.Map
    eventQueue     chan ContextEvent
    maxActions     int           // Keep last N actions (default: 50)
    cacheTimeout   time.Duration // Cache expiry (default: 30min)
    persistInterval time.Duration // Save frequency (default: 5min)
}
```

### Player Context
Complete representation of a player's game state, including character stats, location history, action log, and NPC relationships.

```go
type PlayerContext struct {
    SessionID    string
    Character    CharacterState    // Health, equipment, reputation
    Location     LocationState     // Current/previous locations, visit history
    Actions      []ActionEvent     // Detailed action log with consequences
    NPCStates    map[string]NPCRelationship // Dynamic NPC relationships
    SessionStats SessionMetrics    // Gameplay statistics
}
```

### NPC Relationships
Dynamic tracking of player-NPC interactions with disposition, mood, and memory.

```go
type NPCRelationship struct {
    NPCID            string
    Disposition      int       // -100 (hostile) to +100 (friendly)
    Mood             string    // "friendly", "suspicious", "hostile", etc.
    KnownFacts       []string  // What the NPC knows about the player
    InteractionCount int       // Number of interactions
    LastInteraction  time.Time // When last seen
}
```

## AI Integration

### Contextual Prompt Generation

The system generates rich, contextual prompts for AI agents:

```
GAME MASTER CONTEXT

CURRENT GAME STATE:
- Location: thornwick_village (previously: forest_clearing)
- Player Health: 18/20
- Player Reputation: 35 (Respected)
- Session Duration: 23.4 minutes
- Player Mood: confident

RECENT PLAYER ACTIONS:
- 2 min ago: /attack goblin (combat) -> You defeat the goblin for 8 damage
- 5 min ago: /talk tavern_keeper (social) -> Marcus greets you warmly
- 8 min ago: /look around (examine) -> You see the village square

ACTIVE NPCS IN AREA:
- Marcus (tavern_keeper): friendly mood, ally relationship (last seen 5 min ago)
  - Knows: friendly_conversation, helped_village
- Elena (blacksmith): helpful mood, acquaintance relationship (last seen 15 min ago)

GM INSTRUCTIONS: Respond as the omniscient narrator...
```

### Context Summary API

```go
// Get condensed context for AI
summary, err := contextMgr.GetContextSummary(sessionID)

// Summary includes:
// - Current location and movement history
// - Player health, reputation, and mood
// - Recent actions with outcomes
// - Active NPC relationships and dispositions
// - Session statistics and player behavior patterns
```

## Example Scenarios

### 1. Combat Encounter
```go
// Player attacks goblin
contextMgr.RecordAction(sessionID, "/attack goblin", "combat", "goblin", 
    "forest", "You strike for 8 damage!", 
    []string{"combat_success", "health_damage", "reputation_increase"})

// Automatic consequence processing:
// - Player reputation increases (+5)
// - Player takes damage (-2 health)
// - Combat statistics updated
```

### 2. NPC Interaction
```go
// Player talks to merchant
contextMgr.UpdateNPCRelationship(sessionID, "merchant_tom", "Tom the Merchant", 
    10, []string{"bought_items", "regular_customer"})

// AI prompt will include:
// "Tom the Merchant: friendly mood, ally relationship
//  - Knows: bought_items, regular_customer
//  - Last interaction: 2 minutes ago"
```

### 3. Location Exploration
```go
// Player moves to new area
contextMgr.UpdateLocation(sessionID, "ancient_ruins")
contextMgr.RecordAction(sessionID, "/examine altar", "explore", "altar",
    "ancient_ruins", "You find mysterious runes carved into the stone",
    []string{"exploration_success", "lore_discovered"})

// Tracks: visit count, time spent, exploration history
```

## API Endpoints

The web server example provides a complete REST API:

- `POST /api/session/create` - Create new player session
- `POST /api/game/action` - Execute game action
- `GET /api/game/status` - Get current game status
- `GET /api/ai/prompt` - Generate AI prompt with context
- `GET /api/metrics` - System performance metrics

## Configuration

### Context Manager Settings
```go
contextMgr := context.NewContextManager(storage)
// Customize settings:
// - maxActions: Number of actions to keep in history (default: 50)
// - cacheTimeout: How long to keep contexts in memory (default: 30min)
// - persistInterval: How often to save to database (default: 5min)
```

### Storage Options
```go
// Development: In-memory storage
storage := context.NewMemoryStorage()

// Production: PostgreSQL storage
storage, err := context.NewPostgreSQLStorage(connectionString)
```

## Performance Metrics

The system provides comprehensive metrics:
- Active sessions count
- Cache hit rate
- Event queue size
- Average context size
- Background processing stats

Target Performance:
- **Context Retrieval**: <50ms
- **Memory Usage**: <1MB per session
- **Cache Hit Rate**: >90%
- **Event Processing**: Non-blocking

## Use Cases

### 1. AI Game Master
```go
// Generate contextual GM responses
prompt, _ := contextMgr.GenerateAIPrompt(sessionID)
gmResponse := callAIService(prompt)
```

### 2. Dynamic NPCs
```go
// NPCs react based on relationship history
npcContext := contextMgr.GetNPCRelationship(sessionID, "blacksmith")
if npcContext.Disposition > 50 {
    return "Welcome back, my friend! I have something special for you."
}
```

### 3. Persistent World
```go
// World events based on player actions
if ctx.Character.Reputation < -50 {
    // Trigger consequence events
    contextMgr.RecordAction(sessionID, "reputation_consequence", "event", 
        "village", "Guards approach you suspiciously", []string{"social_penalty"})
}
```

## Next Steps

This context tracking system provides the foundation for:
1. **Phase 2**: Evolving AI agents that learn and adapt
2. **Phase 3**: Web3 integration with persistent NFT assets  
3. **Phase 4**: AI-generated visual content based on context

The system is designed to scale and support the advanced AI agent behaviors described in your MVP roadmap.

## License

This project is part of the AI RPG MVP development roadmap.
