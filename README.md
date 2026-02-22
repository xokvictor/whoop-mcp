# WHOOP MCP Server

[![CI](https://github.com/xokvictor/whoop-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/xokvictor/whoop-mcp/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/xokvictor/whoop-mcp)](https://goreportcard.com/report/github.com/xokvictor/whoop-mcp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

MCP (Model Context Protocol) server for WHOOP API integration. Access your sleep, recovery, workouts, and physiological cycle data directly from Claude.

## Features

- 🏃 **Workouts** - View your workout history with strain, heart rate, and duration
- 😴 **Sleep** - Access detailed sleep data including stages and performance
- 💪 **Recovery** - Get recovery scores, HRV, resting heart rate, and SpO2
- 📊 **Cycles** - View physiological cycles with strain data
- 👤 **Profile** - Access user profile and body measurements

## Requirements

- Go 1.22 or later
- WHOOP membership with API access
- Claude Desktop or Claude Code

## Installation

### Download Binary (Recommended)

Download the latest binary for your platform from [Releases](https://github.com/xokvictor/whoop-mcp/releases):

| Platform | Binary |
|----------|--------|
| macOS (Apple Silicon) | `whoop-mcp-darwin-arm64` |
| macOS (Intel) | `whoop-mcp-darwin-amd64` |
| Linux | `whoop-mcp-linux-amd64` |
| Windows | `whoop-mcp-windows-amd64.exe` |

```bash
# Example for macOS Apple Silicon
curl -L https://github.com/xokvictor/whoop-mcp/releases/latest/download/whoop-mcp-darwin-arm64 -o whoop-mcp
chmod +x whoop-mcp
```

### Go Install

```bash
go install github.com/xokvictor/whoop-mcp@latest
```

### From Source

```bash
git clone https://github.com/xokvictor/whoop-mcp.git
cd whoop-mcp
make build
```

## OAuth Setup

WHOOP uses OAuth 2.0 for authentication. You'll need to create an application and obtain an access token.

### Step 1: Create WHOOP Application

1. Go to [WHOOP Developer Dashboard](https://developer-dashboard.whoop.com)
2. Sign in with your WHOOP account
3. Click "Create New Application"
4. Fill in the details:
   - **Application Name**: `WHOOP MCP` (or any name)
   - **Description**: `MCP Server for Claude`
   - **Redirect URI**: `http://localhost:8080/callback`
5. Copy your **Client ID** and **Client Secret**

### Step 2: Get Access Token

```bash
# Set your credentials
export WHOOP_CLIENT_ID="your_client_id"
export WHOOP_CLIENT_SECRET="your_client_secret"

# Run the OAuth helper
make auth
```

This will:
1. Start a local server at http://localhost:8080
2. Open the authorization page in your browser
3. After you authorize, display your access token

### Step 3: Verify Token (Optional)

```bash
export WHOOP_ACCESS_TOKEN="your_token"
make verify
```

This verifies the token works and shows your profile.

📖 **Detailed instructions:** [OAUTH_SETUP.md](./OAUTH_SETUP.md)

## Configuration

### Claude Desktop

Add to your `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `~/.config/claude/claude_desktop_config.json` (Linux):

```json
{
  "mcpServers": {
    "whoop": {
      "command": "/path/to/whoop-mcp",
      "env": {
        "WHOOP_ACCESS_TOKEN": "your_access_token_here"
      }
    }
  }
}
```

### Claude Code

```bash
claude mcp add whoop /path/to/whoop-mcp -e WHOOP_ACCESS_TOKEN="your_token"
```

## Available Tools

### User Profile
| Tool | Description |
|------|-------------|
| `get_user_profile` | Get basic profile (name, email) |
| `get_body_measurements` | Get body measurements (height, weight, max HR) |

### Cycles
| Tool | Description |
|------|-------------|
| `get_cycles` | List physiological cycles with pagination |
| `get_cycle_by_id` | Get a specific cycle by ID |

### Sleep
| Tool | Description |
|------|-------------|
| `get_sleeps` | List sleep records with pagination |
| `get_sleep_by_id` | Get a specific sleep record by UUID |
| `get_sleep_for_cycle` | Get sleep for a specific cycle |

### Recovery
| Tool | Description |
|------|-------------|
| `get_recoveries` | List recovery records with pagination |
| `get_recovery_for_cycle` | Get recovery for a specific cycle |

### Workouts
| Tool | Description |
|------|-------------|
| `get_workouts` | List workouts with pagination |
| `get_workout_by_id` | Get a specific workout by UUID |

### Utilities
| Tool | Description |
|------|-------------|
| `get_activity_mapping` | Convert V1 Activity ID to V2 UUID |

## Usage Examples

Once configured, you can ask Claude:

```
Show my WHOOP profile
What were my last 5 workouts?
Show my sleep data for the past week
What's my recovery score today?
How has my HRV trended this month?
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Run all CI checks
make ci
```

### Project Structure

```
whoop-mcp/
├── cmd/
│   ├── auth/       # OAuth helper tool
│   └── verify/     # Token verification tool
├── pkg/
│   └── whoop/      # WHOOP API client
│       ├── client.go
│       ├── methods.go
│       └── types.go
├── main.go         # MCP server entry point
├── Makefile
└── README.md
```

### Building

```bash
# Build main binary
make build

# Build all binaries
go build -o whoop-mcp .
go build -o auth ./cmd/auth
go build -o verify ./cmd/verify
```

## API Documentation

- [WHOOP Developer API](https://developer.whoop.com/api)
- [Model Context Protocol](https://modelcontextprotocol.io)

## Token Expiration

- Access tokens typically expire after **30 days**
- When expired, run `make auth` again to get a new token
- Future versions may support automatic token refresh

## Security

⚠️ **Important:**
- Never commit your access token to version control
- Keep your Client Secret secure
- The token provides read access to all your WHOOP data

## Troubleshooting

### "Invalid client credentials"
- Verify your CLIENT_ID and CLIENT_SECRET are correct
- Make sure they're copied completely without extra spaces

### "Redirect URI mismatch"
- In Developer Dashboard, set exactly: `http://localhost:8080/callback`
- No trailing slash!

### "Access token expired"
- Token has expired, run `make auth` to get a new one

### "API error (status 429)"
- Rate limited - wait a moment and try again

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [WHOOP](https://www.whoop.com/) for the API
- [mcp-go](https://github.com/mark3labs/mcp-go) for the MCP Go SDK
