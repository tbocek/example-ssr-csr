#

# Simple Webserver

```
docker run -p 8080:80 -v /tmp/web:/usr/share/caddy caddy:latest
```
with

```
<html>
<body>
<h1>Hallo OST</h1>
</body>
</html>
```
# Caddy with CGI-BIN

```
docker build -t caddy-cgi .
docker run -p 8080:8080 -v /tmp/web:/usr/share/caddy caddy-cgi
```

# JavaScript

```
<html>
<body>
  <h1 id="text">Hello World</h1>
  <button onclick="document.getElementById('text').innerHTML = 'You clicked me!'">
    Click me
  </button>
</body>
</html>
```
