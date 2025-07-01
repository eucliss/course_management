# PostgreSQL Database Setup

## Prerequisites

1. **PostgreSQL installed** on your system
2. **Database created** for the application

## Quick Setup

### 1. Create PostgreSQL Database

```bash
# Connect to PostgreSQL as superuser
psql -U postgres

# Create database
CREATE DATABASE course_management;

# Create user (optional but recommended)
CREATE USER course_user WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE course_management TO course_user;

# Exit PostgreSQL
\q
```

### 2. Configure Environment Variables

Create a `.env` file in your project root (copy from `.env.example`):

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=course_user
DB_PASSWORD=your_secure_password
DB_NAME=course_management
DB_SSLMODE=disable

# Other required variables...
SESSION_SECRET=your-very-secure-session-secret-key-32-characters-minimum
GOOGLE_CLIENT_ID=your_google_client_id_here
MAPBOX_ACCESS_TOKEN=your_mapbox_token_here
```

### 3. Run the Application

```bash
# The application will automatically:
# 1. Connect to PostgreSQL
# 2. Create necessary tables
# 3. Migrate existing JSON course data to database
go run .
```

## How It Works

### Hybrid Mode
The application operates in **hybrid mode** during the transition:

- ✅ **Reads** from database first, falls back to JSON files
- ✅ **Writes** to both database AND JSON files (for backup)
- ✅ **Auto-migrates** existing JSON courses to database on first run

### Database Status Check

Visit `http://localhost:8080/api/status/database` to check connection status:

```json
{
  "database_connected": true,
  "message": "Database connected successfully",
  "stats": {
    "courses": 3,
    "users": 0
  }
}
```

## Database Schema

The application creates these tables automatically:

### `users` table
- Stores Google OAuth user information
- Links users to courses they create

### `course_dbs` table  
- Stores course data in JSONB format
- Maintains compatibility with existing JSON structure
- Tracks course creator

## Troubleshooting

### Connection Issues

1. **Check PostgreSQL is running:**
   ```bash
   # macOS (Homebrew)
   brew services status postgresql
   
   # Linux
   systemctl status postgresql
   ```

2. **Verify database exists:**
   ```bash
   psql -U postgres -l
   ```

3. **Test connection manually:**
   ```bash
   psql -h localhost -U course_user -d course_management
   ```

### Common Errors

**"failed to connect to database"**
- Check DB_HOST, DB_PORT, DB_USER, DB_PASSWORD in .env
- Ensure PostgreSQL is running
- Verify user has access to database

**"relation does not exist"**
- Tables are created automatically on first run
- Check application logs for migration errors

### Fallback Mode

If database connection fails, the application will:
- ⚠️ Log warning messages
- 📁 Continue using JSON files
- 🔄 Retry database connection on restart

## Production Considerations

1. **Use SSL:** Set `DB_SSLMODE=require` for production
2. **Connection pooling:** GORM handles this automatically
3. **Backups:** Both database and JSON files should be backed up
4. **Environment:** Set `ENV=production` to reduce database logging 