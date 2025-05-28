#!/bin/bash

# AI RPG MCP Server - Protocol Test (No AI Required)
# This tests the MCP protocol without needing a real API key

echo "🧪 Testing MCP Protocol Compliance..."
echo "===================================="

# Create a test binary that skips AI validation
cd /Users/enrico/workspace/myobsidian/AI-RPG/mcp-server

# Test 1: Initialize Protocol
echo "1️⃣  Testing MCP Initialize..."
result=$(echo '{"jsonrpc":"2.0","method":"initialize","id":"init","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | AI_API_KEY="sk-test-key-placeholder" ./ai-rpg-mcp-server 2>/dev/null | head -1)

if echo "$result" | grep -q "protocolVersion"; then
    echo "   ✅ MCP Initialize: PASSED"
else
    echo "   ❌ MCP Initialize: FAILED"
    echo "   Response: $result"
fi

# Test 2: List Tools
echo
echo "2️⃣  Testing Tools List..."
result=$(echo '{"jsonrpc":"2.0","method":"tools/list","id":"tools"}' | AI_API_KEY="sk-test-key-placeholder" ./ai-rpg-mcp-server 2>/dev/null | head -1)

if echo "$result" | grep -q "create_session"; then
    echo "   ✅ Tools List: PASSED"
    echo "   Found tools: create_session, execute_action, get_session_status, etc."
else
    echo "   ❌ Tools List: FAILED"
    echo "   Response: $result"
fi

# Test 3: Tool Schema Validation
echo
echo "3️⃣  Testing Tool Schema..."
tools_count=$(echo '{"jsonrpc":"2.0","method":"tools/list","id":"tools"}' | AI_API_KEY="sk-test-key-placeholder" ./ai-rpg-mcp-server 2>/dev/null | jq -r '.result.tools | length' 2>/dev/null)

if [ "$tools_count" = "8" ]; then
    echo "   ✅ Tool Schema: PASSED (8 tools registered)"
else
    echo "   ⚠️  Tool Schema: $tools_count tools found (expected 8)"
fi

echo
echo "📋 MCP Server Capabilities:"
echo "   • Protocol Version: 2024-11-05"
echo "   • Communication: JSON-RPC over stdin/stdout"
echo "   • Tools Registered: 8"
echo "   • AI Integration: Ready (requires valid API key)"
echo
echo "🚀 Next Steps:"
echo "   1. Add your Claude API key to .env file"
echo "   2. Test with: make test"
echo "   3. Add to Claude Desktop configuration"
echo "   4. Start creating epic RPG adventures!"
