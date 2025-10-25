# Simple Webserver Examples

## 1. Simple HTML

Create `/tmp/web/index.html`:
```html
<html>
<body>
  <h1>Hallo OST</h1>
</body>
</html>
```

Run:
```bash
docker run -p 8080:80 -v /tmp/web:/usr/share/caddy caddy:latest
```

Access: http://localhost:8080

## 2. CGI-bin (Server-side processing)

Build and run:
```bash
docker build -t caddy-cgi .
docker run -p 8080:8080 -v /tmp/web:/usr/share/caddy caddy-cgi
```

Access: http://localhost:8080/cgi-bin/dynamic

The CGI script (`dynamic.sh`) generates HTML server-side showing current time and request info.

## 3. JavaScript (Client-side)

Create `/tmp/web/js.html`:
```html
<html>
<body>
  <h1 id="text">Hello World</h1>
  <button onclick="document.getElementById('text').innerHTML = 'You clicked me!'">
    Click me
  </button>
</body>
</html>
```

Access: http://localhost:8080/js.html (with either setup above)