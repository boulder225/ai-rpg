# AI RPG Project

This repository contains the AI RPG project, which includes a game server and an MCP server.

## Quick Start

### Game Server
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

### MCP Server
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

## Project Structure

- `ai-rpg-mvp/`: Contains the game server and context tracking system.
- `mcp-server/`: Contains the MCP server.

For more detailed information, refer to the README files in each directory. 