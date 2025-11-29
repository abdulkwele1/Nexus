#!/bin/bash

# Setup VPC, Subnets, and Application Load Balancer for Nexus ECS
# This script creates the networking infrastructure needed for ECS deployment

set -e

REGION="us-west-2"
VPC_CIDR="10.0.0.0/16"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}Setting up VPC and Application Load Balancer for Nexus...${NC}"
echo ""

# Step 1: Create VPC
echo -e "${YELLOW}Step 1: Creating VPC...${NC}"
VPC_ID=$(aws ec2 create-vpc \
  --cidr-block $VPC_CIDR \
  --region $REGION \
  --query 'Vpc.VpcId' \
  --output text)

aws ec2 create-tags \
  --resources $VPC_ID \
  --tags Key=Name,Value=nexus-vpc \
  --region $REGION

echo -e "${GREEN}✅ VPC created: $VPC_ID${NC}"

# Step 2: Enable DNS hostnames and DNS resolution
echo -e "${YELLOW}Step 2: Enabling DNS...${NC}"
aws ec2 modify-vpc-attribute \
  --vpc-id $VPC_ID \
  --enable-dns-hostnames \
  --region $REGION

aws ec2 modify-vpc-attribute \
  --vpc-id $VPC_ID \
  --enable-dns-support \
  --region $REGION

# Step 3: Create Internet Gateway
echo -e "${YELLOW}Step 3: Creating Internet Gateway...${NC}"
IGW_ID=$(aws ec2 create-internet-gateway \
  --region $REGION \
  --query 'InternetGateway.InternetGatewayId' \
  --output text)

aws ec2 create-tags \
  --resources $IGW_ID \
  --tags Key=Name,Value=nexus-igw \
  --region $REGION

aws ec2 attach-internet-gateway \
  --internet-gateway-id $IGW_ID \
  --vpc-id $VPC_ID \
  --region $REGION

echo -e "${GREEN}✅ Internet Gateway created and attached: $IGW_ID${NC}"

# Step 4: Create Public Subnets
echo -e "${YELLOW}Step 4: Creating Public Subnets...${NC}"
PUBLIC_SUBNET_1=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.1.0/24 \
  --availability-zone ${REGION}a \
  --region $REGION \
  --query 'Subnet.SubnetId' \
  --output text)

aws ec2 create-tags \
  --resources $PUBLIC_SUBNET_1 \
  --tags Key=Name,Value=nexus-public-subnet-1 \
  --region $REGION

PUBLIC_SUBNET_2=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.2.0/24 \
  --availability-zone ${REGION}b \
  --region $REGION \
  --query 'Subnet.SubnetId' \
  --output text)

aws ec2 create-tags \
  --resources $PUBLIC_SUBNET_2 \
  --tags Key=Name,Value=nexus-public-subnet-2 \
  --region $REGION

echo -e "${GREEN}✅ Public Subnets created:${NC}"
echo "   - Public Subnet 1 (${REGION}a): $PUBLIC_SUBNET_1"
echo "   - Public Subnet 2 (${REGION}b): $PUBLIC_SUBNET_2"

# Step 5: Create Private Subnets
echo -e "${YELLOW}Step 5: Creating Private Subnets...${NC}"
PRIVATE_SUBNET_1=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.3.0/24 \
  --availability-zone ${REGION}a \
  --region $REGION \
  --query 'Subnet.SubnetId' \
  --output text)

aws ec2 create-tags \
  --resources $PRIVATE_SUBNET_1 \
  --tags Key=Name,Value=nexus-private-subnet-1 \
  --region $REGION

PRIVATE_SUBNET_2=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.4.0/24 \
  --availability-zone ${REGION}b \
  --region $REGION \
  --query 'Subnet.SubnetId' \
  --output text)

aws ec2 create-tags \
  --resources $PRIVATE_SUBNET_2 \
  --tags Key=Name,Value=nexus-private-subnet-2 \
  --region $REGION

echo -e "${GREEN}✅ Private Subnets created:${NC}"
echo "   - Private Subnet 1 (${REGION}a): $PRIVATE_SUBNET_1"
echo "   - Private Subnet 2 (${REGION}b): $PRIVATE_SUBNET_2"

# Step 6: Create Route Table for Public Subnets
echo -e "${YELLOW}Step 6: Creating Route Table for Public Subnets...${NC}"
PUBLIC_RT=$(aws ec2 create-route-table \
  --vpc-id $VPC_ID \
  --region $REGION \
  --query 'RouteTable.RouteTableId' \
  --output text)

aws ec2 create-tags \
  --resources $PUBLIC_RT \
  --tags Key=Name,Value=nexus-public-rt \
  --region $REGION

aws ec2 create-route \
  --route-table-id $PUBLIC_RT \
  --destination-cidr-block 0.0.0.0/0 \
  --gateway-id $IGW_ID \
  --region $REGION

aws ec2 associate-route-table \
  --subnet-id $PUBLIC_SUBNET_1 \
  --route-table-id $PUBLIC_RT \
  --region $REGION

aws ec2 associate-route-table \
  --subnet-id $PUBLIC_SUBNET_2 \
  --route-table-id $PUBLIC_RT \
  --region $REGION

echo -e "${GREEN}✅ Public Route Table created and associated${NC}"

# Step 7: Create NAT Gateway (for private subnet internet access)
echo -e "${YELLOW}Step 7: Creating NAT Gateway (this may take a few minutes)...${NC}"
# Allocate Elastic IP
EIP_ALLOCATION=$(aws ec2 allocate-address \
  --domain vpc \
  --region $REGION \
  --query 'AllocationId' \
  --output text)

echo "   Elastic IP allocated: $EIP_ALLOCATION"

# Create NAT Gateway in first public subnet
NAT_GW_ID=$(aws ec2 create-nat-gateway \
  --subnet-id $PUBLIC_SUBNET_1 \
  --allocation-id $EIP_ALLOCATION \
  --region $REGION \
  --query 'NatGateway.NatGatewayId' \
  --output text)

echo "   Waiting for NAT Gateway to be available (this takes 3-5 minutes)..."
aws ec2 wait nat-gateway-available \
  --nat-gateway-ids $NAT_GW_ID \
  --region $REGION

echo -e "${GREEN}✅ NAT Gateway created: $NAT_GW_ID${NC}"

# Step 8: Create Route Table for Private Subnets
echo -e "${YELLOW}Step 8: Creating Route Table for Private Subnets...${NC}"
PRIVATE_RT=$(aws ec2 create-route-table \
  --vpc-id $VPC_ID \
  --region $REGION \
  --query 'RouteTable.RouteTableId' \
  --output text)

aws ec2 create-tags \
  --resources $PRIVATE_RT \
  --tags Key=Name,Value=nexus-private-rt \
  --region $REGION

aws ec2 create-route \
  --route-table-id $PRIVATE_RT \
  --destination-cidr-block 0.0.0.0/0 \
  --gateway-id $NAT_GW_ID \
  --region $REGION

aws ec2 associate-route-table \
  --subnet-id $PRIVATE_SUBNET_1 \
  --route-table-id $PRIVATE_RT \
  --region $REGION

aws ec2 associate-route-table \
  --subnet-id $PRIVATE_SUBNET_2 \
  --route-table-id $PRIVATE_RT \
  --region $REGION

echo -e "${GREEN}✅ Private Route Table created and associated${NC}"

# Step 9: Create Security Groups
echo -e "${YELLOW}Step 9: Creating Security Groups...${NC}"

# ALB Security Group
ALB_SG=$(aws ec2 create-security-group \
  --group-name nexus-alb-sg \
  --description "Security group for Nexus Application Load Balancer" \
  --vpc-id $VPC_ID \
  --region $REGION \
  --query 'GroupId' \
  --output text)

aws ec2 authorize-security-group-ingress \
  --group-id $ALB_SG \
  --protocol tcp \
  --port 80 \
  --cidr 0.0.0.0/0 \
  --region $REGION

aws ec2 authorize-security-group-ingress \
  --group-id $ALB_SG \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0 \
  --region $REGION

echo -e "${GREEN}✅ ALB Security Group created: $ALB_SG${NC}"

# ECS Tasks Security Group
ECS_SG=$(aws ec2 create-security-group \
  --group-name nexus-ecs-sg \
  --description "Security group for Nexus ECS tasks" \
  --vpc-id $VPC_ID \
  --region $REGION \
  --query 'GroupId' \
  --output text)

# Allow traffic from ALB to ECS tasks
aws ec2 authorize-security-group-ingress \
  --group-id $ECS_SG \
  --protocol tcp \
  --port 8080 \
  --source-group $ALB_SG \
  --region $REGION

aws ec2 authorize-security-group-ingress \
  --group-id $ECS_SG \
  --protocol tcp \
  --port 80 \
  --source-group $ALB_SG \
  --region $REGION

echo -e "${GREEN}✅ ECS Security Group created: $ECS_SG${NC}"

# Step 10: Create Application Load Balancer
echo -e "${YELLOW}Step 10: Creating Application Load Balancer...${NC}"
ALB_ARN=$(aws elbv2 create-load-balancer \
  --name nexus-alb \
  --subnets $PUBLIC_SUBNET_1 $PUBLIC_SUBNET_2 \
  --security-groups $ALB_SG \
  --scheme internet-facing \
  --type application \
  --ip-address-type ipv4 \
  --region $REGION \
  --query 'LoadBalancers[0].LoadBalancerArn' \
  --output text)

echo "   Waiting for ALB to be active..."
aws elbv2 wait load-balancer-available \
  --load-balancer-arns $ALB_ARN \
  --region $REGION

ALB_DNS=$(aws elbv2 describe-load-balancers \
  --load-balancer-arns $ALB_ARN \
  --region $REGION \
  --query 'LoadBalancers[0].DNSName' \
  --output text)

echo -e "${GREEN}✅ Application Load Balancer created${NC}"
echo "   ALB ARN: $ALB_ARN"
echo "   ALB DNS: $ALB_DNS"

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ Setup Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "VPC ID: $VPC_ID"
echo "Internet Gateway: $IGW_ID"
echo "NAT Gateway: $NAT_GW_ID"
echo ""
echo "Public Subnets:"
echo "  - $PUBLIC_SUBNET_1 (${REGION}a)"
echo "  - $PUBLIC_SUBNET_2 (${REGION}b)"
echo ""
echo "Private Subnets:"
echo "  - $PRIVATE_SUBNET_1 (${REGION}a)"
echo "  - $PRIVATE_SUBNET_2 (${REGION}b)"
echo ""
echo "Security Groups:"
echo "  - ALB: $ALB_SG"
echo "  - ECS: $ECS_SG"
echo ""
echo "Application Load Balancer:"
echo "  - DNS: $ALB_DNS"
echo "  - ARN: $ALB_ARN"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Create target groups for your services"
echo "2. Configure ALB listeners and rules"
echo "3. Update your ECS task definitions with subnet and security group IDs"
echo ""
echo -e "${YELLOW}Save these values - you'll need them for ECS service creation!${NC}"

