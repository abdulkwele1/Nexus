# Quick Start: Deploy Nexus to ECS

## Prerequisites Checklist

- [ ] AWS CLI installed and configured
- [ ] ECR repositories exist (`nexus-api`, `nexus-ui`)
- [ ] Docker images pushed to ECR
- [ ] IAM roles created (`ecsTaskExecutionRole`, `ecsTaskRole`)
- [ ] VPC and subnets configured
- [ ] Security groups created
- [ ] Secrets stored in AWS Secrets Manager
- [ ] Application Load Balancer created (optional)

## Quick Commands

### 1. Register Task Definitions

```bash
cd ecs/
aws ecs register-task-definition \
  --cli-input-json file://nexus-api-task-definition.json \
  --region us-west-2

aws ecs register-task-definition \
  --cli-input-json file://nexus-ui-task-definition.json \
  --region us-west-2
```

### 2. Create Services

**API Service:**
```bash
aws ecs create-service \
  --cluster nexus-cluster \
  --service-name nexus-api-service \
  --task-definition nexus-api \
  --desired-count 1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx,subnet-yyy],securityGroups=[sg-xxx],assignPublicIp=DISABLED}" \
  --region us-west-2
```

**UI Service:**
```bash
aws ecs create-service \
  --cluster nexus-cluster \
  --service-name nexus-ui-service \
  --task-definition nexus-ui \
  --desired-count 1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx,subnet-yyy],securityGroups=[sg-xxx],assignPublicIp=DISABLED}" \
  --region us-west-2
```

### 3. Update Services (After Pushing New Images)

```bash
# Update API
aws ecs update-service \
  --cluster nexus-cluster \
  --service nexus-api-service \
  --force-new-deployment \
  --region us-west-2

# Update UI
aws ecs update-service \
  --cluster nexus-cluster \
  --service nexus-ui-service \
  --force-new-deployment \
  --region us-west-2
```

### 4. View Logs

```bash
# API logs
aws logs tail /ecs/nexus-api --follow --region us-west-2

# UI logs
aws logs tail /ecs/nexus-ui --follow --region us-west-2
```

### 5. Check Service Status

```bash
aws ecs describe-services \
  --cluster nexus-cluster \
  --services nexus-api-service nexus-ui-service \
  --region us-west-2
```

## Important Notes

1. **Update ARNs**: Before using the task definitions, update the IAM role ARNs and Secrets Manager ARNs in the JSON files
2. **Subnet IDs**: Replace `subnet-xxx` and `subnet-yyy` with your actual subnet IDs
3. **Security Group**: Replace `sg-xxx` with your actual security group ID
4. **Secrets**: Ensure all secrets exist in Secrets Manager with the correct names

## Next Steps

See `ECS_DEPLOYMENT.md` for detailed setup instructions including:
- Creating IAM roles
- Setting up VPC and networking
- Creating security groups
- Configuring load balancer
- Storing secrets

