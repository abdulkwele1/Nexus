# Quick Setup Checklist

## ✅ What We've Set Up

- [x] Production Dockerfiles (`api.production.Dockerfile`, `ui.production.Dockerfile`)
- [x] ECR repositories (`nexus-api`, `nexus-ui`)
- [x] Images pushed to ECR
- [x] Production docker-compose file (`docker-compose.production.yml`)
- [x] Build script (`scripts/build-and-push.sh`)
- [x] Makefile command (`make refresh-production`)

## 🚀 Quick Start (3 Steps)

### Step 1: Build & Push Images (Local Machine)
```bash
./scripts/build-and-push.sh
```

### Step 2: Deploy on Server
```bash
ssh -i nexus-demo-server.pem ubuntu@35.94.111.25
cd Nexus
sudo make refresh-production
```

### Step 3: Verify
```bash
sudo docker compose -f docker-compose.production.yml ps
curl http://localhost:8080/healthcheck
```

## 📋 Full Setup (First Time Only)

### On Your Local Machine:
1. [ ] Ensure AWS CLI is configured: `aws configure`
2. [ ] Test ECR access: `aws ecr describe-repositories --region us-west-2`
3. [ ] Build and push images: `./scripts/build-and-push.sh`

### On Your Server:
1. [ ] SSH to server: `ssh -i nexus-demo-server.pem ubuntu@35.94.111.25`
2. [ ] Install AWS CLI (if needed): `sudo apt-get install -y awscli`
3. [ ] Configure AWS credentials: `aws configure`
4. [ ] Test ECR login: `aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 654654589486.dkr.ecr.us-west-2.amazonaws.com`
5. [ ] Navigate to project: `cd Nexus && git pull`
6. [ ] Deploy: `sudo make refresh-production`

## 🔄 Regular Deployment Workflow

**When you make code changes:**

1. **Local**: Build and push
   ```bash
   ./scripts/build-and-push.sh
   ```

2. **Server**: Deploy
   ```bash
   ssh -i nexus-demo-server.pem ubuntu@35.94.111.25
   cd Nexus
   sudo make refresh-production
   ```

## 🆘 Common Issues

| Issue | Solution |
|-------|----------|
| ECR login expired | Run ECR login command again |
| Can't pull image | Check AWS credentials and IAM permissions |
| Service won't start | Check logs: `sudo docker compose -f docker-compose.production.yml logs` |
| Out of space | Run: `docker system prune -a` |

## 📚 More Help

- Full guide: `SETUP_GUIDE.md`
- Production setup details: `ecs/EC2_PRODUCTION_SETUP.md`
- ECS deployment (if needed later): `ecs/ECS_DEPLOYMENT.md`

