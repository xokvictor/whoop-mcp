package whoop

import "time"

// User types
type UserBasicProfile struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UserBodyMeasurement struct {
	HeightMeter    float64 `json:"height_meter"`
	WeightKilogram float64 `json:"weight_kilogram"`
	MaxHeartRate   int     `json:"max_heart_rate"`
}

// Cycle types
type Cycle struct {
	ID             int64       `json:"id"`
	UserID         int64       `json:"user_id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Start          time.Time   `json:"start"`
	End            *time.Time  `json:"end,omitempty"`
	TimezoneOffset string      `json:"timezone_offset"`
	ScoreState     string      `json:"score_state"`
	Score          *CycleScore `json:"score,omitempty"`
}

type CycleScore struct {
	Strain           float64 `json:"strain"`
	Kilojoule        float64 `json:"kilojoule"`
	AverageHeartRate int     `json:"average_heart_rate"`
	MaxHeartRate     int     `json:"max_heart_rate"`
}

type PaginatedCycleResponse struct {
	Records   []Cycle `json:"records"`
	NextToken *string `json:"next_token,omitempty"`
}

// Sleep types
type Sleep struct {
	ID             string      `json:"id"`
	CycleID        int64       `json:"cycle_id"`
	V1ID           *int64      `json:"v1_id,omitempty"`
	UserID         int64       `json:"user_id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Start          time.Time   `json:"start"`
	End            time.Time   `json:"end"`
	TimezoneOffset string      `json:"timezone_offset"`
	Nap            bool        `json:"nap"`
	ScoreState     string      `json:"score_state"`
	Score          *SleepScore `json:"score,omitempty"`
}

type SleepScore struct {
	StageSummary               SleepStageSummary `json:"stage_summary"`
	SleepNeeded                SleepNeeded       `json:"sleep_needed"`
	RespiratoryRate            *float64          `json:"respiratory_rate,omitempty"`
	SleepPerformancePercentage *float64          `json:"sleep_performance_percentage,omitempty"`
	SleepConsistencyPercentage *float64          `json:"sleep_consistency_percentage,omitempty"`
	SleepEfficiencyPercentage  *float64          `json:"sleep_efficiency_percentage,omitempty"`
}

type SleepStageSummary struct {
	TotalInBedTimeMilli         int64 `json:"total_in_bed_time_milli"`
	TotalAwakeTimeMilli         int64 `json:"total_awake_time_milli"`
	TotalNoDataTimeMilli        int64 `json:"total_no_data_time_milli"`
	TotalLightSleepTimeMilli    int64 `json:"total_light_sleep_time_milli"`
	TotalSlowWaveSleepTimeMilli int64 `json:"total_slow_wave_sleep_time_milli"`
	TotalRemSleepTimeMilli      int64 `json:"total_rem_sleep_time_milli"`
	SleepCycleCount             int   `json:"sleep_cycle_count"`
	DisturbanceCount            int   `json:"disturbance_count"`
}

type SleepNeeded struct {
	BaselineMilli             int64 `json:"baseline_milli"`
	NeedFromSleepDebtMilli    int64 `json:"need_from_sleep_debt_milli"`
	NeedFromRecentStrainMilli int64 `json:"need_from_recent_strain_milli"`
	NeedFromRecentNapMilli    int64 `json:"need_from_recent_nap_milli"`
}

type PaginatedSleepResponse struct {
	Records   []Sleep `json:"records"`
	NextToken *string `json:"next_token,omitempty"`
}

// Recovery types
type Recovery struct {
	CycleID    int64          `json:"cycle_id"`
	SleepID    string         `json:"sleep_id"`
	UserID     int64          `json:"user_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	ScoreState string         `json:"score_state"`
	Score      *RecoveryScore `json:"score,omitempty"`
}

type RecoveryScore struct {
	UserCalibrating  bool     `json:"user_calibrating"`
	RecoveryScore    float64  `json:"recovery_score"`
	RestingHeartRate float64  `json:"resting_heart_rate"`
	HrvRmssdMilli    float64  `json:"hrv_rmssd_milli"`
	Spo2Percentage   *float64 `json:"spo2_percentage,omitempty"`
	SkinTempCelsius  *float64 `json:"skin_temp_celsius,omitempty"`
}

type RecoveryCollection struct {
	Records   []Recovery `json:"records"`
	NextToken *string    `json:"next_token,omitempty"`
}

// Workout types
type WorkoutV2 struct {
	ID             string        `json:"id"`
	V1ID           *int64        `json:"v1_id,omitempty"`
	UserID         int64         `json:"user_id"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Start          time.Time     `json:"start"`
	End            time.Time     `json:"end"`
	TimezoneOffset string        `json:"timezone_offset"`
	SportName      string        `json:"sport_name"`
	SportID        *int          `json:"sport_id,omitempty"`
	ScoreState     string        `json:"score_state"`
	Score          *WorkoutScore `json:"score,omitempty"`
}

type WorkoutScore struct {
	Strain              float64       `json:"strain"`
	AverageHeartRate    int           `json:"average_heart_rate"`
	MaxHeartRate        int           `json:"max_heart_rate"`
	Kilojoule           float64       `json:"kilojoule"`
	PercentRecorded     float64       `json:"percent_recorded"`
	DistanceMeter       *float64      `json:"distance_meter,omitempty"`
	AltitudeGainMeter   *float64      `json:"altitude_gain_meter,omitempty"`
	AltitudeChangeMeter *float64      `json:"altitude_change_meter,omitempty"`
	ZoneDurations       ZoneDurations `json:"zone_durations"`
}

type ZoneDurations struct {
	ZoneZeroMilli  int64 `json:"zone_zero_milli"`
	ZoneOneMilli   int64 `json:"zone_one_milli"`
	ZoneTwoMilli   int64 `json:"zone_two_milli"`
	ZoneThreeMilli int64 `json:"zone_three_milli"`
	ZoneFourMilli  int64 `json:"zone_four_milli"`
	ZoneFiveMilli  int64 `json:"zone_five_milli"`
}

type WorkoutCollection struct {
	Records   []WorkoutV2 `json:"records"`
	NextToken *string     `json:"next_token,omitempty"`
}

// Activity mapping
type ActivityIdMappingResponse struct {
	V2ActivityID string `json:"v2_activity_id"`
}

// Query parameters
type CycleParams struct {
	Start     string
	End       string
	Limit     int
	NextToken string
}

type SleepParams struct {
	Start     string
	End       string
	Limit     int
	NextToken string
}

type RecoveryParams struct {
	Start     string
	End       string
	Limit     int
	NextToken string
}

type WorkoutParams struct {
	Start     string
	End       string
	Limit     int
	NextToken string
}
