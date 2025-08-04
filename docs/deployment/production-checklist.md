# Production Deployment Checklist

> **Essential validation checklist before deploying FFprobe API to production**

## Pre-Deployment Validation

### 🔐 Security Configuration

- [ ] **API Keys**
  - [ ] Changed default API key (32+ characters)
  - [ ] Stored securely in environment variables
  - [ ] Not committed to version control

- [ ] **JWT Configuration**
  - [ ] JWT secret configured (32+ characters)
  - [ ] Token expiry times set appropriately
  - [ ] Refresh token mechanism tested

- [ ] **Database Security**
  - [ ] Strong PostgreSQL password set
  - [ ] Database user permissions restricted
  - [ ] Connection SSL enabled for production

- [ ] **Network Security**
  - [ ] HTTPS/TLS certificates configured
  - [ ] Firewall rules configured
  - [ ] CORS policy properly set
  - [ ] Rate limiting enabled

### ⚙️ Environment Configuration

- [ ] **Server Settings**
  ```bash
  API_PORT=8080
  API_HOST=0.0.0.0
  BASE_URL=https://your-domain.com
  LOG_LEVEL=info
  ```

- [ ] **Database Configuration**
  ```bash
  POSTGRES_HOST=postgres
  POSTGRES_PORT=5432
  POSTGRES_DB=ffprobe_api
  POSTGRES_USER=ffprobe_user
  POSTGRES_PASSWORD=[secure-password]
  ```

- [ ] **Redis Configuration**
  ```bash
  REDIS_HOST=redis
  REDIS_PORT=6379
  REDIS_PASSWORD=[redis-password]
  ```

### 📊 Monitoring Setup

- [ ] **Health Checks**
  - [ ] `/health` endpoint accessible
  - [ ] Database connectivity verified
  - [ ] Redis connectivity verified
  - [ ] FFprobe binary available

- [ ] **Logging**
  - [ ] Log rotation configured
  - [ ] Log retention policy set
  - [ ] Error tracking configured
  - [ ] Audit logging enabled

- [ ] **Metrics**
  - [ ] Prometheus endpoint enabled
  - [ ] Key metrics identified
  - [ ] Alert thresholds configured
  - [ ] Dashboard created

### 🔧 Resource Configuration

- [ ] **Container Resources**
  ```yaml
  resources:
    requests:
      memory: "2Gi"
      cpu: "1000m"
    limits:
      memory: "4Gi"
      cpu: "2000m"
  ```

- [ ] **Storage**
  - [ ] Upload directory configured
  - [ ] Reports directory configured
  - [ ] Adequate disk space allocated
  - [ ] Backup storage configured

- [ ] **Database**
  - [ ] Connection pool size configured
  - [ ] Query timeout set
  - [ ] Indexes optimized
  - [ ] Backup schedule configured

### 🚀 Deployment Verification

- [ ] **Service Health**
  ```bash
  # Check all services
  curl https://your-domain.com/health
  
  # Verify FFprobe
  curl https://your-domain.com/api/v1/probe/health
  ```

- [ ] **Authentication Test**
  ```bash
  # Test API key
  curl -H "X-API-Key: your-api-key" \
       https://your-domain.com/api/v1/probe/file
  
  # Test JWT
  curl -H "Authorization: Bearer your-jwt" \
       https://your-domain.com/api/v1/probe/file
  ```

- [ ] **Basic Functionality**
  ```bash
  # Analyze test video
  curl -X POST https://your-domain.com/api/v1/probe/file \
    -H "X-API-Key: your-api-key" \
    -d '{"file_path": "/test/sample.mp4"}'
  ```

### 📋 Performance Validation

- [ ] **Load Testing**
  - [ ] Concurrent request handling verified
  - [ ] Response time within SLA
  - [ ] Memory usage stable
  - [ ] No resource leaks detected

- [ ] **Scalability**
  - [ ] Horizontal scaling tested
  - [ ] Load balancing configured
  - [ ] Session persistence handled
  - [ ] Cache performance verified

### 🔄 Backup & Recovery

- [ ] **Data Backup**
  - [ ] Database backup automated
  - [ ] Configuration backup created
  - [ ] Recovery procedure documented
  - [ ] Restore process tested

- [ ] **Disaster Recovery**
  - [ ] RTO/RPO defined
  - [ ] Failover procedure documented
  - [ ] Recovery time tested
  - [ ] Communication plan ready

### 📄 Documentation

- [ ] **Operational Docs**
  - [ ] Runbook created
  - [ ] Troubleshooting guide updated
  - [ ] Configuration documented
  - [ ] Architecture diagram current

- [ ] **Team Readiness**
  - [ ] Team trained on system
  - [ ] On-call rotation setup
  - [ ] Escalation path defined
  - [ ] Access permissions granted

## Post-Deployment Verification

### Immediate Checks (First Hour)

- [ ] All services running
- [ ] No error spikes in logs
- [ ] Response times normal
- [ ] Authentication working
- [ ] Basic operations successful

### Day 1 Monitoring

- [ ] 24-hour stability verified
- [ ] Performance metrics reviewed
- [ ] Error rate acceptable
- [ ] Resource usage stable
- [ ] No security alerts

### Week 1 Review

- [ ] Performance baseline established
- [ ] Capacity planning reviewed
- [ ] Optimization opportunities identified
- [ ] Team feedback collected
- [ ] Documentation updated

## Rollback Plan

### Quick Rollback Procedure

1. **Identify Issue**
   ```bash
   # Check service health
   docker compose ps
   docker compose logs --tail=100
   ```

2. **Initiate Rollback**
   ```bash
   # Stop current deployment
   docker compose down
   
   # Restore previous version
   git checkout previous-tag
   docker compose up -d
   ```

3. **Verify Rollback**
   ```bash
   # Check health
   curl http://localhost:8080/health
   
   # Verify functionality
   curl -X POST http://localhost:8080/api/v1/probe/file \
     -H "X-API-Key: your-api-key" \
     -d '{"file_path": "/test/sample.mp4"}'
   ```

## Sign-off

### Deployment Approval

- [ ] Security team approval
- [ ] Operations team approval
- [ ] Development team approval
- [ ] Business stakeholder approval

**Deployment Date**: _______________  
**Deployed By**: _______________  
**Version**: _______________  
**Environment**: _______________

---

## Emergency Contacts

- **On-Call Engineer**: [Contact Info]
- **Team Lead**: [Contact Info]
- **Security Team**: [Contact Info]
- **Database Admin**: [Contact Info]

---

*This checklist must be completed and signed off before production deployment.*