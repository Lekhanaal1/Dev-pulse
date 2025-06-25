FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o analytics-service

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/analytics-service ./analytics-service
EXPOSE 8000
CMD ["./analytics-service"] 