# Course Geocoding Setup

This document explains how to use the geocoding script to pre-process course coordinates and improve map performance.

## Overview

The geocoding script solves the performance problem where the map makes thousands of simultaneous API calls to Mapbox's geocoding service. Instead, it pre-processes all course addresses and stores the latitude/longitude coordinates directly in the database.

## Benefits

- **Faster Map Loading**: No real-time geocoding API calls
- **Reduced API Costs**: Geocoding happens once, not on every map load
- **Better User Experience**: Maps load instantly with pre-calculated coordinates
- **Reduced Rate Limiting**: Avoids hitting Mapbox API rate limits

## Prerequisites

1. **Mapbox Access Token**: You need a valid Mapbox access token (can be in `.env` file)
2. **Database Connection**: The script connects to your PostgreSQL database
3. **Go Environment**: Go must be installed to run the script

## Setup Instructions

### 1. Environment Variables

The script automatically loads environment variables from your `.env` file. Make sure your `.env` file contains:

```bash
# Required: Mapbox access token
MAPBOX_ACCESS_TOKEN=your_mapbox_token_here

# Optional: Database connection (uses defaults if not set)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=course_management_dev
DB_SSLMODE=disable
```

**Note**: The script looks for `.env` files in the current directory and parent directories automatically.

### 2. Run the Geocoding Script

```bash
# Navigate to the scripts directory
cd course_management/scripts

# Run the geocoding script (easiest method)
./geocode_courses.sh
```

Or run the Go script directly:

```bash
# Build and run manually
go run geocode_courses.go
```

**No manual environment variable setup needed!** The script will automatically load your `.env` file.

## What the Script Does

1. **Environment Loading**: Automatically loads environment variables from `.env` file
2. **Database Migration**: Adds `latitude` and `longitude` columns to the `course_dbs` table if they don't exist
3. **Fetches Courses**: Retrieves all courses from the database
4. **Geocodes Addresses**: For each course with an address:
   - Calls Mapbox Geocoding API
   - Extracts latitude/longitude coordinates
   - Stores coordinates in the database
5. **Rate Limiting**: Includes 100ms delays between API calls to respect Mapbox limits
6. **Error Handling**: Continues processing even if some courses fail to geocode

## Script Output

The script provides detailed logging:

```
🚀 Starting course geocoding script...
✅ Loaded environment variables from ../.env
✅ Connected to PostgreSQL database: course_management_dev
🔄 Adding latitude and longitude columns to courses table...
✅ Added latitude column
✅ Added longitude column
📊 Found 150 courses to geocode
🔄 Processing course 1/150: Pebble Beach Golf Links
✅ Geocoded Pebble Beach Golf Links: lat=36.571429, lng=-121.948314
📈 Progress: 10/150 processed (✅ 9 successful, ⏭️ 1 skipped, ❌ 0 failed)
...
🎉 Geocoding complete!
📊 Final Statistics:
   Total courses: 150
   Processed: 150
   ✅ Successfully geocoded: 145
   ⏭️ Skipped (already geocoded or no address): 3
   ❌ Failed: 2
```

## Frontend Integration

After running the script, the frontend map code will automatically:

1. **Check for Stored Coordinates**: First tries to use `course.Latitude` and `course.Longitude`
2. **Fallback to Geocoding**: If no stored coordinates, falls back to real-time geocoding
3. **Improved Performance**: Maps load much faster with pre-calculated coordinates

## Re-running the Script

The script is safe to run multiple times:

- **Skips Already Geocoded**: Courses with existing coordinates are skipped
- **Processes New Courses**: Only geocodes courses without coordinates
- **Updates Failed Courses**: Re-attempts previously failed geocoding

## Troubleshooting

### Common Issues

1. **Missing Mapbox Token**:
   ```
   ❌ MAPBOX_ACCESS_TOKEN environment variable is required. Please set it in your .env file or environment.
   ```
   Solution: Add `MAPBOX_ACCESS_TOKEN=your_token_here` to your `.env` file

2. **Database Connection Failed**:
   ```
   ❌ Failed to connect to database: connection refused
   ```
   Solution: Check database connection settings in your `.env` file and ensure PostgreSQL is running

3. **Geocoding API Errors**:
   ```
   ❌ Failed to geocode Course Name: geocoding API returned status 401
   ```
   Solution: Check that your Mapbox token in `.env` is valid and has geocoding permissions

4. **No .env File Found**:
   ```
   ℹ️ No .env file found, using system environment variables
   ```
   This is fine if you have environment variables set system-wide, but using a `.env` file is recommended.

### Debugging

- Check the script logs for detailed error messages
- Verify your `.env` file contains the correct Mapbox token
- Verify database connection with `psql` or your database client
- Test Mapbox token with a simple curl request:
  ```bash
  curl "https://api.mapbox.com/geocoding/v5/mapbox.places/test.json?access_token=YOUR_TOKEN"
  ```

## Database Schema Changes

The script adds these columns to the `course_dbs` table:

```sql
ALTER TABLE course_dbs ADD COLUMN latitude DOUBLE PRECISION;
ALTER TABLE course_dbs ADD COLUMN longitude DOUBLE PRECISION;
```

## Performance Impact

- **Before**: 10,000+ simultaneous API calls on map load
- **After**: 0 API calls on map load (uses stored coordinates)
- **Initial Setup**: One-time geocoding of all courses (takes 5-10 minutes for 1000 courses)

## Maintenance

- **New Courses**: Run the script after adding new courses to the database
- **Address Changes**: Run the script if course addresses are updated
- **Periodic Updates**: Consider running monthly to catch any new or updated courses

## Cost Considerations

- **Mapbox Geocoding**: ~$0.50 per 1,000 requests
- **One-time Cost**: Geocoding 1,000 courses costs ~$0.50
- **Ongoing Savings**: Eliminates repeated geocoding costs on every map load 