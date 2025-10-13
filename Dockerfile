FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o /kvstore ./cmd/kvstore

FROM alpine:latest

WORKDIR /app

COPY --from=builder /kvstore /app/kvstore

EXPOSE 9090
EXPOSE 8080

CMD ["/app/kvstore"]
