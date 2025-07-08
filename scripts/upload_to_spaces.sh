#!/bin/bash

# DigitalOcean Spaces Upload Script
# This script uploads the tools directory to DigitalOcean Spaces

echo "☁️ DIGITALOCEAN SPACES UPLOADER"
echo "═══════════════════════════════════"
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    echo "💡 Please install Go first: https://golang.org/download/"
    exit 1
fi

# Check if we have the required dependencies
echo "🔧 Checking dependencies..."
if ! go list github.com/aws/aws-sdk-go/service/s3 &> /dev/null; then
    echo "📦 Installing AWS SDK..."
    go get github.com/aws/aws-sdk-go/service/s3
fi

if ! go list github.com/joho/godotenv &> /dev/null; then
    echo "📦 Installing godotenv..."
    go get github.com/joho/godotenv
fi

echo "✅ Dependencies ready"
echo

# Run the upload script
echo "🚀 Starting upload..."
go run upload_to_spaces.go

echo
echo "✨ Upload script completed" 