# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
COPY main.go .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway main.go

# Run stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/api-gateway .
EXPOSE 8080
RUN adduser -D -u 10001 appuser
USER appuser
ENTRYPOINT ["./api-gateway"] 