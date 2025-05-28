#!/bin/bash

# Debug script to capture Claude Desktop communication
# This will help us see exactly what Claude Desktop is sending

echo "🔍 Creating debug version of MCP server..."

# Create a debug binary that logs everything
cd /Users/enrico/workspace/myobsidian/AI-RPG/mcp-server

echo "📝 Testing MCP server with verbose logging..."

# Test with a simple initialize message
echo '{"jsonrpc":"2.0","method":"initialize","id":"debug-test","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"claude-desktop","version":"1.0.0"}}}' | AI_API_KEY="test-key" ./ai-rpg-mcp-server 2>&1

echo
echo "🔍 If the above works, the issue is likely in Claude Desktop configuration."
echo
echo "📋 Check your Claude Desktop config file:"
echo "   Location: ~/Library/Application Support/Claude/claude_desktop_config.json"
echo
echo "✅ Expected config format:"
cat << 'EOF'
{
  "mcpServers": {
    "ai-rpg": {
      "command": "/Users/enrico/workspace/myobsidian/AI-RPG/mcp-server/ai-rpg-mcp-server",
      "args": [],
      "env": {
        "AI_API_KEY": "your_real_claude_api_key_here"
      }
    }
  }
}
EOF

echo
echo "🚨 Common issues:"
echo "   1. Wrong file path in 'command' field"
echo "   2. Missing or invalid API key"
echo "   3. File permission issues"
echo "   4. Claude Desktop caching old config"
echo
echo "🔧 Quick fixes:"
echo "   1. Verify file exists: ls -la /Users/enrico/workspace/myobsidian/AI-RPG/mcp-server/ai-rpg-mcp-server"
echo "   2. Set executable: chmod +x /Users/enrico/workspace/myobsidian/AI-RPG/mcp-server/ai-rpg-mcp-server"
echo "   3. Restart Claude Desktop completely"
echo "   4. Clear Claude Desktop cache if needed"
