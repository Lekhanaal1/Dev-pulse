FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o task-service

FROM alpine:3.18
RUN adduser -D -H appuser
WORKDIR /app
COPY --from=builder /app/task-service ./task-service
USER appuser
EXPOSE 8000
CMD ["./task-service"]
