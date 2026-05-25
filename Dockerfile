# syntax=docker/dockerfile:1.6

# ---------- Build stage ----------
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

ARG SERVICE=api
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/app ./cmd/app

# ---------- Runtime stage ----------
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/app /app/app
COPY migrations /app/migrations

USER app

EXPOSE 8080 50051

ENTRYPOINT ["/app/app"]
