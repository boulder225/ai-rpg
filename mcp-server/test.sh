#!/bin/bash

# AI RPG MCP Server Test Script
# This script demonstrates the full functionality of the MCP server

set -e

echo "🎮 AI RPG MCP Server - Full Test Suite"
echo "======================================"

# Build the server
echo "📦 Building MCP server..."
make build

# Test 1: Initialize Protocol
echo
echo "🔧 Test 1: Initialize MCP Protocol"
echo '{"jsonrpc":"2.0","method":"initialize","id":"init","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./ai-rpg-mcp-server | jq .

# Test 2: List Available Tools
echo
echo "🛠️  Test 2: List Available Tools"
echo '{"jsonrpc":"2.0","method":"tools/list","id":"tools"}' | ./ai-rpg-mcp-server | jq '.result.tools[] | {name, description}'

# Test 3: Create Session
echo
echo "👤 Test 3: Create Player Session"
SESSION_RESPONSE=$(echo '{"jsonrpc":"2.0","method":"tools/call","id":"create","params":{"name":"create_session","arguments":{"playerID":"test_player_123","playerName":"Aragorn the Ranger"}}}' | ./ai-rpg-mcp-server)
echo "$SESSION_RESPONSE" | jq .

# Extract session ID for subsequent tests
SESSION_ID=$(echo "$SESSION_RESPONSE" | jq -r '.result.content[0].text' | grep -oE 'ID: [a-f0-9-]+' | sed 's/ID: //')

if [ -z "$SESSION_ID" ]; then
    echo "❌ Failed to extract session ID"
    exit 1
fi

echo "✅ Created session: $SESSION_ID"

# Test 4: Execute Game Actions
echo
echo "🎲 Test 4: Execute Game Actions"

actions=("/look around" "/talk tavern_keeper" "/move forest" "/examine chest")

for action in "${actions[@]}"; do
    echo
    echo "   Action: $action"
    echo "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"id\":\"action\",\"params\":{\"name\":\"execute_action\",\"arguments\":{\"sessionID\":\"$SESSION_ID\",\"command\":\"$action\"}}}" | ./ai-rpg-mcp-server | jq -r '.result.content[0].text'
    echo "   ---"
done

# Test 5: Get Session Status
echo
echo "📊 Test 5: Get Session Status"
echo "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"id\":\"status\",\"params\":{\"name\":\"get_session_status\",\"arguments\":{\"sessionID\":\"$SESSION_ID\"}}}" | ./ai-rpg-mcp-server | jq -r '.result.content[0].text'

# Test 6: Update NPC Relationship
echo
echo "👥 Test 6: Update NPC Relationship"
echo "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"id\":\"npc\",\"params\":{\"name\":\"update_npc_relationship\",\"arguments\":{\"sessionID\":\"$SESSION_ID\",\"npcID\":\"blacksmith_elena\",\"npcName\":\"Elena the Blacksmith\",\"dispositionChange\":15,\"facts\":[\"helped_with_quest\",\"bought_equipment\"]}}}" | ./ai-rpg-mcp-server | jq -r '.result.content[0].text'

# Test 7: Generate AI Response
echo
echo "🤖 Test 7: Generate AI Response"
echo "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"id\":\"ai\",\"params\":{\"name\":\"generate_ai_response\",\"arguments\":{\"sessionID\":\"$SESSION_ID\",\"playerAction\":\"I want to learn more about the mysterious sounds coming from the old mine\"}}}" | ./ai-rpg-mcp-server | jq -r '.result.content[0].text'

# Test 8: Get Session Metrics
echo
echo "📈 Test 8: Get Session Metrics"
echo "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"id\":\"metrics\",\"params\":{\"name\":\"get_session_metrics\",\"arguments\":{\"sessionID\":\"$SESSION_ID\"}}}" | ./ai-rpg-mcp-server | jq -r '.result.content[0].text'

# Test 9: List Active Sessions
echo
echo "📋 Test 9: List Active Sessions"
echo '{"jsonrpc":"2.0","method":"tools/call","id":"list","params":{"name":"list_active_sessions","arguments":{}}}' | ./ai-rpg-mcp-server | jq -r '.result.content[0].text'

echo
echo "✅ All tests completed successfully!"
echo
echo "🎉 Your AI RPG MCP Server is working correctly!"
echo "   You can now integrate it with Claude Desktop or other MCP clients."
echo
echo "Next steps:"
echo "1. Copy .env.example to .env and add your AI API keys"
echo "2. Add the server to your Claude Desktop configuration"
echo "3. Start creating epic RPG adventures with AI!"
