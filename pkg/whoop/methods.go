package whoop

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

const (
	// MaxLimit is the maximum number of records per request
	MaxLimit = 25
)

// UUID validation pattern
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// isValidUUID checks if a string is a valid UUID
func isValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// User methods

// GetUserProfile returns the user's basic profile (name, email).
func (c *Client) GetUserProfile(ctx context.Context) (*UserBasicProfile, error) {
	var profile UserBasicProfile
	if err := c.get(ctx, "/v2/user/profile/basic", &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetBodyMeasurements returns the user's body measurements (height, weight, max HR).
func (c *Client) GetBodyMeasurements(ctx context.Context) (*UserBodyMeasurement, error) {
	var measurements UserBodyMeasurement
	if err := c.get(ctx, "/v2/user/measurement/body", &measurements); err != nil {
		return nil, err
	}
	return &measurements, nil
}

// Cycle methods

// GetCycles returns the user's physiological cycles.
func (c *Client) GetCycles(ctx context.Context, params CycleParams) (*PaginatedCycleResponse, error) {
	path := "/v2/cycle"
	query := buildQuery(params.Start, params.End, params.Limit, params.NextToken)
	if query != "" {
		path += "?" + query
	}

	var response PaginatedCycleResponse
	if err := c.get(ctx, path, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetCycleByID returns a specific cycle by its ID.
func (c *Client) GetCycleByID(ctx context.Context, cycleID int) (*Cycle, error) {
	if cycleID <= 0 {
		return nil, fmt.Errorf("invalid cycle ID: %d", cycleID)
	}

	path := fmt.Sprintf("/v2/cycle/%d", cycleID)
	var cycle Cycle
	if err := c.get(ctx, path, &cycle); err != nil {
		return nil, err
	}
	return &cycle, nil
}

// Sleep methods

// GetSleeps returns the user's sleep records.
func (c *Client) GetSleeps(ctx context.Context, params SleepParams) (*PaginatedSleepResponse, error) {
	path := "/v2/activity/sleep"
	query := buildQuery(params.Start, params.End, params.Limit, params.NextToken)
	if query != "" {
		path += "?" + query
	}

	var response PaginatedSleepResponse
	if err := c.get(ctx, path, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetSleepByID returns a specific sleep record by its UUID.
func (c *Client) GetSleepByID(ctx context.Context, sleepID string) (*Sleep, error) {
	if !isValidUUID(sleepID) {
		return nil, fmt.Errorf("invalid sleep ID: must be a valid UUID")
	}

	path := fmt.Sprintf("/v2/activity/sleep/%s", sleepID)
	var sleep Sleep
	if err := c.get(ctx, path, &sleep); err != nil {
		return nil, err
	}
	return &sleep, nil
}

// GetSleepForCycle returns the sleep record for a specific cycle.
func (c *Client) GetSleepForCycle(ctx context.Context, cycleID int) (*Sleep, error) {
	if cycleID <= 0 {
		return nil, fmt.Errorf("invalid cycle ID: %d", cycleID)
	}

	path := fmt.Sprintf("/v2/cycle/%d/sleep", cycleID)
	var sleep Sleep
	if err := c.get(ctx, path, &sleep); err != nil {
		return nil, err
	}
	return &sleep, nil
}

// Recovery methods

// GetRecoveries returns the user's recovery records.
func (c *Client) GetRecoveries(ctx context.Context, params RecoveryParams) (*RecoveryCollection, error) {
	path := "/v2/recovery"
	query := buildQuery(params.Start, params.End, params.Limit, params.NextToken)
	if query != "" {
		path += "?" + query
	}

	var response RecoveryCollection
	if err := c.get(ctx, path, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetRecoveryForCycle returns the recovery record for a specific cycle.
func (c *Client) GetRecoveryForCycle(ctx context.Context, cycleID int) (*Recovery, error) {
	if cycleID <= 0 {
		return nil, fmt.Errorf("invalid cycle ID: %d", cycleID)
	}

	path := fmt.Sprintf("/v2/cycle/%d/recovery", cycleID)
	var recovery Recovery
	if err := c.get(ctx, path, &recovery); err != nil {
		return nil, err
	}
	return &recovery, nil
}

// Workout methods

// GetWorkouts returns the user's workout records.
func (c *Client) GetWorkouts(ctx context.Context, params WorkoutParams) (*WorkoutCollection, error) {
	path := "/v2/activity/workout"
	query := buildQuery(params.Start, params.End, params.Limit, params.NextToken)
	if query != "" {
		path += "?" + query
	}

	var response WorkoutCollection
	if err := c.get(ctx, path, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetWorkoutByID returns a specific workout by its UUID.
func (c *Client) GetWorkoutByID(ctx context.Context, workoutID string) (*WorkoutV2, error) {
	if !isValidUUID(workoutID) {
		return nil, fmt.Errorf("invalid workout ID: must be a valid UUID")
	}

	path := fmt.Sprintf("/v2/activity/workout/%s", workoutID)
	var workout WorkoutV2
	if err := c.get(ctx, path, &workout); err != nil {
		return nil, err
	}
	return &workout, nil
}

// Activity mapping

// GetActivityMapping returns the V2 UUID for a V1 Activity ID.
func (c *Client) GetActivityMapping(ctx context.Context, activityV1ID int) (*ActivityIdMappingResponse, error) {
	if activityV1ID <= 0 {
		return nil, fmt.Errorf("invalid activity V1 ID: %d", activityV1ID)
	}

	path := fmt.Sprintf("/v1/activity-mapping/%d", activityV1ID)
	var mapping ActivityIdMappingResponse
	if err := c.get(ctx, path, &mapping); err != nil {
		return nil, err
	}
	return &mapping, nil
}

// buildQuery builds a query string from pagination parameters.
func buildQuery(start, end string, limit int, nextToken string) string {
	values := url.Values{}

	if start != "" {
		values.Set("start", start)
	}
	if end != "" {
		values.Set("end", end)
	}
	if limit > 0 {
		if limit > MaxLimit {
			limit = MaxLimit
		}
		values.Set("limit", strconv.Itoa(limit))
	}
	if nextToken != "" {
		values.Set("nextToken", nextToken)
	}

	return values.Encode()
}
