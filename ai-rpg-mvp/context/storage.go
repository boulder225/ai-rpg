package context

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// MemoryContextStorage provides in-memory storage for development
type MemoryContextStorage struct {
	contexts map[string]*PlayerContext
	mutex    sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryContextStorage {
	return &MemoryContextStorage{
		contexts: make(map[string]*PlayerContext),
	}
}

// LoadContext loads a context from memory
func (s *MemoryContextStorage) LoadContext(sessionID string) (*PlayerContext, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	ctx, exists := s.contexts[sessionID]
	if !exists {
		return nil, fmt.Errorf("context not found for session %s", sessionID)
	}

	// Return a copy to avoid concurrent modification
	contextCopy := *ctx
	return &contextCopy, nil
}

// SaveContext saves a context to memory
func (s *MemoryContextStorage) SaveContext(ctx *PlayerContext) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create a copy to avoid sharing references
	contextCopy := *ctx
	s.contexts[ctx.SessionID] = &contextCopy
	return nil
}

// DeleteContext removes a context from memory
func (s *MemoryContextStorage) DeleteContext(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.contexts, sessionID)
	return nil
}

// ListActiveSessions returns all active session IDs
func (s *MemoryContextStorage) ListActiveSessions() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sessions := make([]string, 0, len(s.contexts))
	for sessionID := range s.contexts {
		sessions = append(sessions, sessionID)
	}
	return sessions, nil
}

// GetStats returns storage statistics
func (s *MemoryContextStorage) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"total_contexts": len(s.contexts),
		"storage_type":   "memory",
	}
}

// PostgreSQLContextStorage provides PostgreSQL storage for production
type PostgreSQLContextStorage struct {
	db *sql.DB
}

// NewPostgreSQLStorage creates a new PostgreSQL storage instance
func NewPostgreSQLStorage(connectionString string) (*PostgreSQLContextStorage, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgreSQLContextStorage{db: db}

	// Initialize database schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the necessary database tables
func (s *PostgreSQLContextStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS player_contexts (
		session_id VARCHAR(255) PRIMARY KEY,
		player_id VARCHAR(255) NOT NULL,
		context_data JSONB NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		last_update TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		CONSTRAINT valid_session_id CHECK (session_id != '')
	);

	CREATE INDEX IF NOT EXISTS idx_player_contexts_player_id ON player_contexts(player_id);
	CREATE INDEX IF NOT EXISTS idx_player_contexts_last_update ON player_contexts(last_update);
	CREATE INDEX IF NOT EXISTS idx_player_contexts_context_data ON player_contexts USING GIN(context_data);

	-- Clean up old contexts periodically (optional)
	CREATE TABLE IF NOT EXISTS context_cleanup_log (
		id SERIAL PRIMARY KEY,
		cleanup_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		contexts_removed INTEGER DEFAULT 0
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

// LoadContext loads a context from PostgreSQL
func (s *PostgreSQLContextStorage) LoadContext(sessionID string) (*PlayerContext, error) {
	query := "SELECT context_data FROM player_contexts WHERE session_id = $1"

	var contextJSON []byte
	err := s.db.QueryRow(query, sessionID).Scan(&contextJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("context not found for session %s", sessionID)
		}
		return nil, fmt.Errorf("failed to load context: %w", err)
	}

	var ctx PlayerContext
	if err := json.Unmarshal(contextJSON, &ctx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return &ctx, nil
}

// SaveContext saves a context to PostgreSQL
func (s *PostgreSQLContextStorage) SaveContext(ctx *PlayerContext) error {
	query := `
		INSERT INTO player_contexts (session_id, player_id, context_data, last_update)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (session_id)
		DO UPDATE SET 
			context_data = EXCLUDED.context_data,
			last_update = EXCLUDED.last_update
	`

	contextJSON, err := json.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	_, err = s.db.Exec(query, ctx.SessionID, ctx.PlayerID, contextJSON, ctx.LastUpdate)
	if err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	return nil
}

// DeleteContext removes a context from PostgreSQL
func (s *PostgreSQLContextStorage) DeleteContext(sessionID string) error {
	query := "DELETE FROM player_contexts WHERE session_id = $1"
	
	result, err := s.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete context: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("context not found for session %s", sessionID)
	}

	return nil
}

// ListActiveSessions returns all active session IDs
func (s *PostgreSQLContextStorage) ListActiveSessions() ([]string, error) {
	query := "SELECT session_id FROM player_contexts ORDER BY last_update DESC"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []string
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return nil, fmt.Errorf("failed to scan session ID: %w", err)
		}
		sessions = append(sessions, sessionID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return sessions, nil
}

// GetContextsByPlayer returns all contexts for a specific player
func (s *PostgreSQLContextStorage) GetContextsByPlayer(playerID string) ([]PlayerContext, error) {
	query := "SELECT context_data FROM player_contexts WHERE player_id = $1 ORDER BY last_update DESC"

	rows, err := s.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contexts by player: %w", err)
	}
	defer rows.Close()

	var contexts []PlayerContext
	for rows.Next() {
		var contextJSON []byte
		if err := rows.Scan(&contextJSON); err != nil {
			return nil, fmt.Errorf("failed to scan context data: %w", err)
		}

		var ctx PlayerContext
		if err := json.Unmarshal(contextJSON, &ctx); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context: %w", err)
		}

		contexts = append(contexts, ctx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return contexts, nil
}

// CleanupOldContexts removes contexts older than the specified duration
func (s *PostgreSQLContextStorage) CleanupOldContexts(olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	
	query := "DELETE FROM player_contexts WHERE last_update < $1"
	
	result, err := s.db.Exec(query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old contexts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Log cleanup
	logQuery := "INSERT INTO context_cleanup_log (contexts_removed) VALUES ($1)"
	s.db.Exec(logQuery, rowsAffected)

	return int(rowsAffected), nil
}

// GetStats returns storage statistics
func (s *PostgreSQLContextStorage) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total contexts
	var totalContexts int
	err := s.db.QueryRow("SELECT COUNT(*) FROM player_contexts").Scan(&totalContexts)
	if err != nil {
		return nil, fmt.Errorf("failed to get total contexts: %w", err)
	}
	stats["total_contexts"] = totalContexts

	// Active contexts (updated in last hour)
	var activeContexts int
	cutoff := time.Now().Add(-1 * time.Hour)
	err = s.db.QueryRow("SELECT COUNT(*) FROM player_contexts WHERE last_update > $1", cutoff).Scan(&activeContexts)
	if err != nil {
		return nil, fmt.Errorf("failed to get active contexts: %w", err)
	}
	stats["active_contexts"] = activeContexts

	// Storage type
	stats["storage_type"] = "postgresql"

	// Average context size (approximate)
	var avgSize sql.NullFloat64
	err = s.db.QueryRow("SELECT AVG(LENGTH(context_data::text)) FROM player_contexts").Scan(&avgSize)
	if err == nil && avgSize.Valid {
		stats["avg_context_size_bytes"] = int(avgSize.Float64)
	}

	return stats, nil
}

// Close closes the database connection
func (s *PostgreSQLContextStorage) Close() error {
	return s.db.Close()
}

// BackupContexts exports all contexts to JSON for backup
func (s *PostgreSQLContextStorage) BackupContexts() ([]byte, error) {
	query := "SELECT context_data FROM player_contexts ORDER BY session_id"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to backup contexts: %w", err)
	}
	defer rows.Close()

	var contexts []PlayerContext
	for rows.Next() {
		var contextJSON []byte
		if err := rows.Scan(&contextJSON); err != nil {
			return nil, fmt.Errorf("failed to scan context data: %w", err)
		}

		var ctx PlayerContext
		if err := json.Unmarshal(contextJSON, &ctx); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context: %w", err)
		}

		contexts = append(contexts, ctx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return json.MarshalIndent(contexts, "", "  ")
}
