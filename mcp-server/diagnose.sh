#!/bin/bash

echo "🔍 Claude Desktop MCP Integration Diagnostic"
echo "============================================="

# Check 1: Server binary
echo "1️⃣  Checking MCP server binary..."
SERVER_PATH="/Users/enrico/workspace/myobsidian/AI-RPG/mcp-server/ai-rpg-mcp-server"
if [ -f "$SERVER_PATH" ]; then
    echo "   ✅ Server binary exists"
    echo "   📊 Permissions: $(ls -la "$SERVER_PATH" | awk '{print $1}')"
    echo "   📏 Size: $(ls -lh "$SERVER_PATH" | awk '{print $5}')"
else
    echo "   ❌ Server binary NOT found at: $SERVER_PATH"
    echo "   🔧 Run: cd /Users/enrico/workspace/myobsidian/AI-RPG/mcp-server && go build -o ai-rpg-mcp-server main.go"
    exit 1
fi

# Check 2: Configuration file
echo
echo "2️⃣  Checking Claude Desktop configuration..."
CONFIG_PATH="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
if [ -f "$CONFIG_PATH" ]; then
    echo "   ✅ Config file exists"
    echo "   📄 Location: $CONFIG_PATH"
    
    # Validate JSON
    if jq . "$CONFIG_PATH" > /dev/null 2>&1; then
        echo "   ✅ JSON format is valid"
        
        # Check structure
        if jq -e '.mcpServers.["ai-rpg"].command' "$CONFIG_PATH" > /dev/null 2>&1; then
            CONFIGURED_PATH=$(jq -r '.mcpServers.["ai-rpg"].command' "$CONFIG_PATH")
            echo "   ✅ ai-rpg server configured"
            echo "   📁 Configured path: $CONFIGURED_PATH"
            
            if [ "$CONFIGURED_PATH" = "$SERVER_PATH" ]; then
                echo "   ✅ Server path matches binary location"
            else
                echo "   ❌ Server path MISMATCH!"
                echo "   🔧 Expected: $SERVER_PATH"
                echo "   🔧 Configured: $CONFIGURED_PATH"
            fi
            
            # Check API key
            if jq -e '.mcpServers.["ai-rpg"].env.AI_API_KEY' "$CONFIG_PATH" > /dev/null 2>&1; then
                API_KEY=$(jq -r '.mcpServers.["ai-rpg"].env.AI_API_KEY' "$CONFIG_PATH")
                if [[ "$API_KEY" == sk-ant-api03-* ]]; then
                    echo "   ✅ API key format looks correct"
                else
                    echo "   ❌ API key format incorrect (should start with sk-ant-api03-)"
                fi
            else
                echo "   ❌ No API key configured"
            fi
        else
            echo "   ❌ ai-rpg server not found in configuration"
        fi
    else
        echo "   ❌ JSON format is INVALID"
        echo "   🔧 Fix JSON syntax errors first"
    fi
else
    echo "   ❌ Config file NOT found"
    echo "   📁 Expected location: $CONFIG_PATH"
    echo "   🔧 Create the configuration file first"
fi

# Check 3: Test server directly
echo
echo "3️⃣  Testing MCP server directly..."
cd /Users/enrico/workspace/myobsidian/AI-RPG/mcp-server

# Test basic initialization
echo '{"jsonrpc":"2.0","method":"initialize","id":"test","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"diagnostic","version":"1.0.0"}}}' | AI_API_KEY="test-key" ./ai-rpg-mcp-server 2>/dev/null | head -1 > /tmp/mcp_test_result

if grep -q "protocolVersion" /tmp/mcp_test_result; then
    echo "   ✅ Server responds to initialize correctly"
else
    echo "   ❌ Server not responding correctly"
    echo "   🔧 Check server logs for errors"
fi

# Test tools list
echo '{"jsonrpc":"2.0","method":"tools/list","id":"test"}' | AI_API_KEY="test-key" ./ai-rpg-mcp-server 2>/dev/null | head -1 > /tmp/mcp_tools_result

if grep -q "create_session" /tmp/mcp_tools_result; then
    echo "   ✅ Server returns tools list correctly"
    TOOL_COUNT=$(jq -r '.result.tools | length' /tmp/mcp_tools_result 2>/dev/null)
    echo "   📊 Available tools: $TOOL_COUNT"
else
    echo "   ❌ Server not returning tools correctly"
fi

# Check 4: Environment test
echo
echo "4️⃣  Testing with API key..."
if [ -f "$CONFIG_PATH" ] && jq -e '.mcpServers.["ai-rpg"].env.AI_API_KEY' "$CONFIG_PATH" > /dev/null 2>&1; then
    TEST_API_KEY=$(jq -r '.mcpServers.["ai-rpg"].env.AI_API_KEY' "$CONFIG_PATH")
    if [[ "$TEST_API_KEY" != "your_"* && "$TEST_API_KEY" != "test"* ]]; then
        echo '{"jsonrpc":"2.0","method":"tools/list","id":"test"}' | AI_API_KEY="$TEST_API_KEY" ./ai-rpg-mcp-server 2>/dev/null | head -1 > /tmp/mcp_api_test
        
        if grep -q "create_session" /tmp/mcp_api_test; then
            echo "   ✅ Server works with configured API key"
        else
            echo "   ❌ Server fails with configured API key"
            echo "   🔧 Check API key validity"
        fi
    else
        echo "   ⚠️  Placeholder API key detected - replace with real key"
    fi
else
    echo "   ⚠️  No API key to test with"
fi

echo
echo "📋 Summary:"
echo "=========="

# Overall diagnosis
ISSUES=0

if [ ! -f "$SERVER_PATH" ]; then
    echo "❌ Server binary missing"
    ((ISSUES++))
fi

if [ ! -f "$CONFIG_PATH" ]; then
    echo "❌ Configuration file missing"
    ((ISSUES++))
fi

if [ -f "$CONFIG_PATH" ] && ! jq . "$CONFIG_PATH" > /dev/null 2>&1; then
    echo "❌ Configuration JSON invalid"
    ((ISSUES++))
fi

if [ $ISSUES -eq 0 ]; then
    echo "✅ All basic components are present"
    echo "🎯 If Claude Desktop still fails, try:"
    echo "   1. Completely restart Claude Desktop (Cmd+Q, then reopen)"
    echo "   2. Clear Claude Desktop cache"
    echo "   3. Check Claude Desktop logs for specific errors"
    echo "   4. Try with a minimal test configuration first"
else
    echo "🚨 Found $ISSUES critical issues - fix these first"
fi

echo
echo "🔧 Quick fix command:"
echo "   1. Fix server: cd /Users/enrico/workspace/myobsidian/AI-RPG/mcp-server && go build -o ai-rpg-mcp-server main.go"
echo "   2. Restart Claude: killall Claude && sleep 2 && open /Applications/Claude.app"

# Cleanup
rm -f /tmp/mcp_test_result /tmp/mcp_tools_result /tmp/mcp_api_test
