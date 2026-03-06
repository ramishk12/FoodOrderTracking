# Timezone Handling in Food Order Tracking

This document describes how timezones are handled throughout the application, particularly for the `scheduled_date` field.

## Overview

All timestamps in the application are stored and processed in **UTC (Coordinated Universal Time)** to ensure consistency across different timezones and prevent issues when the application runs in different server locations.

## Database

### Column Definition
The `scheduled_date` column in the `orders` table is defined as:
```sql
scheduled_date TIMESTAMP WITH TIME ZONE
```

Using `TIMESTAMP WITH TIME ZONE` (also known as `TIMESTAMPTZ`) ensures that:
- Timezone information is explicitly stored with the timestamp
- PostgreSQL automatically converts to UTC for storage
- PostgreSQL handles timezone-aware operations correctly

### Default Timezone
All system timestamps use UTC:
```sql
created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
```

## Backend (Go)

### DateTime Handling
When receiving a `scheduled_date` from the frontend:
1. The value comes as an ISO 8601 string (e.g., `"2026-03-05T10:30Z"`)
2. The Go backend parses this as a `time.Time` object (automatically in UTC when string contains Z)
3. Before saving, the backend calls `.UTC()` on the timestamp to ensure it's in UTC
4. The value is saved to PostgreSQL, which stores it with timezone information

### Example Flow
```go
// Frontend sends: "2026-03-05T10:30Z"
var scheduledDate *time.Time = "2026-03-05T10:30:00Z"

// Ensure UTC (idempotent operation)
*scheduledDate = scheduledDate.UTC()

// Save to database (automatically stored as UTC by PostgreSQL)
db.Exec("INSERT INTO orders (scheduled_date) VALUES ($1)", scheduledDate)
```

## Frontend (React)

### Input Format
The `<input type="datetime-local">` HTML element provides a convenient way for users to select dates and times:
- Returns format: `"2026-03-05T10:30"` (without timezone indicator)
- Represents the user's local time

### Conversion to UTC
To convert the datetime-local value to UTC for sending to the backend:

```javascript
// Get value from datetime-local input
const datetimeLocal = "2026-03-05T10:30"

// Treat it as UTC by appending 'Z'
const scheduledDateISO = datetimeLocal + 'Z'  // "2026-03-05T10:30Z"

// Send to backend
await api.createOrder({
  scheduled_date: scheduledDateISO
})
```

### Important: datetime-local Behavior
**Critical understanding:** The `datetime-local` input:
- Does NOT include timezone information
- Returns local time as a string
- Must be explicitly treated as UTC when converting to ISO string

The pattern used in this application is:
```javascript
// Wrong: new Date(datetimeLocal).toISOString()
// This interprets datetimeLocal as local time and converts, losing the intended time

// Correct: datetimeLocal + 'Z'
// This treats datetimeLocal as UTC directly
```

### Display in Frontend
When displaying a scheduled date received from the backend:
```javascript
// Backend returns: "2026-03-05T10:30:00Z" (stored as TIMESTAMPTZ)
const scheduledDate = new Date("2026-03-05T10:30:00Z")

// Display (will show the same time, interpreted as UTC)
scheduledDate.toLocaleString('en-US', { timeZone: 'America/Los_Angeles' })
// Could show: "3/5/2026, 2:30:00 AM" if user is in LA timezone
```

## Affected Pages

- **Items.jsx**: Creates orders with `scheduled_date`
- **OrderEdit.jsx**: Updates orders with `scheduled_date`
- **Orders.jsx**: Displays `scheduled_date` on order cards
- **Schedule.jsx**: Groups and filters orders by `scheduled_date`

## Migration Notes

If upgrading from a version that used `TIMESTAMP` (without timezone):
1. The migration script automatically updates the column type to `TIMESTAMP WITH TIME ZONE`
2. Existing data is preserved (PostgreSQL converts without data loss)
3. All new inserts will use the corrected timezone-aware type

## Testing Timezone Handling

To verify timezone handling works correctly:

1. Create an order with scheduled_date from the Items page
2. Set the scheduled_date to a specific time (e.g., 10:30 AM)
3. Verify in the database that it's stored as UTC
4. Verify on the Schedule page that the date is grouped correctly
5. Verify on the Orders page that the displayed time is correct for your timezone

## References

- [PostgreSQL TIMESTAMP vs TIMESTAMPTZ](https://www.postgresql.org/docs/current/datatype-datetime.html)
- [MDN: HTML datetime-local input](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input/datetime-local)
- [ISO 8601 Standard](https://en.wikipedia.org/wiki/ISO_8601)
- [Go time.Time in UTC](https://golang.org/pkg/time/)
