import { createServer } from 'node:http';
import { readFileSync } from 'node:fs';
import { renderToString } from 'react-dom/server';
import App from './App.jsx';

const clientJs = readFileSync('dist/client.js');

createServer((req, res) => {
    if (req.url === '/client.js') {
        res.writeHead(200, { 'Content-Type': 'text/javascript' });
        res.end(clientJs);
        return;
    }

    res.writeHead(200, { 'Content-Type': 'text/html' });
    res.end(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Hydration Error</title></head>
<body>
    <div id="root">${renderToString(<App />)}</div>
    <script src="/client.js"></script>
</body>
</html>`);
}).listen(8080);
