#!/bin/bash

set -e

echo "🏃 WHOOP MCP Server - Quick Setup"
echo "=================================="
echo ""

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Install Go 1.22 or later."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: $GO_VERSION"
echo ""

# Install dependencies
echo "📦 Installing dependencies..."
go mod download
go mod tidy
echo ""

# Build
echo "🔨 Building project..."
go build -o whoop-mcp .
echo "✅ Built: ./whoop-mcp"
echo ""

# Next steps
echo "📝 Next steps:"
echo ""
echo "1️⃣  Create an application on WHOOP Developer Dashboard:"
echo "    https://developer-dashboard.whoop.com"
echo ""
echo "2️⃣  Get Access Token:"
echo "    export WHOOP_CLIENT_ID=\"your_client_id\""
echo "    export WHOOP_CLIENT_SECRET=\"your_client_secret\""
echo "    make auth"
echo ""
echo "3️⃣  Add configuration to Claude Desktop:"
echo "    ~/.config/claude/claude_desktop_config.json"
echo ""
echo "    {"
echo "      \"mcpServers\": {"
echo "        \"whoop\": {"
echo "          \"command\": \"$(pwd)/whoop-mcp\","
echo "          \"env\": {"
echo "            \"WHOOP_ACCESS_TOKEN\": \"your_token_here\""
echo "          }"
echo "        }"
echo "      }"
echo "    }"
echo ""
echo "4️⃣  Restart Claude Desktop"
echo ""
echo "📚 Documentation:"
echo "   - OAuth setup: OAUTH_SETUP.md"
echo "   - Examples: EXAMPLES.md"
echo ""
echo "✅ Ready to use!"
