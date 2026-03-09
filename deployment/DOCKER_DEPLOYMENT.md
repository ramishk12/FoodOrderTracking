# Docker Deployment Guide

## Overview

This guide explains how to deploy the Food Order Tracking application using Docker.

## Project Structure

```
FoodOrderTracking/
├── deployment/
│   ├── docker-compose.yml    # Local development
│   └── DOCKER_DEPLOYMENT.md # This file
├── Dockerfile.backend       # Go backend Docker image
├── Dockerfile.frontend     # React frontend Docker image
├── nginx.conf             # Nginx configuration
├── .dockerignore          # Build optimization
├── cmd/                   # Backend entry point
├── internal/              # Backend source code
├── web/                   # Frontend source code
├── .github/workflows/     # CI/CD pipelines
└── AGENTS.md            # Development guidelines
```

## Quick Start (Local Development)

### Prerequisites
- Docker Desktop (Windows/Mac/Linux)
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

## Production Deployment

### Option 1: GitHub Container Registry (Recommended)

The CI workflow automatically builds and pushes Docker images to GHCR on push to main:

```
ghcr.io/ramishk12/food-order-tracking/backend:latest
ghcr.io/ramishk12/food-order-tracking/frontend:latest
```

#### Pull on Production Server
```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_ACTOR --password-stdin

# Pull images
docker pull ghcr.io/ramishk12/food-order-tracking/backend:latest
docker pull ghcr.io/ramishk12/food-order-tracking/frontend:latest
```

#### Run with Docker Compose
Create `docker-compose.yml` on your server:

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
# Build images
docker build -t food-order-backend:latest .
docker build -t food-order-frontend:latest -f Dockerfile.frontend .

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
docker-compose logs -f [service name]
```

### Database Connection Issues
```bash
# Check database is ready
docker-compose exec db pg_isready -U postgres
```

### Port Already in Use
Edit ports in docker-compose.yml:
```yaml
ports:
  - "8081:8080"  # Change host port
```

## CI/CD

The project uses GitHub Actions for automated builds:

- **Push to main**: Builds Docker images and pushes to GHCR
- **Pull requests**: Runs tests and builds

See `.github/workflows/ci.yml` for details.
