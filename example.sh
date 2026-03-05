#!/bin/bash

# Example script showing how to test the Massdriver MCP server

set -e

echo "🚀 Building Massdriver MCP Server..."
go build -o mcp-server

echo "✅ Build successful!"

echo ""
echo "📋 To use this MCP server, you need to set these environment variables:"
echo ""
echo "export MASSDRIVER_API_KEY=\"your-api-key-here\""
echo "export MASSDRIVER_ORGANIZATION_ID=\"your-organization-id\""
echo "export MASSDRIVER_URL=\"https://api.massdriver.cloud\"  # Optional"
echo ""

echo "🔧 Available MCP Tools:"
echo "  - list_projects: Lists all projects in your Massdriver organization"
echo ""

echo "📖 Usage Examples:"
echo ""
echo "1. With Claude Desktop - Add to claude_desktop_config.json:"
echo "{"
echo "  \"mcpServers\": {"
echo "    \"massdriver\": {"
echo "      \"command\": \"$(pwd)/mcp-server\","
echo "      \"env\": {"
echo "        \"MASSDRIVER_API_KEY\": \"your-api-key\","
echo "        \"MASSDRIVER_ORGANIZATION_ID\": \"your-org-id\""
echo "      }"
echo "    }"
echo "  }"
echo "}"
echo ""

echo "2. Direct testing (with proper env vars set):"
echo "   ./mcp-server"
echo ""

echo "3. Testing the server protocol:"
echo "   echo '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2024-11-05\",\"capabilities\":{\"tools\":{}},\"clientInfo\":{\"name\":\"test\",\"version\":\"1.0.0\"}}}' | ./mcp-server"
echo ""

echo "🎯 Next Steps:"
echo "1. Set your environment variables with your Massdriver credentials"
echo "2. Test the server with a real MCP client"
echo "3. Add additional tools as needed"
echo ""
echo "For more information, see MCP_README.md"