# Optimized EC2 Production Setup for Nexus

## Recommended Setup for 5-10 Users

**Use EC2 with ECR images** - Best balance of cost, simplicity, and reliability.

## Benefits of This Approach

✅ **Cost**: ~$30-50/month (vs $65-80/month with ECS)  
✅ **Speed**: Faster deployments (pull images vs building)  
✅ **Reliability**: Pre-built, tested images  
✅ **Simplicity**: Keep your current workflow  
✅ **Scalability**: Easy to upgrade instance if needed  

## Setup Instructions

### 1. On Your Local Machine (or CI/CD)

Build and push images to ECR:

```bash
# Build production images
docker build -f api.production.Dockerfile -t nexus-api:latest .
docker build -f ui.production.Dockerfile -t nexus-ui:latest .

# Tag for ECR
docker tag nexus-api:latest 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-api:latest
docker tag nexus-ui:latest 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-ui:latest

# Login to ECR
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com

# Push images
docker push 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-api:latest
docker push 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-ui:latest
```

### 2. On Your EC2 Server

#### Initial Setup (One Time)

```bash
# SSH to server
ssh -i nexus-demo-server.pem ubuntu@35.94.111.25

# Install AWS CLI (if not already installed)
sudo apt-get update
sudo apt-get install -y awscli

# Configure AWS credentials (or use IAM role)
aws configure

# Login to ECR
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com

# Pull the production docker-compose file
cd Nexus
git pull
```

#### Deploy Using ECR Images

```bash
# Pull latest images from ECR
docker pull 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-api:latest
docker pull 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-ui:latest

# Deploy using production compose file
sudo docker compose -f docker-compose.production.yml up -d --force-recreate

# Or update just one service
sudo docker compose -f docker-compose.production.yml up -d nexus-api --force-recreate
sudo docker compose -f docker-compose.production.yml up -d nexus-ui --force-recreate
```

### 3. Update Your Makefile (Optional)

Add a production refresh target:

```makefile
.PHONY: refresh-production
# Pull latest images from ECR and restart services
refresh-production:
	aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com
	docker pull 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-api:latest
	docker pull 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-ui:latest
	docker compose -f docker-compose.production.yml up -d --force-recreate
```

Then deploy with:
```bash
sudo make refresh-production
```

## Deployment Workflow

### Current Workflow (Building on Server)
1. SSH to server
2. `git pull`
3. `sudo make refresh` (builds on server - slow)

### Optimized Workflow (Using ECR)
1. **On your machine**: Build images, push to ECR
2. **On server**: Pull images, restart containers (fast!)

## Cost Breakdown

| Component | Cost |
|-----------|------|
| EC2 t3.medium | ~$30/month |
| ECR storage | ~$0.10/month (minimal) |
| Data transfer | ~$1-5/month |
| **Total** | **~$30-35/month** |

## When to Consider ECS

Consider ECS Fargate when:
- You have **50+ concurrent users**
- You need **automatic scaling**
- You need **multi-region deployment**
- You have **multiple environments** (dev/staging/prod)
- Budget allows for **2-3x cost increase**

For 5-10 users: **EC2 is perfect!**

## Monitoring & Maintenance

### Check Service Status
```bash
sudo docker compose -f docker-compose.production.yml ps
```

### View Logs
```bash
sudo docker compose -f docker-compose.production.yml logs -f
```

### Update Images
```bash
# Re-authenticate (tokens expire after 12 hours)
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com

# Pull and restart
sudo docker compose -f docker-compose.production.yml pull
sudo docker compose -f docker-compose.production.yml up -d
```

## Security Best Practices

1. **Use IAM Role** instead of AWS credentials on server
2. **Keep images updated** - rebuild and push regularly
3. **Use secrets management** for sensitive env vars
4. **Enable CloudWatch logs** (optional but recommended)
5. **Regular backups** of database volume

## Troubleshooting

**Images won't pull:**
- Check ECR login token (expires after 12 hours)
- Verify IAM permissions

**Services won't start:**
- Check environment variables in `.env`
- Check logs: `sudo docker compose -f docker-compose.production.yml logs`

**Out of disk space:**
- Clean old images: `docker system prune -a`

