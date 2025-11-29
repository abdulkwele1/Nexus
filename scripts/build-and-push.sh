#!/bin/bash

# Build and push Nexus images to ECR
# Run this from your local machine or CI/CD

set -e

REGION="us-west-2"
ACCOUNT_ID="654654589486"
ECR_REGISTRY="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Building and pushing Nexus images to ECR...${NC}"

# Step 1: Login to ECR
echo -e "${YELLOW}Logging into ECR...${NC}"
aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $ECR_REGISTRY

# Step 2: Build API image
echo -e "${YELLOW}Building API image...${NC}"
docker build -f api.production.Dockerfile -t nexus-api:latest .

# Step 3: Build UI image
echo -e "${YELLOW}Building UI image...${NC}"
docker build -f ui.production.Dockerfile -t nexus-ui:latest .

# Step 4: Tag images
echo -e "${YELLOW}Tagging images for ECR...${NC}"
docker tag nexus-api:latest ${ECR_REGISTRY}/nexus-api:latest
docker tag nexus-ui:latest ${ECR_REGISTRY}/nexus-ui:latest

# Step 5: Push images
echo -e "${YELLOW}Pushing API image to ECR...${NC}"
docker push ${ECR_REGISTRY}/nexus-api:latest

echo -e "${YELLOW}Pushing UI image to ECR...${NC}"
docker push ${ECR_REGISTRY}/nexus-ui:latest

echo -e "${GREEN}✅ Successfully built and pushed images to ECR!${NC}"
echo -e "${GREEN}You can now deploy on your server using: sudo make refresh-production${NC}"

