# Shop - HTMX + Node.js + PostgreSQL

Simple CRUD app demonstrating server-side rendering with HTMX. No client-side JavaScript framework needed.

## Stack

- **Frontend**: HTMX (hypermedia-driven)
- **Backend**: Node.js 26 + Express
- **Database**: PostgreSQL 18
- **Reverse Proxy**: Caddy
- **Container**: Docker Compose

## Architecture

```
Browser → Caddy (port 8083) → Backend (port 8080) → PostgreSQL (port 5432)
          |
          └→ Serves index.html
```

- Caddy serves `index.html` and proxies `/api/*` to backend
- Backend returns HTML fragments, not JSON
- HTMX swaps HTML directly into DOM
- Zero JavaScript bundle

## File Structure

```
project/
├── docker-compose.yml
├── Dockerfile
├── .dockerignore
├── Caddyfile
├── index.html
├── api.js
├── package.json
└── README.md
```

## Quick Start

```bash
# Start all services
docker compose up --build

# Access app
open http://localhost:8083
```

## Features

- ✅ Add products (name + description)
- ✅ List products with stock
- ✅ Increment stock (optimistic update)
- ✅ Fast shutdown (0s grace period for backend)
- ✅ Persistent PostgreSQL storage

## Development

```bash
# Rebuild after code changes
docker compose up --build

# Stop and remove volumes
docker compose down -v

# View logs
docker compose logs -f backend
```

## API Endpoints

All endpoints return HTML fragments:

- `GET /api/products` - Returns `<table>` with all products
- `POST /api/products` - Creates product, returns updated `<table>`
- `POST /api/products/:id/buy` - Decrements stock, returns updated `<tr>`

## Database

PostgreSQL data persists in `./.db/` directory. Delete it to reset:

```bash
rm -rf .db
```

## Configuration

**Port changes**: Edit `docker-compose.yml`
```yaml
caddy:
  ports:
    - "8083:80"  # Change 8083 to your port
```

**Database URL**: Set in `docker-compose.yml`
```yaml
backend:
  environment:
    DATABASE_URL: "postgres://user:pass@host:5432/db"
```

## Why HTMX?

- No build step
- No hydration
- Smaller payload (3KB vs 100KB+ for React)
- Progressive enhancement
- Server controls UI state

## Performance

Initial page load: ~5KB (HTML + HTMX)
- vs Vue/React: 100KB+ (framework + bundle)
