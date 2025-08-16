FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o key-value-engine ./cmd/kvstore

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/key-value-store .

EXPOSE 8080

CMD ["./key-value-store"] 