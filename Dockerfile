# Multi-stage Dockerfile for Confluent knowledge base
# 1. Build Backend Binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compile static Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/confluent ./cmd/server

# 2. Minimal Production Runtime
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata sqlite

# Create data directory for SQLite
RUN mkdir -p /data

COPY --from=builder /app/bin/confluent /app/confluent

ENV PORT=8080
ENV DB_PATH=/data/confluent.db

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["/app/confluent"]
