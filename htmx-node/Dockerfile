FROM node:24-alpine
WORKDIR /app
COPY package*.json ./
RUN corepack enable pnpm && pnpm install --production
COPY api.js .
EXPOSE 8080
CMD ["node", "api.js"]