#!/bin/bash

# Cleanup script for Course Management application
echo "🧹 Cleaning up temporary files..."

# Remove build artifacts
rm -f tmp/main tmp/course_management tmp/build-errors.log

# Clean up large log files (keep last 1000 lines)
if [ -f "logs/app.log" ]; then
    tail -n 1000 logs/app.log > logs/app.log.tmp && mv logs/app.log.tmp logs/app.log
    echo "✅ Cleaned up log files"
fi

# Remove any backup files
find . -name "*.bak" -delete
find . -name "*~" -delete

# Clean up Go build cache
go clean -cache

echo "✅ Cleanup complete!" 