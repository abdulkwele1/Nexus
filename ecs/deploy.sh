#!/bin/bash

# Nexus ECS Deployment Script
# This script sets up ECS cluster, services, and task definitions

set -e

REGION="us-west-2"
CLUSTER_NAME="nexus-cluster"
ACCOUNT_ID="654654589486"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Nexus ECS Deployment...${NC}"

# Step 1: Create CloudWatch Log Groups
echo -e "${YELLOW}Creating CloudWatch Log Groups...${NC}"
aws logs create-log-group --log-group-name /ecs/nexus-api --region $REGION 2>/dev/null || echo "Log group /ecs/nexus-api already exists"
aws logs create-log-group --log-group-name /ecs/nexus-ui --region $REGION 2>/dev/null || echo "Log group /ecs/nexus-ui already exists"

# Step 2: Create ECS Cluster
echo -e "${YELLOW}Creating ECS Cluster...${NC}"
aws ecs create-cluster \
  --cluster-name $CLUSTER_NAME \
  --region $REGION \
  2>/dev/null || echo "Cluster $CLUSTER_NAME already exists"

# Step 3: Register Task Definitions
echo -e "${YELLOW}Registering Task Definitions...${NC}"
aws ecs register-task-definition \
  --cli-input-json file://nexus-api-task-definition.json \
  --region $REGION

aws ecs register-task-definition \
  --cli-input-json file://nexus-ui-task-definition.json \
  --region $REGION

echo -e "${GREEN}Task definitions registered successfully!${NC}"
echo -e "${YELLOW}Note: You still need to:${NC}"
echo -e "  1. Create IAM roles (ecsTaskExecutionRole and ecsTaskRole)"
echo -e "  2. Create VPC, subnets, and security groups"
echo -e "  3. Create Application Load Balancer (optional but recommended)"
echo -e "  4. Create ECS services using the task definitions"
echo -e "  5. Store secrets in AWS Secrets Manager"
echo ""
echo -e "See ECS_DEPLOYMENT.md for detailed instructions."

