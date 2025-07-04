#!/bin/bash

echo "🔍 Course Reviews Report"
echo "======================="
echo

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the course_management directory"
    echo "   cd course_management && ./scripts/list_reviews.sh"
    exit 1
fi

# Check if database environment variables are set
if [ -z "$DB_PASSWORD" ]; then
    echo "⚠️  Warning: DB_PASSWORD environment variable not set"
    echo "   You may need to set your database credentials:"
    echo "   export DB_PASSWORD='your_password'"
    echo "   export DB_HOST='localhost'"
    echo "   export DB_USER='postgres'"
    echo "   export DB_NAME='course_management'"
    echo
fi

echo "🚀 Running review listing script..."
echo

# Run the Go script
cd scripts
go run list_reviews.go

echo
echo "✅ Script completed!" 