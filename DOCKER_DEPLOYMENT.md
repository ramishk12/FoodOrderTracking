# Docker Deployment Guide

## Overview

This guide explains how to deploy the Food Order Tracking application using Docker on your local Windows machine.

## Prerequisites

### Required Software

1. **Docker Desktop for Windows**
   - Download from: https://www.docker.com/products/docker-desktop
   - Install and run
   - Ensure it's running before deploying

2. **Git** (if not already installed)
   - Download from: https://git-scm.com/download/win

### System Requirements

- Windows 10 Pro/Enterprise or Windows 11
- 4GB+ RAM (8GB recommended)
- 10GB free disk space

## Quick Start (5 minutes)

### Step 1: Clone/Navigate to Project
```bash
cd F:\Development\FoodOrderTracking
```

### Step 2: Start All Services
```bash
docker-compose up -d
```

This command will:
- Pull/build all required images
- Start PostgreSQL database (port 5432)
- Start Go backend API (port 8080)
- Start React frontend with Nginx (port 80)

### Step 3: Access the Application

- **Frontend**: http://localhost
- **Backend API**: http://localhost:8080/api
- **Database**: localhost:5432 (postgres/postgres)

### Step 4: View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f db
```

### Step 5: Stop All Services
```bash
docker-compose down
```

## Common Commands

### View Running Containers
```bash
docker-compose ps
```

### Restart a Service
```bash
docker-compose restart backend
```

### View Full Logs with Timestamps
```bash
docker-compose logs --timestamps -f
```

### Remove Everything (including database)
```bash
docker-compose down -v
```
**Warning**: This deletes the database volume. Use only if you want to start fresh.

### Rebuild After Code Changes
```bash
# Rebuild all images
docker-compose build

# Rebuild and restart
docker-compose up -d --build
```

## Architecture

```
┌─────────────────────────────────────────────┐
│         Docker Network (bridge)             │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────────────────────────────────┐   │
│  │  Nginx (Frontend)                    │   │
│  │  :80 -> Serves React SPA             │   │
│  │  Proxies /api/* to backend:8080      │   │
│  └──────────────────────────────────────┘   │
│                                             │
│  ┌──────────────────────────────────────┐   │
│  │  Go Backend (Gin)                    │   │
│  │  :8080 -> REST API                   │   │
│  └──────────────────────────────────────┘   │
│                                             │
│  ┌──────────────────────────────────────┐   │
│  │  PostgreSQL Database                 │   │
│  │  :5432 -> food_order_tracking DB     │   │
│  └──────────────────────────────────────┘   │
│                                             │
└─────────────────────────────────────────────┘
```

## Services Details

### Database (PostgreSQL)
- **Container Name**: food-order-db
- **Port**: 5432
- **Database**: food_order_tracking
- **Username**: postgres
- **Password**: postgres
- **Volume**: postgres_data (persists across restarts)

### Backend (Go + Gin)
- **Container Name**: food-order-backend
- **Port**: 8080
- **Health Check**: GET http://localhost:8080/health
- **API Base**: http://localhost:8080/api
- **Depends On**: PostgreSQL

### Frontend (React + Nginx)
- **Container Name**: food-order-frontend
- **Port**: 80
- **Build**: Multi-stage (builds React, serves with Nginx)
- **Depends On**: Backend

## Troubleshooting

### Cannot Connect to Docker
```bash
# Ensure Docker Desktop is running
# Check status:
docker ps
```

### Port Already in Use

If you see "bind: address already in use", another service is using the port:

```bash
# Use different ports in docker-compose.yml:
# Change:
#   ports:
#     - "8080:8080"
# To:
#   ports:
#     - "8081:8080"
```

### Database Connection Failing

1. Check database is running:
```bash
docker-compose ps
```

2. Check database logs:
```bash
docker-compose logs db
```

3. Verify database is ready:
```bash
docker-compose exec db pg_isready -U postgres
```

### Frontend Shows "Cannot Find API"

1. Ensure backend is running:
```bash
docker-compose logs backend
```

2. Check Nginx configuration:
```bash
docker-compose exec frontend cat /etc/nginx/nginx.conf
```

3. Test backend directly:
```bash
# From your host machine (Windows)
curl http://localhost:8080/health
```

### Build Fails

Clear Docker cache and rebuild:
```bash
docker-compose down -v
docker system prune -f
docker-compose build --no-cache
docker-compose up -d
```

## Performance Optimization

### Reduce Memory Usage

Edit `docker-compose.yml` to limit resources:

```yaml
backend:
  # ... existing config ...
  deploy:
    resources:
      limits:
        cpus: '0.5'
        memory: 256M
      reservations:
        cpus: '0.25'
        memory: 128M
```

### Database Backup

```bash
# Backup database
docker-compose exec db pg_dump -U postgres food_order_tracking > backup.sql

# Restore from backup
docker-compose exec -T db psql -U postgres food_order_tracking < backup.sql
```

## Development Tips

### Run in Foreground (for debugging)
```bash
# Useful during development to see all logs
docker-compose up
```

### Connect to Running Container Shell
```bash
# Backend
docker-compose exec backend sh

# Database (psql)
docker-compose exec db psql -U postgres -d food_order_tracking
```

### Update Frontend Code (without rebuild)

For rapid development, you can mount the source directly:

Edit `docker-compose.yml` backend service:
```yaml
volumes:
  - ./web/src:/app/src
```

This allows hot-reload during development.

### Check Database Schema
```bash
docker-compose exec db psql -U postgres -d food_order_tracking -c "\dt"
```

## File Structure

```
FoodOrderTracking/
├── Dockerfile.backend        # Go backend build instructions
├── Dockerfile.frontend       # React frontend build instructions
├── nginx.conf               # Nginx configuration
├── docker-compose.yml       # Orchestration config
├── .dockerignore             # Files to exclude from build context
├── DOCKER_DEPLOYMENT.md     # This file
├── cmd/
│   └── main.go              # Backend entry point
├── internal/
│   ├── database/
│   ├── handlers/
│   └── models/
├── web/
│   ├── src/                 # React source
│   ├── index.html
│   ├── vite.config.js
│   └── package.json
├── go.mod
└── go.sum
```

## Security Considerations

⚠️ **For Local Development Only**

The current setup uses:
- Plain text database passwords
- No SSL/TLS encryption
- Default credentials

**For Production**, you should:
1. Use strong, random passwords
2. Store secrets in environment files (not in repo)
3. Enable SSL/TLS for all services
4. Use a reverse proxy (Traefik, nginx)
5. Implement proper authentication
6. Set up monitoring and logging

Example secure setup with `.env` file:

```bash
# Create .env file (add to .gitignore)
DB_PASSWORD=your_secure_password_here
GIN_MODE=release
```

Then in `docker-compose.yml`:
```yaml
environment:
  POSTGRES_PASSWORD: ${DB_PASSWORD}
  DB_PASSWORD: ${DB_PASSWORD}
```

Run with:
```bash
docker-compose --env-file .env up -d
```

## Next Steps

### For Local Production (on Windows):

1. **Create a `.env` file** with production values
2. **Update docker-compose.yml** to use `.env`
3. **Set up automatic restarts** (handled by `restart: unless-stopped`)
4. **Configure backup strategy** (daily database backups)
5. **Monitor logs** (optional: set up log aggregation)

### For Remote Hosting:

1. Use Docker on a Linux VPS
2. Add Traefik for SSL/TLS and routing
3. Use Docker Swarm or Kubernetes for orchestration
4. Set up CI/CD pipeline for automatic deployments

## Support

For issues or questions:

1. Check Docker Desktop logs
2. Review container logs: `docker-compose logs -f`
3. Verify all services are healthy: `docker-compose ps`
4. Check firewall rules for port access

## File Changes Summary

New files created for Docker deployment:
- `Dockerfile.backend` - Multi-stage build for Go
- `Dockerfile.frontend` - Multi-stage build for React with Nginx
- `docker-compose.yml` - Service orchestration
- `nginx.conf` - Nginx configuration for frontend
- `.dockerignore` - Build optimization
- `DOCKER_DEPLOYMENT.md` - This guide
