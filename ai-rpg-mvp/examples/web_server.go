package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"ai-rpg-mvp/context"
)

// GameServer represents our RPG game server
type GameServer struct {
	contextMgr *context.ContextManager
}

// PlayerCommand represents a command from the player
type PlayerCommand struct {
	SessionID string `json:"session_id"`
	Command   string `json:"command"`
	PlayerID  string `json:"player_id,omitempty"`
	PlayerName string `json:"player_name,omitempty"`
}

// GameResponse represents the server's response
type GameResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	SessionID string      `json:"session_id,omitempty"`
	Context   interface{} `json:"context,omitempty"`
	Error     string      `json:"error,omitempty"`
}

func main() {
	// Initialize context manager with in-memory storage
	// In production, you would use PostgreSQL storage
	storage := context.NewMemoryStorage()
	contextMgr := context.NewContextManager(storage)
	defer contextMgr.Shutdown()

	server := &GameServer{
		contextMgr: contextMgr,
	}

	// Setup HTTP routes
	http.HandleFunc("/api/session/create", server.handleCreateSession)
	http.HandleFunc("/api/game/action", server.handleGameAction)
	http.HandleFunc("/api/game/status", server.handleGameStatus)
	http.HandleFunc("/api/ai/prompt", server.handleAIPrompt)
	http.HandleFunc("/api/metrics", server.handleMetrics)

	// Serve static files for a simple web interface
	http.HandleFunc("/", server.handleIndex)

	fmt.Println("Starting game server on http://localhost:8080")
	fmt.Println("API Endpoints:")
	fmt.Println("  POST /api/session/create - Create new session")
	fmt.Println("  POST /api/game/action - Execute game action")
	fmt.Println("  GET  /api/game/status/:session_id - Get game status")
	fmt.Println("  GET  /api/ai/prompt/:session_id - Get AI prompt")
	fmt.Println("  GET  /api/metrics - Get system metrics")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s *GameServer) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmd PlayerCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		s.sendErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if cmd.PlayerID == "" || cmd.PlayerName == "" {
		s.sendErrorResponse(w, "PlayerID and PlayerName are required", http.StatusBadRequest)
		return
	}

	sessionID, err := s.contextMgr.CreateSession(cmd.PlayerID, cmd.PlayerName)
	if err != nil {
		s.sendErrorResponse(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	// Set initial location
	s.contextMgr.UpdateLocation(sessionID, "starting_village")

	response := GameResponse{
		Success:   true,
		Message:   fmt.Sprintf("Welcome to the adventure, %s! Your journey begins in a small village.", cmd.PlayerName),
		SessionID: sessionID,
	}

	s.sendJSONResponse(w, response)
}

func (s *GameServer) handleGameAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmd PlayerCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		s.sendErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if cmd.SessionID == "" || cmd.Command == "" {
		s.sendErrorResponse(w, "SessionID and Command are required", http.StatusBadRequest)
		return
	}

	// Process the command and generate response
	response, err := s.processGameCommand(cmd.SessionID, cmd.Command)
	if err != nil {
		s.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.sendJSONResponse(w, response)
}

func (s *GameServer) handleGameStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		s.sendErrorResponse(w, "session_id parameter is required", http.StatusBadRequest)
		return
	}

	summary, err := s.contextMgr.GetContextSummary(sessionID)
	if err != nil {
		s.sendErrorResponse(w, fmt.Sprintf("Failed to get context: %v", err), http.StatusNotFound)
		return
	}

	response := GameResponse{
		Success: true,
		Message: "Context retrieved successfully",
		Context: summary,
	}

	s.sendJSONResponse(w, response)
}

func (s *GameServer) handleAIPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		s.sendErrorResponse(w, "session_id parameter is required", http.StatusBadRequest)
		return
	}

	prompt, err := s.contextMgr.GenerateAIPrompt(sessionID)
	if err != nil {
		s.sendErrorResponse(w, fmt.Sprintf("Failed to generate prompt: %v", err), http.StatusNotFound)
		return
	}

	response := GameResponse{
		Success: true,
		Message: prompt,
	}

	s.sendJSONResponse(w, response)
}

func (s *GameServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := s.contextMgr.GetContextMetrics()
	
	response := GameResponse{
		Success: true,
		Message: "Metrics retrieved successfully",
		Context: metrics,
	}

	s.sendJSONResponse(w, response)
}

func (s *GameServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>AI RPG Context Tracking Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        button { padding: 10px 20px; margin: 5px; cursor: pointer; }
        input, textarea { width: 300px; padding: 8px; margin: 5px; }
        .output { background: #f5f5f5; padding: 15px; border-radius: 5px; white-space: pre-wrap; }
        .hidden { display: none; }
    </style>
</head>
<body>
    <div class="container">
        <h1>AI RPG Context Tracking Demo</h1>
        
        <div class="section">
            <h3>Create New Session</h3>
            <input type="text" id="playerName" placeholder="Player Name" value="Aragorn">
            <button onclick="createSession()">Create Session</button>
            <div id="sessionInfo" class="output hidden"></div>
        </div>
        
        <div class="section">
            <h3>Game Actions</h3>
            <input type="text" id="sessionId" placeholder="Session ID">
            <input type="text" id="command" placeholder="Command (e.g., /look around)" value="/look around">
            <button onclick="executeAction()">Execute Action</button>
            <div id="actionResult" class="output hidden"></div>
        </div>
        
        <div class="section">
            <h3>Context & AI Integration</h3>
            <button onclick="getStatus()">Get Game Status</button>
            <button onclick="getAIPrompt()">Generate AI Prompt</button>
            <button onclick="getMetrics()">Get System Metrics</button>
            <div id="contextResult" class="output hidden"></div>
        </div>
        
        <div class="section">
            <h3>Quick Test Scenario</h3>
            <button onclick="runTestScenario()">Run Complete Test Scenario</button>
            <div id="testResult" class="output hidden"></div>
        </div>
    </div>

    <script>
        let currentSessionId = '';

        async function createSession() {
            const playerName = document.getElementById('playerName').value;
            if (!playerName) {
                alert('Please enter a player name');
                return;
            }

            try {
                const response = await fetch('/api/session/create', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        player_id: 'player_' + Date.now(),
                        player_name: playerName
                    })
                });

                const data = await response.json();
                if (data.success) {
                    currentSessionId = data.session_id;
                    document.getElementById('sessionId').value = currentSessionId;
                    document.getElementById('sessionInfo').innerText = 
                        'Session ID: ' + data.session_id + '\nMessage: ' + data.message;
                    show('sessionInfo');
                } else {
                    alert('Error: ' + (data.error || 'Unknown error'));
                }
            } catch (error) {
                alert('Network error: ' + error.message);
            }
        }

        async function executeAction() {
            const sessionId = document.getElementById('sessionId').value;
            const command = document.getElementById('command').value;
            
            if (!sessionId || !command) {
                alert('Please enter session ID and command');
                return;
            }

            try {
                const response = await fetch('/api/game/action', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        session_id: sessionId,
                        command: command
                    })
                });

                const data = await response.json();
                document.getElementById('actionResult').innerText = JSON.stringify(data, null, 2);
                show('actionResult');
            } catch (error) {
                alert('Network error: ' + error.message);
            }
        }

        async function getStatus() {
            const sessionId = document.getElementById('sessionId').value || currentSessionId;
            if (!sessionId) {
                alert('Please enter session ID');
                return;
            }

            try {
                const response = await fetch('/api/game/status?session_id=' + sessionId);
                const data = await response.json();
                document.getElementById('contextResult').innerText = JSON.stringify(data, null, 2);
                show('contextResult');
            } catch (error) {
                alert('Network error: ' + error.message);
            }
        }

        async function getAIPrompt() {
            const sessionId = document.getElementById('sessionId').value || currentSessionId;
            if (!sessionId) {
                alert('Please enter session ID');
                return;
            }

            try {
                const response = await fetch('/api/ai/prompt?session_id=' + sessionId);
                const data = await response.json();
                document.getElementById('contextResult').innerText = data.message;
                show('contextResult');
            } catch (error) {
                alert('Network error: ' + error.message);
            }
        }

        async function getMetrics() {
            try {
                const response = await fetch('/api/metrics');
                const data = await response.json();
                document.getElementById('contextResult').innerText = JSON.stringify(data, null, 2);
                show('contextResult');
            } catch (error) {
                alert('Network error: ' + error.message);
            }
        }

        async function runTestScenario() {
            const output = document.getElementById('testResult');
            output.innerText = 'Running test scenario...\n';
            show('testResult');

            try {
                // Create session
                const sessionResp = await fetch('/api/session/create', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        player_id: 'test_player',
                        player_name: 'Test Hero'
                    })
                });
                const sessionData = await sessionResp.json();
                const testSessionId = sessionData.session_id;
                
                output.innerText += 'Created session: ' + testSessionId + '\n\n';

                // Execute several actions
                const actions = [
                    '/look around',
                    '/talk tavern_keeper',
                    '/attack goblin',
                    '/move forest',
                    '/examine chest'
                ];

                for (const action of actions) {
                    const actionResp = await fetch('/api/game/action', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            session_id: testSessionId,
                            command: action
                        })
                    });
                    const actionData = await actionResp.json();
                    output.innerText += 'Action: ' + action + '\nResponse: ' + actionData.message + '\n\n';
                }

                // Get final AI prompt
                const promptResp = await fetch('/api/ai/prompt?session_id=' + testSessionId);
                const promptData = await promptResp.json();
                output.innerText += '=== Final AI Prompt ===\n' + promptData.message;

            } catch (error) {
                output.innerText += 'Error: ' + error.message;
            }
        }

        function show(elementId) {
            document.getElementById(elementId).classList.remove('hidden');
        }
    </script>
</body>
</html>
    `
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *GameServer) processGameCommand(sessionID, command string) (GameResponse, error) {
	// Simple command processing logic
	// In a real game, this would be much more sophisticated
	
	ctx, err := s.contextMgr.GetContext(sessionID)
	if err != nil {
		return GameResponse{}, fmt.Errorf("session not found")
	}

	var actionType, target, outcome string
	var consequences []string

	switch {
	case command == "/look around" || command == "/look":
		actionType = "examine"
		target = "environment"
		outcome = fmt.Sprintf("You observe your surroundings in %s. The area is filled with interesting details.", ctx.Location.Current)
		consequences = []string{"exploration_success"}

	case command == "/talk tavern_keeper":
		actionType = "social"
		target = "tavern_keeper"
		outcome = "The tavern keeper greets you warmly and offers information about local happenings."
		consequences = []string{"social_success", "npc_noticed"}
		
		// Update NPC relationship
		s.contextMgr.UpdateNPCRelationship(sessionID, "tavern_keeper", "Marcus the Tavern Keeper", 5, 
			[]string{"friendly_conversation", "willing_to_help"})

	case command == "/attack goblin":
		actionType = "combat"
		target = "goblin"
		outcome = "You successfully defeat the goblin! You gain experience and reputation."
		consequences = []string{"combat_success", "reputation_increase"}
		
		// Apply consequences
		s.contextMgr.UpdateReputation(sessionID, 10)
		s.contextMgr.UpdateCharacterHealth(sessionID, -2) // Small damage taken

	case command == "/move forest" || command == "/go forest":
		actionType = "move"
		target = "forest"
		outcome = "You travel to the nearby forest. Tall trees surround you."
		consequences = []string{"location_change"}
		
		// Update location
		s.contextMgr.UpdateLocation(sessionID, "thornwick_forest")

	case command == "/examine chest" || command == "/search chest":
		actionType = "examine"
		target = "chest"
		outcome = "You find a valuable item in the chest! A magical ring is now in your inventory."
		consequences = []string{"item_gained", "exploration_success"}

	case command == "/inventory" || command == "/inv":
		actionType = "examine"
		target = "inventory"
		outcome = fmt.Sprintf("Health: %d/%d, Reputation: %d, Location: %s", 
			ctx.Character.Health.Current, ctx.Character.Health.Max, 
			ctx.Character.Reputation, ctx.Location.Current)
		consequences = []string{}

	default:
		actionType = "unknown"
		target = "unknown"
		outcome = "You try something, but nothing obvious happens. Try commands like /look, /talk tavern_keeper, /attack goblin, /move forest"
		consequences = []string{}
	}

	// Record the action
	err = s.contextMgr.RecordAction(sessionID, command, actionType, target, ctx.Location.Current, outcome, consequences)
	if err != nil {
		return GameResponse{}, fmt.Errorf("failed to record action: %v", err)
	}

	// Get updated context for response
	summary, err := s.contextMgr.GetContextSummary(sessionID)
	if err != nil {
		return GameResponse{}, fmt.Errorf("failed to get updated context: %v", err)
	}

	return GameResponse{
		Success: true,
		Message: outcome,
		Context: map[string]interface{}{
			"location":    summary.CurrentLocation,
			"health":      summary.PlayerHealth,
			"reputation":  summary.PlayerReputation,
			"mood":        summary.PlayerMood,
			"session_time": fmt.Sprintf("%.1f minutes", summary.SessionDuration),
		},
	}, nil
}

func (s *GameServer) sendJSONResponse(w http.ResponseWriter, response GameResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func (s *GameServer) sendErrorResponse(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := GameResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(response)
}
