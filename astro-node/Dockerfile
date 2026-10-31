FROM node:26-alpine
RUN npm install -g pnpm@latest
WORKDIR /app
COPY package.json ./
RUN pnpm install --frozen-lockfile || pnpm install
COPY . .
RUN pnpm run build
EXPOSE 4321
CMD ["node", "./dist/server/entry.mjs"]