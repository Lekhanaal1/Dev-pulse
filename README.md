DevPulse
========

DevPulse is a work-in-progress microservices task tracker, built to practice
Go, Docker, and distributed-systems patterns. It is a scaffold, not a
finished product — most services currently only expose a health check.

Planned architecture
---------------------

- user-service: Go + Gin + PostgreSQL
- task-service: Go + Gin + MongoDB
- analytics-service: Go + Gin + NATS
- notification-service: Go + Gin + NATS
- frontend: Next.js (React)
- NATS for async messaging between services
- Prometheus + Grafana for monitoring

Current status
---------------

- [x] Docker Compose wiring for all services + Postgres/Mongo/NATS/Prometheus/Grafana
- [x] Each Go service builds and serves `/healthz`
- [ ] Actual task/user/notification business logic (not yet implemented)
- [ ] API gateway
- [ ] Automated tests
- [ ] CI/CD

Getting started
----------------

**Prerequisites:**
- Docker and Docker Compose
- Go 1.21 or newer
- Node.js 18 or newer

**Quickstart:**

    git clone <repo-url>
    cd Dev-pulse
    cp .env.example .env   # then edit .env if you want non-default credentials
    docker-compose up --build

This brings up all services, Postgres, MongoDB, NATS, Prometheus (port 9090),
and Grafana (port 3001). Each Go service responds to `GET /healthz` on its
mapped port (8001-8004).

Repo structure
---------------

    Dev-pulse/
      user-service/
      task-service/
      analytics-service/
      notification-service/
      frontend/
      deployments/    # Prometheus config
      docker-compose.yml

Next steps
-----------
- Implement real handlers for each service (currently health-check stubs only)
- Add an API gateway in front of the individual services
- Write unit/integration tests
- Add a GitHub Actions workflow for build + test on push
