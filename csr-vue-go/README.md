# Minimal CSR Example - Vue + Go

## Structure
```
csr-vue-go/
├── frontend/
│   ├── index.html      # HTML shell (minimal, no content)
│   └── app.js          # Vue app (fetches data, renders UI)
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
2. Caddy returns `index.html` (empty shell with `<div id="app">`)
3. Browser downloads Vue (from unpkg) and `app.js` (app code)
4. Vue executes, shows "Loading..."
5. Vue makes API call to `/api/products` (Caddy proxies it to the backend)
6. Go backend returns JSON data
7. Vue renders product list in the DOM

## Key CSR Characteristics
- **Empty HTML shell**: No content in initial HTML
- **JavaScript does everything**: Rendering, routing, state
- **Separate API**: Backend only serves JSON data
- **Independent deployment**: Frontend (static) and backend (API) are separate containers