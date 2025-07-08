package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

type SpacesConfig struct {
	AccessKey  string
	SecretKey  string
	Region     string
	Endpoint   string
	BucketName string
}

func loadConfig() (*SpacesConfig, error) {
	// Try to load .env from current directory, parent, or grandparent
	envPaths := []string{".env", "../.env", "../../.env"}
	var envLoaded bool
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			fmt.Printf("📄 Loaded environment variables from %s\n", path)
			envLoaded = true
			break
		}
	}
	if !envLoaded {
		log.Printf("Warning: .env file not found")
	}

	config := &SpacesConfig{
		AccessKey:  os.Getenv("DO_SPACES_ACCESS_KEY"),
		SecretKey:  os.Getenv("DO_SPACES_SECRET_KEY"),
		Region:     getEnvOrDefault("DO_SPACES_REGION", "nyc3"),
		Endpoint:   getEnvOrDefault("DO_SPACES_ENDPOINT", "https://nyc3.digitaloceanspaces.com"),
		BucketName: os.Getenv("DO_SPACES_BUCKET"),
	}

	if config.AccessKey == "" || config.SecretKey == "" || config.BucketName == "" {
		return nil, fmt.Errorf("missing required environment variables:\n" +
			"  DO_SPACES_ACCESS_KEY\n" +
			"  DO_SPACES_SECRET_KEY\n" +
			"  DO_SPACES_BUCKET\n" +
			"Optional:\n" +
			"  DO_SPACES_REGION (default: nyc3)\n" +
			"  DO_SPACES_ENDPOINT (default: https://nyc3.digitaloceanspaces.com)")
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createS3Session(config *SpacesConfig) (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
		Endpoint:    aws.String(config.Endpoint),
		Region:      aws.String(config.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return s3.New(sess), nil
}

func uploadFile(svc *s3.S3, bucketName, localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", localPath, err)
	}
	defer file.Close()

	// Get file info for content type detection
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// Determine content type based on file extension
	contentType := getContentType(localPath)

	// Set ACL to public-read for web accessibility
	acl := "public-read"

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(remotePath),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         aws.String(acl),
	})

	if err != nil {
		return fmt.Errorf("failed to upload %s: %v", remotePath, err)
	}

	fmt.Printf("✅ Uploaded: %s → %s (%d bytes)\n", localPath, remotePath, fileInfo.Size())
	return nil
}

func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pbf":
		return "application/x-protobuf"
	case ".json":
		return "application/json"
	case ".geojson":
		return "application/geo+json"
	case ".mbtiles":
		return "application/vnd.mapbox-vector-tile"
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".html":
		return "text/html"
	case ".txt":
		return "text/plain"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

func uploadDirectory(svc *s3.S3, bucketName, localDir, remotePrefix string) error {
	fmt.Printf("🔄 Scanning directory: %s\n", localDir)

	var uploadCount int
	var totalSize int64

	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path from the base directory
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Convert to forward slashes for S3 key (works on all platforms)
		remotePath := filepath.ToSlash(filepath.Join(remotePrefix, relPath))

		// Upload the file
		if err := uploadFile(svc, bucketName, path, remotePath); err != nil {
			return err
		}

		uploadCount++
		totalSize += info.Size()
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("\n📊 Upload Summary:\n")
	fmt.Printf("   Files uploaded: %d\n", uploadCount)
	fmt.Printf("   Total size: %.2f MB\n", float64(totalSize)/(1024*1024))

	return nil
}

func main() {
	fmt.Println("☁️ DIGITALOCEAN SPACES UPLOADER")
	fmt.Println("═══════════════════════════════════")
	fmt.Println()

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Configuration error: %v\n", err)
		fmt.Println("\n💡 Add these to your .env file:")
		fmt.Println("   DO_SPACES_ACCESS_KEY=your_access_key")
		fmt.Println("   DO_SPACES_SECRET_KEY=your_secret_key")
		fmt.Println("   DO_SPACES_BUCKET=your_bucket_name")
		fmt.Println("   DO_SPACES_REGION=nyc3  # optional")
		fmt.Println("   DO_SPACES_ENDPOINT=https://nyc3.digitaloceanspaces.com  # optional")
		os.Exit(1)
	}

	fmt.Printf("🔧 Configuration:\n")
	fmt.Printf("   Bucket: %s\n", config.BucketName)
	fmt.Printf("   Region: %s\n", config.Region)
	fmt.Printf("   Endpoint: %s\n", config.Endpoint)
	fmt.Println()

	// Create S3 session
	svc, err := createS3Session(config)
	if err != nil {
		fmt.Printf("❌ Failed to create session: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Connected to DigitalOcean Spaces")

	// Determine what to upload
	sourceDir := "tiles"
	remotePrefix := "tiles"

	// Check if we're in scripts directory
	if _, err := os.Stat("tiles"); os.IsNotExist(err) {
		// Try parent directory
		if _, err := os.Stat("../tiles"); err == nil {
			sourceDir = "../tiles"
		} else if _, err := os.Stat("scripts/tiles"); err == nil {
			sourceDir = "scripts/tiles"
		} else {
			fmt.Printf("❌ Could not find 'tiles' directory\n")
			fmt.Println("💡 Please run this script from:")
			fmt.Println("   • Project root (if scripts/tiles/ exists)")
			fmt.Println("   • Scripts directory (if tiles/ exists)")
			os.Exit(1)
		}
	}

	fmt.Printf("📁 Source directory: %s\n", sourceDir)
	fmt.Printf("🎯 Remote prefix: %s\n", remotePrefix)
	fmt.Println()

	// Upload the directory
	if err := uploadDirectory(svc, config.BucketName, sourceDir, remotePrefix); err != nil {
		fmt.Printf("❌ Upload failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("🎉 Upload completed successfully!")
	fmt.Printf("🌐 Your files are now available at:\n")
	fmt.Printf("   https://%s.%s/%s/\n",
		config.BucketName,
		strings.TrimPrefix(config.Endpoint, "https://"),
		remotePrefix)
	fmt.Println()
	fmt.Println("💡 To use vector tiles, set in your .env:")
	fmt.Printf("   VECTOR_TILE_URL=https://%s.%s/%s/{z}/{x}/{y}.pbf\n",
		config.BucketName,
		strings.TrimPrefix(config.Endpoint, "https://"),
		remotePrefix)
}
