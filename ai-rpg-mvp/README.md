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
- **Claude AI Integration**: Contextual Game Master responses with personality consistency
- **Structured Prompts**: Generate rich contextual AI prompts with full game state
- **NPC Dialogue Generation**: Character-specific dialogue with personality traits
- **Scene Descriptions**: Dynamic environmental descriptions based on context
- **Caching & Rate Limiting**: Optimized AI API usage with intelligent caching
- **Multiple Providers**: Pluggable AI provider system (Claude, OpenAI)

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

### 2. Starting the Project

#### Game Server
1. Navigate to the `ai-rpg-mvp` directory:
   ```bash
   cd ai-rpg-mvp
   ```
2. Copy the example environment file and update it with your settings:
   ```bash
   cp .env.example .env
   # Edit .env and add your API keys and configuration
   ```
3. Run the game server:
   ```bash
   make run
   ```

#### MCP Server
1. Navigate to the `mcp-server` directory:
   ```bash
   cd mcp-server
   ```
2. Copy the example environment file and update it with your settings:
   ```bash
   cp .env.example .env
   # Edit .env and add your API keys and configuration
   ```
3. Run the MCP server:
   ```bash
   make run
   ```

### 3. Claude AI Integration

```bash
# Set up your Claude API key
cp .env.example .env
# Edit .env and add your Claude API key:
# AI_API_KEY=your_claude_api_key_here

# Run with Claude AI GM
make run
```

The system will automatically generate contextual responses using Claude:

```
🎮 Player: /look around
🤖 AI GM: You find yourself in the heart of Thornwick village, where cobblestone 
paths wind between timber-framed houses. The evening air carries the scent of 
woodsmoke and fresh bread from the nearby tavern, while merchants pack up their 
stalls for the night. A hooded figure near the well quickly averts their gaze 
when they notice you watching.

🎮 Player: /talk to tavern keeper  
🤖 AI GM: Marcus, the burly tavern keeper, looks up from polishing a pewter mug. 
His weathered face creases into a cautious smile. "Welcome, traveler. I'm Marcus - 
been keeping this place running for nigh on twenty years. What brings you to our 
humble village?" His eyes briefly flick to your travel-worn gear with keen interest.
```

### 4. Database Setup (Production)

```go
// Use PostgreSQL for production
storage, err := context.NewPostgreSQLStorage("postgres://user:pass@localhost/rpgdb?sslmode=disable")
contextMgr := context.NewContextManager(storage)
```

## 🤖 AI Game Master Features

### Claude Integration
The system uses Claude's advanced language understanding for:

**🎭 Game Master Responses**
- Contextual storytelling based on player history
- Consistent world-building and lore maintenance  
- Dynamic difficulty adjustment based on player skill
- Rich environmental descriptions with sensory details

**👥 NPC Dialogue Generation**
```go
// Generate character-specific dialogue
npcDialogue, err := aiService.GenerateNPCDialogue(
    "Marcus the Tavern Keeper",
    "Gruff but helpful, suspicious of strangers, loves local gossip",
    "Player asks about strange noises from the old mine"
)
// Result: "Aye, been hearing those sounds myself. Started three nights back..."
```

**🌍 Scene Descriptions**
```go
// Generate atmospheric scene descriptions
sceneDesc, err := aiService.GenerateSceneDescription(
    "Ancient Elven Ruins",
    "Player discovered hidden chamber after solving puzzle",
    "Mysterious and awe-inspiring"
)
```

**⚡ Performance Features**
- **Intelligent Caching**: Reduce API calls with context-aware caching
- **Rate Limiting**: Built-in rate limiting to prevent API overuse
- **Retry Logic**: Automatic retry with exponential backoff
- **Fallback Responses**: Graceful degradation if AI service is unavailable

---

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