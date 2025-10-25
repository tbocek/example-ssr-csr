FROM caddy:builder AS builder
RUN xcaddy build --with=github.com/aksdb/caddy-cgi/v2

FROM caddy:latest
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY Caddyfile /etc/caddy/Caddyfile
COPY --chmod=755 dynamic.sh /usr/local/bin/dynamic.sh