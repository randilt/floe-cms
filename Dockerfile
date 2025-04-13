# Dockerfile
FROM node:18 AS frontend-builder

WORKDIR /app/web/admin

COPY web/admin/package*.json ./
RUN npm install

COPY web/admin/ ./
RUN npm run build

FROM golang:1.22-alpine AS backend-builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Copy frontend build from previous stage
COPY --from=frontend-builder /app/web/admin/dist /app/web/admin/dist

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-s -w" -o floe-cms ./cmd/floe-cms

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install necessary dependencies for runtime
RUN apk add --no-cache ca-certificates tzdata libc6-compat

# Copy the binary
COPY --from=backend-builder /app/floe-cms /app/floe-cms
COPY --from=backend-builder /app/config.yaml /app/config.yaml

# Create uploads directory
RUN mkdir -p /app/uploads && chmod 755 /app/uploads

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget -qO- http://localhost:8080/api/health || exit 1

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/floe-cms", "--config", "/app/config.yaml"]