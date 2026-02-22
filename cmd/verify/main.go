package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/xokvictor/whoop-mcp/pkg/whoop"
)

func main() {
	token := os.Getenv("WHOOP_ACCESS_TOKEN")
	if token == "" {
		fmt.Println("❌ Error: WHOOP_ACCESS_TOKEN is not set")
		fmt.Println("\nSet the token:")
		fmt.Println("export WHOOP_ACCESS_TOKEN=\"your_token\"")
		os.Exit(1)
	}

	fmt.Println("🔍 Verifying WHOOP token...")
	fmt.Println()
	fmt.Printf("Token: %s...%s\n\n", token[:10], token[len(token)-10:])

	client := whoop.NewClient()
	ctx := context.Background()

	// Verify profile
	fmt.Println("📱 Fetching user profile...")
	profile, err := client.GetUserProfile(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Println("\nPossible reasons:")
		fmt.Println("- Token is invalid or expired")
		fmt.Println("- Missing scope read:profile")
		fmt.Println("- Network issues")
		os.Exit(1)
	}

	fmt.Println("✅ Token is valid!")
	fmt.Println()
	fmt.Println("👤 User profile:")
	printJSON(profile)

	// Verify body measurements
	fmt.Println("\n📊 Fetching body measurements...")
	measurements, err := client.GetBodyMeasurements(ctx)
	if err != nil {
		fmt.Printf("⚠️  Failed to fetch measurements: %v\n", err)
	} else {
		fmt.Println("✅ Body measurements:")
		printJSON(measurements)
	}

	// Verify cycles
	fmt.Println("\n🔄 Verifying cycles access...")
	cycles, err := client.GetCycles(ctx, whoop.CycleParams{Limit: 1})
	if err != nil {
		fmt.Printf("⚠️  Failed to fetch cycles: %v\n", err)
	} else {
		fmt.Printf("✅ Available cycles: %d\n", len(cycles.Records))
	}

	// Verify sleep
	fmt.Println("\n😴 Verifying sleep data access...")
	sleeps, err := client.GetSleeps(ctx, whoop.SleepParams{Limit: 1})
	if err != nil {
		fmt.Printf("⚠️  Failed to fetch sleep data: %v\n", err)
	} else {
		fmt.Printf("✅ Available sleep records: %d\n", len(sleeps.Records))
	}

	// Verify recovery
	fmt.Println("\n💪 Verifying recovery access...")
	recoveries, err := client.GetRecoveries(ctx, whoop.RecoveryParams{Limit: 1})
	if err != nil {
		fmt.Printf("⚠️  Failed to fetch recovery: %v\n", err)
	} else {
		fmt.Printf("✅ Available recovery records: %d\n", len(recoveries.Records))
	}

	// Verify workouts
	fmt.Println("\n🏋️  Verifying workouts access...")
	workouts, err := client.GetWorkouts(ctx, whoop.WorkoutParams{Limit: 1})
	if err != nil {
		fmt.Printf("⚠️  Failed to fetch workouts: %v\n", err)
	} else {
		fmt.Printf("✅ Available workouts: %d\n", len(workouts.Records))
	}

	fmt.Println("\n✅ All verifications complete!")
	fmt.Println("\n💡 You can now add the token to claude_desktop_config.json")
}

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}
