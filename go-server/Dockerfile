# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-bookworm AS builder
ENV CGO_ENABLED=1
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ./frontend/dist
RUN go build -o wiretap ./cmd/wiretap/

# Stage 3: Runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/wiretap .

EXPOSE 8080 9090 1025

VOLUME /data

ENV WIRETAP_STORAGE_PATH=/data/wiretap.duckdb

CMD ["./wiretap"]
