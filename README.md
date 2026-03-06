# Food Order Tracking System

A full-stack food order tracking application with Go backend and React frontend.

## Tech Stack

- **Backend**: Go with Gin web framework
- **Database**: PostgreSQL
- **Frontend**: React with Vite
- **Styling**: Plain CSS

## Project Structure

```
FoodOrderTracking/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── database/
│   │   ├── database.go      # DB connection
│   │   ├── migrate.go       # Schema migrations
│   │   └── seed.go          # Sample data
│   ├── handlers/
│   │   ├── customer.go      # Customer API
│   │   ├── order.go         # Order API
│   │   └── item.go          # Menu Item API
│   └── models/
│       └── models.go        # Data models
├── pkg/
└── web/                    # React frontend
    ├── src/
    │   ├── pages/
    │   │   ├── Home.jsx
    │   │   ├── Orders.jsx
    │   │   ├── OrderEdit.jsx
    │   │   ├── Customers.jsx
    │   │   └── Items.jsx
    │   ├── services/
    │   │   └── api.js       # API client
    │   ├── App.jsx
    │   └── index.css
    └── package.json
```

## Getting Started

### Quick Start with Docker (Recommended - 5 minutes)

**Prerequisites**: Docker Desktop for Windows

1. Start the application:
   ```bash
   cd F:\Development\FoodOrderTracking
   docker-compose up -d
   ```

2. Access the application:
   - Frontend: http://localhost
   - API: http://localhost:8080/api

3. View logs:
   ```bash
   docker-compose logs -f
   ```

For detailed instructions, see [QUICK_START.md](QUICK_START.md)

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

3. Run Frontend
   ```powershell
   cd F:\Development\FoodOrderTracking\web
   npm install
   npm run dev
   ```
   App runs on `http://localhost:3000`

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
| Column           | Type         | Description           |
|------------------|--------------|----------------------|
| id               | SERIAL       | Primary key          |
| customer_id      | INTEGER      | FK to customers     |
| delivery_address | TEXT         | Delivery address     |
| status           | VARCHAR(50)  | Order status        |
| total_amount     | DECIMAL(10,2)| Order total         |
| notes            | TEXT         | Order notes          |
| created_at       | TIMESTAMP    | Creation date        |
| updated_at       | TIMESTAMP    | Last update          |

### Order Items Table
| Column      | Type          | Description          |
|-------------|---------------|---------------------|
| id          | SERIAL        | Primary key         |
| order_id    | INTEGER       | FK to orders        |
| item_id     | INTEGER       | FK to items         |
| quantity    | INTEGER       | Item quantity       |
| unit_price  | DECIMAL(10,2) | Price per unit      |
| subtotal    | DECIMAL(10,2) | Quantity * unit_price|

## API Endpoints

### Customers
| Method | Endpoint         | Description        |
|--------|-----------------|-------------------|
| GET    | /api/customers | List all customers |
| GET    | /api/customers/:id | Get customer   |
| POST   | /api/customers | Create customer    |
| PUT    | /api/customers/:id | Update customer |
| DELETE | /api/customers/:id | Delete customer |

### Orders
| Method | Endpoint      | Description        |
|--------|--------------|-------------------|
| GET    | /api/orders  | List all orders   |
| GET    | /api/orders/:id | Get order      |
| POST   | /api/orders  | Create order      |
| PUT    | /api/orders/:id | Update order    |
| DELETE | /api/orders/:id | Delete order    |

### Items
| Method | Endpoint   | Description          |
|--------|-----------|---------------------|
| GET    | /api/items | List all menu items |
| GET    | /api/items/:id | Get item         |
| POST   | /api/items | Create menu item    |
| PUT    | /api/items/:id | Update menu item |
| DELETE | /api/items/:id | Delete menu item |

## Order Statuses

- `pending` - Order received
- `preparing` - Being prepared
- `ready` - Ready for pickup/delivery
- `delivered` - Completed
- `cancelled` - Cancelled

## Features

- **Home Page**: Navigation to all sections
- **Orders Page**: View, create, edit, delete orders; update status; filter by status
- **Order Edit Page**: Edit order details and items with full quantity management
- **Customers Page**: Manage customer database; add, edit, delete customers
- **Menu Items Page**: Manage menu items; add, edit, delete items; set availability by category
- **Search/Filter**: Search orders by customer name; filter orders by status

## Docker Deployment

Complete Docker setup for local deployment with PostgreSQL, Go backend, and React frontend.

### Files Included
- `Dockerfile.backend` - Multi-stage build for Go backend
- `Dockerfile.frontend` - Multi-stage build for React frontend with Nginx
- `docker-compose.yml` - Services orchestration
- `nginx.conf` - Nginx configuration
- `QUICK_START.md` - Quick start guide
- `DOCKER_DEPLOYMENT.md` - Full deployment documentation

See [QUICK_START.md](QUICK_START.md) for details.

## Recent Updates

- Added collapsible sections by order status on Orders page (Issue #53)
- Implemented Docker deployment for easy local setup
- Added menu items management with categories and availability
- Added order edit page with item quantity management
- Added search and filter functionality on orders page
- Updated database to use order_items table instead of items text field
- Added automatic updated_at timestamp tracking via database triggers
