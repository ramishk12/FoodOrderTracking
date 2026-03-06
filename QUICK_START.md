# Food Order Tracking - Docker Quick Start

Get the application running locally in under 5 minutes!

## Prerequisites

Before you start, ensure you have:
- **Docker Desktop for Windows** installed and running
  - Download: https://www.docker.com/products/docker-desktop
  - This includes Docker and docker-compose

## Quick Start (3 Steps)

### Step 1: Navigate to Project
```bash
cd F:\Development\FoodOrderTracking
```

### Step 2: Start Everything
Choose one option based on your preference:

**Option A: Using PowerShell (Recommended)**
```powershell
.\docker-start.ps1
```

**Option B: Using Command Prompt**
```cmd
docker-start up
```

**Option C: Using Docker Compose directly**
```bash
docker-compose up -d
```

### Step 3: Access Your Application
Open your browser and go to:
- **Application**: http://localhost
- **API Health Check**: http://localhost:8080/health

## What Just Happened?

Docker started three containers:

1. **PostgreSQL Database** - Running on port 5432
   - Database: `food_order_tracking`
   - Username: `postgres`
   - Password: `postgres`

2. **Go Backend API** - Running on port 8080
   - API endpoints at: `http://localhost:8080/api`

3. **React Frontend** - Running on port 80
   - Served with Nginx
   - Automatically proxies API requests to backend

All three services are in a Docker network and communicate with each other.

## Common Tasks

### View Logs
```bash
# All services
docker-compose logs -f

# Just backend
docker-compose logs -f backend

# Just frontend
docker-compose logs -f frontend

# Just database
docker-compose logs -f db
```

### Check Status
```bash
docker-compose ps
```

Expected output:
```
NAME                COMMAND                  SERVICE      STATUS      PORTS
food-order-backend  "./food-order-tracker"   backend      Up 2 mins   0.0.0.0:8080->8080/tcp
food-order-db       "docker-entrypoint.s…"   db           Up 2 mins   0.0.0.0:5432->5432/tcp
food-order-frontend "nginx -g daemon off;"   frontend     Up 1 min    0.0.0.0:80->80/tcp
```

### Restart a Service
```bash
docker-compose restart backend
```

### Stop Everything
```bash
docker-compose down
```

### Restart Everything
```bash
docker-compose restart
```

### Full Cleanup (removes database!)
```bash
docker-compose down -v
```

## Making Code Changes

### Backend Changes (Go)

1. Make your code changes in `cmd/` or `internal/`
2. Rebuild and restart:
   ```bash
   docker-compose restart backend
   ```
   Or rebuild from scratch:
   ```bash
   docker-compose up -d --build
   ```

### Frontend Changes (React)

1. Make your code changes in `web/src/`
2. Rebuild and restart:
   ```bash
   docker-compose up -d --build frontend
   ```

## Troubleshooting

### "Cannot connect to Docker daemon"
- **Solution**: Start Docker Desktop and wait for it to fully load

### "Ports already in use"
- **Solution**: Stop other services using ports 80, 8080, or 5432
  ```bash
  # Find what's using port 80
  netstat -ano | findstr :80
  
  # Kill the process (replace PID with the actual number)
  taskkill /PID <PID> /F
  ```

### "Frontend shows connection error"
- **Solution**: Wait 10-15 seconds for backend to start and database to be ready
- Check logs: `docker-compose logs backend`

### "Database connection refused"
1. Verify database is running:
   ```bash
   docker-compose logs db
   ```
2. Restart database:
   ```bash
   docker-compose restart db
   ```

### Application won't start
- Clear everything and restart:
  ```bash
  docker-compose down -v
  docker-compose up -d
  ```

## Understanding the Setup

### File Structure
```
FoodOrderTracking/
├── Dockerfile.backend          # Go backend containerization
├── Dockerfile.frontend         # React frontend containerization
├── nginx.conf                  # Nginx configuration
├── docker-compose.yml          # Services orchestration
├── docker-start.ps1            # PowerShell helper script
├── docker-start.bat            # Batch file helper script
├── DOCKER_DEPLOYMENT.md        # Full documentation
├── QUICK_START.md              # This file
│
├── cmd/
│   └── main.go                 # Backend entry point
├── internal/
│   ├── database/               # DB connection & migrations
│   ├── handlers/               # API endpoints
│   └── models/                 # Data models
├── web/
│   ├── src/                    # React components
│   ├── index.html
│   └── package.json
├── go.mod                      # Go dependencies
└── go.sum
```

### How Communication Works

```
Browser (http://localhost)
    ↓
Nginx (port 80)
    ├─→ Serves React SPA
    └─→ Proxies /api/* to Backend (port 8080)
        ↓
    Go API (port 8080)
        ↓
    PostgreSQL (port 5432)
```

## Performance Notes

For 5 users on a local machine:

- **Minimal Resource Usage**: ~500MB RAM total
- **CPU**: Minimal (< 5% at rest)
- **Startup Time**: 15-20 seconds
- **Response Time**: < 100ms typically

## Security Notice

⚠️ **This setup is for local development/testing only!**

Default credentials are hardcoded:
- Database username: `postgres`
- Database password: `postgres`

**Never use this in production!** For production, you would:
- Use strong passwords
- Store secrets in environment variables
- Enable SSL/TLS
- Set up proper authentication

## Next Steps

### Learn More
- Full documentation: See `DOCKER_DEPLOYMENT.md`
- Docker basics: https://docs.docker.com/get-started/

### Deployment
- When ready for production, review `DOCKER_DEPLOYMENT.md` for hardening the setup
- Consider cloud options like AWS, Azure, or DigitalOcean

## Quick Command Reference

```bash
# Start
docker-compose up -d

# Stop
docker-compose down

# View logs
docker-compose logs -f

# Check status
docker-compose ps

# Restart one service
docker-compose restart backend

# Rebuild after code changes
docker-compose up -d --build

# Full cleanup (removes database)
docker-compose down -v

# Connect to database
docker-compose exec db psql -U postgres -d food_order_tracking

# Connect to backend container
docker-compose exec backend sh

# View database tables
docker-compose exec db psql -U postgres -d food_order_tracking -c "\dt"
```

## Getting Help

1. Check the logs: `docker-compose logs -f`
2. See if containers are running: `docker-compose ps`
3. Test if backend is responding: `curl http://localhost:8080/health`
4. Read the full documentation: `DOCKER_DEPLOYMENT.md`

---

**Enjoy your containerized Food Order Tracking application! 🐳**
