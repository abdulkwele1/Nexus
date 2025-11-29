# Production Setup Guide - Step by Step

This guide will help you set up the optimized EC2 production deployment using ECR images.

## Prerequisites Checklist

- [x] ECR repositories created (`nexus-api`, `nexus-ui`)
- [x] Docker images already pushed to ECR (we did this earlier)
- [ ] AWS CLI configured on your local machine
- [ ] Access to your EC2 server

## Step 1: Build and Push Images (Local Machine)

### Option A: Use the Script (Easiest)

```bash
# Make script executable
chmod +x scripts/build-and-push.sh

# Run the script
./scripts/build-and-push.sh
```

### Option B: Manual Commands

```bash
# Login to ECR
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com

# Build API image
docker build -f api.production.Dockerfile -t nexus-api:latest .

# Build UI image  
docker build -f ui.production.Dockerfile -t nexus-ui:latest .

# Tag images
docker tag nexus-api:latest 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-api:latest
docker tag nexus-ui:latest 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-ui:latest

# Push images
docker push 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-api:latest
docker push 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-ui:latest
```

**✅ Checkpoint**: Images are now in ECR and ready to deploy.

---

## Step 2: Set Up Your EC2 Server

### 2.1 SSH to Your Server

```bash
ssh -i nexus-demo-server.pem ubuntu@35.94.111.25
```

### 2.2 Install/Verify AWS CLI

```bash
# Check if AWS CLI is installed
aws --version

# If not installed:
sudo apt-get update
sudo apt-get install -y awscli

# Configure AWS credentials
aws configure
# Enter your:
# - AWS Access Key ID
# - AWS Secret Access Key  
# - Default region: us-west-2
# - Default output format: json
```

**Note**: For better security, consider using an IAM role on the EC2 instance instead of storing credentials.

### 2.3 Navigate to Your Project

```bash
cd Nexus
git pull  # Get latest code including docker-compose.production.yml
```

### 2.4 Test ECR Login

```bash
# Login to ECR (tokens expire after 12 hours)
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com

# Test pulling an image
docker pull 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-api:latest
```

**✅ Checkpoint**: Server can access ECR.

---

## Step 3: Deploy Using Production Images

### 3.1 Stop Current Services (If Running)

```bash
# Stop current services
sudo docker compose down

# Or if using the old compose file:
sudo docker compose -f docker-compose.yml down
```

### 3.2 Deploy with Production Images

**Option A: Using Make (Recommended)**

```bash
sudo make refresh-production
```

**Option B: Manual Commands**

```bash
# Login to ECR
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com

# Pull latest images
docker pull 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-api:latest
docker pull 654654589486.dkr.ecr.us-west-2.amazonaws.com/nexus-ui:latest

# Start services
sudo docker compose -f docker-compose.production.yml up -d
```

### 3.3 Verify Services Are Running

```bash
# Check status
sudo docker compose -f docker-compose.production.yml ps

# Check logs
sudo docker compose -f docker-compose.production.yml logs -f

# Test API health
curl http://localhost:8080/healthcheck
```

**✅ Checkpoint**: Services are running from ECR images!

---

## Step 4: Update Your Deployment Workflow

### New Deployment Process

**On Your Local Machine:**
```bash
# 1. Make code changes and commit
git add .
git commit -m "Your changes"
git push

# 2. Build and push new images
./scripts/build-and-push.sh
```

**On Your Server:**
```bash
# 1. SSH to server
ssh -i nexus-demo-server.pem ubuntu@35.94.111.25

# 2. Pull latest code (optional, for config changes)
cd Nexus
git pull

# 3. Deploy new images
sudo make refresh-production
```

---

## Step 5: Verify Everything Works

### Check Services

```bash
# View running containers
sudo docker compose -f docker-compose.production.yml ps

# Check API health
curl http://localhost:8080/healthcheck

# View logs
sudo docker compose -f docker-compose.production.yml logs nexus-api
sudo docker compose -f docker-compose.production.yml logs nexus-ui
```

### Test Your Application

1. Visit `https://nexus.eternalrelayrace.com`
2. Verify the UI loads
3. Test login functionality
4. Check API endpoints

---

## Troubleshooting

### Issue: "denied: Your authorization token has expired"

**Solution**: Re-login to ECR
```bash
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com
```

### Issue: "Cannot pull image"

**Solution**: Check IAM permissions and ECR repository exists
```bash
# Verify repositories
aws ecr describe-repositories --region us-west-2

# Check your AWS credentials
aws sts get-caller-identity
```

### Issue: "Service won't start"

**Solution**: Check environment variables and logs
```bash
# Check logs
sudo docker compose -f docker-compose.production.yml logs

# Verify .env file exists and has correct values
cat .env
```

### Issue: "Out of disk space"

**Solution**: Clean up old images
```bash
# Remove unused images
docker system prune -a

# Remove old ECR images (optional, from local machine)
aws ecr list-images --repository-name nexus-api --region us-west-2
```

---

## Maintenance

### Regular Tasks

1. **Update Images** (when you push new code):
   ```bash
   # On local: build and push
   ./scripts/build-and-push.sh
   
   # On server: pull and restart
   sudo make refresh-production
   ```

2. **Check Logs** (weekly):
   ```bash
   sudo docker compose -f docker-compose.production.yml logs --tail=100
   ```

3. **Update ECR Login** (every 12 hours if needed):
   ```bash
   aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com
   ```

4. **Backup Database** (regularly):
   ```bash
   sudo docker compose -f docker-compose.production.yml exec nexus-db pg_dump -U postgres postgres > backup.sql
   ```

---

## Next Steps

- [ ] Set up automated builds (GitHub Actions, etc.)
- [ ] Configure IAM role for EC2 (instead of credentials)
- [ ] Set up CloudWatch monitoring
- [ ] Configure automated backups
- [ ] Set up SSL certificate renewal

---

## Quick Reference

| Task | Command |
|------|---------|
| Build & push images | `./scripts/build-and-push.sh` |
| Deploy on server | `sudo make refresh-production` |
| View logs | `sudo docker compose -f docker-compose.production.yml logs -f` |
| Check status | `sudo docker compose -f docker-compose.production.yml ps` |
| Restart service | `sudo docker compose -f docker-compose.production.yml restart nexus-api` |
| ECR login | `aws ecr get-login-password --region us-west-2 \| docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com` |

---

Need help? Check the troubleshooting section or review `ecs/EC2_PRODUCTION_SETUP.md` for more details.

