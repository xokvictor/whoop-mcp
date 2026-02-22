# WHOOP MCP Usage Examples

## Setup

Before using, make sure you have:
1. Obtained a WHOOP Access Token
2. Added the configuration to Claude Desktop
3. Restarted Claude Desktop

## Example Queries

### User Profile

```
Show my WHOOP profile
```

Result includes:
- User ID
- Email
- First and last name

```
What are my body measurements?
```

Result includes:
- Height (meters)
- Weight (kilograms)
- Maximum heart rate

### Cycles

```
Show my last 5 cycles
```

```
What's my strain score for today?
```

```
Show detailed information about cycle 12345
```

### Sleep

```
Show my sleep data for the last week
```

```
How much did I sleep last night?
```

```
What's my sleep efficiency for the past month?
```

```
Show sleep stage breakdown for last night
```

### Recovery

```
What's my recovery score today?
```

```
Show my HRV for the last 7 days
```

```
What's my resting heart rate?
```

```
Show my SPO2 and skin temperature (for WHOOP 4.0)
```

### Workouts

```
What are my last 10 workouts?
```

```
Show detailed stats for workout [workout_id]
```

```
How many calories did I burn in my last workout?
```

```
Show heart rate zone distribution for my last workout
```

```
What was my average heart rate during yesterday's run?
```

### Analytics

```
Compare my strain this week vs last week
```

```
What's the relationship between my sleep and recovery score?
```

```
Show my HRV trend for the past month
```

```
Which days do I have the best recovery score?
```

## Data Filtering

### By Date

All collections support date filtering in ISO 8601 format:

```
Show sleep from February 1 to February 10, 2024
start: 2024-02-01T00:00:00Z
end: 2024-02-10T23:59:59Z
```

### Pagination

Use `limit` and `next_token` to retrieve large datasets:

```
Show the first 5 workouts, then the next 5
```

## Complex Queries

### Combining Data

```
Find correlation between my sleep, recovery, and strain for the past month
```

```
Compare my workouts on high recovery days vs low recovery days
```

### Statistics

```
What's my average sleep duration for the past week?
```

```
How many workouts did I do this month?
```

```
What's my average strain score?
```

## Technical Queries

### ID Mapping

If you have an old V1 Activity ID:

```
Convert V1 activity ID 12345678 to V2 UUID
```

### Cycle-Specific Data

```
Show sleep for cycle 93845
```

```
Show recovery for cycle 93845
```

## Date Format

WHOOP API uses ISO 8601 format:

- `2024-02-17T00:00:00Z` - UTC time
- `2024-02-17T12:00:00-05:00` - With timezone offset

Example with timezone:

```
Show data with my local time (EST)
start: 2024-02-01T00:00:00-05:00
end: 2024-02-10T23:59:59-05:00
```

## Error Handling

If you get an error:

| Error | Solution |
|-------|----------|
| **401 Invalid authorization** | Check WHOOP_ACCESS_TOKEN |
| **404 No resource found** | Verify the resource ID |
| **429 Rate limiting** | Wait before the next request |
| **500 Server error** | Try again later |

## Scope Requirements

Make sure your token has the required scopes:

| Scope | Used For |
|-------|----------|
| `read:profile` | User profile |
| `read:body_measurement` | Body measurements |
| `read:cycles` | Physiological cycles |
| `read:sleep` | Sleep data |
| `read:recovery` | Recovery data |
| `read:workout` | Workout data |
