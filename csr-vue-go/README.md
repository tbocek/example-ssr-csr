# Minimal CSR Example - Vue + Go

## Structure
```
csr-vue-go/
├── frontend/
│   ├── index.html      # HTML page with the Vue template inside <div id="app">
│   └── app.js          # Vue app (compiles the template, fetches data, renders UI)
├── backend/
│   ├── main.go         # Go API server (returns JSON)
│   └── Dockerfile      # Go API container
├── Caddyfile           # Serves frontend/, proxies /api/* to the backend
└── docker-compose.yml  # Orchestrates Caddy, backend and PostgreSQL
```

## Run
```bash
docker-compose up --build
```

Then open: http://localhost:8081

## How It Works (CSR Flow)
1. Browser requests `http://localhost:8081`
2. Caddy returns `index.html` (Vue template inside `<div id="app">`, no product data)
3. Browser downloads Vue (from unpkg) and `app.js` (app code)
4. Until `app.js` runs, the browser may briefly show the raw template, including `{{ ... }}` placeholders
5. Vue reads the template from `<div id="app">`, compiles it, replaces the div's content with newly created DOM nodes and shows "Loading..."
6. Vue makes API call to `/api/products` (Caddy proxies it to the backend)
7. Go backend returns JSON data
8. Vue renders product list in the DOM

This is not hydration: hydration reuses HTML rendered on the server (`createSSRApp`), while `createApp` here discards the initial HTML and renders it again.

## Key CSR Characteristics
- **No data in initial HTML**: The HTML only contains the uncompiled template
- **JavaScript does the rendering**: Template compilation, rendering and state all happen in the browser
- **Separate API**: Backend only serves JSON data
- **Independent deployment**: Frontend (static) and backend (API) are separate containers