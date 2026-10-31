FROM node:26-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install -g pnpm@latest && pnpm install --production
COPY api.js .
EXPOSE 8080
CMD ["node", "api.js"]