FROM node:24-alpine
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app
COPY package.json ./
RUN pnpm install --frozen-lockfile || pnpm install
COPY . .
RUN pnpm run build
EXPOSE 4321
CMD ["node", "./dist/server/entry.mjs"]