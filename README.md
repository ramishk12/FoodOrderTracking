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
│   │   └── order.go        # Order API
│   └── models/
│       └── models.go        # Data models
├── pkg/
└── web/                    # React frontend
    ├── src/
    │   ├── pages/
    │   │   ├── Home.jsx
    │   │   ├── Orders.jsx
    │   │   └── Customers.jsx
    │   ├── services/
    │   │   └── api.js      # API client
    │   ├── App.jsx
    │   └── index.css
    └── package.json
```

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+

### 1. Set Up PostgreSQL

```sql
-- Create database
CREATE DATABASE food_order_tracking;
```

### 2. Run Backend

```powershell
cd F:\Development\FoodOrderTracking
go mod tidy
go run cmd/main.go
```

The server runs on `http://localhost:8080`

### 3. Run Frontend

```powershell
cd F:\Development\FoodOrderTracking\web
npm install
npm run dev
```

The app runs on `http://localhost:3000`

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

### Orders Table
| Column           | Type         | Description           |
|------------------|--------------|----------------------|
| id               | SERIAL       | Primary key          |
| customer_id      | INTEGER      | FK to customers     |
| delivery_address | TEXT         | Delivery address     |
| status           | VARCHAR(50)  | Order status        |
| total_amount     | DECIMAL(10,2)| Order total         |
| items            | TEXT         | Ordered items        |
| notes            | TEXT         | Order notes          |
| created_at       | TIMESTAMP    | Creation date        |
| updated_at       | TIMESTAMP    | Last update          |

## API Endpoints

### Customers
| Method | Endpoint      | Description        |
|--------|---------------|-------------------|
| GET    | /api/customers      | List all customers |
| GET    | /api/customers/:id  | Get customer      |
| POST   | /api/customers      | Create customer   |
| PUT    | /api/customers/:id  | Update customer   |
| DELETE | /api/customers/:id  | Delete customer   |

### Orders
| Method | Endpoint      | Description        |
|--------|---------------|-------------------|
| GET    | /api/orders         | List all orders   |
| GET    | /api/orders/:id     | Get order         |
| POST   | /api/orders         | Create order      |
| PUT    | /api/orders/:id     | Update order      |
| DELETE | /api/orders/:id     | Delete order      |

## Order Statuses

- `pending` - Order received
- `preparing` - Being prepared
- `ready` - Ready for pickup/delivery
- `delivered` - Completed
- `cancelled` - Cancelled

## Features

- **Orders Page**: View, create, edit, delete orders; update status
- **Customers Page**: Manage customer database
- **Home Page**: Navigation to Orders and Customers
