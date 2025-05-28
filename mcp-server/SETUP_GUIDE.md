# AI RPG MCP Server - Complete Setup Guide

This guide walks you through setting up and using the AI RPG MCP Server with Claude Desktop.

## What is This?

The AI RPG MCP Server is a **Model Context Protocol (MCP) server** that wraps your AI RPG MVP game engine, allowing Claude (or other AI assistants) to:

- 🎮 Create and manage RPG player sessions
- 🤖 Generate contextual AI Game Master responses  
- 👥 Track NPC relationships and player reputation
- 🗺️ Manage locations and world state
- 📊 Monitor gameplay metrics and statistics
- ⚡ Execute game actions with real-time AI responses

## Quick Start

### 1. Build the Server

```bash
cd /Users/enrico/workspace/myobsidian/AI-RPG/mcp-server
make setup
```

This will:
- Install Go dependencies
- Create `.env` configuration file
- Build the MCP server binary
- Run basic tests

### 2. Configure Your API Keys

Edit the `.env` file with your AI provider credentials:

```bash
# For Claude (Anthropic)
AI_PROVIDER=claude
AI_API_KEY=your_claude_api_key_here

# Or for OpenAI
AI_PROVIDER=openai  
AI_API_KEY=your_openai_api_key_here
```

### 3. Test the Server

```bash
make test
# Or run the full test suite
./test.sh
```

### 4. Add to Claude Desktop

Add this configuration to your Claude Desktop settings:

**Location**: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)

```json
{
  "mcpServers": {
    "ai-rpg": {
      "command": "/Users/enrico/workspace/myobsidian/AI-RPG/mcp-server/ai-rpg-mcp-server",
      "args": []
    }
  }
}
```

Restart Claude Desktop and you'll see the AI RPG tools available!

## Available Tools

### Core Game Management

| Tool | Purpose | Example Use |
|------|---------|-------------|
| `create_session` | Create new player | "Create a session for player 'Gandalf'" |
| `execute_action` | Execute game commands | "Have the player look around the tavern" |
| `get_session_status` | Check current state | "What's the current status of this session?" |
| `update_location` | Move player | "Move the player to the ancient forest" |

### AI & NPC Management  

| Tool | Purpose | Example Use |
|------|---------|-------------|
| `generate_ai_response` | AI GM responses | "Generate a response to the player's question" |
| `update_npc_relationship` | Manage NPCs | "The tavern keeper now trusts the player more" |

### Analytics & Monitoring

| Tool | Purpose | Example Use |
|------|---------|-------------|
| `get_session_metrics` | View statistics | "Show me the session statistics" |
| `list_active_sessions` | List all sessions | "What sessions are currently active?" |

## Game Commands

When using `execute_action`, these commands are supported:

### Movement
- `/move forest` - Move to the forest
- `/go village` - Return to village
- `/look around` - Examine current area

### Social Interaction  
- `/talk tavern_keeper` - Speak with NPCs
- `/speak blacksmith` - Dialogue with characters

### Combat & Action
- `/attack goblin` - Combat encounters
- `/examine chest` - Investigate objects
- `/inventory` - Check equipment

### Exploration
- `/search room` - Look for hidden items
- `/listen` - Listen for sounds
- `/climb tower` - Physical actions

## Example RPG Session

Here's how Claude can use the tools to run an RPG session:

```
🧙‍♂️ Claude: I'll create a new RPG session for you as "Aria the Rogue"

[Uses create_session tool]

🎮 Session Created! You begin in the village of Thornwick...

👤 Player: "I want to look around the village"

🧙‍♂️ Claude: [Uses execute_action with "/look around"]

The village square bustles with activity. Merchants hawk their wares while children play near the old stone fountain. To the north stands "The Prancing Pony" tavern, its warm glow inviting. An elderly blacksmith works at her forge to the east, and mysterious forest paths wind south into the unknown...

👤 Player: "I'll talk to the tavern keeper"

🧙‍♂️ Claude: [Uses execute_action with "/talk tavern_keeper"]

Marcus the tavern keeper looks up from polishing a pewter mug. His weathered face breaks into a cautious smile. "Welcome, traveler! Name's Marcus - been keeping this place running for twenty years. What brings you to our humble village?" He glances at your rogue's gear with keen interest...

[Updates NPC relationship automatically]
```

## Advanced Features

### Persistent World State
- NPCs remember previous interactions
- Player reputation affects all encounters  
- Locations maintain history of visits
- Actions have lasting consequences

### AI-Powered Responses
- Context-aware Game Master responses
- Character-specific NPC dialogue
- Dynamic scene descriptions
- Adaptive difficulty based on player behavior

### Analytics Dashboard
```bash
# View detailed session metrics
make docs
```

## Troubleshooting

### Common Issues

**"Command not found"**
- Make sure Go is installed: `go version`
- Run `make build` to compile the server

**"API Key Error"**  
- Check your `.env` file has the correct API key
- Verify the AI provider is set correctly

**"Session not found"**
- Sessions are stored in memory - they reset when server restarts
- For persistence, configure PostgreSQL storage

**"Rate limit exceeded"**
- Adjust rate limiting in `.env`:
  ```
  AI_RATE_LIMIT_REQUESTS=30
  AI_RATE_LIMIT_DURATION=60s
  ```

### Debug Mode

```bash
# Enable detailed logging
LOG_LEVEL=debug make run
```

### Performance Tuning

```bash
# Enable caching for better performance
AI_ENABLE_CACHING=true
AI_CACHE_TTL=1800s  # 30 minutes
```

## Development

### Adding Custom Tools

1. Define tool schema in `handleToolsList()`
2. Implement logic in `executeToolCall()` 
3. Add helper functions as needed
4. Test with `make test`

### Custom Game Commands

Extend `parseGameCommand()` to support new actions:

```go
case strings.HasPrefix(command, "/cast"):
    actionType = "magic"
    target = "spell"
    consequences = []string{"magic_success", "mana_drain"}
```

## Production Deployment

### Database Storage
```bash
# Use PostgreSQL for persistence
DATABASE_URL=postgres://user:pass@localhost/airgp?sslmode=disable
```

### Scaling
- Multiple server instances for load balancing
- Redis for session sharing
- Rate limiting per user/session

## What's Next?

This MCP server provides the foundation for:

1. **🔮 Advanced AI Agents**: NPCs that learn and evolve
2. **💰 Web3 Integration**: NFT items and stablecoin economies  
3. **🎨 Visual Generation**: AI-generated scene artwork
4. **🌐 Multiplayer Support**: Shared persistent worlds
5. **📱 Mobile Integration**: Cross-platform gameplay

## Support

- Check the logs: Server outputs to stdout/stderr
- Test individual tools: `make test-create-session`
- Full test suite: `./test.sh`
- Issues: Check the AI-RPG repository

---

**🎉 You're now ready to create epic AI-powered RPG adventures!**

Start by asking Claude: *"Create a new RPG session for me as a brave knight named Sir Galahad, and let's begin an adventure!"*
