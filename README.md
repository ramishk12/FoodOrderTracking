# Food Order Tracking System

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-20+-2496ED?style=flat&logo=docker)

A full-stack food order tracking application with Go backend and React frontend.

## Tech Stack

- **Backend**: Go with Gin web framework
- **Database**: PostgreSQL
- **Frontend**: React with Vite
- **Styling**: Plain CSS (no UI library dependencies)

## Project Structure

```
FoodOrderTracking/
├── cmd/
│   └── main.go                  # Entry point
├── internal/
│   ├── database/
│   │   ├── database.go          # DB connection
│   │   ├── migrate.go           # Schema migrations
│   │   └── seed.go              # Sample data
│   ├── handlers/
│   │   ├── customer.go          # Customer API
│   │   ├── dashboard.go         # Dashboard & analytics API
│   │   ├── item.go              # Menu item API
│   │   ├── order.go             # Order API
│   │   ├── helpers_test.go      # Shared test helpers
│   │   ├── customer_test.go     # Customer handler tests
│   │   ├── dashboard_test.go    # Dashboard handler tests
│   │   ├── item_test.go         # Item handler tests
│   │   └── order_test.go        # Order handler tests
│   └── models/
│       └── models.go            # Data models
├── pkg/
└── web/                         # React frontend
    ├── src/
    │   ├── pages/
    │   │   ├── Home.jsx
    │   │   ├── Dashboard.jsx
    │   │   ├── Orders.jsx
    │   │   ├── OrderEdit.jsx
    │   │   ├── Customers.jsx
    │   │   ├── Items.jsx
    │   │   └── Schedule.jsx
    │   ├── services/
    │   │   └── api.js           # API client
    │   ├── App.jsx
    │   └── index.css
    └── package.json
```

## Getting Started

### Quick Start with Docker (Recommended)

**Prerequisites**: Docker Desktop for Windows

1. Start the application:
   ```bash
   cd F:\Development\FoodOrderTracking\deployment
   docker-compose up -d
   ```

2. Access the application:
   - Frontend: http://localhost
   - API: http://localhost:8080/api

3. View logs:
   ```bash
   docker-compose logs -f
   ```

For detailed instructions, see [deployment/DOCKER_DEPLOYMENT.md](deployment/DOCKER_DEPLOYMENT.md)

### Traditional Setup (Local Development)

**Prerequisites**
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+

**Setup Steps**

1. Set Up PostgreSQL
   ```sql
   CREATE DATABASE food_order_tracking;
   ```

2. Run Backend
   ```powershell
   cd F:\Development\FoodOrderTracking
   go mod tidy
   go run cmd/main.go
   ```
   Server runs on `http://localhost:8080`

   **Default Ports:**
   - Backend API: `http://localhost:8080`
   - Frontend Dev: `http://localhost:3000`

3. Run Frontend
   ```powershell
   cd F:\Development\FoodOrderTracking\web
   npm install
   npm run dev
   ```
   App runs on `http://localhost:3000`

### Running Tests

```powershell
cd F:\Development\FoodOrderTracking
go test ./internal/handlers/... -v
```

## Database Schema

### Customers Table
| Column     | Type         | Description        |
|------------|--------------|-------------------|
| id         | SERIAL       | Primary key        |
| name       | VARCHAR(255) | Customer name      |
| phone      | VARCHAR(50)  | Phone number       |
| email      | VARCHAR(255) | Email address      |
| address    | TEXT         | Home address       |
| created_at | TIMESTAMP    | Creation date      |
| updated_at | TIMESTAMP    | Last update        |

### Items Table
| Column      | Type          | Description          |
|-------------|---------------|---------------------|
| id          | SERIAL        | Primary key         |
| name        | VARCHAR(255)  | Item name           |
| description | TEXT          | Item description    |
| price       | DECIMAL(10,2) | Item price          |
| category    | VARCHAR(100)  | Item category       |
| available   | BOOLEAN       | Is available        |
| created_at  | TIMESTAMP     | Creation date       |
| updated_at  | TIMESTAMP     | Last update         |

### Orders Table
| Column           | Type         | Description              |
|------------------|--------------|-------------------------|
| id               | SERIAL       | Primary key             |
| customer_id      | INTEGER      | FK to customers         |
| delivery_address | TEXT         | Delivery address        |
| status           | VARCHAR(50)  | Order status            |
| total_amount     | DECIMAL(10,2)| Order total             |
| payment_method   | VARCHAR(50)  | cash or e-transfer      |
| scheduled_date   | TIMESTAMP    | Scheduled delivery time |
| notes            | TEXT         | Order notes             |
| created_at       | TIMESTAMP    | Creation date           |
| updated_at       | TIMESTAMP    | Last update             |

### Order Items Table
| Column      | Type          | Description           |
|-------------|---------------|----------------------|
| id          | SERIAL        | Primary key          |
| order_id    | INTEGER       | FK to orders         |
| item_id     | INTEGER       | FK to items          |
| quantity    | INTEGER       | Item quantity        |
| unit_price  | DECIMAL(10,2) | Price per unit       |
| subtotal    | DECIMAL(10,2) | Quantity × unit_price|

## API Endpoints

### Customers
| Method | Endpoint              | Description                          |
|--------|----------------------|--------------------------------------|
| GET    | /api/customers       | List all customers                   |
| GET    | /api/customers/:id   | Get customer by ID                   |
| POST   | /api/customers       | Create new customer                  |
| PUT    | /api/customers/:id   | Update customer                      |
| DELETE | /api/customers/:id   | Delete customer (blocked if has orders) |

### Orders
| Method | Endpoint                              | Description                   |
|--------|--------------------------------------|-------------------------------|
| GET    | /api/orders                          | List all orders               |
| GET    | /api/orders/:id                      | Get order by ID               |
| POST   | /api/orders                          | Create new order              |
| PUT    | /api/orders/:id                      | Update order                  |
| DELETE | /api/orders/:id                      | Delete order                  |
| GET    | /api/orders/scheduled                | Get upcoming scheduled orders |
| GET    | /api/orders/customer/:customerId     | Get orders by customer        |

### Menu Items
| Method | Endpoint                     | Description                  |
|--------|------------------------------|------------------------------|
| GET    | /api/items                   | List all menu items          |
| GET    | /api/items/:id               | Get item by ID               |
| POST   | /api/items                   | Create menu item             |
| PUT    | /api/items/:id               | Update menu item             |
| PATCH  | /api/items/:id/deactivate    | Mark item as unavailable     |
| PATCH  | /api/items/:id/activate      | Mark item as available       |

### Dashboard
| Method | Endpoint        | Description                   |
|--------|----------------|-------------------------------|
| GET    | /api/dashboard | Get dashboard statistics      |

The dashboard response includes a `warnings` array (omitted when empty) that lists any supplementary queries that failed while still returning partial data. Clients can check `response.warnings?.length` to show a degraded-data notice.

### Health Check
| Method | Endpoint | Description       |
|--------|----------|-------------------|
| GET    | /health  | Check API health  |

## Order Statuses

- `pending` — Order received, awaiting preparation
- `preparing` — Being prepared
- `ready` — Ready for pickup/delivery
- `delivered` — Completed
- `cancelled` — Cancelled

### Order Workflow

```
┌──────────┐    ┌───────────┐    ┌─────────┐    ┌────────────┐
│ PENDING  │───▶│ PREPARING │───▶│  READY  │───▶│ DELIVERED  │
└──────────┘    └───────────┘    └─────────┘    └────────────┘
      │
      │              ┌────────────┐
      └─────────────▶│ CANCELLED  │
                     └────────────┘
```

### Payment Methods

- `cash` — Cash on delivery
- `e-transfer` — Electronic transfer

## Features

### Orders Management
- **View Orders**: Orders grouped by status with collapsible sections
  - Default expanded: pending, preparing, ready
  - Default collapsed: delivered, cancelled
  - Order count badges per section
  - Expand All / Collapse All buttons
- **Create Orders**: Add new orders with customer selection and item quantities
- **Edit Orders**: Full item quantity management on a dedicated edit page; requires at least one item before saving
- **Delete Orders**: Remove orders from the system
- **Status Updates**: Inline status dropdown; preserves payment method on update
- **Search**: Find orders by customer name or delivery address
- **Filter by Status**: View orders grouped by status category
- **Filter by Payment Method**: Filter by Cash or e-Transfer

### Customers Management
- **View Customers**: List all customers with contact information and order count
- **Create / Edit / Delete**: Full CRUD with confirmation dialogs
- **Search**: Live filtering by name, phone, email, or address
- **Customer History**: View order history per customer in the Items and Order Edit pages
- **Protected Delete**: Customers with existing orders cannot be deleted

### Menu Items Management
- **View Items**: Items grouped by category with pill-filter navigation
- **Create / Edit**: Full CRUD with inline validation; whitespace trimmed from all fields
- **Availability Toggle**: Activate or deactivate individual items without deleting them
- **Sticky Order Bar**: Appears at the bottom when items are selected; clears on cancel or submit
- **Category Filters**: Filter the menu grid by category pill

### Dashboard & Analytics
- **KPI Strip**: Total revenue, monthly revenue, daily revenue, average order value, total orders, today's orders — all with animated count-up
- **Sales Trend Chart**: Smooth bezier SVG chart over the past 30 days, toggleable between revenue and orders; zero-filled for days with no activity
- **Orders by Status**: Visual breakdown of all active statuses
- **Best Selling Items**: Top 10 items by quantity ordered
- **Top Customers**: Top 5 customers by order count
- **Partial Data Warnings**: Dashboard still loads with a `warnings` field if supplementary queries fail

### Order Scheduling
- **Schedule View**: Upcoming orders grouped as Overdue / Today / Tomorrow / This Week / date-labelled far-future groups
- **Day Window**: Filter view to next 3 / 7 / 14 / 30 days
- **Timezone Handling**: All dates stored as UTC; converted to local time for display and input
- **Summary Strip**: Total, today, and overdue counts at a glance

### User Interface
- **Design System**: Playfair Display headings, IBM Plex Mono data labels, cream/amber/espresso palette
- **Responsive**: Mobile-friendly layout; navbar wraps on small screens
- **Active Nav Links**: Highlights the correct link for nested routes (e.g. `/orders/5/edit` highlights Orders)
- **Toast Notifications**: Replace all browser `alert()` calls
- **Confirm Dialogs**: Replace all browser `confirm()` calls
- **Animations**: Fade-up card entrances, slide-down forms, count-up KPI values

## Backend Architecture

### Shared Patterns Across All Handlers
- `make([]T, 0)` used for all slices to ensure `[]` not `null` in JSON responses
- `strings.TrimSpace` applied to all string inputs before validation and storage via `trimCustomer` / `trimItem` helpers
- `rows.Err()` checked after every scan loop
- `tx.Commit()` errors always returned
- `id <= 0` guard on all single-ID endpoints — zero and negative IDs return `400`
- `errors.Is(err, sql.ErrNoRows)` used to distinguish `404` from `500` on single-row queries
- `result.RowsAffected() == 0` checked after every `UPDATE` and `DELETE` — returns `404` rather than silent `200` when the record doesn't exist
- POST handlers fetch the full record after INSERT to return DB-generated timestamps in the response

### Dashboard
- Single `GROUP BY` query for the 30-day sales trend (not 30 individual queries)
- `now` passed as a parameter to `fetchSalesTrend` to guarantee the SQL `WHERE` window and the day-grid loop use the same instant
- `COALESCE(SUM(...), 0)` on all revenue aggregates including `fetchTopCustomers`
- `Warnings []string` field on `DashboardStats` populated on non-fatal supplementary query failures; omitted from JSON when empty

## Docker Deployment

Complete Docker setup for local deployment with PostgreSQL, Go backend, and React frontend.

### Deployment Files (in `deployment/` folder)
- `docker-compose.yml` — Services orchestration
- `Dockerfile.backend` — Multi-stage build for Go backend
- `Dockerfile.frontend` — Multi-stage build for React frontend with Nginx
- `nginx.conf` — Nginx configuration
- `DOCKER_DEPLOYMENT.md` — Full deployment documentation

See [deployment/DOCKER_DEPLOYMENT.md](deployment/DOCKER_DEPLOYMENT.md) for details.

## Configuration

### Environment Variables

| Variable    | Default              | Description                     |
|------------|----------------------|---------------------------------|
| DB_HOST    | localhost            | Database host address           |
| DB_PORT    | 5432                 | Database port                   |
| DB_USER    | postgres             | Database username               |
| DB_PASSWORD| postgres             | Database password               |
| DB_NAME    | food_order_tracking  | Database name                   |
| GIN_MODE   | debug                | Gin server mode (debug/release) |

## Technologies Used

### Backend
- **Go** — Programming language
- **Gin** — Web framework
- **lib/pq** — PostgreSQL driver
- **testify** — Test assertions
- **go-sqlmock** — SQL mock for unit tests

### Frontend
- **React 18** — UI library
- **Vite** — Build tool
- **React Router** — Client-side routing

### Database
- **PostgreSQL 15** — Relational database

### DevOps
- **Docker** — Containerization
- **Nginx** — Web server / reverse proxy

## Recent Updates

### Bug Fixes & Hardening (2025 — this session)

**Backend**
- All single-ID handlers now reject ID ≤ 0 with `400` instead of silently querying the DB
- `GetCustomer` and `GetItem` now return `500` for genuine DB errors instead of masking them as `404`
- `UpdateCustomer`, `UpdateItem`, `DeactivateItem`, `ActivateItem`, `DeleteCustomer`, `DeleteOrder` now return `404` when the record doesn't exist instead of silent `200`
- `CreateCustomer` and `CreateItem` now fetch the full record after INSERT to return DB-generated timestamps
- `trimCustomer` and `trimItem` helpers added — all string fields whitespace-trimmed before validation and storage
- `fetchTopCustomers` dashboard query: added `COALESCE` to `SUM(o.total_amount)` to prevent null scan errors dropping customers from the list
- `fetchSalesTrend` now receives `now` as a parameter, eliminating the clock-drift race between the SQL window and the day-grid loop

**Frontend**
- Active nav link now uses `pathname.startsWith(path)` — correctly highlights Orders on `/orders/:id/edit`
- Dashboard trend chart: null revenue/orders values coerced to `0` before SVG coordinate calculation; single data point guard added
- Dashboard KPI: `monthly_orders` now formatted with thousand separator via `fmt.num()`
- Items page: closing the order form now resets all quantities; unused `useRef` import removed
- OrderEdit: validates at least one item is selected before saving
- Schedule page: far-future date groups now sort chronologically (previously sorted by insertion order)
- Order history panels (Items, OrderEdit): items now render one per line instead of a comma-joined string
- `payment_method` included in Orders status-change payload to prevent it being wiped on update

**Tests**
- `customer_test.go`: 35 test cases — added zero/negative ID cases, split 404/500 DB error cases, timestamp assertions on create, `WithArgs` on update error mock, non-existent resource 404 cases for update and delete
- `item_test.go`: 46 test cases — full rewrite using `withMockDB`; added `TestActivateItem`, trim test cases, `WithArgs` on follow-up SELECT, zero/negative ID cases across all handlers
- `dashboard_test.go`: replaced `WithArgs(startOfMonth/startOfDay)` with `sqlmock.AnyArg()` to eliminate time-drift races; added `Warnings` assertion to partial-data test; derived `AverageOrderValue` assertion from inputs; robust zero-filled day index

### Previous Features
- **Issue #53**: Collapsible order sections by status
- **Docker Deployment**: One-command local setup
- **Issue #44**: Payment method selection and filtering
- **Issues #38, #41**: Scheduled orders with UTC timezone handling
- **Issue #43**: Customer order history panel on Items page
- **Order Edit Page**: Full quantity management
- **Dashboard**: Sales analytics with trend chart

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is for educational/personal use.

## Support

For issues or questions, please open a GitHub issue.
