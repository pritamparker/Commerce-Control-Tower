# Ecommerce Discount Store

A lightweight Go service that exposes cart, checkout, and admin APIs for an ecommerce experience. It keeps state in-memory and showcases the "every *n*th order gets 10% off" promotion with a tiny UI.

## Features

- Add items to a per-user cart and view its contents.
- Checkout endpoint validates optional discount codes and records orders.
- Admin endpoints to generate the next eligible discount code and to view live stats (orders, revenue, discounts, code history).
- React front-end (`frontend/`) built with Tailwind CSS + shadcn/ui to visualize carts, checkout, and admin stats.
- Unit tests covering the store logic.

## Getting Started

```bash
# Install dependencies
cd /Users/pritam/Documents/pre

# Fetch modules
go mod tidy

# Run tests
go test ./...

# Start the server (defaults to port 8080)
go run ./cmd/server
```

Optional environment variables:

- `PORT`: override HTTP port (default `8080`).
- `NTH_ORDER_DISCOUNT`: change how frequently discount codes are unlocked (default `3`).

The Go server exposes only the API. Run the React app below (or use your favorite REST client) to interact with it.

### React Frontend (optional but recommended)

```bash
cd /Users/pritam/Documents/pre/frontend
npm install        # installs CRA deps + tailwind/shadcn stack
npm start          # runs on http://localhost:3000
```

The CRA dev server proxies `/api` calls to the Go backend (see `package.json`), so keep `go run ./cmd/server` running on port `8080`.

## API Overview

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/cart/{userID}/items` | Add/merge an item into the user's cart. |
| `GET` | `/api/cart/{userID}/items` | Inspect the current cart snapshot. |
| `POST` | `/api/cart/{userID}/checkout` | Checkout the cart. Accepts `{ "discountCode": "DISC-..." }`. |
| `POST` | `/api/admin/discounts/generate` | Create the next discount code if the nth-order rule is satisfied. |
| `GET` | `/api/admin/stats` | Aggregated counts (orders, revenue, discounts, codes). |

All responses are JSON. Error payloads look like `{ "error": "message" }`.

## Testing With REST Clients

Example `curl` sequence:

```bash
# Add item
curl -X POST http://localhost:8080/api/cart/user-1/items \
  -H 'Content-Type: application/json' \
  -d '{"sku":"SKU1","name":"Widget","price":49.99,"quantity":2}'

# Checkout
curl -X POST http://localhost:8080/api/cart/user-1/checkout -d '{}'

# Generate discount (admin)
curl -X POST http://localhost:8080/api/admin/discounts/generate
```

## Next Steps

- Persist carts/orders in a database.
- Add authentication for admin endpoints.
- Build a polished frontend or mobile client.
