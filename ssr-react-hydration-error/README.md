# Minimal Hydration Error Example - React + Node

## Structure
```
ssr-react-hydration-error/
├── App.jsx             # Component that renders differently on server and client
├── server.jsx          # Node server, renders App to HTML
├── client.jsx          # Browser code, hydrates the server-rendered HTML
├── Dockerfile          # Bundles the code with esbuild and runs the server
└── docker-compose.yml  # Runs the app
```

## Run
```bash
docker compose up --build
```

Then open http://localhost:8084 and the browser console.

## How It Works
1. The server renders `App`, `typeof window` is `undefined`: `Rendered by: server`
2. The browser shows this HTML and downloads `client.js`
3. `hydrateRoot` renders `App` again, `window` exists now: `Rendered by: client`
4. The text differs from the server HTML, React logs the error and renders the tree again on the client

The console shows:
```
Hydration failed because the server rendered text didn't match the client.
...
  <App>
    <p>
+     client
-     server
```

The page ends up showing `Rendered by: client`. The client bundle uses the React development build, a production build only logs a minified error.

## Fix
The first client render must match the server, so change the value after hydration:
```jsx
import { useEffect, useState } from 'react';

export default function App() {
    const [renderedBy, setRenderedBy] = useState('server');

    useEffect(() => {
        setRenderedBy('client');
    }, []);

    return <p>Rendered by: {renderedBy}</p>;
}
```
