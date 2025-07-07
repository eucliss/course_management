#!/bin/bash

# Geocode all courses in the database
# This script will add latitude and longitude to all courses by calling the Mapbox Geocoding API
# The script will automatically load environment variables from .env file

echo "🚀 Starting course geocoding process..."

# Navigate to the scripts directory
cd "$(dirname "$0")"

# Check if the geocode_courses.go file exists
if [ ! -f "geocode_courses.go" ]; then
    echo "❌ Error: geocode_courses.go not found in current directory"
    exit 1
fi

echo "🔄 Building geocoding script..."

# Build the Go script
go build -o geocode_courses geocode_courses.go

if [ $? -ne 0 ]; then
    echo "❌ Failed to build geocoding script"
    echo "💡 Make sure you have Go installed and dependencies are available"
    exit 1
fi

echo "✅ Script built successfully"

# Run the geocoding script
echo "🗺️ Starting geocoding process..."
echo "📝 The script will automatically load MAPBOX_ACCESS_TOKEN from your .env file"
echo "⏰ This may take a while depending on the number of courses..."

./geocode_courses

# Check the exit status
if [ $? -eq 0 ]; then
    echo "🎉 Geocoding completed successfully!"
    echo "📍 All courses now have latitude and longitude coordinates stored in the database"
    echo "🚀 Your maps will now load much faster!"
else
    echo "❌ Geocoding failed with errors"
    echo "💡 Check the output above for error details"
    exit 1
fi

# Clean up the binary
rm -f geocode_courses

echo "🧹 Cleanup completed"
echo "✅ Geocoding process finished!" 