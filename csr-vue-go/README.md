# Minimal CSR Example - Vue + Go

## Structure
```
csr-example/
├── frontend/
│   ├── index.html      # HTML shell (minimal, no content)
│   ├── app.js          # Vue app (fetches data, renders UI)
│   └── Dockerfile      # Nginx serves static files
├── backend/
│   ├── main.go         # Go API server (returns JSON)
│   └── Dockerfile      # Go API container
└── docker-compose.yml  # Orchestrates both services
```

## Run
```bash
docker-compose up --build
```

Then open: http://localhost:3000

## How It Works (CSR Flow)
1. Browser requests `http://localhost:3000`
2. Nginx returns `index.html` (empty shell with `<div id="app">`)
3. Browser downloads `app.js` (Vue framework + app code)
4. Vue executes, shows "Loading..."
5. Vue makes API call to `http://localhost:8080/api/users`
6. Go backend returns JSON data
7. Vue renders user list in the DOM

## Key CSR Characteristics
- **Empty HTML shell**: No content in initial HTML
- **JavaScript does everything**: Rendering, routing, state
- **Separate API**: Backend only serves JSON data
- **Independent deployment**: Frontend (static) and backend (API) are separate containers