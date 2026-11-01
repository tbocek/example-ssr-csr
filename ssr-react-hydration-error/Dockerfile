FROM node:26-alpine
WORKDIR /app
COPY package.json ./
RUN npm install
COPY App.jsx client.jsx server.jsx ./
RUN npm run build
EXPOSE 8080
CMD ["npm", "start"]
