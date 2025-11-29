#!/bin/bash

# Cleanup script for ECS resources (if you're using EC2 instead)
# This removes ECS clusters and task definitions, but KEEPS ECR repositories

set -e

REGION="us-west-2"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}ECS Cleanup Script${NC}"
echo "This will delete ECS clusters and task definitions."
echo -e "${RED}WARNING: This will NOT delete ECR repositories (you're using those!)${NC}"
echo ""
read -p "Continue? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# List and delete clusters
echo -e "${YELLOW}Checking for ECS clusters...${NC}"
CLUSTERS=$(aws ecs list-clusters --region $REGION --query 'clusterArns' --output text)

if [ -z "$CLUSTERS" ]; then
    echo -e "${GREEN}No clusters found.${NC}"
else
    for CLUSTER_ARN in $CLUSTERS; do
        CLUSTER_NAME=$(echo $CLUSTER_ARN | cut -d'/' -f2)
        echo -e "${YELLOW}Deleting cluster: $CLUSTER_NAME${NC}"
        
        # Delete all services first (if any)
        SERVICES=$(aws ecs list-services --cluster $CLUSTER_NAME --region $REGION --query 'serviceArns' --output text)
        if [ ! -z "$SERVICES" ]; then
            for SERVICE_ARN in $SERVICES; do
                SERVICE_NAME=$(echo $SERVICE_ARN | cut -d'/' -f3)
                echo "  Stopping service: $SERVICE_NAME"
                aws ecs update-service --cluster $CLUSTER_NAME --service $SERVICE_NAME --desired-count 0 --region $REGION > /dev/null
                aws ecs delete-service --cluster $CLUSTER_NAME --service $SERVICE_NAME --region $REGION > /dev/null
            done
        fi
        
        # Delete cluster
        aws ecs delete-cluster --cluster $CLUSTER_NAME --region $REGION > /dev/null
        echo -e "${GREEN}  ✅ Deleted cluster: $CLUSTER_NAME${NC}"
    done
fi

# List and deregister task definitions
echo -e "${YELLOW}Checking for task definitions...${NC}"
TASK_DEFS=$(aws ecs list-task-definitions --region $REGION --query 'taskDefinitionArns' --output text)

if [ -z "$TASK_DEFS" ]; then
    echo -e "${GREEN}No task definitions found.${NC}"
else
    for TASK_DEF_ARN in $TASK_DEFS; do
        TASK_DEF_FAMILY=$(echo $TASK_DEF_ARN | cut -d'/' -f2 | cut -d':' -f1)
        echo -e "${YELLOW}Deregistering task definition family: $TASK_DEF_FAMILY${NC}"
        
        # Get all revisions
        REVISIONS=$(aws ecs list-task-definitions --family-prefix $TASK_DEF_FAMILY --region $REGION --query 'taskDefinitionArns' --output text)
        for REVISION_ARN in $REVISIONS; do
            REVISION=$(echo $REVISION_ARN | cut -d':' -f7)
            aws ecs deregister-task-definition --task-definition $TASK_DEF_FAMILY:$REVISION --region $REGION > /dev/null
            echo "  ✅ Deregistered: $TASK_DEF_FAMILY:$REVISION"
        done
    done
fi

echo ""
echo -e "${GREEN}✅ Cleanup complete!${NC}"
echo ""
echo -e "${YELLOW}Note: ECR repositories were NOT deleted (you're using those with EC2!)${NC}"
echo "To verify ECR repos: aws ecr describe-repositories --region $REGION"

