# Build stage - Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Build stage - Backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o nft-ui .

# Final stage
FROM alpine:latest
RUN apk --no-cache add nftables ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/nft-ui .

# Default environment variables
ENV NFT_UI_LISTEN_ADDR=:8080

EXPOSE 8080

ENTRYPOINT ["/app/nft-ui"]
