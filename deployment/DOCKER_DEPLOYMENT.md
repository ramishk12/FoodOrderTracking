# Docker Deployment Guide

## Overview

This guide explains how to deploy the Food Order Tracking application using Docker.

## Project Structure

```
FoodOrderTracking/
├── deployment/
│   ├── docker-compose.yml      # Local development orchestration
│   ├── Dockerfile.backend     # Go backend Docker image
│   ├── Dockerfile.frontend   # React frontend Docker image
│   ├── nginx.conf            # Nginx configuration
│   ├── DOCKER_DEPLOYMENT.md  # This file
│   └── DEPLOYMENT_SUMMARY.md # Quick reference
├── .dockerignore              # Build optimization
├── cmd/                       # Backend entry point
├── internal/                  # Backend source code
├── web/                       # Frontend source code
├── .github/workflows/         # CI/CD pipelines
└── AGENTS.md                  # Development guidelines
```

## Quick Start (Local Development)

### Prerequisites
- **Docker Desktop** (Windows/Mac/Linux)
  - Download: https://www.docker.com/products/docker-desktop
  - Includes Docker and docker-compose
- Git

### Start All Services
```bash
cd deployment
docker-compose up -d
```

This will start:
- **PostgreSQL** (port 5432)
- **Go Backend API** (port 8080)
- **React Frontend** (port 80)

### Access the Application
- **Frontend**: http://localhost
- **Backend API**: http://localhost:8080/api
- **Database**: localhost:5432 (postgres/postgres)

### Common Commands
```bash
# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild after code changes
docker-compose build --no-cache
docker-compose up -d

# Remove everything including database
docker-compose down -v
```

## Architecture

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

All three services run in a Docker network and communicate with each other internally.

### Services Overview

| Service | Port | Description |
|---------|------|-------------|
| PostgreSQL | 5432 | Database: `food_order_tracking`, User: `postgres`, Password: `postgres` |
| Go Backend | 8080 | API endpoints at `/api/*` |
| React Frontend | 80 | Served with Nginx, proxies API requests to backend |

## Production Deployment

### Option 1: GitHub Container Registry (Recommended)

The CI workflow automatically builds and pushes Docker images to GHCR on push to main:

```
# amd64 (Intel/AMD)
ghcr.io/ramishk12/food-order-tracking/backend:latest
ghcr.io/ramishk12/food-order-tracking/frontend:latest

# armv7 (Raspberry Pi 3)
ghcr.io/ramishk12/food-order-tracking/backend:armv7
ghcr.io/ramishk12/food-order-tracking/frontend:armv7
```

#### Create GitHub Token for Pulling Images

1. Go to: https://github.com/settings/tokens/new
2. Select scopes: `read:packages`
3. Generate token and copy it

#### Login to GHCR on Production Server

```bash
# SSH to your server/Pi
ssh pi@10.0.0.122

# Login to GHCR (paste your token when prompted)
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_ACTOR --password-stdin

# Or interactively
docker login ghcr.io
# Username: your GitHub username
# Password: your GitHub token (with read:packages scope)
```

#### Pull Images

```bash
# For Raspberry Pi 3 (armv7)
docker pull ghcr.io/ramishk12/food-order-tracking/backend:armv7
docker pull ghcr.io/ramishk12/food-order-tracking/frontend:armv7

# For x86_64 servers (amd64)
docker pull ghcr.io/ramishk12/food-order-tracking/backend:latest
docker pull ghcr.io/ramishk12/food-order-tracking/frontend:latest
```

#### Run with Docker Compose
Create `docker-compose.yml` on your server:

> **For Raspberry Pi 3:** Use `armv7` tag instead of `latest`
> ```yaml
> image: ghcr.io/ramishk12/food-order-tracking/backend:armv7
> image: ghcr.io/ramishk12/food-order-tracking/frontend:armv7
> ```

```yaml
version: '3.8'

services:
  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: food_order_tracking
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: your_secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  backend:
    image: ghcr.io/ramishk12/food-order-tracking/backend:latest
    environment:
      DB_HOST: db
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: your_secure_password
      DB_NAME: food_order_tracking
      GIN_MODE: release
    depends_on:
      - db
    restart: unless-stopped

  frontend:
    image: ghcr.io/ramishk12/food-order-tracking/frontend:latest
    ports:
      - "80:80"
    depends_on:
      - backend
    restart: unless-stopped

volumes:
  postgres_data:
```

### Option 2: Build Locally and Transfer

```bash
# Build images (from project root)
docker build -t food-order-backend:latest -f deployment/Dockerfile.backend .
docker build -t food-order-frontend:latest -f deployment/Dockerfile.frontend .

# Save images
docker save food-order-backend:latest -o backend.tar
docker save food-order-frontend:latest -o frontend.tar

# Transfer to Raspberry Pi and load
docker load -i backend.tar
docker load -i frontend.tar
```

## Raspberry Pi 3 Deployment

### Build for ARMv7 (32-bit)

The Pi 3 Model B uses ARMv7 architecture. Build locally:

```bash
# Install Docker on Pi
curl -sSL https://get.docker.com | sh
sudo usermod -aG docker pi

# Or use Docker Buildx for cross-compilation
docker buildx create --name mybuilder
docker buildx use mybuilder
docker buildx build --platform linux/arm/v7 -t food-order-backend:latest .
```

### Running on Pi

```bash
# SSH to Pi
ssh pi@raspberrypi.local

# Create docker-compose.yml (see above)
# Use postgres:15-alpine (check ARM support)

# Start services
docker-compose up -d
```

## Database Backup

```bash
# Backup
docker-compose exec db pg_dump -U postgres food_order_tracking > backup.sql

# Restore
docker-compose exec -T db psql -U postgres food_order_tracking < backup.sql
```

## Security Notes

⚠️ **For Production:**
- Use strong passwords (not default)
- Store secrets in environment variables
- Enable SSL/TLS
- Use a reverse proxy (nginx/Caddy)
- Set up firewall rules

## Troubleshooting

### Check Running Containers
```bash
docker-compose ps
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f db
```

### Check Backend Health
```bash
curl http://localhost:8080/health
```

### Database Connection Issues
```bash
# Check database is ready
docker-compose exec db pg_isready -U postgres

# View database tables
docker-compose exec db psql -U postgres -d food_order_tracking -c "\dt"

# Check network connectivity
docker-compose exec backend ping db
```

### Connect to Running Containers
```bash
# Shell access to backend
docker-compose exec backend sh

# View frontend files
docker-compose exec frontend ls -la /usr/share/nginx/html

# View nginx config
docker-compose exec frontend cat /etc/nginx/nginx.conf

# Check backend environment
docker-compose exec backend env

# View backend directory
docker-compose exec backend pwd
docker-compose exec backend ls -la
```

### Port Already in Use
Edit ports in docker-compose.yml:
```yaml
ports:
  - "8081:8080"  # Change host port
```

Or find and kill the process (Windows):
```bash
netstat -ano | findstr :80
taskkill /PID <PID> /F
```

### Rebuild Everything Fresh
```bash
docker-compose down -v
docker system prune -f
docker-compose build --no-cache
docker-compose up -d
```

### View Full Container Output with Timestamps
```bash
docker-compose logs -f --timestamps
```

### "Cannot connect to Docker daemon"
- **Solution**: Start Docker Desktop and wait for it to fully load

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

### Quick Troubleshooting Reference

| Issue | Solution |
|-------|----------|
| Docker not found | Install Docker Desktop for Windows |
| Port already in use | Stop other services using ports 80, 8080, 5432 |
| Can't connect to API | Wait 15s for backend to start, check `docker-compose logs backend` |
| Database not ready | Check `docker-compose logs db`, restart with `docker-compose restart db` |
| Frontend shows error | Clear browser cache, wait for backend to fully start |
| Code changes not showing | Rebuild with `docker-compose up -d --build` |

## Service Management

### Restart Services
```bash
# Restart single service
docker-compose restart backend

# Restart all services
docker-compose restart
```

### Start/Stop Single Service
```bash
docker-compose stop backend
docker-compose start backend
```

### Rebuild After Code Changes
```bash
# Rebuild specific service
docker-compose up -d --build backend
docker-compose up -d --build frontend

# Rebuild all
docker-compose up -d --build
```

## Data Management

### Check Database Size
```bash
docker-compose exec db psql -U postgres -d food_order_tracking -c "SELECT pg_size_pretty(pg_database_size('food_order_tracking'));"
```

### Connect to Database
```bash
# Interactive SQL shell
docker-compose exec db psql -U postgres -d food_order_tracking

# Quick query
docker-compose exec db psql -U postgres -d food_order_tracking -c "SELECT * FROM customers;"
```

### Clear Database (Full Reset)
```bash
docker-compose down -v
docker-compose up -d
```

## Cleanup

### Remove Stopped Containers
```bash
docker-compose rm
```

### Remove All Docker Images
```bash
docker image prune -a
```

### Full System Cleanup
```bash
docker-compose down -v
docker system prune -a --volumes
```

## Performance Monitoring

### Check Container Resource Usage
```bash
docker stats
```

### Limit Backend Memory
Edit `docker-compose.yml`:
```yaml
backend:
  deploy:
    resources:
      limits:
        memory: 256M
```

## Useful One-Liners

```bash
# Start and watch logs
docker-compose up -d && docker-compose logs -f

# Rebuild and watch logs
docker-compose up -d --build && docker-compose logs -f

# Check everything is healthy
docker-compose exec db pg_isready -U postgres && \
  curl http://localhost:8080/health && \
  echo "All services OK"

# Database backup with timestamp
docker-compose exec db pg_dump -U postgres food_order_tracking > backup_$(date +%Y%m%d_%H%M%S).sql
```

## CI/CD

The project uses GitHub Actions for automated builds:

- **Push to main**: Builds Docker images and pushes to GHCR
- **Pull requests**: Runs tests and builds

See `.github/workflows/ci.yml` for details.

## Performance

For 5 users on a local machine:

- **Memory**: ~500MB total (~300MB DB, ~100MB backend, ~100MB frontend)
- **CPU**: < 5% at idle, < 20% under load
- **Storage**: ~2GB (mostly database capacity)
- **Startup Time**: 15-20 seconds
- **Response Time**: < 100ms typically

## Key Features

- **One Command Start** - `docker-compose up -d`
- **Automatic Networking** - Services find each other automatically
- **Database Persistence** - Data survives container restarts
- **Health Checks** - Database readiness verified before backend starts
- **Proper Logging** - All output captured for debugging
- **Easy Code Updates** - Rebuild with `docker-compose up -d --build`
- **Low Resource Usage** - ~500MB RAM total for 5 users

## Docker vs Traditional Development

### Traditional Development
```
Your Machine
├── Go backend running directly (go run)
├── Node dev server running directly (npm run dev)
└── PostgreSQL running locally
```

### Docker Development
```
Your Machine
└── Docker
    ├── PostgreSQL container
    ├── Go backend container
    └── Nginx + React container
```

**Benefits:**
- No dependency conflicts
- Easier to share environment
- Closer to production setup
- Easy to start/stop/restart

## Support Resources

- Docker docs: https://docs.docker.com
- Gin (Go framework): https://gin-gonic.com
- React: https://react.dev
- PostgreSQL: https://www.postgresql.org/docs
- Nginx: https://nginx.org/en/docs

## Success Checklist

- Docker Desktop installed and running
- Application starts with `docker-compose up -d`
- Can access http://localhost
- Can access http://localhost:8080/api/customers
- Database shows data: `docker-compose exec db psql -U postgres -d food_order_tracking -c "SELECT COUNT(*) FROM customers;"`

If all checks pass, you're ready to go!

## Getting Help

1. Check the logs: `docker-compose logs -f`
2. See if containers are running: `docker-compose ps`
3. Test if backend is responding: `curl http://localhost:8080/health`
4. Verify database is ready: `docker-compose exec db pg_isready -U postgres`

---

**For more quick commands, see `DEPLOYMENT_SUMMARY.md`**
