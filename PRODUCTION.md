# DevPulse Production Deployment Guide

## 🚀 Production Checklist

### 1. Security Configuration

#### JWT Secret
- [ ] Generate a strong JWT secret (at least 32 characters)
- [ ] Set `JWT_SECRET` environment variable
- [ ] Never commit secrets to version control

```bash
# Generate a secure JWT secret
openssl rand -base64 32
```

#### Database Security
- [ ] Use strong passwords for PostgreSQL and MongoDB
- [ ] Enable SSL/TLS for database connections
- [ ] Restrict database access to application servers only
- [ ] Use connection pooling for better performance

#### Network Security
- [ ] Enable HTTPS with valid SSL certificates
- [ ] Configure firewall rules
- [ ] Use reverse proxy (nginx/traefik) for SSL termination
- [ ] Implement rate limiting per IP/user

### 2. Environment Configuration

Create a `.env` file based on `env.example`:

```bash
# Copy and customize the environment file
cp env.example .env

# Edit with production values
nano .env
```

#### Required Environment Variables

```bash
# JWT Configuration
JWT_SECRET=your-production-jwt-secret-here

# Database Configuration
POSTGRES_USER=devpulse_user
POSTGRES_PASSWORD=your-secure-postgres-password
POSTGRES_DB=devpulse_users

# MongoDB Configuration
MONGO_URI=mongodb://your-mongo-instance:27017/devpulse_tasks

# Frontend Configuration
NEXT_PUBLIC_API_URL=https://your-api-domain.com
```

### 3. Infrastructure Setup

#### Option A: Docker Compose (Development/Staging)
```bash
# Build and start services
docker-compose up --build -d

# Check service health
docker-compose ps
```

#### Option B: Kubernetes (Production)
```bash
# Deploy to Kubernetes
kubectl apply -f deployments/helm/devpulse/

# Check deployment status
kubectl get pods -n devpulse
```

### 4. Monitoring & Observability

#### Prometheus Configuration
- [ ] Configure Prometheus to scrape all services
- [ ] Set up alerting rules
- [ ] Configure retention policies

#### Grafana Dashboards
- [ ] Import default dashboards
- [ ] Configure alerts
- [ ] Set up user authentication

#### Logging
- [ ] Configure structured logging
- [ ] Set up log aggregation (ELK stack)
- [ ] Configure log retention policies

### 5. Performance Optimization

#### Database Optimization
```sql
-- PostgreSQL indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- MongoDB indexes
db.tasks.createIndex({ "user_id": 1 });
db.tasks.createIndex({ "status": 1 });
db.tasks.createIndex({ "created_at": -1 });
```

#### Application Optimization
- [ ] Enable connection pooling
- [ ] Configure appropriate timeouts
- [ ] Set up caching (Redis)
- [ ] Optimize database queries

### 6. Backup & Recovery

#### Database Backups
```bash
# PostgreSQL backup
pg_dump -h localhost -U postgres devpulse_users > backup.sql

# MongoDB backup
mongodump --db devpulse_tasks --out /backup/
```

#### Automated Backups
- [ ] Set up automated daily backups
- [ ] Test backup restoration
- [ ] Store backups in secure location

### 7. Security Hardening

#### Container Security
```bash
# Run security scans
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image devpulse/user-service:latest
```

#### Network Security
- [ ] Use private networks for inter-service communication
- [ ] Implement service mesh (Istio/Linkerd)
- [ ] Configure network policies

### 8. CI/CD Pipeline

#### GitHub Actions Example
```yaml
name: Deploy to Production
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build and push images
        run: |
          docker build -t devpulse/user-service ./user-service
          docker push your-registry/devpulse/user-service
      - name: Deploy to production
        run: |
          kubectl set image deployment/user-service user-service=your-registry/devpulse/user-service:latest
```

### 9. Health Checks

#### Service Health Endpoints
```bash
# Check all services
curl http://api-gateway:8080/healthz
curl http://user-service:8000/healthz
curl http://task-service:8000/healthz
curl http://analytics-service:8000/healthz
curl http://notification-service:8000/healthz
```

#### Load Balancer Health Checks
- [ ] Configure health check endpoints
- [ ] Set appropriate timeouts
- [ ] Configure failure thresholds

### 10. Disaster Recovery

#### High Availability
- [ ] Deploy multiple instances of each service
- [ ] Use load balancers
- [ ] Configure auto-scaling

#### Data Recovery
- [ ] Document recovery procedures
- [ ] Test recovery scenarios
- [ ] Maintain runbooks

## 🔧 Troubleshooting

### Common Issues

1. **JWT Token Issues**
   ```bash
   # Check JWT secret is set
   echo $JWT_SECRET
   ```

2. **Database Connection Issues**
   ```bash
   # Test PostgreSQL connection
   psql -h localhost -U postgres -d devpulse_users
   
   # Test MongoDB connection
   mongo mongodb://localhost:27017/devpulse_tasks
   ```

3. **Service Communication Issues**
   ```bash
   # Check service discovery
   docker-compose exec api-gateway ping user-service
   ```

### Log Analysis
```bash
# View service logs
docker-compose logs -f user-service
docker-compose logs -f api-gateway

# Check Prometheus metrics
curl http://localhost:9090/api/v1/query?query=up
```

## 📊 Performance Monitoring

### Key Metrics to Monitor
- Request latency (p95, p99)
- Error rates
- Database connection pool usage
- Memory and CPU usage
- Disk I/O

### Alerting Rules
```yaml
# Example Prometheus alerting rules
groups:
  - name: devpulse_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
```

## 🔐 Security Checklist

- [ ] All secrets are in environment variables
- [ ] HTTPS is enabled
- [ ] Rate limiting is configured
- [ ] Input validation is implemented
- [ ] SQL injection protection is in place
- [ ] JWT tokens are properly validated
- [ ] Database connections use SSL
- [ ] Logs don't contain sensitive data
- [ ] Regular security updates are applied
- [ ] Access controls are implemented

## 📞 Support

For production issues:
1. Check service logs
2. Verify environment configuration
3. Test health endpoints
4. Review monitoring dashboards
5. Contact the development team

---

**Remember**: Always test changes in a staging environment before deploying to production! 