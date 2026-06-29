# ---- build stage ----
FROM node:20-alpine AS build
WORKDIR /app

# Install deps first (cached unless package files change).
COPY package.json package-lock.json ./
RUN npm ci

# Build the static site. The NATS WebSocket URL is baked in at build
# time; the default (ws://localhost:8080) is correct for local Docker.
# To target another broker, build with --build-arg or set VITE_NATS_URL.
COPY . .
RUN npm run build

# ---- serve stage ----
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
# nginx:alpine runs nginx in the foreground by default.
