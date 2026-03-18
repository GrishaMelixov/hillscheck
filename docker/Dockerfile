# Stage 1: Build React frontend
FROM node:24-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/src ./src
COPY web/public ./public
COPY web/index.html web/vite.config.ts web/tailwind.config.ts web/tsconfig.json web/postcss.config.js ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o hillscheck ./cmd/server

# Stage 3: Alpine runtime with Tesseract OCR (free, offline, no external API)
FROM alpine:3.21
RUN apk add --no-cache \
      tesseract-ocr \
      tesseract-ocr-data-rus \
      ca-certificates \
    && rm -rf /var/cache/apk/*
COPY --from=builder /app/hillscheck /hillscheck
COPY --from=builder /app/migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/hillscheck"]
