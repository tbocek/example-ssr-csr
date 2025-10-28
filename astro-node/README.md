# Game Management - Astro Islands

Conversion from HTMX to Astro demonstrating Islands Architecture.

## Key Differences from HTMX Version

**HTMX (original)**
- SSR: Server returns HTML fragments
- Zero JS on initial load (3KB HTMX library)
- Every interaction = HTTP request
- Server-driven UI updates

**Astro (this version)**
- SSR + Hydration: Server renders, client hydrates islands
- ~10-15KB JS (only interactive components)
- Form submission + star clicks use fetch API
- Client-side updates (no full reload for stars)

## Architecture

```
Browser → Astro SSR → API Routes → PostgreSQL
          ↓
      Islands (JS only where needed)
```

**What ships JS:**
- GameList component (`<script>` tag)
- Nothing else

**What's static:**
- Page layout
- Form HTML
- Styles

## Run

```bash
docker compose up --build
# Access: http://localhost:3000
```

## Islands Architecture

Only `GameList.astro` has `<script>` tag, so only that component's JS ships to browser. Rest is static HTML.

**Bundle comparison:**
- HTMX version: ~3KB (htmx.org)
- Astro version: ~12KB (island hydration)
- React/Vue equivalent: 100KB+

## Trade-offs

**HTMX wins:**
- Smaller payload (3KB vs 12KB)
- Simpler mental model
- No build step

**Astro wins:**
- Better UX (no full reload on star click)
- Can mix frameworks (React islands, Vue islands)
- Better for complex client-side logic
- Static build option for blogs/docs

## Fresh Comparison

Fresh (Deno) is similar but:
- Uses Preact (not React)
- Built-in edge rendering
- No build step (like HTMX)
- Islands by default
- TypeScript-first

Choose Fresh if you're on Deno. Choose Astro if you want framework flexibility and npm ecosystem.