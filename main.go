package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/xokvictor/whoop-mcp/pkg/whoop"
)

const (
	serverName    = "whoop-mcp"
	serverVersion = "0.1.0"
)

func main() {
	// Configure logging to stderr (required for STDIO servers)
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Initialize WHOOP client
	client := whoop.NewClient()

	// Validate token on startup
	if !client.HasToken() {
		log.Println("Warning: WHOOP_ACCESS_TOKEN not set. API calls will fail.")
	}

	// Create MCP server
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(false),
	)

	// Register tools
	registerTools(s, client)

	// Register OAuth configuration resource
	registerResources(s)

	// Start server
	log.Printf("Starting %s v%s", serverName, serverVersion)
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func registerResources(s *server.MCPServer) {
	oauthResource := mcp.NewResource(
		"oauth://config",
		"WHOOP OAuth Configuration",
		mcp.WithResourceDescription("OAuth 2.0 configuration for WHOOP API authentication. Use this to understand required scopes and endpoints."),
		mcp.WithMIMEType("application/json"),
	)

	s.AddResource(oauthResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		config := map[string]interface{}{
			"authorization_url": whoop.AuthURL,
			"token_url":         whoop.TokenURL,
			"scopes": map[string]string{
				"read:recovery":         "Read Recovery data (HRV, resting HR, recovery score)",
				"read:cycles":           "Read physiological cycles (strain, day boundaries)",
				"read:workout":          "Read workout data (activities, heart rate zones)",
				"read:sleep":            "Read sleep data (stages, efficiency, duration)",
				"read:profile":          "Read user profile (name, email)",
				"read:body_measurement": "Read body measurements (height, weight, max HR)",
			},
		}
		data, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      request.Params.URI,
					MIMEType: "application/json",
				},
				Text: string(data),
			},
		}, nil
	})
}

func registerTools(s *server.MCPServer, client *whoop.Client) {
	// User profile tools
	s.AddTool(
		mcp.NewTool("get_user_profile",
			mcp.WithDescription("Get the authenticated user's basic profile information. Returns user ID, email, first name, and last name. Requires scope: read:profile"),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			profile, err := client.GetUserProfile(ctx)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(profile)
		},
	)

	s.AddTool(
		mcp.NewTool("get_body_measurements",
			mcp.WithDescription("Get the user's body measurements including height (meters), weight (kilograms), and maximum heart rate. Requires scope: read:body_measurement"),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			measurements, err := client.GetBodyMeasurements(ctx)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(measurements)
		},
	)

	// Cycle tools
	s.AddTool(
		mcp.NewTool("get_cycles",
			mcp.WithDescription("Get the user's physiological cycles. Each cycle represents a day's worth of strain data with start/end times. Returns paginated results with cycle ID, timestamps, strain score, and heart rate data. Requires scope: read:cycles"),
			mcp.WithString("start",
				mcp.Description("Start date/time in ISO 8601 format (e.g., 2024-01-01T00:00:00Z). Filters cycles starting on or after this time."),
			),
			mcp.WithString("end",
				mcp.Description("End date/time in ISO 8601 format (e.g., 2024-12-31T23:59:59Z). Filters cycles ending on or before this time."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of records to return (1-25, default: 10). Use with next_token for pagination."),
			),
			mcp.WithString("next_token",
				mcp.Description("Pagination token from previous response. Use to fetch the next page of results."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			params := whoop.CycleParams{
				Start:     getStringArg(request.Params.Arguments, "start"),
				End:       getStringArg(request.Params.Arguments, "end"),
				Limit:     getIntArg(request.Params.Arguments, "limit", 10),
				NextToken: getStringArg(request.Params.Arguments, "next_token"),
			}
			cycles, err := client.GetCycles(ctx, params)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(cycles)
		},
	)

	s.AddTool(
		mcp.NewTool("get_cycle_by_id",
			mcp.WithDescription("Get a specific physiological cycle by its numeric ID. Returns detailed cycle data including strain, heart rate stats, and timestamps. Requires scope: read:cycles"),
			mcp.WithNumber("cycle_id",
				mcp.Required(),
				mcp.Description("The numeric cycle ID (e.g., 1325792966). Can be obtained from get_cycles response."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			cycleID := getIntArg(request.Params.Arguments, "cycle_id", 0)
			if cycleID == 0 {
				return mcp.NewToolResultError("cycle_id is required and must be a positive integer"), nil
			}
			cycle, err := client.GetCycleByID(ctx, cycleID)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(cycle)
		},
	)

	// Sleep tools
	s.AddTool(
		mcp.NewTool("get_sleeps",
			mcp.WithDescription("Get the user's sleep records. Each record includes sleep stages (light, deep, REM), efficiency percentage, disturbances, and respiratory rate. Requires scope: read:sleep"),
			mcp.WithString("start",
				mcp.Description("Start date/time in ISO 8601 format. Filters sleep records starting on or after this time."),
			),
			mcp.WithString("end",
				mcp.Description("End date/time in ISO 8601 format. Filters sleep records ending on or before this time."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of records to return (1-25, default: 10)."),
			),
			mcp.WithString("next_token",
				mcp.Description("Pagination token from previous response."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			params := whoop.SleepParams{
				Start:     getStringArg(request.Params.Arguments, "start"),
				End:       getStringArg(request.Params.Arguments, "end"),
				Limit:     getIntArg(request.Params.Arguments, "limit", 10),
				NextToken: getStringArg(request.Params.Arguments, "next_token"),
			}
			sleeps, err := client.GetSleeps(ctx, params)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(sleeps)
		},
	)

	s.AddTool(
		mcp.NewTool("get_sleep_by_id",
			mcp.WithDescription("Get a specific sleep record by its UUID. Returns detailed sleep data including all stages, efficiency, and performance metrics. Requires scope: read:sleep"),
			mcp.WithString("sleep_id",
				mcp.Required(),
				mcp.Description("The sleep record UUID (e.g., 89329a72-94e7-486c-a072-342501371575). Can be obtained from get_sleeps response."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sleepID := getStringArg(request.Params.Arguments, "sleep_id")
			if sleepID == "" {
				return mcp.NewToolResultError("sleep_id is required and must be a valid UUID"), nil
			}
			sleep, err := client.GetSleepByID(ctx, sleepID)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(sleep)
		},
	)

	s.AddTool(
		mcp.NewTool("get_sleep_for_cycle",
			mcp.WithDescription("Get the sleep record associated with a specific physiological cycle. Useful for correlating sleep with daily strain. Requires scopes: read:sleep, read:cycles"),
			mcp.WithNumber("cycle_id",
				mcp.Required(),
				mcp.Description("The numeric cycle ID to get sleep data for."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			cycleID := getIntArg(request.Params.Arguments, "cycle_id", 0)
			if cycleID == 0 {
				return mcp.NewToolResultError("cycle_id is required and must be a positive integer"), nil
			}
			sleep, err := client.GetSleepForCycle(ctx, cycleID)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(sleep)
		},
	)

	// Recovery tools
	s.AddTool(
		mcp.NewTool("get_recoveries",
			mcp.WithDescription("Get the user's recovery records. Each record includes recovery score (0-100%), HRV (heart rate variability in ms), resting heart rate, SpO2 percentage, and skin temperature. Requires scope: read:recovery"),
			mcp.WithString("start",
				mcp.Description("Start date/time in ISO 8601 format. Filters recovery records starting on or after this time."),
			),
			mcp.WithString("end",
				mcp.Description("End date/time in ISO 8601 format. Filters recovery records ending on or before this time."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of records to return (1-25, default: 10)."),
			),
			mcp.WithString("next_token",
				mcp.Description("Pagination token from previous response."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			params := whoop.RecoveryParams{
				Start:     getStringArg(request.Params.Arguments, "start"),
				End:       getStringArg(request.Params.Arguments, "end"),
				Limit:     getIntArg(request.Params.Arguments, "limit", 10),
				NextToken: getStringArg(request.Params.Arguments, "next_token"),
			}
			recoveries, err := client.GetRecoveries(ctx, params)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(recoveries)
		},
	)

	s.AddTool(
		mcp.NewTool("get_recovery_for_cycle",
			mcp.WithDescription("Get the recovery record associated with a specific physiological cycle. Requires scopes: read:recovery, read:cycles"),
			mcp.WithNumber("cycle_id",
				mcp.Required(),
				mcp.Description("The numeric cycle ID to get recovery data for."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			cycleID := getIntArg(request.Params.Arguments, "cycle_id", 0)
			if cycleID == 0 {
				return mcp.NewToolResultError("cycle_id is required and must be a positive integer"), nil
			}
			recovery, err := client.GetRecoveryForCycle(ctx, cycleID)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(recovery)
		},
	)

	// Workout tools
	s.AddTool(
		mcp.NewTool("get_workouts",
			mcp.WithDescription("Get the user's workout records. Each record includes sport type, strain, heart rate data (average/max), calories burned, duration, and heart rate zone distribution. Requires scope: read:workout"),
			mcp.WithString("start",
				mcp.Description("Start date/time in ISO 8601 format. Filters workouts starting on or after this time."),
			),
			mcp.WithString("end",
				mcp.Description("End date/time in ISO 8601 format. Filters workouts ending on or before this time."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of records to return (1-25, default: 10)."),
			),
			mcp.WithString("next_token",
				mcp.Description("Pagination token from previous response."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			params := whoop.WorkoutParams{
				Start:     getStringArg(request.Params.Arguments, "start"),
				End:       getStringArg(request.Params.Arguments, "end"),
				Limit:     getIntArg(request.Params.Arguments, "limit", 10),
				NextToken: getStringArg(request.Params.Arguments, "next_token"),
			}
			workouts, err := client.GetWorkouts(ctx, params)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(workouts)
		},
	)

	s.AddTool(
		mcp.NewTool("get_workout_by_id",
			mcp.WithDescription("Get a specific workout by its UUID. Returns detailed workout data including all heart rate zones and metrics. Requires scope: read:workout"),
			mcp.WithString("workout_id",
				mcp.Required(),
				mcp.Description("The workout UUID (e.g., 89329a72-94e7-486c-a072-342501371575). Can be obtained from get_workouts response."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			workoutID := getStringArg(request.Params.Arguments, "workout_id")
			if workoutID == "" {
				return mcp.NewToolResultError("workout_id is required and must be a valid UUID"), nil
			}
			workout, err := client.GetWorkoutByID(ctx, workoutID)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(workout)
		},
	)

	// Utility tools
	s.AddTool(
		mcp.NewTool("get_activity_mapping",
			mcp.WithDescription("Convert a legacy V1 Activity ID to the current V2 UUID format. Useful when migrating from older WHOOP API versions."),
			mcp.WithNumber("activity_v1_id",
				mcp.Required(),
				mcp.Description("The legacy V1 Activity ID (numeric). Returns the corresponding V2 UUID."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			activityID := getIntArg(request.Params.Arguments, "activity_v1_id", 0)
			if activityID == 0 {
				return mcp.NewToolResultError("activity_v1_id is required and must be a positive integer"), nil
			}
			mapping, err := client.GetActivityMapping(ctx, activityID)
			if err != nil {
				return mcp.NewToolResultError(formatError(err)), nil
			}
			return resultFromJSON(mapping)
		},
	)
}

// Helper functions

func getStringArg(args map[string]interface{}, key string) string {
	if args == nil {
		return ""
	}
	if val, ok := args[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if args == nil {
		return defaultVal
	}
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case int64:
			return int(v)
		}
	}
	return defaultVal
}

func formatError(err error) string {
	if apiErr, ok := err.(*whoop.APIError); ok {
		switch {
		case apiErr.IsUnauthorized():
			return "Authentication failed: Invalid or expired access token. Please refresh your WHOOP_ACCESS_TOKEN."
		case apiErr.IsNotFound():
			return "Resource not found: The requested ID does not exist or you don't have access to it."
		case apiErr.IsRateLimited():
			return "Rate limited: Too many requests. Please wait a moment and try again."
		default:
			return fmt.Sprintf("WHOOP API error (status %d): %s", apiErr.StatusCode, apiErr.Message)
		}
	}
	return fmt.Sprintf("Error: %v", err)
}

func resultFromJSON(data interface{}) (*mcp.CallToolResult, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Serialization error: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}
