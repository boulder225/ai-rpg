package context

import (
	"time"
)

// PlayerContext represents the complete context state for a player session
type PlayerContext struct {
	// Identity & Session
	PlayerID   string    `json:"player_id"`
	SessionID  string    `json:"session_id"`
	StartTime  time.Time `json:"start_time"`
	LastUpdate time.Time `json:"last_update"`

	// Character State
	Character CharacterState `json:"character"`

	// Location & Movement
	Location LocationState `json:"location"`

	// Interaction History
	Actions []ActionEvent `json:"actions"`

	// Relationships
	NPCStates map[string]NPCRelationship `json:"npc_states"`

	// Session Metrics
	SessionStats SessionMetrics `json:"session_stats"`
}

// CharacterState represents the player's character information
type CharacterState struct {
	Name       string                 `json:"name"`
	Health     HealthStatus           `json:"health"`
	Equipment  []EquipmentItem        `json:"equipment"`
	Inventory  []InventoryItem        `json:"inventory"`
	Reputation int                    `json:"reputation"` // -100 to 100
	Attributes map[string]int         `json:"attributes"` // strength, charisma, etc.
	Metadata   map[string]interface{} `json:"metadata"`
}

// HealthStatus tracks character health
type HealthStatus struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

// EquipmentItem represents equipped items
type EquipmentItem struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"` // weapon, armor, accessory
	Slot     string                 `json:"slot"` // mainhand, offhand, chest, etc.
	Stats    map[string]int         `json:"stats"`
	Metadata map[string]interface{} `json:"metadata"`
}

// InventoryItem represents items in inventory
type InventoryItem struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Quantity int                    `json:"quantity"`
	Value    int                    `json:"value"`
	Metadata map[string]interface{} `json:"metadata"`
}

// LocationState tracks player movement and location history
type LocationState struct {
	Current        string    `json:"current"`
	Previous       string    `json:"previous"`
	VisitCount     int       `json:"visit_count"`
	FirstVisit     time.Time `json:"first_visit"`
	TimeInLocation int       `json:"time_in_location"` // minutes
	LocationHistory []LocationVisit `json:"location_history"`
}

// LocationVisit represents a visit to a location
type LocationVisit struct {
	Location  string    `json:"location"`
	EntryTime time.Time `json:"entry_time"`
	ExitTime  time.Time `json:"exit_time,omitempty"`
	Duration  int       `json:"duration"` // minutes
}

// ActionEvent represents a player action with context and consequences
type ActionEvent struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Type         string                 `json:"type"` // "move", "talk", "attack", "examine"
	Command      string                 `json:"command"`
	Target       string                 `json:"target,omitempty"`
	Location     string                 `json:"location"`
	Outcome      string                 `json:"outcome"`
	Consequences []string               `json:"consequences"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// NPCRelationship tracks relationship with specific NPCs
type NPCRelationship struct {
	NPCID            string    `json:"npc_id"`
	Name             string    `json:"name"`
	Disposition      int       `json:"disposition"` // -100 to 100
	FirstMet         time.Time `json:"first_met"`
	LastInteraction  time.Time `json:"last_interaction"`
	InteractionCount int       `json:"interaction_count"`
	KnownFacts       []string  `json:"known_facts"`
	Mood             string    `json:"mood"` // "friendly", "hostile", "neutral", "suspicious"
	Location         string    `json:"location"`
	Notes            []string  `json:"notes"`
}

// SessionMetrics tracks session statistics
type SessionMetrics struct {
	TotalActions   int     `json:"total_actions"`
	CombatActions  int     `json:"combat_actions"`
	SocialActions  int     `json:"social_actions"`
	ExploreActions int     `json:"explore_actions"`
	SessionTime    float64 `json:"session_time_minutes"`
	LocationsVisited int   `json:"locations_visited"`
	NPCsInteracted   int   `json:"npcs_interacted"`
}

// ContextSummary provides a condensed view for AI integration
type ContextSummary struct {
	CurrentLocation    string           `json:"current_location"`
	PreviousLocation   string           `json:"previous_location"`
	PlayerHealth       string           `json:"player_health"`
	PlayerReputation   int              `json:"player_reputation"`
	RecentActions      []string         `json:"recent_actions"`
	ActiveNPCs         []NPCContextInfo `json:"active_npcs"`
	SessionDuration    float64          `json:"session_duration_minutes"`
	PlayerMood         string           `json:"player_mood"`
	WorldState         map[string]interface{} `json:"world_state"`
}

// NPCContextInfo provides NPC information for AI context
type NPCContextInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Disposition  int      `json:"disposition"`
	Mood         string   `json:"mood"`
	KnownFacts   []string `json:"known_facts"`
	LastSeen     string   `json:"last_seen"`
	Location     string   `json:"location"`
	Relationship string   `json:"relationship"` // "stranger", "acquaintance", "friend", "enemy"
}

// ContextEvent represents an event to be processed by the context manager
type ContextEvent struct {
	SessionID string      `json:"session_id"`
	Event     ActionEvent `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
}

// ContextStorage interface for different storage implementations
type ContextStorage interface {
	LoadContext(sessionID string) (*PlayerContext, error)
	SaveContext(ctx *PlayerContext) error
	DeleteContext(sessionID string) error
	ListActiveSessions() ([]string, error)
}

// AIPromptData contains structured data for AI prompt generation
type AIPromptData struct {
	SessionContext  *ContextSummary `json:"session_context"`
	RecentEvents    []ActionEvent   `json:"recent_events"`
	WorldKnowledge  map[string]interface{} `json:"world_knowledge"`
	PlayerProfile   map[string]interface{} `json:"player_profile"`
	GMPersonality   map[string]interface{} `json:"gm_personality"`
}
