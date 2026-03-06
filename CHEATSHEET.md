# Docker Deployment - Cheat Sheet

Quick reference for common Docker commands.

## Essential Commands

### Start Everything
```bash
docker-compose up -d
```

### Stop Everything
```bash
docker-compose down
```

### Check Status
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

## Working with Code

### After Backend (Go) Code Changes
```bash
# Rebuild and restart
docker-compose up -d --build backend
```

### After Frontend (React) Code Changes
```bash
# Rebuild and restart
docker-compose up -d --build frontend
```

### After Both
```bash
# Rebuild and restart everything
docker-compose up -d --build
```

## Accessing Services

### Connect to Database
```bash
# Interactive SQL shell
docker-compose exec db psql -U postgres -d food_order_tracking

# Quick query
docker-compose exec db psql -U postgres -d food_order_tracking -c "SELECT * FROM customers;"
```

### Connect to Backend Container
```bash
# Get shell access
docker-compose exec backend sh
```

### View Frontend Container Files
```bash
docker-compose exec frontend ls -la /usr/share/nginx/html
```

## Debugging

### Check Backend Health
```bash
curl http://localhost:8080/health
```

### Check Database Status
```bash
docker-compose exec db pg_isready -U postgres
```

### View Database Tables
```bash
docker-compose exec db psql -U postgres -d food_order_tracking -c "\dt"
```

### Check Network
```bash
docker-compose exec backend ping db
```

## Service Management

### Restart Single Service
```bash
docker-compose restart backend
```

### Restart All Services
```bash
docker-compose restart
```

### Stop Single Service
```bash
docker-compose stop backend
```

### Start Single Service
```bash
docker-compose start backend
```

## Data Management

### Backup Database
```bash
docker-compose exec db pg_dump -U postgres food_order_tracking > backup.sql
```

### Restore Database
```bash
docker-compose exec -T db psql -U postgres food_order_tracking < backup.sql
```

### Check Database Size
```bash
docker-compose exec db psql -U postgres -d food_order_tracking -c "SELECT pg_size_pretty(pg_database_size('food_order_tracking'));"
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

## Performance

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

## Viewing Files

### List React Build Files
```bash
docker-compose exec frontend ls -la /usr/share/nginx/html
```

### View Nginx Config
```bash
docker-compose exec frontend cat /etc/nginx/nginx.conf
```

### View Backend Current Directory
```bash
docker-compose exec backend pwd
docker-compose exec backend ls -la
```

## Environment Variables

### Check Running Environment
```bash
docker-compose exec backend env
```

### Pass Custom Environment Variables
Edit `docker-compose.yml` or create `.env` file:
```
DB_PASSWORD=custom_password
GIN_MODE=release
```

## Troubleshooting

### Port Already in Use
```bash
# Find what's using port 80 (Windows)
netstat -ano | findstr :80

# Kill the process
taskkill /PID <PID> /F
```

### Rebuild Everything Fresh
```bash
docker-compose down -v
docker system prune -f
docker-compose build --no-cache
docker-compose up -d
```

### View Full Container Output
```bash
docker-compose logs -f --timestamps
```

### Force Restart
```bash
docker-compose restart --timeout 5
```

## One-Liners

```bash
# Start and watch logs
docker-compose up -d && docker-compose logs -f

# Rebuild and watch logs
docker-compose up -d --build && docker-compose logs -f

# Stop and clean
docker-compose down -v

# Check everything is healthy
docker-compose exec db pg_isready -U postgres && \
  curl http://localhost:8080/health && \
  echo "All services OK"

# Database backup with timestamp
docker-compose exec db pg_dump -U postgres food_order_tracking > backup_$(date +%Y%m%d_%H%M%S).sql
```

## Quick Navigation

| Need... | Command |
|---------|---------|
| Start app | `docker-compose up -d` |
| Stop app | `docker-compose down` |
| View logs | `docker-compose logs -f` |
| Check status | `docker-compose ps` |
| Database shell | `docker-compose exec db psql -U postgres -d food_order_tracking` |
| Backend shell | `docker-compose exec backend sh` |
| View all containers | `docker ps -a` |
| View all images | `docker images` |
| Rebuild app | `docker-compose up -d --build` |
| Full reset | `docker-compose down -v` |

---

💡 **Tip**: Most commands ending with `-f` mean "follow" (live output). Press Ctrl+C to exit.
