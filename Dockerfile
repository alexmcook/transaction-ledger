FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy and download dependencies (cached if go.mod/sum don't change)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 creates a static binary
# -ldflags="-s -w" strips debug info to reduce binary size
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o ledger-api ./cmd/transaction-ledger/main.go

FROM alpine:3

# tini is required for fiber prefork
RUN apk add --no-cache tini

WORKDIR /

COPY --from=builder /app/ledger-api /ledger-api

EXPOSE 8080

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/ledger-api"]
