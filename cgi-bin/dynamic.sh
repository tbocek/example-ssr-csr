#!/bin/sh

echo "Content-Type: text/html"
echo ""
echo "<html><body>"
echo "<h1>CGI Script Output</h1>"
echo "<p>Current time: $(date)</p>"
echo "<p>Server: $SERVER_SOFTWARE</p>"
echo "<p>Request: $REQUEST_METHOD $REQUEST_URI</p>"
echo "</body></html>"