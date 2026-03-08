# Agent Guidelines for FoodOrderTracking

## Project Principles

### No Duplicated Work
- Reuse existing workflows before creating new ones
- Extend existing CI/CD pipelines rather than creating parallel ones
- Consolidate related changes into single PRs when possible

## Git Workflow

### Branch Naming
- Bug fixes: `fix/issue-{number}-{description}`
- Features: `feature/{description}`
- Tests: `feature/{handler}-handler-tests`

### PRs
- Create separate PRs for each fix/feature
- Always base on `main` branch
- Run tests before pushing

## CI/CD Guidelines

### Workflows
- Use existing `.github/workflows/ci.yml` for all CI needs
- Add new jobs to existing workflow, don't create duplicates
- Use path filtering to skip unnecessary jobs

### Docker
- Reuse CI build steps for Docker images
- Push to GitHub Container Registry (GHCR)

## Code Quality

### Backend (Go)
- Run `go vet ./...` before committing
- Run `go test ./...` to verify tests pass

### Frontend (React)
- Ensure `npm run build` succeeds

## Testing
- Add unit tests for new handlers using sqlmock
- Follow existing test patterns in `internal/handlers/*_test.go`
