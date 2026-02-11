# syntax=docker/dockerfile:1.7

# ---- Frontend build ----
FROM node:20-alpine AS web
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend ./
RUN npm run build

# ---- Backend build ----
FROM golang:1.22-alpine AS api
WORKDIR /app/backend
COPY backend/go.* ./
RUN go mod download
COPY backend ./
# Static binary for simplicity
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/server ./cmd/server

# ---- Runtime ----
FROM alpine:3.19
WORKDIR /app

# Bash is used by start.sh
RUN apk add --no-cache bash ca-certificates && update-ca-certificates

# Copy prebuilt artifacts
COPY --from=api /app/bin/server /app/bin/server
COPY --from=web /app/frontend/dist /app/frontend/dist
COPY start.sh /app/start.sh

# Default envs (override in Railway)
ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    FRONTEND_DIST_DIR=/app/frontend/dist \
    UPLOAD_DIR=/app/uploads

# Create upload dirs and make script executable
RUN mkdir -p /app/uploads/tasks /app/uploads/leaves && chmod +x /app/start.sh

EXPOSE 8080
CMD ["/bin/bash", "/app/start.sh"]
