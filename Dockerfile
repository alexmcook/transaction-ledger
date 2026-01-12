FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# -ldflags="-s -w" strips debug info to reduce binary size
RUN go build -ldflags="-s -w" -o tl-api ./cmd/api/main.go
RUN go build -ldflags="-s -w" -o tl-worker ./cmd/worker/main.go

FROM alpine:3 AS api
WORKDIR /
COPY --from=builder /app/tl-api /tl-api
EXPOSE 8080
ENTRYPOINT ["/tl-api"]

FROM alpine:3 AS worker
WORKDIR /
COPY --from=builder /app/tl-worker /tl-worker
ENTRYPOINT ["/tl-worker"]
