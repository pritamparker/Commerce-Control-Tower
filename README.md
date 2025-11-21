# Ecommerce Discount Store

A full-stack ecommerce application featuring a Go backend API and React frontend. The system implements an "every *n*th order gets 10% off" promotion system with cart management, checkout, and admin dashboard capabilities.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Backend (Go)](#backend-go)
  - [Setup](#backend-setup)
  - [API Documentation](#api-documentation)
  - [Testing](#backend-testing)
  - [Configuration](#backend-configuration)
- [Frontend (React)](#frontend-react)
  - [Setup](#frontend-setup)
  - [Features](#frontend-features)
  - [Testing](#frontend-testing)
  - [Build & Deployment](#frontend-build--deployment)
- [Development Workflow](#development-workflow)
- [API Examples](#api-examples)
- [Next Steps](#next-steps)

## Features

### Backend Features

- **Cart Management**: Add, update, and view items in per-user shopping carts
- **Checkout System**: Process orders with optional discount code validation
- **Discount Engine**: Automatic discount code generation every *n*th order (configurable)
- **Admin Dashboard**: Generate discount codes and view real-time statistics
- **In-Memory Store**: Fast, lightweight state management (suitable for development/demos)
- **RESTful API**: Clean, JSON-based API with proper error handling

### Frontend Features

- **Cart Playground**: Interactive interface to add items to any user's cart
- **Checkout Lab**: Complete checkout flow with discount code support
- **Real-time Cart View**: Live cart snapshot with item details and totals
- **Admin Insights**: Dashboard showing orders, revenue, discounts, and code history
- **API Console**: Raw API response viewer for debugging
- **Modern UI**: Built with Tailwind CSS and shadcn/ui components
- **Responsive Design**: Works on desktop and mobile devices

## Architecture

```
┌─────────────────┐
│  React Frontend │  (Port 3000)
│  (Tailwind UI)  │
└────────┬────────┘
         │ HTTP Proxy
         │ /api/*
         ▼
┌─────────────────┐
│   Go Backend    │  (Port 8080)
│  (Chi Router)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Memory Store    │
│  (In-Memory)    │
└─────────────────┘
```

The frontend is a React application that communicates with the Go backend via REST API. The backend uses an in-memory store for state management, making it perfect for development and demonstrations.

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── http/
│   │   └── server.go            # HTTP server and route handlers
│   └── store/
│       ├── store.go              # Business logic and in-memory store
│       └── store_test.go        # Backend unit tests
├── frontend/
│   ├── src/
│   │   ├── App.js                # Main React component
│   │   ├── App.test.js          # Frontend tests
│   │   ├── components/
│   │   │   └── ui/              # Reusable UI components (shadcn)
│   │   └── lib/
│   │       └── utils.js         # Utility functions
│   ├── public/                  # Static assets
│   ├── build/                   # Production build output
│   └── package.json
├── go.mod                        # Go dependencies
├── go.sum                        # Go dependency checksums
└── README.md
```

## Backend (Go)

### Backend Setup

**Prerequisites:**

- Go 1.25.4 or later
- Git

**Installation:**

```bash
# Clone the repository
cd /path/to/project

# Install Go dependencies
go mod tidy

# Run tests
go test ./...

# Start the server
go run ./cmd/server
```

The server will start on `http://localhost:8080` by default.

### Backend Configuration

Environment variables:

| Variable             | Default | Description                           |
| -------------------- | ------- | ------------------------------------- |
| `PORT`               | `8080`  | HTTP server port                      |
| `NTH_ORDER_DISCOUNT` | `3`     | Generate discount code every N orders |

Example:

```bash
PORT=3001 NTH_ORDER_DISCOUNT=5 go run ./cmd/server
```

### API Documentation

#### Base URL

```
http://localhost:8080/api
```

#### Endpoints

##### 1. Add Item to Cart

```http
POST /api/cart/{userID}/items
Content-Type: application/json

{
  "sku": "SKU1",
  "name": "Widget",
  "price": 49.99,
  "quantity": 2
}
```

**Response:** `201 Created`

```json
{
  "items": {
    "SKU1": {
      "sku": "SKU1",
      "name": "Widget",
      "price": 49.99,
      "quantity": 2
    }
  }
}
```

**Errors:**

- `400 Bad Request`: Invalid payload (missing sku, price, or quantity)

---

##### 2. View Cart

```http
GET /api/cart/{userID}/items
```

**Response:** `200 OK`

```json
{
  "items": {
    "SKU1": {
      "sku": "SKU1",
      "name": "Widget",
      "price": 49.99,
      "quantity": 2
    }
  }
}
```

---

##### 3. Checkout

```http
POST /api/cart/{userID}/checkout
Content-Type: application/json

{
  "discountCode": "DISC-ABC123"
}
```

**Response:** `200 OK`

```json
{
  "id": "ord_1234567890",
  "userId": "user-1",
  "items": [
    {
      "sku": "SKU1",
      "name": "Widget",
      "price": 49.99,
      "quantity": 2
    }
  ],
  "totalAmount": 89.98,
  "discountCode": "DISC-ABC123",
  "discountValue": 10.0,
  "createdAt": "2024-01-15T10:30:00Z"
}
```

**Errors:**

- `400 Bad Request`: Cart is empty
- `422 Unprocessable Entity`: Invalid discount code (not active, already used, or mismatch)

---

##### 4. Generate Discount Code

```http
POST /api/admin/discounts/generate
```

**Response:** `201 Created`

```json
{
  "code": "DISC-XYZ789",
  "percentage": 10,
  "generatedAt": "2024-01-15T10:30:00Z",
  "eligibleOrderNum": 3,
  "isRedeemed": false
}
```

**Errors:**

- `409 Conflict`: Not eligible to generate discount (not enough orders or active discount exists)

---

##### 5. Get Admin Stats

```http
GET /api/admin/stats
```

**Response:** `200 OK`

```json
{
  "totalOrders": 10,
  "totalItemsSold": 25,
  "grossRevenue": 500.5,
  "totalDiscountGiven": 50.25,
  "discountCodes": [
    {
      "code": "DISC-ABC123",
      "percentage": 10,
      "generatedAt": "2024-01-15T10:00:00Z",
      "redeemedAt": "2024-01-15T10:05:00Z",
      "isRedeemed": true,
      "eligibleOrderNum": 3
    }
  ],
  "activeDiscount": {
    "code": "DISC-XYZ789",
    "percentage": 10,
    "generatedAt": "2024-01-15T10:30:00Z",
    "isRedeemed": false,
    "eligibleOrderNum": 6
  }
}
```

---

#### Error Response Format

All errors follow this format:

```json
{
  "error": "error message here"
}
```

### Backend Testing

Run all backend tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests for a specific package:

```bash
go test ./internal/store
```

Run tests with coverage:

```bash
go test -cover ./...
```

**Test Coverage:**

- `TestAddItemAndCheckout`: Tests adding items and checkout flow
- `TestDiscountLifecycle`: Tests discount generation and redemption

## Frontend (React)

### Frontend Setup

**Prerequisites:**

- Node.js 16+ and npm
- Go backend running on port 8080

**Installation:**

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm start
```

The frontend will start on `http://localhost:3000` and automatically proxy API requests to the backend.

### Frontend Features

#### 1. Cart Playground

- Add items to any user's cart
- Configure SKU, name, price, and quantity
- Supports multiple users simultaneously

#### 2. Checkout Lab

- Select user for checkout
- Apply discount codes (if available)
- View cart before checkout
- See order confirmation with discount details

#### 3. Cart Snapshot

- Real-time view of selected user's cart
- Itemized list with prices and quantities
- Automatic total calculation
- Empty cart state handling

#### 4. Admin Insights

- **Statistics Dashboard:**
  - Total orders count
  - Total items sold
  - Gross revenue (formatted as currency)
  - Total discount given
- **Discount Code Management:**
  - View all discount codes (active and redeemed)
  - See eligible order numbers
  - Generate next discount code
  - Refresh stats manually

#### 5. API Console

- View raw API responses
- Debug API interactions
- See HTTP status codes and response bodies

### Frontend Testing

Run tests:

```bash
cd frontend
npm test
```

Run tests in watch mode:

```bash
npm test -- --watch
```

Run tests with coverage:

```bash
npm test -- --coverage
```

Run tests once (CI mode):

```bash
npm test -- --watchAll=false
```

**Test Coverage:**

- Initial render and component display
- Form input handling
- Cart operations (add, view)
- Checkout flow (with and without discounts)
- Discount code generation and display
- Admin stats display
- Error handling
- Loading states
- API console output

**Test Suite:** 21 tests covering user interactions, API calls, and UI states.

### Frontend Build & Deployment

**Development Build:**

```bash
npm start
```

**Production Build:**

```bash
npm run build
```

This creates an optimized production build in the `build/` directory.

**Serve Production Build:**

```bash
# Using serve (install globally: npm install -g serve)
serve -s build

# Or using Python
cd build
python -m http.server 8000
```

**Environment Variables:**

- `REACT_APP_API_BASE`: Override API base URL (default: `/api`)

Example:

```bash
REACT_APP_API_BASE=http://api.example.com/api npm start
```

### Frontend Technologies

- **React 19.2**: UI framework
- **Tailwind CSS**: Utility-first CSS framework
- **shadcn/ui**: Reusable component library
- **Lucide React**: Icon library
- **React Testing Library**: Testing utilities
- **Create React App**: Build tooling

## Development Workflow

### Running Both Services

**Terminal 1 - Backend:**

```bash
cd /path/to/project
go run ./cmd/server
```

**Terminal 2 - Frontend:**

```bash
cd /path/to/project/frontend
npm start
```

### Development Tips

1. **Hot Reload**: Both services support hot reload

   - Go: Restart server after code changes
   - React: Automatic via Create React App

2. **API Proxy**: Frontend automatically proxies `/api/*` to `http://localhost:8080`

3. **Debugging**:

   - Backend: Use `fmt.Printf` or a debugger
   - Frontend: Use browser DevTools and React DevTools
   - API Console: View raw responses in the frontend

4. **Testing**:
   - Run backend tests: `go test ./...`
   - Run frontend tests: `cd frontend && npm test`

## API Examples

### Complete Workflow Example

```bash
# 1. Add items to cart
curl -X POST http://localhost:8080/api/cart/user-1/items \
  -H 'Content-Type: application/json' \
  -d '{"sku":"SKU1","name":"Widget","price":49.99,"quantity":2}'

curl -X POST http://localhost:8080/api/cart/user-1/items \
  -H 'Content-Type: application/json' \
  -d '{"sku":"SKU2","name":"Cable","price":19.99,"quantity":1}'

# 2. View cart
curl http://localhost:8080/api/cart/user-1/items

# 3. Checkout (without discount)
curl -X POST http://localhost:8080/api/cart/user-1/checkout \
  -H 'Content-Type: application/json' \
  -d '{}'

# 4. Generate discount code (after N orders)
curl -X POST http://localhost:8080/api/admin/discounts/generate

# 5. Add items for another checkout
curl -X POST http://localhost:8080/api/cart/user-2/items \
  -H 'Content-Type: application/json' \
  -d '{"sku":"SKU3","name":"Gadget","price":99.99,"quantity":1}'

# 6. Checkout with discount code
curl -X POST http://localhost:8080/api/cart/user-2/checkout \
  -H 'Content-Type: application/json' \
  -d '{"discountCode":"DISC-XYZ789"}'

# 7. View admin stats
curl http://localhost:8080/api/admin/stats
```

### Error Handling Examples

```bash
# Empty cart checkout
curl -X POST http://localhost:8080/api/cart/user-1/checkout \
  -H 'Content-Type: application/json' \
  -d '{}'
# Response: {"error":"cart is empty"}

# Invalid discount code
curl -X POST http://localhost:8080/api/cart/user-1/checkout \
  -H 'Content-Type: application/json' \
  -d '{"discountCode":"INVALID-CODE"}'
# Response: {"error":"discount code mismatch"}

# Generate discount before eligibility
curl -X POST http://localhost:8080/api/admin/discounts/generate
# Response: {"error":"not eligible to generate discount code yet"}
```

## License

This project is provided as-is for demonstration and educational purposes.
