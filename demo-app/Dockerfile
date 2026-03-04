# ── Build stage ──────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /demo-app .

# ── Runtime stage ────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates
COPY --from=builder /demo-app /demo-app

EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/demo-app"]
