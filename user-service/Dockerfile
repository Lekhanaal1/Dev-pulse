FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o user-service

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/user-service ./user-service
EXPOSE 8000
CMD ["./user-service"] 