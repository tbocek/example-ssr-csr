FROM caddy:2.11.7-builder-alpine AS builder
RUN xcaddy build --with=github.com/aksdb/caddy-cgi/v2@v2.2.7

FROM caddy:2.11.7-alpine
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY Caddyfile /etc/caddy/Caddyfile
COPY --chmod=755 dynamic.sh /usr/local/bin/dynamic.sh
RUN caddy validate --config /etc/caddy/Caddyfile