package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

const (
	authURL     = "https://api.prod.whoop.com/oauth/oauth2/auth"
	tokenURL    = "https://api.prod.whoop.com/oauth/oauth2/token"
	redirectURL = "http://localhost:8080/callback"
)

var (
	clientID     string
	clientSecret string
	oauthConfig  *oauth2.Config
	oauthState   string
)

func main() {
	// Read Client ID and Client Secret from environment variables
	clientID = os.Getenv("WHOOP_CLIENT_ID")
	clientSecret = os.Getenv("WHOOP_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		fmt.Println("❌ Error: WHOOP_CLIENT_ID and WHOOP_CLIENT_SECRET are required")
		fmt.Println("\nHow to obtain:")
		fmt.Println("1. Register at https://developer-dashboard.whoop.com")
		fmt.Println("2. Create a new application")
		fmt.Println("3. Set Redirect URI: http://localhost:8080/callback")
		fmt.Println("4. Copy Client ID and Client Secret")
		fmt.Println("\nRun:")
		fmt.Println("WHOOP_CLIENT_ID=your_id WHOOP_CLIENT_SECRET=your_secret go run cmd/auth/main.go")
		os.Exit(1)
	}

	// Generate random state for CSRF protection
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate state: %v", err)
	}
	oauthState = base64.URLEncoding.EncodeToString(b)

	// Configure OAuth2
	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"read:profile",
			"read:body_measurement",
			"read:cycles",
			"read:sleep",
			"read:recovery",
			"read:workout",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	// HTTP server for callback handling
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/callback", handleCallback)

	port := "8080"
	fmt.Println("\n🚀 OAuth server running at http://localhost:" + port)
	fmt.Println("\n📋 Open your browser and navigate to http://localhost:" + port)
	fmt.Println("   Or click the authorization link")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server startup error:", err)
	}
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	authURL := oauthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOffline)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>WHOOP OAuth Authorization</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 600px;
            margin: 50px auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            margin-bottom: 20px;
        }
        .btn {
            display: inline-block;
            padding: 12px 24px;
            background: #d36b2f;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            font-weight: bold;
            margin-top: 20px;
        }
        .btn:hover {
            background: #b8571f;
        }
        .scopes {
            background: #f9f9f9;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
        .scopes ul {
            margin: 10px 0;
            padding-left: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏃 WHOOP OAuth Authorization</h1>
        <p>Click the button below to authorize with WHOOP</p>

        <div class="scopes">
            <strong>Access will be granted to:</strong>
            <ul>
                <li>User profile</li>
                <li>Body measurements (height, weight, max HR)</li>
                <li>Physiological cycles</li>
                <li>Sleep data</li>
                <li>Recovery data</li>
                <li>Workout data</li>
            </ul>
        </div>

        <a href="%s" class="btn">🔐 Authorize with WHOOP</a>

        <p style="color: #666; margin-top: 20px; font-size: 14px;">
            After authorization you will be redirected back and receive an access token
        </p>
    </div>
</body>
</html>
`, authURL)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	// Validate state
	state := r.FormValue("state")
	if state != oauthState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.FormValue("code")
	if code == "" {
		errorMsg := r.FormValue("error")
		errorDesc := r.FormValue("error_description")
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Authorization Error</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .error { background: #fee; border: 1px solid #fcc; padding: 20px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="error">
        <h1>❌ Authorization Error</h1>
        <p><strong>Error:</strong> %s</p>
        <p><strong>Description:</strong> %s</p>
    </div>
</body>
</html>
`, errorMsg, errorDesc)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
		return
	}

	// Exchange authorization code for access token
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Token Exchange Error</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .error { background: #fee; border: 1px solid #fcc; padding: 20px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="error">
        <h1>❌ Token Exchange Error</h1>
        <p>%s</p>
    </div>
</body>
</html>
`, err.Error())
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
		return
	}

	// Display access token
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Success!</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .success {
            background: #d4edda;
            border: 1px solid #c3e6cb;
            padding: 20px;
            border-radius: 5px;
            margin: 20px 0;
        }
        .token {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            font-family: monospace;
            word-break: break-all;
            border: 1px solid #ddd;
        }
        .copy-btn {
            background: #d36b2f;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            margin-top: 10px;
        }
        .copy-btn:hover {
            background: #b8571f;
        }
        .instructions {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            padding: 15px;
            border-radius: 5px;
            margin-top: 20px;
        }
        code {
            background: #f4f4f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: monospace;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">
            <h1>✅ Authorization Successful!</h1>
        </div>

        <h2>🔑 Access Token:</h2>
        <div class="token" id="token">%s</div>
        <button class="copy-btn" onclick="copyToken()">📋 Copy Token</button>

        <div class="instructions">
            <h3>📝 Next Steps:</h3>
            <ol>
                <li>Copy the token above</li>
                <li>Add it to your <code>~/.config/claude/claude_desktop_config.json</code>:</li>
            </ol>
            <pre style="background: #f4f4f4; padding: 15px; border-radius: 5px; overflow-x: auto;">
{
  "mcpServers": {
    "whoop": {
      "command": "/path/to/whoop-mcp",
      "env": {
        "WHOOP_ACCESS_TOKEN": "%s"
      }
    }
  }
}</pre>
            <p><strong>3.</strong> Restart Claude Desktop</p>
        </div>

        <p style="color: #666; margin-top: 20px;">
            <strong>Expires:</strong> %s<br>
            <strong>Token type:</strong> %s
        </p>
    </div>

    <script>
        function copyToken() {
            const token = document.getElementById('token').innerText;
            navigator.clipboard.writeText(token).then(() => {
                alert('✅ Token copied to clipboard!');
            });
        }
    </script>
</body>
</html>
`, token.AccessToken, token.AccessToken, token.Expiry.Format(time.RFC3339), token.TokenType)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)

	// Console output
	fmt.Println("\n✅ Authorization successful!")
	fmt.Println("\n🔑 Access Token:")
	fmt.Println(token.AccessToken)
	fmt.Println("\n📅 Expires:", token.Expiry.Format(time.RFC3339))

	if token.RefreshToken != "" {
		fmt.Println("\n🔄 Refresh Token:")
		fmt.Println(token.RefreshToken)
	}

	fmt.Println("\n💾 Save the token to your claude_desktop_config.json")
	fmt.Println("\nServer continues running. Press Ctrl+C to exit.")
}
