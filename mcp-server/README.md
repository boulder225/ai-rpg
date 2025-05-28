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

## Key Player-AI Interactions from Demo Session

### 1. Character Creation and Game Start

```
Player: Start a new game as a character named Conan
AI-RPG: [Creates a new session] Welcome, Conan! Your adventure begins in a small frontier village...
```

### 2. Basic NPC Interaction

```
Player: /talk blacksmith
AI-RPG: You approach the blacksmith's forge... A burly man with a thick beard notices you. "Well met, traveler," he calls out. "I am Thoren, the local blacksmith. What brings you to these parts?"
```

### 3. Personality-Driven Interactions

```
Player: /talk blacksmith "Hey Thoren" and kiss him
AI-RPG: The blacksmith is surprised by your unexpected greeting but laughs heartily. "Well now, you're a bold one, aren't you?" Thoren chuckles. "I don't normally get such... enthusiastic greetings from strangers."
```

### 4. Multi-Language Support

```
Player: "Ti amo Thoren!" e mi siedo per terra
AI-RPG: The blacksmith's eyes go wide as you proclaim your love in a foreign tongue and sit on the ground. "I'm... not sure I follow, traveler," Thoren says slowly. "Your words are strange to me."
```

### 5. Mission Acquisition

```
Player: /talk mayor
AI-RPG: The mayor tells you about beasts plaguing the village - goblins, worgs, and even trolls. "Bring me proof you've culled the menace - ears from goblins, pelts from worgs, and heads of any trolls - and you'll be paid generously."
```

### 6. Equipment Negotiation

```
Player: /talk blacksmith "Unfortunately I don't have money with me, but if you could point me to someone willing to hire me to drive away some nasty monsters, I'll be happy to pay you upon my return!"
AI-RPG: Thoren nods slowly. "Ah, so you're seeking employment as a monster-slayer, are you?" He scratches his beard. "The mayor has put out a bounty on any who can clear those vile creatures from our borders."
```

### 7. Autonomous Simulation

```
Player: Hey GM, run the game autonomously while I'm away
AI-RPG: [Autonomous mode begins]
[AI simulates Conan's journey into the woods, encounters with trolls and worgs, combat sequences, and return to the village with trophies - all without player input]
```

### 8. Combat Sequence Example

```
[During autonomous simulation]
AI-RPG: The hulking troll lumbers forward, its massive arm sweeping down in a brutal overhead strike. But you're ready - you deftly sidestep the blow, the creature's club splintering the ground. Seizing the opening, you bring your longsword arcing up, cutting deep into its neck. The troll reels back, clawing at the ragged gash.

Player: [Returning to game] Use a spell scroll to weaken the dragon, then charge with my potion-enhanced strength
AI-RPG: You unfurl one of the magic scrolls obtained from cautious Mildreth. The arcane glyphs of Draconic Weakening flow from your tongue as the air around the mighty dragon seems to contract like an invisible vise. The fearsome beast roars in impotent fury, its claws and fangs seeming to dull as the magical energy diminishes its terrible power.
```

### 9. Character Memory Demonstration

```
[Later in the game when returning to the blacksmith]
Player: /talk blacksmith
AI-RPG: Thoren looks up from his forge, recognition flashing in his eyes. "Ah, the bold traveler returns! I trust your monster hunting went well? Did you find enough trophies to pay for that armor you admired?"
```

### 10. World State Persistence

```
Player: What monsters have I killed so far?
AI-RPG: You've successfully hunted and slain a troll in the Shalebark Woods, collecting its head as a trophy. You've also defeated a pack of worgs near the foothills of the Tannren Heights, gathering several of their pelts. Most impressively, you've vanquished a fearsome green dragon that was terrorizing the area, earning you significant recognition in the village.
```

### 11. Key Command Examples

```
/look [object/area]       # Examine surroundings or specific object
/talk [npc] "dialogue"    # Speak to an NPC with specific words
/attack [target]          # Initiate combat with a target
/examine [item]           # Closely inspect an item
/inventory                # Check what you're carrying
/cast [spell]             # Cast a magical spell
/use [item]               # Utilize an item in your possession
/buy [item]               # Attempt to purchase something
/sell [item]              # Attempt to sell something you own
```

### 12. GM System Controls

```
Player: Hey GM, update my status
AI-RPG: [Provides complete status update with character stats, inventory, relationships, and quest progress]

Player: Pause the game
AI-RPG: [Saves game state] Your game has been paused. Your session ID is 74c8839d-2ac2-4fec-aac4-539c65263def. You can resume anytime by asking to continue your game with Conan.
```

These examples highlight the system's capabilities for natural language interaction, character memory, autonomous operation, multi-language support, and complex narrative generation. The commands shown represent the core interface through which players interact with the AI-RPG system.
