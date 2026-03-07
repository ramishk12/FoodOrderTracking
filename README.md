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

### Orders Management
- **View Orders**: Display all orders in a responsive card grid layout
- **Create Orders**: Add new orders with customer selection and items
- **Edit Orders**: Modify order details and item quantities on dedicated edit page
- **Delete Orders**: Remove orders from the system
- **Update Status**: Change order status (pending → preparing → ready → delivered/cancelled)
- **Collapsible Sections** ⭐ NEW: Orders grouped by status with expand/collapse functionality
  - Default expanded: pending, preparing, ready
  - Default collapsed: delivered, cancelled
  - Order count badges per section
  - Expand All / Collapse All buttons
- **Search**: Find orders by customer name or delivery address
- **Filter by Status**: View orders in specific status categories
- **Filter by Payment Method**: Filter orders by Cash or e-Transfer

### Customers Management
- **View Customers**: List all customers with contact information
- **Create Customers**: Add new customers to the database
- **Edit Customers**: Update customer information
- **Delete Customers**: Remove customers from the system
- **Customer History**: View order history for each customer on the Items page

### Menu Items Management
- **View Menu Items**: Display all menu items organized by category
- **Create Items**: Add new menu items with price, description, and category
- **Edit Items**: Update item details
- **Delete Items**: Remove items from the menu
- **Availability Control**: Toggle item availability by category
- **Category Organization**: Items grouped by category (starters, mains, desserts, drinks, etc.)

### Dashboard & Analytics
- **Dashboard Page**: View key metrics and statistics
  - Total orders count
  - Orders by status breakdown
  - Top menu items by quantity ordered
  - Top customers by total spent
  - Sales chart by day
- **Real-time Statistics**: Auto-refreshing dashboard data

### Order Scheduling
- **Scheduled Orders**: Schedule orders for future delivery
- **Schedule View**: View upcoming orders by date
- **Timezone Support**: Proper UTC timezone handling with timezone conversion
- **Schedule Filtering**: Filter orders by scheduled date

### Payment Methods
- **Cash Payment**: Option for cash on delivery
- **e-Transfer Payment**: Option for e-Transfer payment
- **Payment Filtering**: Filter and track orders by payment method
- **Payment Display**: Show payment method on all order views

### User Interface
- **Responsive Design**: Mobile-friendly layout
- **Navigation**: Easy navigation between all sections
- **Search & Filter**: Quick access to find specific items
- **Status Badges**: Color-coded status indicators
  - Pending: Orange
  - Preparing: Blue
  - Ready: Purple
  - Delivered: Green
  - Cancelled: Red
- **Card Grid Layout**: Clean, organized card-based presentation
- **Forms & Modals**: Easy-to-use forms for data entry
- **Real-time Updates**: Instant reflection of changes across pages

### Technical Features
- **RESTful API**: Clean REST API endpoints for all operations
- **Error Handling**: Comprehensive error messages and handling
- **Data Validation**: Input validation on both frontend and backend
- **Automatic Timestamps**: Track creation and update times for all records
- **Database Triggers**: Automatic updated_at timestamp management
- **UTC Timezone Handling**: Consistent timezone management throughout the application

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
