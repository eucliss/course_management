#!/bin/bash

echo "🗺️ Golf Course GeoJSON Exporter"
echo "═══════════════════════════════"
echo ""
echo "📝 This script will:"
echo "   • Connect to your PostgreSQL database"
echo "   • Export all courses with coordinates to GeoJSON format"
echo "   • Create a golf_courses.geojson file"
echo ""
echo "📋 Prerequisites:"
echo "   • Database must be running"
echo "   • Courses must have coordinates (run geocode_courses.sh first)"
echo "   • .env file must contain database credentials"
echo ""

# Check if we're in the scripts directory
if [[ $(basename "$PWD") == "scripts" ]]; then
    echo "📂 Running from scripts directory..."
    go run export_geojson.go
else
    echo "📂 Running from project root..."
    cd scripts && go run export_geojson.go
fi

echo ""
echo "✅ GeoJSON export script completed!"
echo ""
echo "📄 Output file: golf_courses.geojson"
echo ""
echo "🔗 You can now use this file with:"
echo "   • GitHub (drag & drop to view on map)"
echo "   • QGIS or ArcGIS"
echo "   • Mapbox Studio"
echo "   • Leaflet or other web mapping libraries"
echo "   • Google My Maps (import as KML after conversion)" 