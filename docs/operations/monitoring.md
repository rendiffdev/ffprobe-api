# Monitoring & Observability Guide

> **Complete guide for monitoring FFprobe API in production environments**

## Overview

FFprobe API provides comprehensive monitoring capabilities through:
- **Health Checks** - Service and dependency health monitoring
- **Prometheus Metrics** - Performance and business metrics
- **Structured Logging** - JSON-formatted logs with correlation IDs
- **Distributed Tracing** - Request flow tracking (optional)

## Health Monitoring

### Health Check Endpoints

#### System Health
```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "checks": {
    "database": "healthy",
    "redis": "healthy",
    "ffprobe": "healthy",
    "storage": "healthy"
  }
}
```

#### FFprobe Service Health
```bash
GET /api/v1/probe/health
```

**Response:**
```json
{
  "status": "healthy",
  "ffprobe_version": "6.1.1",
  "ffmpeg_version": "6.1.1"
}
```

### Health Check Implementation

```go
// Kubernetes liveness probe
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

// Kubernetes readiness probe
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Prometheus Metrics

### Available Metrics

#### HTTP Metrics
```prometheus
# Request rate
http_requests_total{method="POST",endpoint="/api/v1/probe/file",status="200"}

# Request duration
http_request_duration_seconds{method="POST",endpoint="/api/v1/probe/file"}

# Active requests
http_requests_in_flight{endpoint="/api/v1/probe/file"}
```

#### Business Metrics
```prometheus
# Analysis metrics
ffprobe_analyses_total{status="completed"}
ffprobe_analysis_duration_seconds{type="video"}
ffprobe_analysis_file_size_bytes

# Quality metrics
ffprobe_quality_vmaf_score{file_type="mp4"}
ffprobe_quality_checks_total{check_type="enhanced"}
```

#### System Metrics
```prometheus
# Database connections
postgres_connections_active
postgres_connections_idle
postgres_query_duration_seconds

# Redis metrics
redis_commands_total
redis_memory_used_bytes
redis_cache_hits_total
```

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'rendiff-probe'
    static_configs:
      - targets: ['rendiff-probe:8080']
    metrics_path: '/metrics'
```

### Grafana Dashboards

#### API Performance Dashboard
```json
{
  "dashboard": {
    "title": "FFprobe API Performance",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Response Time (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, http_request_duration_seconds)"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])"
          }
        ]
      }
    ]
  }
}
```

## Logging

### Log Format

All logs use structured JSON format:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "service": "rendiff-probe",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user-123",
  "method": "POST",
  "path": "/api/v1/probe/file",
  "status": 200,
  "duration_ms": 1234,
  "message": "Request completed",
  "metadata": {
    "file_size": 10485760,
    "file_type": "mp4",
    "analysis_type": "enhanced"
  }
}
```

### Log Levels

| Level | Usage | Example |
|-------|-------|---------|
| `debug` | Detailed debugging info | SQL queries, detailed traces |
| `info` | Normal operations | Request handling, analysis completion |
| `warn` | Warning conditions | Deprecated API usage, high memory |
| `error` | Error conditions | Failed analysis, database errors |
| `fatal` | Critical failures | Service shutdown, panic recovery |

### Log Aggregation

#### ELK Stack Configuration

```yaml
# logstash.conf
input {
  tcp {
    port => 5000
    codec => json
  }
}

filter {
  json {
    source => "message"
  }
  
  date {
    match => ["timestamp", "ISO8601"]
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "rendiff-probe-%{+YYYY.MM.dd}"
  }
}
```

#### Fluentd Configuration

```yaml
# fluent.conf
<source>
  @type forward
  port 24224
</source>

<filter ffprobe.**>
  @type parser
  format json
  key_name log
</filter>

<match ffprobe.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  logstash_format true
  logstash_prefix ffprobe
</match>
```

## Alerting

### Alert Rules

```yaml
# alerts.yml
groups:
  - name: rendiff-probe
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second"
      
      - alert: SlowResponseTime
        expr: histogram_quantile(0.95, http_request_duration_seconds) > 5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Slow response times"
          description: "95th percentile response time is {{ $value }} seconds"
      
      - alert: DatabaseConnectionPool
        expr: postgres_connections_active / postgres_connections_max > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database connection pool near limit"
          description: "{{ $value }}% of connections in use"
```

### Notification Channels

#### Slack Integration
```yaml
receivers:
  - name: 'slack-notifications'
    slack_configs:
      - api_url: 'YOUR_SLACK_WEBHOOK_URL'
        channel: '#alerts'
        title: 'FFprobe API Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

#### PagerDuty Integration
```yaml
receivers:
  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_KEY'
        description: '{{ .GroupLabels.alertname }}'
```

## Custom Metrics

### Adding Custom Metrics

```go
// Example: Track video processing by codec
var (
    videoProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ffprobe_videos_processed_total",
            Help: "Total number of videos processed by codec",
        },
        []string{"codec", "resolution"},
    )
)

// Register metric
prometheus.MustRegister(videoProcessed)

// Increment counter
videoProcessed.WithLabelValues("h264", "1080p").Inc()
```

### Business KPIs

```prometheus
# Daily active analyses
sum(increase(ffprobe_analyses_total[1d]))

# Average processing time by file size
avg(ffprobe_analysis_duration_seconds) by (file_size_bucket)

# Success rate
sum(rate(ffprobe_analyses_total{status="completed"}[5m])) / 
sum(rate(ffprobe_analyses_total[5m]))
```

## Distributed Tracing (Optional)

### OpenTelemetry Setup

```go
// tracer.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
)

func InitTracer() {
    exp, _ := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
    ))
    
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String("rendiff-probe"),
        )),
    )
    
    otel.SetTracerProvider(tp)
}
```

### Trace Visualization

Access Jaeger UI at `http://localhost:16686` to view:
- Request flow through services
- Latency breakdown by operation
- Error traces and debugging
- Service dependency graph

## Performance Monitoring

### Key Performance Indicators

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| Response Time (p95) | < 2s | > 5s |
| Error Rate | < 1% | > 5% |
| Availability | 99.9% | < 99.5% |
| Throughput | 100 req/s | < 50 req/s |

### Capacity Planning

```prometheus
# Predict storage needs
predict_linear(ffprobe_storage_used_bytes[7d], 86400 * 30)

# Estimate database growth
predict_linear(postgres_database_size_bytes[7d], 86400 * 30)

# Project API usage
predict_linear(rate(http_requests_total[7d])[7d:1h], 86400 * 7)
```

## Monitoring Best Practices

### 1. Dashboard Organization
- **Overview Dashboard**: High-level KPIs and health
- **Performance Dashboard**: Detailed performance metrics
- **Error Dashboard**: Error tracking and debugging
- **Business Dashboard**: Usage and business metrics

### 2. Alert Fatigue Prevention
- Set appropriate thresholds
- Use alert grouping and inhibition
- Implement quiet hours for non-critical alerts
- Regular alert review and tuning

### 3. Log Retention
- **Debug logs**: 3 days
- **Info logs**: 7 days
- **Warning logs**: 30 days
- **Error logs**: 90 days
- **Audit logs**: 1 year

### 4. Metric Cardinality
- Limit label combinations
- Use recording rules for expensive queries
- Implement metric expiry for unused series
- Monitor Prometheus memory usage

## Troubleshooting Monitoring Issues

### Missing Metrics
```bash
# Check Prometheus targets
curl http://prometheus:9090/api/v1/targets

# Verify metrics endpoint
curl http://rendiff-probe:8080/metrics
```

### High Memory Usage
```bash
# Check cardinality
curl http://prometheus:9090/api/v1/label/__name__/values | jq '. | length'

# Top memory consumers
curl http://prometheus:9090/api/v1/query?query=prometheus_tsdb_symbol_table_size_bytes
```

### Log Overflow
```bash
# Check log volume
docker compose -f docker-image/compose.yaml logs rendiff-probe | wc -l

# Adjust log level
export LOG_LEVEL=warn
docker compose -f docker-image/compose.yaml restart rendiff-probe
```

---

## Next Steps

- [Security Guide](security.md)
- [Backup & Recovery](backup.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Production Readiness Checklist](../deployment/PRODUCTION_READINESS_CHECKLIST.md)