# =========================
# Build stage
# =========================

FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o kuflow ./cmd/kuflow

# =========================
# Runtime stage
# =========================

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

    COPY --from=builder /app/kuflow .
    COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./kuflow"]

