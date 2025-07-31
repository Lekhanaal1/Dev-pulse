# DevPulse - Microservices Developer Dashboard

A modern, microservices-based developer dashboard for task tracking, prioritization, and health monitoring built with Go, Next.js, and cloud-native technologies.

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   API Gateway   │    │   User Service  │
│   (Next.js)     │◄──►│   (Go/Gin)      │◄──►│   (Go/PostgreSQL)│
│   Port: 3000    │    │   Port: 8000    │    │   Port: 8001    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Task Service   │    │ Analytics Svc   │    │ Notification Svc│
│  (Go/MongoDB)   │    │ (Go/NATS)       │    │ (Go/NATS)       │
│  Port: 8002     │    │ Port: 8003      │    │ Port: 8004      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Features

- **Microservices Architecture**: Independent services with clear boundaries
- **Polyglot Persistence**: PostgreSQL for users, MongoDB for tasks
- **Event-Driven Messaging**: NATS for asynchronous communication
- **API Gateway**: Centralized routing, authentication, and rate limiting
- **Monitoring**: Prometheus metrics and Grafana dashboards
- **Containerized**: Docker and Docker Compose for easy deployment
- **Security**: JWT authentication and password hashing

## 🛠️ Tech Stack

- **Backend**: Go 1.23+, Gin framework
- **Frontend**: Next.js, React
- **Databases**: PostgreSQL, MongoDB
- **Message Broker**: NATS
- **Monitoring**: Prometheus, Grafana
- **Containerization**: Docker, Docker Compose
- **Authentication**: JWT tokens

## 📋 Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local development)
- Node.js 18+ (for local development)

## 🚀 Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/Lekhanaal1/Dev-pulse.git
   cd Dev-pulse
   ```

2. **Start all services**
   ```bash
   docker-compose up --build
   ```

3. **Access the application**
   - Frontend: http://localhost:3000
   - API Gateway: http://localhost:8000
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3001

## 📚 API Documentation

### User Service (Port 8001)

#### Create User
```bash
POST /users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

#### Get All Users
```bash
GET /users
Authorization: Bearer <jwt-token>
```

#### Get User by ID
```bash
GET /users/:id
Authorization: Bearer <jwt-token>
```

#### Update User
```bash
PUT /users/:id
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "name": "John Updated",
  "email": "john.updated@example.com"
}
```

#### Delete User
```bash
DELETE /users/:id
Authorization: Bearer <jwt-token>
```

### Health Checks

All services expose health endpoints:
```bash
curl http://localhost:8000/healthz  # API Gateway
curl http://localhost:8001/healthz  # User Service
curl http://localhost:8002/healthz  # Task Service
curl http://localhost:8003/healthz  # Analytics Service
curl http://localhost:8004/healthz  # Notification Service
```

## 🔧 Development

### Local Development Setup

1. **Start dependencies only**
   ```bash
   docker-compose up postgres mongodb nats prometheus grafana
   ```

2. **Run services locally**
   ```bash
   # Terminal 1 - User Service
   cd user-service && go run main.go
   
   # Terminal 2 - Task Service
   cd task-service && go run main.go
   
   # Terminal 3 - Frontend
   cd frontend && npm run dev
   ```

### Testing

```bash
# Run all tests
go test ./...

# Test specific service
cd user-service && go test -v
```

## 📊 Monitoring

### Prometheus Metrics
- Service health and availability
- Request rates and response times
- Error rates and status codes

### Grafana Dashboards
- Real-time service metrics
- Performance analytics
- Alert notifications

## 🔒 Security

- JWT-based authentication
- Password hashing with bcrypt
- Rate limiting on API Gateway
- Input validation and sanitization

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with modern microservices best practices
- Inspired by enterprise-grade developer tools
- Uses industry-standard technologies and patterns 