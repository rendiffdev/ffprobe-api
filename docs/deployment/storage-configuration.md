# 💾 Storage Configuration Guide

This guide covers all storage options available in FFprobe API, from local storage to cloud and network-attached storage (NAS) configurations.

## 🎯 Overview

FFprobe API supports multiple storage backends for maximum flexibility:

- **Local Storage** - Direct filesystem access
- **Cloud Storage** - S3, Google Cloud Storage, Azure Blob
- **Network Storage** - SMB/CIFS, NFS, FTP
- **NAS Systems** - Synology, QNAP, TrueNAS, and more

## 🚀 Quick Setup

### Interactive Configuration

The installer will guide you through storage setup:

```bash
cd ffprobe-api
./scripts/setup/install.sh
```

Storage options during installation:
1. **📁 Local filesystem** - Store files locally
2. **☁️ Amazon S3** - AWS cloud storage
3. **🌤️ Google Cloud Storage** - GCP cloud storage  
4. **🔷 Azure Blob Storage** - Microsoft cloud storage
5. **🏠 Network Attached Storage (NAS)** - SMB/CIFS shares
6. **📡 Network File System (NFS)** - Unix/Linux network storage
7. **📤 FTP/SFTP** - File transfer protocol storage
8. **🔗 Custom storage endpoint** - Other S3-compatible services

## 📁 Local Storage

### Configuration

```bash
# Environment variables
STORAGE_PROVIDER=local
STORAGE_BUCKET=./storage          # Local directory path
UPLOAD_DIR=/tmp/uploads           # Upload staging directory
REPORTS_DIR=/tmp/reports          # Generated reports directory
```

### Docker Volume Setup

```yaml
# docker-compose.yml
services:
  ffprobe-api:
    volumes:
      - ./storage:/app/storage:rw
      - ./uploads:/tmp/uploads:rw
      - ./reports:/tmp/reports:rw
```

### Permissions

```bash
# Set proper permissions
sudo chown -R 1000:1000 ./storage ./uploads ./reports
sudo chmod -R 755 ./storage ./uploads ./reports
```

## ☁️ Cloud Storage

### Amazon S3

```bash
# Environment variables
STORAGE_PROVIDER=s3
STORAGE_BUCKET=your-bucket-name
STORAGE_REGION=us-east-1
AWS_ACCESS_KEY_ID=AKIAXXXXXXXXXXXXXXXX
AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Optional: Custom S3 endpoint (for S3-compatible services)
STORAGE_ENDPOINT=https://s3.amazonaws.com
STORAGE_USE_SSL=true
```

#### IAM Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::your-bucket-name",
        "arn:aws:s3:::your-bucket-name/*"
      ]
    }
  ]
}
```

### Google Cloud Storage

```bash
# Environment variables
STORAGE_PROVIDER=gcs
STORAGE_BUCKET=your-gcs-bucket
GCP_SERVICE_ACCOUNT_JSON=/path/to/service-account.json
```

#### Service Account Setup

1. Create service account in Google Cloud Console
2. Grant "Storage Object Admin" role
3. Download JSON key file
4. Mount key file in container:

```yaml
# docker-compose.yml
services:
  ffprobe-api:
    volumes:
      - ./gcp-service-account.json:/app/gcp-service-account.json:ro
    environment:
      - GCP_SERVICE_ACCOUNT_JSON=/app/gcp-service-account.json
```

### Azure Blob Storage

```bash
# Environment variables
STORAGE_PROVIDER=azure
STORAGE_BUCKET=your-container-name
AZURE_STORAGE_ACCOUNT=yourstorageaccount
AZURE_STORAGE_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx==
```

## 🏠 Network Attached Storage (NAS)

### SMB/CIFS Configuration

For Windows shares and most NAS systems:

```bash
# Environment variables
STORAGE_PROVIDER=smb
STORAGE_ENDPOINT=//192.168.1.100/media
STORAGE_ACCESS_KEY=username
STORAGE_SECRET_KEY=password

# Optional: Domain for Active Directory
STORAGE_REGION=DOMAIN
```

#### Docker SMB Mount

```yaml
# docker-compose.yml
services:
  ffprobe-api:
    volumes:
      - smb-storage:/app/storage
    environment:
      - STORAGE_PROVIDER=smb
      - STORAGE_ENDPOINT=//nas.local/ffprobe

volumes:
  smb-storage:
    driver: local
    driver_opts:
      type: cifs
      o: username=ffprobe,password=your-password,uid=1000,gid=1000
      device: //192.168.1.100/media
```

### NFS Configuration

For Unix/Linux network storage:

```bash
# Environment variables
STORAGE_PROVIDER=nfs
STORAGE_ENDPOINT=/mnt/nfs/media    # Mount point inside container
```

#### Docker NFS Mount

```yaml
# docker-compose.yml
services:
  ffprobe-api:
    volumes:
      - nfs-storage:/app/storage
    environment:
      - STORAGE_PROVIDER=nfs

volumes:
  nfs-storage:
    driver: local
    driver_opts:
      type: nfs
      o: addr=192.168.1.100,rw,noatime,rsize=8192,wsize=8192,tcp,timeo=14
      device: ":/export/media"
```

### FTP/SFTP Configuration

```bash
# Environment variables
STORAGE_PROVIDER=ftp
STORAGE_ENDPOINT=ftp.example.com:21
STORAGE_ACCESS_KEY=username
STORAGE_SECRET_KEY=password
STORAGE_USE_SSL=true              # Enable FTPS
```

## 🔧 Popular NAS System Examples

### Synology NAS

```bash
# SMB/CIFS Access
STORAGE_PROVIDER=smb
STORAGE_ENDPOINT=//synology.local/ffprobe
STORAGE_ACCESS_KEY=admin
STORAGE_SECRET_KEY=your-password

# Or NFS Access (if enabled)
STORAGE_PROVIDER=nfs
STORAGE_ENDPOINT=/volume1/ffprobe
```

### QNAP NAS

```bash
# SMB Access
STORAGE_PROVIDER=smb  
STORAGE_ENDPOINT=//qnap.local/Public
STORAGE_ACCESS_KEY=admin
STORAGE_SECRET_KEY=your-password
```

### TrueNAS

```bash
# NFS Access (recommended)
STORAGE_PROVIDER=nfs
STORAGE_ENDPOINT=/mnt/pool1/ffprobe

# SMB Access
STORAGE_PROVIDER=smb
STORAGE_ENDPOINT=//truenas.local/ffprobe
```

### Unraid

```bash
# SMB Access
STORAGE_PROVIDER=smb
STORAGE_ENDPOINT=//unraid.local/media
STORAGE_ACCESS_KEY=your-username
STORAGE_SECRET_KEY=your-password
```

## 📊 Performance Considerations

### Storage Type Performance

| Storage Type | Latency | Throughput | Cost | Best For |
|--------------|---------|------------|------|----------|
| Local SSD | Lowest | Highest | Low | Development, small scale |
| Local HDD | Low | Medium | Low | Large files, archival |
| NAS (Gigabit) | Medium | Medium | Medium | Small teams, home lab |
| NAS (10GbE) | Medium | High | High | Production, large teams |
| Cloud Storage | High | Variable | Variable | Scalability, global access |

### Optimization Tips

#### Local Storage
```bash
# Use SSD for uploads and processing
UPLOAD_DIR=/fast-ssd/uploads
REPORTS_DIR=/fast-ssd/reports

# Use HDD for long-term storage
STORAGE_BUCKET=/bulk-storage/media
```

#### Network Storage
```bash
# Optimize mount options for performance
# NFS example:
o: addr=nas.local,rw,noatime,rsize=65536,wsize=65536,tcp

# SMB example: 
o: username=user,password=pass,vers=3.0,cache=strict
```

#### Cloud Storage
```bash
# Use appropriate region
STORAGE_REGION=us-west-2  # Close to your users/servers

# Enable transfer acceleration (S3)
STORAGE_ENDPOINT=https://s3-accelerate.amazonaws.com
```

## 🔒 Security Best Practices

### Access Control

#### Dedicated Storage User
```bash
# Create dedicated user for FFprobe API
# Linux/NAS systems
useradd -r -s /bin/false ffprobe-api
chown -R ffprobe-api:ffprobe-api /storage/path
```

#### Minimal Permissions
```bash
# Grant only necessary permissions
# Read/write access to storage directory
# No administrative privileges
chmod 750 /storage/ffprobe
```

### Network Security

#### VPN Access
```bash
# Access NAS through VPN for remote deployments
# Use private IP ranges: 192.168.x.x, 10.x.x.x
STORAGE_ENDPOINT=//192.168.1.100/media  # Private IP
```

#### Firewall Rules
```bash
# Restrict storage access to FFprobe API servers only
# Example iptables rule:
iptables -A INPUT -p tcp --dport 445 -s 192.168.1.50 -j ACCEPT
iptables -A INPUT -p tcp --dport 445 -j DROP
```

### Encryption

#### SMB/CIFS Encryption
```bash
# Force SMB encryption
o: username=user,password=pass,seal,vers=3.0
```

#### Cloud Storage Encryption
```bash
# S3 Server-side encryption
STORAGE_SSE=AES256

# Or use KMS key
STORAGE_KMS_KEY_ID=alias/your-key
```

## 🚨 Troubleshooting

### Common Issues

#### Permission Denied
```bash
# Check mount permissions
ls -la /mount/point

# Fix ownership
sudo chown -R 1000:1000 /storage/path

# Check Docker user mapping
docker exec -it ffprobe-api id
```

#### Network Connectivity
```bash
# Test NAS connectivity
ping nas.local
telnet nas.local 445  # SMB
telnet nas.local 2049 # NFS

# Check from container
docker exec ffprobe-api ping nas.local
```

#### Mount Failures
```bash
# Check NFS exports
showmount -e nas.local

# Test SMB connection
smbclient //nas.local/share -U username

# Container logs
docker logs ffprobe-api | grep storage
```

### Debug Commands

```bash
# Check storage configuration
curl -H "X-API-Key: your-key" \
     http://localhost:8080/api/v1/storage/health

# Test file upload
curl -X POST -H "X-API-Key: your-key" \
     -F "file=@test.mp4" \
     http://localhost:8080/api/v1/probe/upload

# Check disk usage
docker exec ffprobe-api df -h
```

## 📈 Monitoring Storage

### Metrics to Monitor

#### Disk Usage
```bash
# Available space
df -h /storage/path

# Inode usage  
df -i /storage/path
```

#### Performance Metrics
```bash
# I/O statistics
iostat -x 1

# Network storage latency
ping -c 10 nas.local
```

### Alerts

Set up monitoring for:
- **Disk space** < 10% available
- **High latency** > 100ms to storage
- **Mount failures** 
- **Permission errors**

## 🔄 Migration Guide

### Local to Cloud
```bash
# 1. Configure cloud storage
STORAGE_PROVIDER=s3
# ... other S3 settings

# 2. Sync existing files
aws s3 sync ./storage/ s3://your-bucket/ --exclude "temp/*"

# 3. Update configuration and restart
docker compose restart ffprobe-api
```

### Between Cloud Providers
```bash
# Example: S3 to GCS
# 1. Install gsutil
# 2. Configure authentication
# 3. Transfer data
gsutil -m cp -r s3://old-bucket gs://new-bucket

# 4. Update configuration
STORAGE_PROVIDER=gcs
STORAGE_BUCKET=new-bucket
```

## 📚 Advanced Configuration

### Multi-Tier Storage

Configure different storage for different data types:

```bash
# Fast storage for active processing
UPLOAD_DIR=/fast-ssd/uploads
REPORTS_DIR=/fast-ssd/reports

# Slow storage for archives  
STORAGE_BUCKET=/slow-storage/archive

# Cloud storage for backups
BACKUP_STORAGE_PROVIDER=s3
BACKUP_STORAGE_BUCKET=backup-bucket
```

### Storage Classes

For cloud storage, use appropriate storage classes:

```bash
# S3 Intelligent Tiering
STORAGE_CLASS=INTELLIGENT_TIERING

# Lifecycle rules for cost optimization
# Transition to cheaper storage after 30 days
```

---

**🚀 Ready to configure storage?** Run the interactive installer:

```bash
./scripts/setup/install.sh
```

The installer will detect your environment and recommend optimal storage configurations!