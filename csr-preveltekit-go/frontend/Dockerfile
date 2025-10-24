FROM node:24-alpine AS base
RUN apk add --no-cache brotli zopfli zstd pnpm
WORKDIR /app
COPY pnpm-lock.yaml package.json  ./
RUN pnpm install --frozen-lockfile
COPY src ./src
RUN pnpm build

FROM caddy:2-alpine
COPY --from=base /app/dist/ /usr/share/caddy
RUN ls -la /usr/share/caddy/