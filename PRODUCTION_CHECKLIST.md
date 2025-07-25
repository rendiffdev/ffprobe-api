# 🚀 Production Deployment Checklist

This checklist ensures your FFprobe API deployment is production-ready and secure.

## 🔐 Security Configuration

### Authentication & Authorization
- [ ] **Change default API key** from `.env.example`
- [ ] **Change default JWT secret** from `.env.example` 
- [ ] **Set strong database passwords** (PostgreSQL & Redis)
- [ ] **Configure allowed CORS origins** (remove `*` wildcard)
- [ ] **Enable authentication** (`ENABLE_AUTH=true`)
- [ ] **Enable rate limiting** (`ENABLE_RATE_LIMIT=true`)
- [ ] **Review rate limit values** for your use case
- [ ] **Configure trusted proxy IPs** if behind load balancer

### Database Security
- [ ] **Use dedicated database user** with minimal privileges
- [ ] **Enable SSL/TLS** for database connections (`POSTGRES_SSL_MODE=require`)
- [ ] **Restrict database network access** (firewall rules)
- [ ] **Set up database backups** (automated)
- [ ] **Test backup restoration procedure**

### Network Security
- [ ] **Configure firewall rules** (allow only necessary ports)
- [ ] **Set up HTTPS/TLS termination** (SSL certificates)
- [ ] **Disable public access** to database ports (5432, 6379)
- [ ] **Configure reverse proxy** (Nginx included)
- [ ] **Set up DDoS protection** if needed

## 🐳 Docker & Infrastructure

### Container Security
- [ ] **Run containers as non-root user** (already configured)
- [ ] **Update base images** to latest stable versions
- [ ] **Scan images for vulnerabilities** (`docker scout`)
- [ ] **Use specific image tags** (not `latest`)
- [ ] **Mount volumes with appropriate permissions**
- [ ] **Limit container resources** (CPU/memory)

### Environment Configuration
- [ ] **Set production log level** (`LOG_LEVEL=warn` or `error`)
- [ ] **Configure proper timezone** (`TZ=UTC` or your region)
- [ ] **Set maximum file size limits** (`MAX_FILE_SIZE`)
- [ ] **Configure storage provider** (S3, GCS, Azure, or local)
- [ ] **Set up monitoring** (Prometheus + Grafana)

## 📊 Monitoring & Observability

### Health Monitoring
- [ ] **Set up health check endpoints** (`/health`)
- [ ] **Configure external monitoring** (Pingdom, UptimeRobot, etc.)
- [ ] **Set up log aggregation** (ELK stack, Loki, etc.)
- [ ] **Configure alerting** (Slack, email, PagerDuty)
- [ ] **Monitor resource usage** (CPU, memory, disk)

### Metrics & Analytics
- [ ] **Enable Prometheus metrics** (`/metrics`)
- [ ] **Configure Grafana dashboards**
- [ ] **Set up alert rules** for critical metrics
- [ ] **Monitor API response times**
- [ ] **Track error rates and patterns**

## 🔄 Backup & Recovery

### Data Protection
- [ ] **Automated database backups** (daily minimum)
- [ ] **Test backup restoration** (monthly)
- [ ] **Off-site backup storage** (cloud storage)
- [ ] **Document recovery procedures**
- [ ] **Set up media file backups** if using local storage

### Disaster Recovery
- [ ] **Document deployment procedures**
- [ ] **Prepare rollback strategy**
- [ ] **Test complete system restore**
- [ ] **Configure multi-region deployment** (if needed)

## ⚡ Performance & Scaling

### Resource Optimization
- [ ] **Configure database connection pooling**
- [ ] **Set up Redis caching** with appropriate TTL
- [ ] **Optimize Docker resource limits**
- [ ] **Configure CDN** for static assets (if applicable)
- [ ] **Set up load balancing** for multiple instances

### Capacity Planning
- [ ] **Estimate concurrent user load**
- [ ] **Calculate storage requirements**
- [ ] **Plan for traffic spikes**
- [ ] **Set up auto-scaling** (if using Kubernetes)

## 🛠️ Operational Procedures

### Deployment Process
- [ ] **Use version tags** for deployments
- [ ] **Implement blue-green deployment** (if needed)
- [ ] **Set up CI/CD pipeline**
- [ ] **Configure automated testing**
- [ ] **Document rollback procedures**

### Maintenance
- [ ] **Schedule regular updates** (security patches)
- [ ] **Set up log rotation** (prevent disk full)
- [ ] **Configure automatic cleanup** (old files, logs)
- [ ] **Plan maintenance windows**
- [ ] **Document operational procedures**

## ☁️ Cloud Provider Specific

### AWS
- [ ] **Configure IAM roles** with minimal permissions
- [ ] **Use VPC** for network isolation
- [ ] **Set up CloudWatch** monitoring
- [ ] **Configure S3 bucket** for file storage
- [ ] **Use RDS** for managed PostgreSQL
- [ ] **Set up ElastiCache** for managed Redis

### Google Cloud
- [ ] **Configure service accounts** with minimal permissions
- [ ] **Use Cloud SQL** for managed PostgreSQL
- [ ] **Set up Cloud Storage** for file storage
- [ ] **Configure Cloud Monitoring**
- [ ] **Use Memorystore** for managed Redis

### Azure
- [ ] **Configure managed identities**
- [ ] **Use Azure Database** for PostgreSQL
- [ ] **Set up Blob Storage** for file storage
- [ ] **Configure Azure Monitor**
- [ ] **Use Azure Cache** for Redis

## 🧪 Testing & Validation

### Pre-Production Testing
- [ ] **Load testing** with expected traffic
- [ ] **Security scanning** (OWASP ZAP, etc.)
- [ ] **API endpoint testing** (Postman, curl)
- [ ] **File upload testing** (various sizes)
- [ ] **Quality analysis testing** (VMAF, PSNR)

### Production Validation
- [ ] **Health check validation**
- [ ] **Authentication testing**
- [ ] **Rate limiting verification**
- [ ] **Monitoring alerts testing**
- [ ] **Backup restoration testing**

## 📋 Final Deployment

### Go-Live Checklist
- [ ] **All security configurations verified**
- [ ] **Monitoring and alerting active**
- [ ] **Backup procedures tested**
- [ ] **Performance baselines established**
- [ ] **Documentation updated**
- [ ] **Team trained on operations**
- [ ] **Emergency contacts configured**
- [ ] **Rollback plan ready**

### Post-Deployment
- [ ] **Monitor for 24-48 hours** after deployment
- [ ] **Verify all functionality** works as expected
- [ ] **Check logs** for errors or warnings
- [ ] **Validate monitoring alerts**
- [ ] **Document any issues** and resolutions

---

## 🚨 Emergency Procedures

### Service Down
1. Check container status: `docker compose ps`
2. Check logs: `docker compose logs -f ffprobe-api`
3. Restart service: `docker compose restart ffprobe-api`
4. Check health endpoint: `curl http://localhost:8080/health`

### Database Issues
1. Check database connectivity: `docker compose logs postgres`
2. Restore from backup if needed: `/app/scripts/backup.sh`
3. Check disk space: `df -h`
4. Restart database: `docker compose restart postgres`

### High Load
1. Check resource usage: `docker stats`
2. Scale horizontally: Add more API instances
3. Check rate limiting: Review `/metrics` endpoint
4. Implement emergency rate limiting

---

✅ **Complete this checklist before going to production!**