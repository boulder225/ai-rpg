package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"ai-rpg-mvp/ai"
	"ai-rpg-mvp/config"
	"ai-rpg-mvp/context"
)

// MCP Protocol Messages (JSON-RPC 2.0 compliant)
type MCPMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`  // Can be string, number, or null
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`  // Can be string, number, or null
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP Tool Definitions
type MCPTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type MCPToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type MCPToolResult struct {
	Content []MCPContent `json:"content"`
}

type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AI RPG MCP Server
type AIRPGMCPServer struct {
	contextMgr *context.ContextManager
	aiService  *ai.AIService
	config     *config.Config
}

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize context manager
	storage := context.NewMemoryStorage()
	contextMgr := context.NewContextManager(storage)
	defer contextMgr.Shutdown()

	// Initialize AI service
	aiConfig := ai.AIConfig{
		Provider:           cfg.AI.Provider,
		APIKey:             cfg.AI.APIKey,
		Model:              cfg.AI.Model,
		MaxTokens:          cfg.AI.MaxTokens,
		Temperature:        cfg.AI.Temperature,
		Timeout:            cfg.AI.Timeout,
		MaxRetries:         cfg.AI.MaxRetries,
		RetryDelay:         cfg.AI.RetryDelay,
		RateLimitRequests:  cfg.AI.RateLimitRequests,
		RateLimitDuration:  cfg.AI.RateLimitDuration,
		EnableCaching:      cfg.AI.EnableCaching,
		CacheTTL:           cfg.AI.CacheTTL,
	}

	aiService, err := ai.NewAIService(aiConfig)
	if err != nil {
		log.Fatalf("Failed to initialize AI service: %v", err)
	}

	server := &AIRPGMCPServer{
		contextMgr: contextMgr,
		aiService:  aiService,
		config:     cfg,
	}

	log.Println("AI RPG MCP Server started - reading from stdin...")
	server.run()
}

func (s *AIRPGMCPServer) run() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Log incoming message for debugging
		log.Printf("Received message: %s", line)

		var msg MCPMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Printf("Parse error: %v", err)
			// Send JSON-RPC 2.0 parse error
			parseErrorResponse := MCPResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &MCPError{
					Code:    -32700,
					Message: "Parse error",
				},
			}
			s.sendMessage(parseErrorResponse)
			continue
		}

		log.Printf("Parsed message - Method: %s, ID: %v", msg.Method, msg.ID)
		s.handleMessage(msg)
	}
}

func (s *AIRPGMCPServer) handleMessage(msg MCPMessage) {
	// Validate JSON-RPC 2.0 format
	if msg.JSONRPC != "2.0" {
		log.Printf("Invalid JSON-RPC version: %s", msg.JSONRPC)
		s.sendError(msg.ID, -32600, "Invalid Request: jsonrpc field must be '2.0'")
		return
	}

	// Validate method is provided
	if msg.Method == "" {
		log.Printf("Missing method field")
		s.sendError(msg.ID, -32600, "Invalid Request: method field is required")
		return
	}

	log.Printf("Handling method: %s", msg.Method)

	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg.ID)
	case "tools/list":
		s.handleToolsList(msg.ID)
	case "tools/call":
		s.handleToolCall(msg.ID, msg.Params)
	case "prompts/list":
		s.handlePromptsList(msg.ID)
	default:
		log.Printf("Unknown method: %s", msg.Method)
		s.sendError(msg.ID, -32601, "Method not found")
	}
}

func (s *AIRPGMCPServer) handleInitialize(id interface{}) {
	log.Printf("Handling initialize request with ID: %v", id)
	
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "ai-rpg-server",
			"version": "1.0.0",
		},
	}
	
	log.Printf("Sending initialize response")
	s.sendResponse(id, result)
}

func (s *AIRPGMCPServer) handleToolsList(id interface{}) {
	log.Printf("Handling tools/list request with ID: %v", id)
	
	tools := []MCPTool{
		{
			Name:        "create_session",
			Description: "Create a new AI RPG player session",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"playerID": map[string]interface{}{
						"type":        "string",
						"description": "Unique player identifier",
					},
					"playerName": map[string]interface{}{
						"type":        "string",
						"description": "Player character name",
					},
				},
				"required": []string{"playerID", "playerName"},
			},
		},
		{
			Name:        "execute_action",
			Description: "Execute a game action for a player",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sessionID": map[string]interface{}{
						"type":        "string",
						"description": "Player session identifier",
					},
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Game command to execute (e.g., '/look around', '/talk tavern_keeper')",
					},
				},
				"required": []string{"sessionID", "command"},
			},
		},
		{
			Name:        "get_session_status",
			Description: "Get current session status and context",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sessionID": map[string]interface{}{
						"type":        "string",
						"description": "Player session identifier",
					},
				},
				"required": []string{"sessionID"},
			},
		},
		{
			Name:        "update_location",
			Description: "Update player location",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sessionID": map[string]interface{}{
						"type":        "string",
						"description": "Player session identifier",
					},
					"location": map[string]interface{}{
						"type":        "string",
						"description": "New location name",
					},
				},
				"required": []string{"sessionID", "location"},
			},
		},
		{
			Name:        "update_npc_relationship",
			Description: "Update relationship with an NPC",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sessionID": map[string]interface{}{
						"type":        "string",
						"description": "Player session identifier",
					},
					"npcID": map[string]interface{}{
						"type":        "string",
						"description": "NPC identifier",
					},
					"npcName": map[string]interface{}{
						"type":        "string",
						"description": "NPC display name",
					},
					"dispositionChange": map[string]interface{}{
						"type":        "integer",
						"description": "Change in disposition (-100 to +100)",
					},
					"facts": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Facts the NPC learns about the player",
					},
				},
				"required": []string{"sessionID", "npcID", "npcName"},
			},
		},
		{
			Name:        "generate_ai_response",
			Description: "Generate AI Game Master response for current context",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sessionID": map[string]interface{}{
						"type":        "string",
						"description": "Player session identifier",
					},
					"playerAction": map[string]interface{}{
						"type":        "string",
						"description": "Player action to respond to",
					},
				},
				"required": []string{"sessionID", "playerAction"},
			},
		},
		{
			Name:        "get_session_metrics",
			Description: "Get session metrics and statistics",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sessionID": map[string]interface{}{
						"type":        "string",
						"description": "Player session identifier",
					},
				},
				"required": []string{"sessionID"},
			},
		},
		{
			Name:        "list_active_sessions",
			Description: "List all active player sessions",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	result := map[string]interface{}{
		"tools": tools,
	}
	
	log.Printf("Sending tools/list response with %d tools", len(tools))
	s.sendResponse(id, result)
}

func (s *AIRPGMCPServer) handlePromptsList(id interface{}) {
	log.Printf("Handling prompts/list request with ID: %v", id)
	
	// Return empty prompts list since we don't use prompts
	result := map[string]interface{}{
		"prompts": []interface{}{},
	}
	
	log.Printf("Sending empty prompts/list response")
	s.sendResponse(id, result)
}

func (s *AIRPGMCPServer) handleToolCall(id interface{}, params interface{}) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		s.sendError(id, -32602, "Invalid params")
		return
	}

	toolName, ok := paramsMap["name"].(string)
	if !ok {
		s.sendError(id, -32602, "Missing tool name")
		return
	}

	arguments, ok := paramsMap["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	result, err := s.executeToolCall(toolName, arguments)
	if err != nil {
		s.sendError(id, -32603, err.Error())
		return
	}

	s.sendResponse(id, result)
}

func (s *AIRPGMCPServer) executeToolCall(toolName string, args map[string]interface{}) (*MCPToolResult, error) {
	switch toolName {
	case "create_session":
		return s.toolCreateSession(args)
	case "execute_action":
		return s.toolExecuteAction(args)
	case "get_session_status":
		return s.toolGetSessionStatus(args)
	case "update_location":
		return s.toolUpdateLocation(args)
	case "update_npc_relationship":
		return s.toolUpdateNPCRelationship(args)
	case "generate_ai_response":
		return s.toolGenerateAIResponse(args)
	case "get_session_metrics":
		return s.toolGetSessionMetrics(args)
	case "list_active_sessions":
		return s.toolListActiveSessions(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// Tool Implementations

func (s *AIRPGMCPServer) toolCreateSession(args map[string]interface{}) (*MCPToolResult, error) {
	playerID, ok := args["playerID"].(string)
	if !ok {
		return nil, fmt.Errorf("playerID is required")
	}

	playerName, ok := args["playerName"].(string)
	if !ok {
		return nil, fmt.Errorf("playerName is required")
	}

	sessionID, err := s.contextMgr.CreateSession(playerID, playerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Set initial location
	s.contextMgr.UpdateLocation(sessionID, "starting_village")

	result := &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Session created for %s with ID: %s\nStarting location: starting_village", playerName, sessionID),
			},
		},
	}

	return result, nil
}

func (s *AIRPGMCPServer) toolExecuteAction(args map[string]interface{}) (*MCPToolResult, error) {
	sessionID, ok := args["sessionID"].(string)
	if !ok {
		return nil, fmt.Errorf("sessionID is required")
	}

	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command is required")
	}

	// Get current context
	ctx, err := s.contextMgr.GetContext(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Determine action type and consequences
	actionType, target, consequences := s.parseGameCommand(command)

	// Generate AI response
	prompt, err := s.contextMgr.GenerateAIPrompt(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI prompt: %w", err)
	}

	fullPrompt := fmt.Sprintf("%s\n\nPlayer Action: %s\n\nAs the Game Master, respond to this player action with an engaging, contextual response.", prompt, command)

	aiResponse, err := s.aiService.GenerateGMResponse(fullPrompt)
	if err != nil {
		log.Printf("AI service error: %v", err)
		aiResponse = fmt.Sprintf("You attempt to %s. The world responds to your action.", command)
	}

	// Record the action
	err = s.contextMgr.RecordAction(sessionID, command, actionType, target, ctx.Location.Current, aiResponse, consequences)
	if err != nil {
		return nil, fmt.Errorf("failed to record action: %w", err)
	}

	// Apply specific consequences
	s.applyActionConsequences(sessionID, command, consequences)

	// Get updated context
	summary, err := s.contextMgr.GetContextSummary(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated context: %w", err)
	}

	resultText := fmt.Sprintf("GM Response: %s\n\nCurrent Status:\n- Location: %s\n- Health: %s\n- Reputation: %d\n- Session Duration: %.1f minutes",
		aiResponse, summary.CurrentLocation, summary.PlayerHealth, summary.PlayerReputation, summary.SessionDuration)

	return &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

func (s *AIRPGMCPServer) toolGetSessionStatus(args map[string]interface{}) (*MCPToolResult, error) {
	sessionID, ok := args["sessionID"].(string)
	if !ok {
		return nil, fmt.Errorf("sessionID is required")
	}

	summary, err := s.contextMgr.GetContextSummary(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	// Format the response
	statusText := fmt.Sprintf(`Session Status for %s:

Current State:
- Location: %s (previously: %s)
- Health: %s
- Reputation: %d (%s)
- Mood: %s
- Session Duration: %.1f minutes

Recent Actions:
%s

Active NPCs:
%s`,
		sessionID,
		summary.CurrentLocation,
		summary.PreviousLocation,
		summary.PlayerHealth,
		summary.PlayerReputation,
		s.getReputationDescription(summary.PlayerReputation),
		summary.PlayerMood,
		summary.SessionDuration,
		strings.Join(summary.RecentActions, "\n"),
		s.formatNPCs(summary.ActiveNPCs),
	)

	return &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: statusText,
			},
		},
	}, nil
}

func (s *AIRPGMCPServer) toolUpdateLocation(args map[string]interface{}) (*MCPToolResult, error) {
	sessionID, ok := args["sessionID"].(string)
	if !ok {
		return nil, fmt.Errorf("sessionID is required")
	}

	location, ok := args["location"].(string)
	if !ok {
		return nil, fmt.Errorf("location is required")
	}

	err := s.contextMgr.UpdateLocation(sessionID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to update location: %w", err)
	}

	return &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Player location updated to: %s", location),
			},
		},
	}, nil
}

func (s *AIRPGMCPServer) toolUpdateNPCRelationship(args map[string]interface{}) (*MCPToolResult, error) {
	sessionID, ok := args["sessionID"].(string)
	if !ok {
		return nil, fmt.Errorf("sessionID is required")
	}

	npcID, ok := args["npcID"].(string)
	if !ok {
		return nil, fmt.Errorf("npcID is required")
	}

	npcName, ok := args["npcName"].(string)
	if !ok {
		return nil, fmt.Errorf("npcName is required")
	}

	dispositionChange := 0
	if val, ok := args["dispositionChange"].(float64); ok {
		dispositionChange = int(val)
	}

	var facts []string
	if factsInterface, ok := args["facts"].([]interface{}); ok {
		for _, fact := range factsInterface {
			if str, ok := fact.(string); ok {
				facts = append(facts, str)
			}
		}
	}

	err := s.contextMgr.UpdateNPCRelationship(sessionID, npcID, npcName, dispositionChange, facts)
	if err != nil {
		return nil, fmt.Errorf("failed to update NPC relationship: %w", err)
	}

	return &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Updated relationship with %s (ID: %s)\nDisposition change: %+d\nNew facts: %s",
					npcName, npcID, dispositionChange, strings.Join(facts, ", ")),
			},
		},
	}, nil
}

func (s *AIRPGMCPServer) toolGenerateAIResponse(args map[string]interface{}) (*MCPToolResult, error) {
	sessionID, ok := args["sessionID"].(string)
	if !ok {
		return nil, fmt.Errorf("sessionID is required")
	}

	playerAction, ok := args["playerAction"].(string)
	if !ok {
		return nil, fmt.Errorf("playerAction is required")
	}

	prompt, err := s.contextMgr.GenerateAIPrompt(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI prompt: %w", err)
	}

	fullPrompt := fmt.Sprintf("%s\n\nPlayer Action: %s\n\nAs the Game Master, respond to this player action with an engaging, contextual response.", prompt, playerAction)

	aiResponse, err := s.aiService.GenerateGMResponse(fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	return &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("AI GM Response: %s", aiResponse),
			},
		},
	}, nil
}

func (s *AIRPGMCPServer) toolGetSessionMetrics(args map[string]interface{}) (*MCPToolResult, error) {
	sessionID, ok := args["sessionID"].(string)
	if !ok {
		return nil, fmt.Errorf("sessionID is required")
	}

	ctx, err := s.contextMgr.GetContext(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	duration, err := s.contextMgr.GetSessionDuration(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session duration: %w", err)
	}

	metricsText := fmt.Sprintf(`Session Metrics for %s:

Statistics:
- Total Actions: %d
- Combat Actions: %d
- Social Actions: %d
- Exploration Actions: %d
- Session Duration: %s
- Locations Visited: %d
- NPCs Interacted: %d

Character State:
- Health: %d/%d
- Reputation: %d
- Equipment Items: %d
- Inventory Items: %d`,
		sessionID,
		ctx.SessionStats.TotalActions,
		ctx.SessionStats.CombatActions,
		ctx.SessionStats.SocialActions,
		ctx.SessionStats.ExploreActions,
		duration.String(),
		ctx.SessionStats.LocationsVisited,
		ctx.SessionStats.NPCsInteracted,
		ctx.Character.Health.Current,
		ctx.Character.Health.Max,
		ctx.Character.Reputation,
		len(ctx.Character.Equipment),
		len(ctx.Character.Inventory),
	)

	return &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: metricsText,
			},
		},
	}, nil
}

func (s *AIRPGMCPServer) toolListActiveSessions(args map[string]interface{}) (*MCPToolResult, error) {
	sessions := s.contextMgr.GetActiveSessions()

	if len(sessions) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{
				{
					Type: "text",
					Text: "No active sessions",
				},
			},
		}, nil
	}

	sessionsList := "Active Sessions:\n"
	for i, sessionID := range sessions {
		ctx, err := s.contextMgr.GetContext(sessionID)
		if err != nil {
			continue
		}

		duration, _ := s.contextMgr.GetSessionDuration(sessionID)
		sessionsList += fmt.Sprintf("%d. %s - %s (Duration: %s, Location: %s)\n",
			i+1, sessionID, ctx.Character.Name, duration.String(), ctx.Location.Current)
	}

	return &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: sessionsList,
			},
		},
	}, nil
}

// Helper functions

func (s *AIRPGMCPServer) parseGameCommand(command string) (string, string, []string) {
	var actionType, target string
	var consequences []string

	switch {
	case strings.HasPrefix(command, "/look") || strings.HasPrefix(command, "/examine"):
		actionType = "examine"
		target = "environment"
		consequences = []string{"exploration_success"}
	case strings.HasPrefix(command, "/talk") || strings.HasPrefix(command, "/speak"):
		actionType = "social"
		parts := strings.Fields(command)
		if len(parts) > 1 {
			target = parts[1]
		}
		consequences = []string{"social_success", "npc_noticed"}
	case strings.HasPrefix(command, "/attack") || strings.HasPrefix(command, "/fight"):
		actionType = "combat"
		parts := strings.Fields(command)
		if len(parts) > 1 {
			target = parts[1]
		}
		consequences = []string{"combat_success", "reputation_increase"}
	case strings.HasPrefix(command, "/move") || strings.HasPrefix(command, "/go"):
		actionType = "move"
		parts := strings.Fields(command)
		if len(parts) > 1 {
			target = parts[1]
		}
		consequences = []string{"location_change"}
	default:
		actionType = "unknown"
		target = "unknown"
		consequences = []string{}
	}

	return actionType, target, consequences
}

func (s *AIRPGMCPServer) applyActionConsequences(sessionID, command string, consequences []string) {
	for _, consequence := range consequences {
		switch consequence {
		case "reputation_increase":
			s.contextMgr.UpdateReputation(sessionID, 5)
		case "combat_success":
			s.contextMgr.UpdateReputation(sessionID, 10)
			s.contextMgr.UpdateCharacterHealth(sessionID, -2)
		case "location_change":
			if strings.Contains(command, "forest") {
				s.contextMgr.UpdateLocation(sessionID, "thornwick_forest")
			} else if strings.Contains(command, "village") {
				s.contextMgr.UpdateLocation(sessionID, "starting_village")
			}
		case "npc_noticed":
			if strings.Contains(command, "tavern_keeper") {
				s.contextMgr.UpdateNPCRelationship(sessionID, "tavern_keeper", "Marcus the Tavern Keeper", 5,
					[]string{"friendly_conversation", "noticed_player"})
			}
		}
	}
}

func (s *AIRPGMCPServer) getReputationDescription(reputation int) string {
	switch {
	case reputation >= 75:
		return "Heroic"
	case reputation >= 50:
		return "Well-regarded"
	case reputation >= 25:
		return "Respected"
	case reputation >= 0:
		return "Neutral"
	case reputation >= -25:
		return "Mistrusted"
	case reputation >= -50:
		return "Disliked"
	default:
		return "Notorious"
	}
}

func (s *AIRPGMCPServer) formatNPCs(npcs []context.NPCContextInfo) string {
	if len(npcs) == 0 {
		return "No active NPCs"
	}

	var result []string
	for _, npc := range npcs {
		result = append(result, fmt.Sprintf("- %s: %s mood, %s relationship (last seen %s)",
			npc.Name, npc.Mood, npc.Relationship, npc.LastSeen))
	}

	return strings.Join(result, "\n")
}

// MCP Protocol helpers

func (s *AIRPGMCPServer) sendResponse(id interface{}, result interface{}) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.sendMessage(response)
}

func (s *AIRPGMCPServer) sendError(id interface{}, code int, message string) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
	s.sendMessage(response)
}

func (s *AIRPGMCPServer) sendMessage(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	
	// Log outgoing message for debugging
	log.Printf("Sending response: %s", string(data))
	
	fmt.Println(string(data))
}
