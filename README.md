# SSR vs CSR vs Hybrid Rendering Examples

This repository contains three implementations of the same Game Management application, demonstrating different rendering approaches: Server-Side Rendering (SSR), Client-Side Rendering (CSR), and Hybrid Rendering with PrevelteKit.

## Projects

### 1. [ssr-java](./ssr-java) - Server-Side Rendering
**Stack:** Java 25, Spring Boot, Thymeleaf, PostgreSQL

Classic server-side rendering where HTML is generated on the server for each request.

**Characteristics:**
- Server renders complete HTML with data on every request
- Fast First Contentful Paint (FCP) - content visible immediately
- Full page reload on every interaction
- No JavaScript required for basic functionality
- Higher server load (rendering on each request)

**Best for:** Content-heavy sites, SEO-critical pages, traditional web applications

---

### 2. [csr-vue-go](./csr-vue-go) - Client-Side Rendering
**Stack:** Vue 3, Go, PostgreSQL

Modern client-side rendering where the browser does all the work.

**Characteristics:**
- Server sends empty HTML shell
- JavaScript fetches data and renders UI client-side
- No page reloads after initial load (client-side routing)
- Requires JavaScript for all functionality
- Slower FCP (must download and execute JS first)
- Subsequent interactions are very fast

**Best for:** Web applications, dashboards, interactive tools

---

### 3. [csr-preveltekit-go](./csr-preveltekit-go) - Hybrid Rendering
**Stack:** Svelte 5, PrevelteKit, Go, PostgreSQL

Hybrid approach combining the best of SSR and CSR using build-time pre-rendering.

**Characteristics:**
- Server sends pre-rendered HTML with visible content
- Fast FCP (content visible immediately like SSR)
- JavaScript hydrates for interactivity (no page reloads like CSR)
- Static files can be deployed to CDN
- Build-time rendering (not per-request like SSR)

**Best for:** Experiments (I'm currently the only user)

---

## Feature Comparison

| Feature | SSR (Java) | CSR (Vue) | Hybrid (PrevelteKit) |
|---------|------------|-----------|----------------------|
| **First Contentful Paint** | ⚡ Fast | 🐌 Slow | ⚡ Fast |
| **Time to Interactive** | ⚡ Fast | 🐌 Slow | ⚡ Fast |
| **Subsequent Navigation** | 🐌 Full reload | ⚡ Instant | ⚡ Instant |
| **SEO** | ✅ Excellent | ❌ Poor* | ✅ Excellent |
| **Server Load** | 🔥 High | ✅ Low | ✅ Low |
| **JavaScript Required** | ❌ No | ✅ Yes | ⚠️ For interactivity |
| **Deployment** | Server | CDN/Server | CDN |
| **Scalability** | Vertical | Horizontal | Horizontal |

\* CSR can be made SEO-friendly with server-side rendering or pre-rendering

---

## Architecture Patterns

### Three-Tier Architecture

All three examples implement a three-tier architecture:

**Presentation Tier:**
- **SSR:** Server-rendered HTML (Thymeleaf templates)
- **CSR:** Client-rendered UI (Vue components)
- **Hybrid:** Pre-rendered HTML + hydrated Svelte components

**Application Tier:**
- **SSR:** Spring Boot MVC controllers
- **CSR:** Go REST API
- **Hybrid:** Go REST API

**Data Tier:**
- All three use PostgreSQL for persistent storage

### Key Architectural Difference

**Traditional SSR:** Presentation logic mixed with application tier (server generates HTML)

**CSR/Hybrid:** Clean separation - Presentation tier entirely separate from application tier (API-first approach)

---

## Request Flow Comparison

### SSR Flow
1. Browser requests `/games`
2. Spring Boot controller queries database
3. Thymeleaf merges data with template server-side
4. Complete HTML sent to browser
5. Browser displays content immediately
6. User interaction → full page reload (back to step 1)

### CSR Flow
1. Browser requests `/`
2. Server sends empty HTML shell + JavaScript bundle
3. Browser downloads and executes JavaScript
4. JavaScript makes API call to `/api/games`
5. API queries database, returns JSON
6. Vue renders HTML in browser
7. User interaction → JavaScript updates DOM (no reload)

### Hybrid Flow
1. Browser requests `/`
2. Server sends pre-rendered HTML with visible content
3. Browser displays content immediately (fast FCP)
4. JavaScript downloads and hydrates page
5. JavaScript makes API call to `/api/games`
6. API queries database, returns JSON
7. Svelte updates DOM with fresh data
8. User interaction → JavaScript updates DOM (no reload)

---

## Running the Examples

Each project has its own `docker-compose.yml` and can be run independently:

```bash
# SSR (Java/Spring Boot)
cd ssr-java
docker compose up --build
# Access at http://localhost:8080

# CSR (Vue/Go)
cd csr-vue-go
docker compose up --build
# Access at http://localhost:3000

# Hybrid (PrevelteKit/Go)
cd csr-preveltekit-go
docker compose up --build
# Access at http://localhost
```

---

## Common Features

All three implementations provide the same functionality:
- ✅ Add games with title and description
- ✅ List all games in a table
- ✅ Click star button to increment count
- ✅ Persistent storage in PostgreSQL
- ✅ Sample data (Zelda, Mario) loaded on startup

---

## Technology Stack Summary

| Component | SSR | CSR | Hybrid |
|-----------|-----|-----|--------|
| **Frontend** | Thymeleaf | Vue 3 | Svelte 5 |
| **Backend** | Spring Boot | Go | Go |
| **Database** | PostgreSQL | PostgreSQL | PostgreSQL |
| **Web Server** | Embedded Tomcat | Caddy | Caddy |
| **Build Tool** | Gradle | - | PrevelteKit |

---

## When to Use Each Approach

### Choose SSR when:
- SEO is critical
- Content changes frequently
- Simple user interactions
- You have powerful servers
- Users may have JavaScript disabled

### Choose CSR when:
- Building a web application (not a website)
- Rich interactivity is needed
- SEO is not a priority
- You want to scale horizontally
- You need offline functionality (with service workers)

### Choose Hybrid when:
- SEO is important AND you want rich interactivity
- You want fast initial load times
- You can pre-render content at build time
- You want to deploy to a CDN
- Content doesn't change per-user on initial load

---

## Learning Objectives

By exploring these three examples, you'll understand:

1. **Architecture:** How SSR and CSR represent different architectural philosophies - SSR keeps presentation logic on the server (blurred tier boundaries, monolithic), while CSR enforces strict three-tier separation with an API-first approach (presentation tier moves entirely to client, application tier becomes stateless REST/GraphQL API). This impacts deployment models, horizontal scaling, team structure (frontend/backend split), technology choices per tier, and authentication.
2. **Rendering Strategies:** How SSR, CSR, and hybrid rendering work
3. **Performance Trade-offs:** Why each approach has different performance characteristics
4. **Architecture Patterns:** How MVC and three-tier architecture apply in different contexts
5. **Modern Web Development:** Current best practices for building web applications
6. **Framework Comparison:** Practical differences between Spring Boot, Vue, and Svelte

---

## Further Reading

- [Server-Side Rendering vs Client-Side Rendering](https://web.dev/rendering-on-the-web/)
- [Spring Boot Documentation](https://spring.io/projects/spring-boot)
- [Vue.js Guide](https://vuejs.org/guide/)
- [Svelte Tutorial](https://svelte.dev/tutorial)
- [PrevelteKit](https://github.com/tbocek/preveltekit)