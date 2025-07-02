#!/bin/bash

echo "🔄 Course Migration Script"
echo "=========================="

# Check if server is running
if ! curl -s http://localhost:8080/api/status/database > /dev/null 2>&1; then
    echo "❌ Server is not running on localhost:8080"
    echo "Please start the server first with: go run ."
    exit 1
fi

echo "✅ Server is running"

# Check database status
echo "🔍 Checking database connection..."
DB_STATUS=$(curl -s http://localhost:8080/api/status/database)
echo "Database status: $DB_STATUS"

if echo "$DB_STATUS" | grep -q '"database_connected":true'; then
    echo "✅ Database is connected"
    
    echo "📤 Starting course migration..."
    MIGRATION_RESULT=$(curl -s -X POST http://localhost:8080/api/migrate/courses)
    
    echo "Migration result:"
    echo "$MIGRATION_RESULT" | jq . 2>/dev/null || echo "$MIGRATION_RESULT"
    
    if echo "$MIGRATION_RESULT" | grep -q "Migration completed successfully"; then
        echo "🎉 Migration completed successfully!"
        
        echo "📊 Final database status:"
        curl -s http://localhost:8080/api/status/database | jq . 2>/dev/null || curl -s http://localhost:8080/api/status/database
    else
        echo "❌ Migration failed"
        exit 1
    fi
else
    echo "❌ Database is not connected"
    echo "Please check your database configuration in .env file"
    exit 1
fi 