# Shop - Astro Islands

Conversion from HTMX to Astro demonstrating Islands Architecture.

## Key Differences from HTMX Version

**HTMX (original)**
- SSR: Server returns HTML fragments
- No app JS on initial load (~13KB gzipped HTMX library)
- Every interaction = HTTP request
- Server-driven UI updates

**Astro (this version)**
- SSR + inline script: Server renders, a `<script>` attaches event listeners
- ~1KB gzipped JS (one inline script, no hydration runtime)
- Form submission + buy clicks use fetch API
- Client-side updates (no full reload for stock)

## Architecture

```
Browser → Astro SSR → API Routes → PostgreSQL
          ↓
      Islands (JS only where needed)
```

**What ships JS:**
- The `<script>` tag in `src/pages/index.astro`
- Nothing else

**What's static:**
- Page layout
- Form HTML
- Styles

## Run

```bash
docker compose up --build
# Access: http://localhost:4321
```

## Islands Architecture

Only `src/pages/index.astro` has a `<script>` tag, so only that script ships to the browser. Rest is static HTML. This example uses no framework island components.

**Bundle comparison:**
- HTMX version: ~13KB gzipped (htmx.org 4.0.0)
- Astro version: ~1KB gzipped (inline script)
- React/Vue equivalent: 100KB+

## Trade-offs

**HTMX wins:**
- Simpler mental model
- No build step

**Astro wins:**
- Smaller payload (~1KB vs ~13KB)
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