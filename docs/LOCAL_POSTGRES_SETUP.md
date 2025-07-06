# Local PostgreSQL Database Setup Guide

## Overview
This guide walks you through setting up a local PostgreSQL v17 database for development with the course management application.

## Prerequisites
- macOS with Homebrew installed
- Terminal access
- Basic familiarity with command line operations

## Installation Steps

### 1. Install PostgreSQL v17
```bash
brew install postgresql@17
```

### 2. Add PostgreSQL to PATH
```bash
echo 'export PATH="/opt/homebrew/opt/postgresql@17/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### 3. Configure PostgreSQL Port
Since you may have other PostgreSQL installations, configure this instance to run on port 5433:
```bash
echo "port = 5433" >> /opt/homebrew/var/postgresql@17/postgresql.conf
```

### 4. Start PostgreSQL Service
```bash
brew services start postgresql@17
```

### 5. Verify Installation
```bash
psql --version
# Should show: psql (PostgreSQL) 17.5 (Homebrew)
```

## Database Setup

### 1. Create Development Database
```bash
createdb -h localhost -p 5433 course_management_dev
```

### 2. Create Test Database (Optional)
```bash
createdb -h localhost -p 5433 course_management_test
```

### 3. Verify Database Creation
```bash
psql -h localhost -p 5433 -d course_management_dev -c "SELECT current_database(), version();"
```

### 4. Start the application
```bash
cd course_management
air
```

### 5. Import Courses
```bash
cd course_management
./scripts/import_courses.sh
```

## Connection Details

### Database Configuration
- **Host**: `localhost`
- **Port**: `5433`
- **Database Name**: `course_management_dev`
- **Username**: `<local user>` (your system username)
- **Password**: None (trust authentication)
- **SSL Mode**: `disable` (for local development)

### Connection Strings

#### For Go Applications
```go
// Standard format
"host=localhost port=5433 user=<local user> dbname=course_management_dev sslmode=disable"

// PostgreSQL URL format
"postgres://<local user>@localhost:5433/course_management_dev?sslmode=disable"

// Minimal format (uses system defaults)
"host=localhost port=5433 dbname=course_management_dev sslmode=disable"
```

#### For Command Line Access
```bash
# Connect to development database
psql -h localhost -p 5433 -d course_management_dev

# Connect with explicit user
psql -h localhost -p 5433 -U <local user> -d course_management_dev
```

## Service Management

### Start/Stop/Restart Service
```bash
# Start PostgreSQL
brew services start postgresql@17

# Stop PostgreSQL
brew services stop postgresql@17

# Restart PostgreSQL
brew services restart postgresql@17

# Check service status
brew services list | grep postgresql
```

### Manual Server Control
```bash
# Start manually (foreground)
postgres -D /opt/homebrew/var/postgresql@17

# Stop manually
pg_ctl -D /opt/homebrew/var/postgresql@17 stop
```

## Application Configuration

### Update Your Go Application
In your Go application, update the database connection configuration:

```go
// config.go or database.go
const (
    DBHost     = "localhost"
    DBPort     = 5433
    DBUser     = "<local user>"
    DBName     = "course_management_dev"
    DBSSLMode  = "disable"
)

// Connection string
dbConnectionString := fmt.Sprintf(
    "host=%s port=%d user=%s dbname=%s sslmode=%s",
    DBHost, DBPort, DBUser, DBName, DBSSLMode,
)
```

### Environment Variables (Recommended)
Create a `.env` file in your project root:
```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=<local user>
DB_NAME=course_management_dev
DB_SSL_MODE=disable
```

## Database Operations

### Common Commands
```bash
# List all databases
psql -h localhost -p 5433 -l

# Connect to specific database
psql -h localhost -p 5433 -d course_management_dev

# Run SQL file
psql -h localhost -p 5433 -d course_management_dev -f schema.sql

# Backup database
pg_dump -h localhost -p 5433 course_management_dev > backup.sql

# Restore database
psql -h localhost -p 5433 -d course_management_dev < backup.sql
```

### Inside psql
```sql
-- List tables
\dt

-- Describe table structure
\d table_name

-- Show current connection info
\conninfo

-- Exit psql
\q
```

## Migration and Schema Management

### Running Migrations
If you have migration files:
```bash
# Navigate to your project directory
cd /Users/<local user>/Projects/empty/course_management

# Run migrations (example using your migrate command)
go run cmd/migrate/main.go
```

### Manual Schema Creation
```sql
-- Example table creation
CREATE TABLE courses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Troubleshooting

### Common Issues

#### Port Already in Use
If you get port 5432 conflicts:
```bash
# Check what's using the port
lsof -i :5432

# Use port 5433 instead (already configured)
psql -h localhost -p 5433 -d course_management_dev
```

#### Service Won't Start
```bash
# Check service status
brew services list | grep postgresql

# View logs
tail -f /opt/homebrew/var/log/postgresql@17.log

# Restart service
brew services restart postgresql@17
```

#### Connection Refused
```bash
# Verify service is running
brew services list | grep postgresql

# Check if PostgreSQL is listening on correct port
lsof -i :5433

# Verify configuration
cat /opt/homebrew/var/postgresql@17/postgresql.conf | grep port
```

### Authentication Issues
The database is configured with "trust" authentication for local connections, so no password is required. If you encounter authentication issues:

```bash
# Check authentication configuration
cat /opt/homebrew/var/postgresql@17/pg_hba.conf
```

## Security Notes

### Local Development
- Trust authentication is safe for local development
- Database is only accessible from localhost
- No external network access by default

### Production Considerations
For production deployment:
- Use password authentication
- Configure SSL/TLS encryption
- Restrict network access
- Use environment variables for credentials

## Quick Reference

### Essential Commands
```bash
# Start PostgreSQL
brew services start postgresql@17

# Connect to database
psql -h localhost -p 5433 -d course_management_dev

# Create new database
createdb -h localhost -p 5433 database_name

# Check service status
brew services list | grep postgresql

# View logs
tail -f /opt/homebrew/var/log/postgresql@17.log
```

### Connection String Template
```
host=localhost port=5433 user=<local user> dbname=course_management_dev sslmode=disable
```

## Support

If you encounter issues:
1. Check the service status: `brew services list | grep postgresql`
2. Review the logs: `tail -f /opt/homebrew/var/log/postgresql@17.log`
3. Verify the connection: `psql -h localhost -p 5433 -d course_management_dev`
4. Restart the service: `brew services restart postgresql@17`

---

**Last Updated**: January 2025  
**PostgreSQL Version**: 17.5 (Homebrew)  
**Platform**: macOS (Apple Silicon) 