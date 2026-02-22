# OAuth Setup for WHOOP API

This guide walks you through the complete OAuth setup process for the WHOOP MCP Server.

## Step 1: Create Application in WHOOP Developer Dashboard

1. Go to https://developer-dashboard.whoop.com
2. Sign in with your WHOOP account
3. Click **"Create New Application"**
4. Fill in the form:
   - **Application Name**: `WHOOP MCP` (or any name you prefer)
   - **Description**: `MCP Server for Claude Desktop`
   - **Redirect URI**: `http://localhost:8080/callback`
5. After creation, copy:
   - **Client ID**
   - **Client Secret**

## Step 2: Get Access Token

### Option A: Using the OAuth Helper (Recommended)

```bash
# Set environment variables
export WHOOP_CLIENT_ID="your_client_id"
export WHOOP_CLIENT_SECRET="your_client_secret"

# Run the OAuth helper
make auth
```

Then:
1. Open your browser at http://localhost:8080
2. Click "Authorize with WHOOP"
3. Sign in to your WHOOP account
4. Grant permission to the application
5. Copy the displayed Access Token

### Option B: Manual Process with curl

```bash
# 1. Open this URL in your browser (replace YOUR_CLIENT_ID):
https://api.prod.whoop.com/oauth/oauth2/auth?client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost:8080/callback&response_type=code&scope=read:profile%20read:body_measurement%20read:cycles%20read:sleep%20read:recovery%20read:workout

# 2. After authorization, you'll be redirected to:
http://localhost:8080/callback?code=AUTHORIZATION_CODE&state=...

# 3. Exchange the code for an access token:
curl -X POST https://api.prod.whoop.com/oauth/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=AUTHORIZATION_CODE" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "redirect_uri=http://localhost:8080/callback"
```

## Step 3: Configure Claude Desktop

Add the token to your Claude Desktop configuration file.

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`

**Linux:** `~/.config/claude/claude_desktop_config.json`

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

## Step 4: Restart Claude Desktop

Completely quit and reopen Claude Desktop for the changes to take effect.

## Verification

In Claude Desktop, try:

```
Show my WHOOP profile
```

If everything is working, you'll see your name and email.

## Scopes (Permissions)

The OAuth token grants access to:

| Scope | Description |
|-------|-------------|
| `read:profile` | User profile (name, email) |
| `read:body_measurement` | Body measurements (height, weight, max HR) |
| `read:cycles` | Physiological cycles and strain |
| `read:sleep` | Sleep data |
| `read:recovery` | Recovery data (HRV, RHR) |
| `read:workout` | Workout data |

## Token Expiration

- Access Token typically lasts **30 days**
- If the token expires, repeat the authorization process
- Future versions may support automatic refresh via Refresh Token

## Security

⚠️ **Important:**
- Do not publish Access Token in code or repositories
- Do not share the token with others
- The token provides full access to your WHOOP data
- Keep Client Secret secure

## Revoking Access

To revoke access:

1. Go to Developer Dashboard
2. Delete the application or generate a new Client Secret

Or via API:

```bash
curl -X DELETE https://api.prod.whoop.com/developer/v2/user/access \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Troubleshooting

### "Invalid client credentials"
- Verify CLIENT_ID and CLIENT_SECRET
- Make sure they're copied completely without extra spaces

### "Redirect URI mismatch"
- In Developer Dashboard, set exactly: `http://localhost:8080/callback`
- No trailing slash at the end!

### "Invalid authorization code"
- Authorization code can only be used once
- If error occurs, get a new code by restarting the auth flow

### "Access token expired"
- Token has expired, need to get a new one
- Repeat the authorization process

### Port 8080 already in use
- Another application is using port 8080
- Stop the other application or wait for it to finish
- Check with: `lsof -i :8080`
- Kill with: `kill -9 <PID>`
