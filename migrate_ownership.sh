#!/bin/bash

echo "🔄 Course Ownership Migration Script"
echo "===================================="

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the course_management directory"
    exit 1
fi

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "⚠️ Warning: .env file not found. Make sure database credentials are set as environment variables."
fi

echo "📋 This script will:"
echo "   1. Add UpdatedBy field to CourseDB table"
echo "   2. Add Updater relationship for tracking course edits"
echo "   3. Ensure proper foreign key constraints"
echo ""

# Ask for confirmation
read -p "🤔 Do you want to proceed with the migration? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Migration cancelled"
    exit 0
fi

echo "🚀 Running migration..."
echo ""

# Run the migration
cd scripts
go run migrate_ownership.go

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 Migration completed successfully!"
    echo ""
    echo "📝 What changed:"
    echo "   - Added 'updated_by' column to courses table"
    echo "   - Added foreign key relationship to users table"
    echo ""
    echo "🔄 Don't forget to:"
    echo "   1. Restart your application to pick up the schema changes"
    echo "   2. Update handlers to use the new ownership fields"
else
    echo ""
    echo "❌ Migration failed. Please check the error messages above."
    exit 1
fi 