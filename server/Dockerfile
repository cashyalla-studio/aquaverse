# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o aquaverse-api ./cmd/server/...

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/aquaverse-api .
COPY config.yaml ./config.yaml
EXPOSE 8080
ENTRYPOINT ["./aquaverse-api"]
