# PrevelteKit Game Management - Hybrid Rendering Demo

A game management application demonstrating **hybrid rendering** (build-time pre-rendering + hydration) using PrevelteKit, Svelte 5, Go, and PostgreSQL.

## What is PrevelteKit?

PrevelteKit is a minimalistic framework (<500 lines of code) that combines:
- **Build-time pre-rendering**: Renders Svelte components to HTML during build using jsdom
- **Hydration**: Attaches JavaScript event handlers to pre-rendered HTML
- **Static deployment**: No Node.js runtime needed in production

**The best of both worlds:** Fast initial load (like SSR) + instant navigation (like CSR).

## Project Structure
```
csr-preveltekit-go/
├── src/
│   └── Index.svelte        # Main Svelte 5 component with reactive state
├── backend/
│   ├── main.go             # Go REST API (game CRUD + stars)
│   ├── go.mod              # Go dependencies
│   ├── go.sum              # Dependency checksums
│   └── Dockerfile          # Multi-stage build for Go API
├── package.json            # PrevelteKit + Svelte dependencies
├── Dockerfile              # Multi-stage: npm build → Caddy
├── Caddyfile               # Reverse proxy config (static + API)
├── docker-compose.yml      # Orchestrates all 3 services
└── .db/                    # PostgreSQL data volume
```

## Quick Start

```bash
docker compose up --build
```

Access the application at: **http://localhost**

## Features

- ✅ Add games with title and description
- ✅ List all games in a table
- ✅ Click star (⭐) button to increment count
- ✅ **Build-time pre-rendering** - content visible immediately
- ✅ **Hydration** - JavaScript adds interactivity without page reloads
- ✅ **Persistent storage** in PostgreSQL
- ✅ **Sample data** (Zelda, Mario) loaded on startup

## How PrevelteKit Works

### Build Time (during `docker build`)
1. `npm run build` executes PrevelteKit
2. PrevelteKit uses **jsdom** to create a fake browser environment
3. Svelte components render to HTML in this fake browser
4. Pre-rendered HTML + JavaScript bundle saved to `/dist`
5. Static files copied to Caddy container

### Runtime (when user visits site)
1. Browser requests `http://localhost`
2. Caddy returns **pre-rendered HTML** with visible layout/content
3. Browser displays content immediately (**fast FCP**)
4. JavaScript downloads and **hydrates** the page (attaches event handlers)
5. JavaScript makes API call to `/api/games` (proxied to Go backend)
6. Go backend queries PostgreSQL, returns JSON
7. Svelte updates DOM with fresh data
8. User interactions update DOM **without page reloads**

### Key Implementation Detail

The `window?.__isBuildTime` check prevents API calls during pre-rendering:

```javascript
// Fetch on mount (only if not pre-rendering)
if (!window?.__isBuildTime) {
    fetchGames();
}
```

## Architecture

### Three-Tier Architecture with Clean Separation

**Presentation Tier** (Client-Side)
- Pre-rendered Svelte components served as static files
- Caddy web server (no Node.js runtime needed)
- Deployed to CDN or any static host

**Application Tier** (Backend API)
- Go REST API with JSON endpoints
- Stateless design (no sessions)
- Handles business logic and data operations

**Data Tier** (Database)
- PostgreSQL for persistent storage
- Auto-created schema from Go structs

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/games` | List all games |
| POST | `/api/games` | Create new game |
| POST | `/api/games/{id}/star` | Increment star count |

### Database Schema

```sql
CREATE TABLE games (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    stars INTEGER DEFAULT 0
);
```

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Frontend Framework** | Svelte 5 | Reactive UI components |
| **Build Tool** | PrevelteKit | Build-time pre-rendering |
| **Web Server** | Caddy 2 | Serves static files + reverse proxy |
| **Backend** | Go 1.21 | REST API server |
| **Database** | PostgreSQL 18 | Persistent storage |
| **Containerization** | Docker + Compose | Multi-service orchestration |

## Comparison: Hybrid vs SSR vs CSR

| Aspect | Hybrid (PrevelteKit) | SSR (Spring Boot) | CSR (Vue) |
|--------|---------------------|-------------------|-----------|
| **Initial HTML** | Pre-rendered with content | Rendered per-request | Empty shell |
| **FCP** | ⚡ Fast (~200ms) | ⚡ Fast (~200ms) | 🐌 Slow (~800ms) |
| **TTI** | ⚡ Fast (~400ms) | ⚡ Fast (~300ms) | 🐌 Slow (~1000ms) |
| **Navigation** | ⚡ No reload | 🐌 Full reload | ⚡ No reload |
| **SEO** | ✅ Excellent | ✅ Excellent | ❌ Poor |
| **Server Load** | ✅ Low (static) | 🔥 High (per-request) | ✅ Low (static) |
| **Deployment** | CDN | Application Server | CDN |
| **JS Required** | ⚠️ For interactivity | ❌ No | ✅ Yes |

## Key Advantages of Hybrid Rendering

### vs Traditional SSR
- ✅ No server-side rendering on each request → lower server costs
- ✅ Deploy to CDN → better global performance
- ✅ Simpler infrastructure (just static files)
- ✅ No server-side state to manage

### vs Pure CSR
- ✅ Fast FCP → content visible immediately
- ✅ Better SEO → search engines see content
- ✅ Works without JavaScript (partially) → accessibility
- ✅ Faster TTI → less JavaScript to execute initially

### Trade-offs
- ⚠️ Content pre-rendered at build time (not per-user)
- ⚠️ Dynamic per-user content requires API calls
- ⚠️ Build time increases with number of routes

## When to Use PrevelteKit (Hybrid Rendering)

### ✅ Great For:
- Marketing sites with dynamic features
- Documentation sites with search/filtering
- E-commerce product pages
- Landing pages with forms
- Blogs with comments/interactions
- Company websites with dashboards

### ❌ Not Ideal For:
- Per-user personalized content on initial load
- Real-time collaborative applications
- Admin panels with complex auth flows
- Applications with thousands of dynamic routes

## Development Workflow

### Local Development (without Docker)
```bash
# Install dependencies
npm install

# Start dev server with hot reload
npm run dev

# Start backend separately
cd backend && go run main.go

# Start PostgreSQL
docker run -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:18-alpine
```

### Production Build
```bash
# Build static files
npm run build

# Files generated in ./dist/
# - index.html (pre-rendered)
# - index.html.br (Brotli compressed)
# - index.html.gz (Gzip compressed)
# - index.html.zst (Zstandard compressed)
# - static/ (JS, CSS bundles)
```

## Docker Services

```yaml
services:
  frontend:  # Caddy with pre-rendered static files
    ports: 80
    
  backend:   # Go API server
    ports: 8080
    
  postgres:  # Database
    ports: 5432
```

## Learning Resources

- [PrevelteKit GitHub](https://github.com/tbocek/preveltekit)
- [PrevelteKit Documentation](https://tbocek.github.io/preveltekit/doc)
- [Svelte 5 Tutorial](https://svelte.dev/tutorial)
- [Rendering on the Web (Google)](https://web.dev/rendering-on-the-web/)

## License

Educational project for demonstrating hybrid rendering concepts.