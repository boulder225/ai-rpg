-- AI RPG Database Initialization Script
-- This script sets up the database schema for the AI RPG Context Tracking System

-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create player_contexts table (main context storage)
CREATE TABLE IF NOT EXISTS player_contexts (
    session_id VARCHAR(255) PRIMARY KEY,
    player_id VARCHAR(255) NOT NULL,
    context_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_update TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT valid_session_id CHECK (session_id != ''),
    CONSTRAINT valid_player_id CHECK (player_id != '')
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_player_contexts_player_id ON player_contexts(player_id);
CREATE INDEX IF NOT EXISTS idx_player_contexts_last_update ON player_contexts(last_update);
CREATE INDEX IF NOT EXISTS idx_player_contexts_context_data ON player_contexts USING GIN(context_data);
CREATE INDEX IF NOT EXISTS idx_player_contexts_created_at ON player_contexts(created_at);

-- Create context_cleanup_log table for tracking cleanup operations
CREATE TABLE IF NOT EXISTS context_cleanup_log (
    id SERIAL PRIMARY KEY,
    cleanup_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    contexts_removed INTEGER DEFAULT 0,
    cleanup_criteria TEXT,
    execution_time_ms INTEGER DEFAULT 0
);

-- Create context_metrics table for storing system metrics
CREATE TABLE IF NOT EXISTS context_metrics (
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(100) NOT NULL,
    metric_value NUMERIC NOT NULL,
    metric_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    session_id VARCHAR(255),
    metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_context_metrics_name_timestamp ON context_metrics(metric_name, metric_timestamp);
CREATE INDEX IF NOT EXISTS idx_context_metrics_session_id ON context_metrics(session_id);

-- Create AI prompt templates table for storing reusable prompts
CREATE TABLE IF NOT EXISTS ai_prompt_templates (
    id SERIAL PRIMARY KEY,
    template_name VARCHAR(100) NOT NULL UNIQUE,
    template_content TEXT NOT NULL,
    template_version INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB
);

-- Insert default AI prompt templates
INSERT INTO ai_prompt_templates (template_name, template_content, metadata) VALUES
('default_gm_prompt', 
'GAME MASTER CONTEXT

CURRENT GAME STATE:
- Location: {{.CurrentLocation}} (previously: {{.PreviousLocation}})
- Player Health: {{.PlayerHealth}}
- Player Reputation: {{.PlayerReputation}} ({{.ReputationDescription}})
- Session Duration: {{.SessionDuration}} minutes
- Player Mood: {{.PlayerMood}}

RECENT PLAYER ACTIONS:
{{.RecentActions}}

ACTIVE NPCS IN AREA:
{{.ActiveNPCs}}

GM INSTRUCTIONS:
You are the AI Game Master for this fantasy RPG session. Based on the current context:
1. Respond as the omniscient narrator and world
2. Maintain consistency with previous interactions
3. React appropriately to the player''s reputation and recent actions
4. Consider NPC relationships and dispositions
5. Provide immersive, contextual descriptions
6. Balance challenge with player agency

Current situation requires your response as Game Master.',
'{"type": "gm_prompt", "version": "1.0", "category": "default"}'::jsonb),

('npc_interaction_prompt',
'NPC INTERACTION CONTEXT

NPC DETAILS:
- Name: {{.NPCName}}
- Disposition towards player: {{.NPCDisposition}} ({{.NPCMood}})
- Known facts about player: {{.NPCKnownFacts}}
- Last interaction: {{.LastInteraction}}
- Current location: {{.CurrentLocation}}

PLAYER CONTEXT:
- Reputation: {{.PlayerReputation}}
- Recent actions: {{.RecentActions}}
- Current equipment: {{.PlayerEquipment}}

INTERACTION INSTRUCTIONS:
1. Respond as {{.NPCName}} based on their current disposition
2. Reference known facts about the player naturally
3. Maintain personality consistency
4. Consider the location and context
5. Provide meaningful dialogue options',
'{"type": "npc_prompt", "version": "1.0", "category": "social"}'::jsonb);

-- Create functions for common operations

-- Function to get context summary
CREATE OR REPLACE FUNCTION get_context_summary(p_session_id VARCHAR)
RETURNS JSON AS $$
DECLARE
    result JSON;
BEGIN
    SELECT json_build_object(
        'session_id', session_id,
        'player_id', player_id,
        'last_update', last_update,
        'created_at', created_at,
        'context_size', length(context_data::text),
        'current_location', context_data->'location'->>'current',
        'player_health', context_data->'character'->'health',
        'player_reputation', context_data->'character'->>'reputation'
    ) INTO result
    FROM player_contexts
    WHERE session_id = p_session_id;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup old contexts
CREATE OR REPLACE FUNCTION cleanup_old_contexts(older_than_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
    start_time TIMESTAMP;
    end_time TIMESTAMP;
    execution_time INTEGER;
BEGIN
    start_time := clock_timestamp();
    
    DELETE FROM player_contexts 
    WHERE last_update < (NOW() - INTERVAL '1 day' * older_than_days);
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    end_time := clock_timestamp();
    execution_time := EXTRACT(MILLISECONDS FROM (end_time - start_time));
    
    -- Log the cleanup operation
    INSERT INTO context_cleanup_log (contexts_removed, cleanup_criteria, execution_time_ms)
    VALUES (deleted_count, 'older_than_' || older_than_days || '_days', execution_time);
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to record metric
CREATE OR REPLACE FUNCTION record_metric(
    p_metric_name VARCHAR(100),
    p_metric_value NUMERIC,
    p_session_id VARCHAR(255) DEFAULT NULL,
    p_metadata JSONB DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO context_metrics (metric_name, metric_value, session_id, metadata)
    VALUES (p_metric_name, p_metric_value, p_session_id, p_metadata);
END;
$$ LANGUAGE plpgsql;

-- Create views for common queries

-- View for active sessions (updated in last hour)
CREATE OR REPLACE VIEW active_sessions AS
SELECT 
    session_id,
    player_id,
    context_data->'character'->>'name' as character_name,
    context_data->'location'->>'current' as current_location,
    context_data->'character'->>'reputation' as reputation,
    last_update,
    EXTRACT(MINUTES FROM (NOW() - created_at)) as session_duration_minutes
FROM player_contexts
WHERE last_update > (NOW() - INTERVAL '1 hour')
ORDER BY last_update DESC;

-- View for context statistics
CREATE OR REPLACE VIEW context_statistics AS
SELECT 
    COUNT(*) as total_contexts,
    COUNT(CASE WHEN last_update > (NOW() - INTERVAL '1 hour') THEN 1 END) as active_contexts,
    COUNT(CASE WHEN last_update > (NOW() - INTERVAL '1 day') THEN 1 END) as daily_active_contexts,
    AVG(LENGTH(context_data::text)) as avg_context_size_bytes,
    MIN(created_at) as oldest_context,
    MAX(last_update) as most_recent_update
FROM player_contexts;

-- Create a trigger to automatically update last_update timestamp
CREATE OR REPLACE FUNCTION update_last_update()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_update = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_last_update
    BEFORE UPDATE ON player_contexts
    FOR EACH ROW
    EXECUTE FUNCTION update_last_update();

-- Grant permissions (adjust as needed for your security requirements)
-- These are basic permissions for the application user
-- GRANT SELECT, INSERT, UPDATE, DELETE ON player_contexts TO rpguser;
-- GRANT SELECT, INSERT ON context_cleanup_log TO rpguser;
-- GRANT SELECT, INSERT ON context_metrics TO rpguser;
-- GRANT SELECT ON active_sessions TO rpguser;
-- GRANT SELECT ON context_statistics TO rpguser;

-- Insert some sample data for testing (optional)
-- Uncomment the following if you want sample data

/*
INSERT INTO player_contexts (session_id, player_id, context_data) VALUES
('sample-session-1', 'player-123', '{
    "character": {
        "name": "Aragorn",
        "health": {"current": 20, "max": 20},
        "reputation": 15,
        "equipment": [],
        "inventory": []
    },
    "location": {
        "current": "starting_village",
        "previous": "",
        "visit_count": 1
    },
    "actions": [],
    "npc_states": {},
    "session_stats": {
        "total_actions": 0,
        "combat_actions": 0,
        "social_actions": 0,
        "explore_actions": 0
    }
}'::jsonb),
('sample-session-2', 'player-456', '{
    "character": {
        "name": "Legolas", 
        "health": {"current": 18, "max": 20},
        "reputation": 25,
        "equipment": [{"name": "Elven Bow", "type": "weapon"}],
        "inventory": []
    },
    "location": {
        "current": "elven_forest",
        "previous": "starting_village", 
        "visit_count": 2
    },
    "actions": [
        {
            "type": "combat",
            "command": "/attack orc",
            "outcome": "Victory",
            "timestamp": "2024-01-01T12:00:00Z"
        }
    ],
    "npc_states": {
        "elf_lord": {
            "disposition": 30,
            "mood": "friendly",
            "known_facts": ["skilled_archer", "orc_slayer"]
        }
    },
    "session_stats": {
        "total_actions": 5,
        "combat_actions": 2, 
        "social_actions": 2,
        "explore_actions": 1
    }
}'::jsonb);
*/

-- Create a scheduled job to cleanup old contexts (requires pg_cron extension)
-- Uncomment if you have pg_cron installed and want automatic cleanup
-- SELECT cron.schedule('cleanup-old-contexts', '0 2 * * *', 'SELECT cleanup_old_contexts(30);');

COMMIT;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'AI RPG Database initialization completed successfully!';
    RAISE NOTICE 'Tables created: player_contexts, context_cleanup_log, context_metrics, ai_prompt_templates';
    RAISE NOTICE 'Views created: active_sessions, context_statistics';
    RAISE NOTICE 'Functions created: get_context_summary, cleanup_old_contexts, record_metric';
END $$;
