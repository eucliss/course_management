# DigitalOcean Spaces Upload Script

This script uploads your local `tiles` directory (containing vector tiles) to DigitalOcean Spaces for CDN distribution.

## Prerequisites

1. **DigitalOcean Spaces Account**: You need a DigitalOcean Spaces bucket
2. **API Keys**: Access key and secret key for your Spaces bucket
3. **Go Environment**: Go must be installed and in your PATH

## Environment Variables

Add these to your `.env` file:

```env
# Required
DO_SPACES_ACCESS_KEY=your_access_key_here
DO_SPACES_SECRET_KEY=your_secret_key_here
DO_SPACES_BUCKET=your_bucket_name_here

# Optional (defaults shown)
DO_SPACES_REGION=nyc3
DO_SPACES_ENDPOINT=https://nyc3.digitaloceanspaces.com
```

## Getting DigitalOcean Spaces Credentials

1. **Create a Spaces Bucket**:
   - Go to DigitalOcean Control Panel → Spaces
   - Create a new Space or use existing one
   - Note the bucket name and region

2. **Generate API Keys**:
   - Go to API → Spaces Keys
   - Click "Generate New Key"
   - Copy the Access Key and Secret Key
   - Add them to your `.env` file

## Usage

### Method 1: Shell Script (Recommended)
```bash
cd scripts
./upload_to_spaces.sh
```

### Method 2: Direct Go Execution
```bash
cd scripts
go run upload_to_spaces.go
```

## What It Does

1. **Scans Directory**: Looks for `tiles` directory containing your vector tiles
2. **Uploads Files**: Recursively uploads all files to your Spaces bucket
3. **Sets Permissions**: Makes files publicly accessible via CDN
4. **Content Types**: Automatically sets correct MIME types for different file formats:
   - `.pbf` → `application/x-protobuf` (vector tiles)
   - `.json` → `application/json`
   - `.geojson` → `application/geo+json`
   - `.mbtiles` → `application/vnd.mapbox-vector-tile`
   - And more...

## File Structure

The script uploads to the `tiles` prefix in your bucket:
```
Local: tiles/0/0/0.pbf
Remote: tiles/0/0/0.pbf
URL: https://your-bucket.nyc3.digitaloceanspaces.com/tiles/0/0/0.pbf
```

## After Upload

1. **Test Access**: Visit your CDN URL to ensure files are accessible
2. **Update Environment**: Add the vector tile URL to your `.env`:
   ```env
   VECTOR_TILE_URL=https://your-bucket.nyc3.digitaloceanspaces.com/tiles/{z}/{x}/{y}.pbf
   ```
3. **Restart Application**: Air should auto-reload and use the vector tiles

## Troubleshooting

### Authentication Errors
- Verify your access key and secret key are correct
- Check that your API keys have Spaces permissions
- Ensure bucket name matches exactly (case-sensitive)

### File Not Found Errors
- Script looks for `tiles` directory in current location
- Run from project root or scripts directory
- Ensure your vector tiles are in the `tiles` folder

### Upload Failures
- Check internet connection
- Verify bucket permissions allow uploads
- Ensure sufficient storage space in your Spaces bucket

## Features

- **Smart Path Detection**: Automatically finds `tiles` directory
- **Progress Tracking**: Shows upload progress with file counts and sizes
- **Error Handling**: Detailed error messages for troubleshooting
- **Content Type Detection**: Proper MIME types for web serving
- **Public Access**: Files are uploaded with public-read permissions
- **Environment Loading**: Automatically loads `.env` files

## Security Notes

- Never commit your API keys to version control
- Use environment variables for all sensitive data
- Consider using IAM roles for production deployments
- Regularly rotate your API keys for security

## Performance

- Uploads are sequential (not parallel) to avoid rate limits
- Large files may take time depending on your internet connection
- Consider using DigitalOcean's upload tools for very large datasets

## Cost Considerations

- DigitalOcean Spaces charges for storage and bandwidth
- Vector tiles are typically small but numerous
- Monitor your usage in the DigitalOcean control panel 