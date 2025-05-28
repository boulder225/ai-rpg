# AI RPG MCP Server

A Model Context Protocol (MCP) server that wraps the AI RPG MVP game server functionalities, enabling AI assistants to interact with the autonomous RPG system.

## Overview

This MCP server provides tools for AI assistants to:
- Create and manage player sessions
- Execute game actions with AI-generated responses
- Track player context, NPCs, and world state
- Generate contextual AI responses as Game Master
- Monitor session metrics and statistics

## Features

### Core Tools

- **create_session**: Create new player session with character name
- **execute_action**: Execute game actions with AI GM responses
- **get_session_status**: Retrieve current session context and state
- **update_location**: Move player to different locations
- **update_npc_relationship**: Manage NPC relationships and disposition
- **generate_ai_response**: Generate contextual AI Game Master responses
- **get_session_metrics**: View session statistics and metrics
- **list_active_sessions**: List all currently active player sessions

### AI Integration

- **Claude/OpenAI Support**: Integrated AI providers for GM responses
- **Contextual Responses**: Rich context-aware AI responses based on game state
- **NPC Dialogue**: Character-specific dialogue generation
- **Scene Descriptions**: Dynamic environmental descriptions

## Installation

1. Clone the repository and navigate to the MCP server directory:
```bash
cd ai-rpg-mvp/../mcp-server
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up configuration:
```bash
cp ../.env.example .env
# Edit .env with your AI API keys
```

4. Build the server:
```bash
go build -o ai-rpg-mcp-server main.go
```

## Usage

### Running the MCP Server

The MCP server communicates via JSON-RPC over stdin/stdout:

```bash
./ai-rpg-mcp-server
```

### Example Tool Calls

#### Creating a Session
```json
{
  "method": "tools/call",
  "params": {
    "name": "create_session",
    "arguments": {
      "playerID": "player123",
      "playerName": "Aragorn"
    }
  }
}
```

#### Executing Actions
```json
{
  "method": "tools/call",
  "params": {
    "name": "execute_action",
    "arguments": {
      "sessionID": "session-uuid-here",
      "command": "/look around"
    }
  }
}
```

#### Getting Session Status
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_session_status",
    "arguments": {
      "sessionID": "session-uuid-here"
    }
  }
}
```

## Game Commands

The server supports various game commands:

- **Movement**: `/move forest`, `/go village`
- **Interaction**: `/talk tavern_keeper`, `/speak npc_name`
- **Combat**: `/attack goblin`, `/fight monster`
- **Exploration**: `/look around`, `/examine chest`
- **Inventory**: `/inventory`, `/inv`

## Configuration

Environment variables:
```bash
AI_PROVIDER=claude          # or openai
AI_API_KEY=your_api_key
AI_MODEL=claude-3-sonnet-20240229
AI_MAX_TOKENS=1000
AI_TEMPERATURE=0.7
```

## Architecture

```
MCP Server
├── JSON-RPC Protocol Handler
├── Tool Registry (8 core tools)
├── AI RPG Context Manager
├── AI Service Integration
└── Game State Management
```

## Integration with Claude Desktop

### Configuration
1. Open Claude Desktop and navigate to the settings.
2. Locate the MCP server configuration section.
3. Add the following configuration:
   ```json
   {
     "mcpServers": {
       "ai-rpg": {
         "command": "/path/to/ai-rpg-mcp-server",
         "args": []
       }
     }
   }
   ```
4. Replace `/path/to/ai-rpg-mcp-server` with the actual path to your MCP server executable.
5. Save the configuration and restart Claude Desktop.

### Interacting with Claude MCP
You can interact with the MCP server using natural language queries. Here are some practical examples:

#### Example 1: Creating a Session
- **Query**: "Create a new session for player 'Aragorn'."
- **Response**: The MCP server will create a new session and return a session ID.

#### Example 2: Executing an Action
- **Query**: "Execute the command '/look around' in session 'session-uuid-here'."
- **Response**: The MCP server will execute the command and return the AI-generated response.

#### Example 3: Getting Session Status
- **Query**: "What is the current status of session 'session-uuid-here'?"
- **Response**: The MCP server will return the current session context and state.

#### Example 4: Updating Location
- **Query**: "Move the player to the 'forest' location in session 'session-uuid-here'."
- **Response**: The MCP server will update the player's location and return a confirmation.

#### Example 5: Updating NPC Relationship
- **Query**: "Update the relationship with 'tavern_keeper' to 'friendly' in session 'session-uuid-here'."
- **Response**: The MCP server will update the NPC relationship and return a confirmation.

### Additional Tips
- Ensure that the MCP server is running before sending queries.
- Use clear and specific queries to get accurate responses.
- Check the MCP server logs for any errors or issues.

## Development

### Adding New Tools

1. Define tool schema in `handleToolsList()`
2. Implement tool logic in `executeToolCall()`
3. Add helper functions as needed

### Testing

```bash
# Build and test
go build && echo '{"method":"tools/list","id":"test"}' | ./ai-rpg-mcp-server
```

## License

Part of the AI RPG MVP development roadmap.
