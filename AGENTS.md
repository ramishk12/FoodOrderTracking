# Agent Guidelines for FoodOrderTracking

## Project Overview
Full-stack food order tracking application with Go backend (Gin framework) and React frontend (Vite).

## Build, Lint, and Test Commands

### Backend (Go)
```bash
# Run all tests
go test ./...

# Run a single test (use -run with regex pattern)
go test ./internal/handlers/... -run "TestGetCustomer"

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Lint/vet code
go vet ./...

# Format code
go fmt ./...

# Build the application
go build ./...
```

### Frontend (React)
```bash
cd web

# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Docker
```bash
# Start PostgreSQL database
docker run -d --name food-order-db \
  -e POSTGRES_DB=food_order_tracking \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 postgres:15-alpine

# Run SQL queries
docker exec -it food-order-db psql -U postgres -d food_order_tracking -c "SELECT * FROM customers;"
```

## Git Workflow

### Branch Naming
- Bug fixes: `fix/issue-{number}-{description}` (e.g., `fix/customer-handler-bugs`)
- Features: `feature/{description}` (e.g., `feature/add-modifiers`)
- Test improvements: `test/{handler}-handler-tests-update` (e.g., `test/customer-handler-tests`)
- Enhancements: `enhancement/{description}`

### Creating PRs
1. Always base on `main` branch
2. Create separate PRs for each fix/feature
3. Run all tests before pushing
4. Link related GitHub issues in PR description
5. Use "Closes #123" in PR body to auto-close issues

## Code Style Guidelines

### Backend (Go)

#### Imports
- Use standard library first, then third-party
- Group: standard library → external packages → internal packages
- Example:
  ```go
  import (
      "database/sql"
      "fmt"
      "net/http"
      "time"

      "github.com/gin-gonic/gin"
      "github.com/DATA-DOG/go-sqlmock"

      "food-order-tracking/internal/database"
      "food-order-tracking/internal/models"
  )
  ```

#### Naming Conventions
- **Functions/Variables**: camelCase (e.g., `getCustomer`, `totalAmount`)
- **Constants**: PascalCase (e.g., `MaxResults`, `DefaultTimeout`)
- **Types/Interfaces**: PascalCase (e.g., `Customer`, `OrderService`)
- **Package names**: short, lowercase, no underscores (e.g., `handlers`, `models`)
- **Database tables**: snake_case (e.g., `order_items`, `customer_id`)

#### Error Handling
- Return errors with context, don't log and continue
- Use helper functions for common validations
- Distinguish between error types:
  ```go
  // For Not Found - use sql.ErrNoRows check
  if errors.Is(err, sql.ErrNoRows) {
      c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
      return
  }
  // For other errors
  if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
  }
  ```
- Validate input IDs with zero/negative checks:
  ```go
  id, err := strconv.Atoi(c.Param("id"))
  if err != nil || id <= 0 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
      return
  }
  ```

#### Database Patterns
- Use `sqlmock` for handler tests
- Use `withMockDB` helper function consistently
- Use `sqlmock.AnyArg()` for time-dependent queries to avoid race conditions
- Return full records after INSERT to get DB-generated timestamps
- Use `COALESCE` for SUM aggregations to handle NULL:
  ```sql
  COALESCE(SUM(total_amount), 0)
  ```
- Check `RowsAffected()` after UPDATE/DELETE for 404 responses

#### Handler Structure
```go
// Handler function signature
func GetCustomer(c *gin.Context) {
    // 1. Parse and validate input
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil || id <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
        return
    }

    // 2. Query database
    row := database.DB.QueryRow(customerQuery+" WHERE id = $1", id)
    customer, err := scanCustomer(row)

    // 3. Handle errors with proper status codes
    if errors.Is(err, sql.ErrNoRows) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 4. Return success response
    c.JSON(http.StatusOK, customer)
}
```

### Frontend (React)

#### File Organization
- Pages in `web/src/pages/`
- Components in `web/src/components/`
- Services/API in `web/src/services/`
- Styles in `web/src/index.css` (centralized CSS)

#### Naming
- Components: PascalCase (e.g., `OrderEdit.jsx`)
- Hooks: camelCase starting with use (e.g., `useOrders`)
- CSS classes: lowercase with hyphens (e.g., `.order-card`)

#### State Management
- Use React hooks (`useState`, `useEffect`)
- Use `api.js` service for HTTP calls

## Testing Guidelines

### Handler Tests
- Use `sqlmock` for database mocking
- Use `withMockDB` helper from `test_helpers_test.go`
- Test both success and error cases
- Include zero/negative ID validation tests
- Include 404 for not found scenarios
- Include 500 for database errors
- Use `ExpectRollback()` when testing error paths in transactions

### Test Naming
```
Test{HandlerName}/{Description}
TestGetCustomer/Returns_customer_by_ID
TestGetCustomer/Returns_404_when_not_found
```

### Test Structure
```go
func TestGetCustomer(t *testing.T) {
    tests := []struct {
        name           string
        setupMock      func(sqlmock.Sqlmock)
        expectedStatus int
        checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
    }{
        {
            name: "Returns customer by ID",
            setupMock: func(m sqlmock.Sqlmock) {
                m.ExpectQuery(...).WillReturnRows(...)
            },
            expectedStatus: http.StatusOK,
            checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
                var customer models.Customer
                json.Unmarshal(w.Body.Bytes(), &customer)
                assert.Equal(t, 1, customer.ID)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            withMockDB(t, func(mock sqlmock.Sqlmock) {
                tt.setupMock(mock)
                w := httptest.NewRecorder()
                c, _ := gin.CreateTestContext(w)
                c.Params = gin.Params{{Key: "id", Value: "1"}}
                GetCustomer(c)

                assert.Equal(t, tt.expectedStatus, w.Code)
                mock.ExpectationsWereMet()
            })
        })
    }
}
```

## Database

### Connecting via Docker
```bash
# Start database
docker run -d --name food-order-db \
  -e POSTGRES_DB=food_order_tracking \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 postgres:15-alpine

# Run queries interactively
docker exec -it food-order-db psql -U postgres -d food_order_tracking
```

### Migrations
- Database migrations in `internal/database/migrate.go`
- Seed data in `internal/database/seed.go`

## Common Patterns

### Creating Issues
- Label as `bug` for fixes
- Label as `enhancement` for new features
- Include code examples showing the issue
- Reference similar implementations if available

### Updating Dependencies
- Go: `go get -u ./...` then `go mod tidy`
- npm: `npm update` in web directory
