# Configuration Management

This document explains how to configure the Course Management System for different environments.

## Overview

The Course Management System uses a comprehensive configuration system that supports:

- Environment-specific configuration files
- Environment variable overrides
- Secure secret management
- Configuration validation
- Docker deployment support

## Quick Start

### 1. Choose Your Environment

Set the environment variable:
```bash
export ENV=development  # or testing, staging, production
```

### 2. Load Configuration

The application automatically loads configuration from:
1. `config/{environment}.env` file
2. `.env` file (as fallback)
3. Environment variables (highest priority)

### 3. Generate Secrets (Optional)

For production environments, generate secure secrets:
```bash
go run cmd/generate-secrets/main.go --env=production
```

## Configuration Structure

### Environment Files

- `config/development.env` - Development environment settings
- `config/testing.env` - Testing environment settings  
- `config/production.env` - Production environment settings
- `config/{env}.secrets.env` - Secure secrets (generated, not in git)

### Configuration Sections

#### Server Configuration
```bash
# Server settings
PORT=8080
HOST=localhost
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_SHUTDOWN_TIMEOUT=10s
MAX_REQUEST_SIZE=33554432  # 32MB
```

#### Database Configuration
```bash
# Database connection
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=course_management
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=5m
```

#### Security Configuration
```bash
# Security settings
SESSION_SECRET=your-very-secure-session-secret-key-32-characters-minimum
CSRF_SECRET=your-csrf-secret-key
JWT_SECRET=your-jwt-secret-key
SESSION_TIMEOUT=24h
SECURE_COOKIES=false  # true in production
RATE_LIMIT_PER_MIN=60
BCRYPT_COST=12
TRUSTED_PROXIES=192.168.1.0/24,10.0.0.0/8
```

#### Google OAuth Configuration
```bash
# Google OAuth
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
```

#### Logging Configuration
```bash
# Logging settings
LOG_LEVEL=info          # debug, info, warn, error
LOG_FORMAT=json         # json, text
LOG_OUTPUT=stdout       # stdout, file
LOG_MAX_SIZE=100        # MB
LOG_MAX_BACKUPS=3
LOG_MAX_AGE=28          # days
LOG_COMPRESS=true
```

#### Path Configuration
```bash
# File paths
VIEWS_DIR=views
STATIC_DIR=static
UPLOADS_DIR=uploads
TEMPLATES_DIR=templates
```

## Environment-Specific Settings

### Development
- Detailed logging (DEBUG level)
- Hot reloading support
- Relaxed security settings
- Local database connection

### Testing
- Minimal logging (WARN level)
- In-memory database
- Fast startup/shutdown
- Mock external services

### Production
- Secure defaults
- Required secret validation
- Performance optimization
- HTTPS enforcement

## Secret Management

### Generating Secrets

Use the built-in secret generator:
```bash
# Generate secrets for production
go run cmd/generate-secrets/main.go --env=production

# Generate secrets for development
go run cmd/generate-secrets/main.go --env=development
```

### Manual Secret Generation

For custom secrets, ensure they meet minimum requirements:
- `SESSION_SECRET`: 32+ characters
- `CSRF_SECRET`: 16+ characters  
- `JWT_SECRET`: 32+ characters

Example using OpenSSL:
```bash
# Generate 32-byte base64 encoded secret
openssl rand -base64 32

# Generate 64-character hex secret
openssl rand -hex 32
```

### Security Best Practices

1. **Never commit secrets to version control**
2. **Use environment variables in production**
3. **Rotate secrets regularly**
4. **Use different secrets per environment**
5. **Monitor access to secret files**

## Docker Configuration

### Development with Docker Compose

```bash
# Start development environment
docker-compose up

# With custom environment file
docker-compose --env-file config/development.env up
```

### Production Deployment

```bash
# Build production image
docker build -t course-management .

# Run with environment variables
docker run -d \
  --name course-management \
  -p 8080:8080 \
  -e ENV=production \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  -e SESSION_SECRET=your-session-secret \
  course-management
```

## Validation

The configuration system includes comprehensive validation:

### Required Fields
- `SESSION_SECRET` (32+ characters in production)
- `DB_HOST` and `DB_NAME`
- `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` (in production)

### Environment Validation
- Valid environment names: `development`, `testing`, `staging`, `production`
- Path validation (directories must exist except in testing)
- Database connection validation

### Security Validation
- Secret length requirements
- Weak password detection
- Production-specific validations

## Troubleshooting

### Common Issues

#### Configuration Not Loading
```bash
# Check if environment file exists
ls -la config/

# Verify environment variable is set
echo $ENV

# Check for syntax errors in .env file
cat config/development.env
```

#### Database Connection Failed
```bash
# Test database connection
psql -h localhost -U postgres -d course_management

# Check database configuration
echo $DB_HOST $DB_PORT $DB_NAME
```

#### Secret Validation Failed
```bash
# Check secret length
echo -n "$SESSION_SECRET" | wc -c

# Generate new secret if needed
go run cmd/generate-secrets/main.go --env=development
```

### Debugging Configuration

Enable debug logging to see configuration loading:
```bash
export LOG_LEVEL=debug
./course_management
```

### Health Check

The application provides a health check endpoint:
```bash
curl http://localhost:8080/health
```

Response includes:
- Application status
- Current environment
- Configuration validation status

## Migration from Old Configuration

If migrating from the old `config.go` system:

1. **Backup existing configuration**
2. **Set environment**: `export ENV=development`
3. **Create new environment file** from template
4. **Copy existing values** to new format
5. **Test configuration** with health check
6. **Update deployment scripts** to use new system

## Configuration Reference

For a complete list of all configuration options, see:
- `config/config.go` - Configuration structure
- `config/development.env` - Development example
- `config/production.env` - Production template

## Support

For configuration issues:
1. Check this documentation
2. Validate configuration with health check
3. Review application logs
4. Verify environment file syntax
5. Test database connectivity