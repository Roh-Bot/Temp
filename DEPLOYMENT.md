# Deployment Guide

## Local Development

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- Make (optional)

### Setup
```bash
# Clone repository
cd /mnt/d/Rohbot/task-manager

# Install dependencies
go mod download

# Start PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=taskmanager \
  --name postgres \
  postgres:latest

# Run migrations
psql -h localhost -U postgres -d taskmanager -f migrations/001_init.sql

# Update config.yaml (change db host to localhost)
# Run application
go run cmd/blog-api/main.go
```

## Docker Deployment

### Using Docker Compose (Recommended)
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild after code changes
docker-compose up -d --build
```

### Using Docker Only
```bash
# Build image
docker build -t task-manager:latest .

# Run PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=taskmanager \
  --name postgres \
  postgres:latest

# Run application
docker run -d -p 8000:8000 \
  --link postgres:db \
  -e DB_HOST=db \
  task-manager:latest
```

## Cloud Deployment

### AWS ECS

1. **Build and push image**
```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build and tag
docker build -t task-manager:latest .
docker tag task-manager:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/task-manager:latest

# Push
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/task-manager:latest
```

2. **Create RDS PostgreSQL instance**
```bash
aws rds create-db-instance \
  --db-instance-identifier task-manager-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --master-username postgres \
  --master-user-password <password> \
  --allocated-storage 20
```

3. **Create ECS task definition**
```json
{
  "family": "task-manager",
  "containerDefinitions": [
    {
      "name": "task-manager",
      "image": "<account-id>.dkr.ecr.us-east-1.amazonaws.com/task-manager:latest",
      "portMappings": [
        {
          "containerPort": 8000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DB_HOST",
          "value": "<rds-endpoint>"
        }
      ]
    }
  ]
}
```

4. **Create ECS service**
```bash
aws ecs create-service \
  --cluster default \
  --service-name task-manager \
  --task-definition task-manager \
  --desired-count 2 \
  --launch-type FARGATE
```

### Google Cloud Run

```bash
# Build and push to GCR
gcloud builds submit --tag gcr.io/<project-id>/task-manager

# Deploy
gcloud run deploy task-manager \
  --image gcr.io/<project-id>/task-manager \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars DB_HOST=<cloud-sql-ip>
```

### Kubernetes

1. **Create deployment**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: task-manager
spec:
  replicas: 3
  selector:
    matchLabels:
      app: task-manager
  template:
    metadata:
      labels:
        app: task-manager
    spec:
      containers:
      - name: task-manager
        image: task-manager:latest
        ports:
        - containerPort: 8000
        env:
        - name: DB_HOST
          value: postgres-service
```

2. **Create service**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: task-manager-service
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8000
  selector:
    app: task-manager
```

3. **Deploy**
```bash
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

## Environment Variables

### Required
- `DB_HOST` - Database host
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name

### Optional
- `SERVER_ADDRESS` - Server address (default: 0.0.0.0:8000)
- `JWT_SECRET` - JWT signing secret
- `JWT_TOKEN_TTL` - Token TTL in minutes (default: 15)
- `AUTO_COMPLETE_MIN` - Auto-complete delay (default: 5)
- `LOG_LEVEL` - Log level (default: info)

## Database Migrations

### Manual Migration
```bash
psql -h <host> -U <user> -d <database> -f migrations/001_init.sql
```

### Automated Migration (on startup)
Migrations are automatically applied when using Docker Compose.

## Monitoring

### Health Check
```bash
curl http://localhost:8000/api/health
```

### Logs
```bash
# Docker Compose
docker-compose logs -f api-go

# Kubernetes
kubectl logs -f deployment/task-manager

# AWS ECS
aws logs tail /ecs/task-manager --follow
```

### Metrics
The application logs include:
- Request latency
- Error rates
- Request counts
- Background worker activity

## Scaling

### Horizontal Scaling
```bash
# Docker Compose
docker-compose up -d --scale api-go=3

# Kubernetes
kubectl scale deployment task-manager --replicas=5

# AWS ECS
aws ecs update-service \
  --cluster default \
  --service task-manager \
  --desired-count 5
```

### Database Scaling
- Use read replicas for read-heavy workloads
- Enable connection pooling (already configured)
- Add database indexes (already included)

## Security Checklist

- [ ] Change default JWT secret
- [ ] Use strong database passwords
- [ ] Enable SSL/TLS for database connections
- [ ] Use HTTPS in production
- [ ] Set up firewall rules
- [ ] Enable rate limiting
- [ ] Regular security updates
- [ ] Implement API key rotation
- [ ] Set up monitoring and alerts
- [ ] Enable audit logging

## Backup and Recovery

### Database Backup
```bash
# Manual backup
pg_dump -h <host> -U <user> taskmanager > backup.sql

# Restore
psql -h <host> -U <user> taskmanager < backup.sql
```

### Automated Backups
- AWS RDS: Enable automated backups
- Google Cloud SQL: Enable automated backups
- Self-hosted: Use cron jobs with pg_dump

## Troubleshooting

### Application won't start
```bash
# Check logs
docker-compose logs api-go

# Verify database connection
docker-compose exec api-go ping db

# Check environment variables
docker-compose exec api-go env
```

### Database connection issues
```bash
# Test connection
psql -h <host> -U <user> -d taskmanager

# Check network
docker-compose exec api-go nc -zv db 5432
```

### High memory usage
- Reduce connection pool size in config.yaml
- Adjust logger buffer sizes
- Check for goroutine leaks

## Performance Tuning

### Database
```yaml
Database:
  max_connection_idle_time: 30m
  max_connection_lifetime: 1h
```

### Logger
```yaml
Logger:
  buffer_size: 200000
  batch_size: 10000
  flush_delay: 30ms
```

### Worker
```yaml
AutoCompleteMin: 5  # Adjust based on requirements
```

## Rollback Strategy

### Docker Compose
```bash
# Tag current version
docker tag task-manager:latest task-manager:v1.0.0

# Rollback
docker-compose down
docker-compose up -d task-manager:v1.0.0
```

### Kubernetes
```bash
# Rollback to previous version
kubectl rollout undo deployment/task-manager

# Rollback to specific revision
kubectl rollout undo deployment/task-manager --to-revision=2
```

## Support

For issues or questions:
1. Check logs for error messages
2. Verify configuration settings
3. Test database connectivity
4. Review API documentation at /swagger/index.html
5. Check IMPLEMENTATION.md for architecture details
