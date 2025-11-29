# VPC and Application Load Balancer Setup Guide

This guide explains how to set up VPC networking and an Application Load Balancer for your ECS deployment.

## What You're Setting Up

### 1. VPC (Virtual Private Cloud)
- Your isolated network in AWS
- Contains all your resources (subnets, load balancer, ECS tasks)

### 2. Public Subnets
- Accessible from the internet
- Used for: Application Load Balancer, NAT Gateway
- **Subnets**: `10.0.1.0/24` and `10.0.2.0/24`

### 3. Private Subnets
- Not directly accessible from internet
- Used for: ECS tasks (more secure)
- **Subnets**: `10.0.3.0/24` and `10.0.4.0/24`

### 4. Application Load Balancer (ALB)
- Distributes traffic to your ECS services
- Handles SSL/TLS termination
- Provides health checks
- Routes traffic based on paths (e.g., `/api/*` → API, `/` → UI)

## Quick Setup (Automated)

### Option 1: Use the Script (Recommended)

```bash
# Make script executable
chmod +x ecs/setup-vpc-and-alb.sh

# Run the script
./ecs/setup-vpc-and-alb.sh
```

This script will:
- ✅ Create VPC
- ✅ Create public and private subnets
- ✅ Set up Internet Gateway
- ✅ Create NAT Gateway (for private subnet internet access)
- ✅ Configure route tables
- ✅ Create security groups
- ✅ Create Application Load Balancer

**Time**: ~5-10 minutes (mostly waiting for NAT Gateway)

### Option 2: Manual Setup

Follow the step-by-step commands below.

---

## Manual Setup (Step-by-Step)

### Step 1: Create VPC

```bash
REGION="us-west-2"

# Create VPC
VPC_ID=$(aws ec2 create-vpc \
  --cidr-block 10.0.0.0/16 \
  --region $REGION \
  --query 'Vpc.VpcId' \
  --output text)

# Tag it
aws ec2 create-tags \
  --resources $VPC_ID \
  --tags Key=Name,Value=nexus-vpc \
  --region $REGION

echo "VPC ID: $VPC_ID"
```

### Step 2: Enable DNS

```bash
# Enable DNS hostnames and resolution
aws ec2 modify-vpc-attribute \
  --vpc-id $VPC_ID \
  --enable-dns-hostnames \
  --region $REGION

aws ec2 modify-vpc-attribute \
  --vpc-id $VPC_ID \
  --enable-dns-support \
  --region $REGION
```

### Step 3: Create Internet Gateway

```bash
# Create Internet Gateway
IGW_ID=$(aws ec2 create-internet-gateway \
  --region $REGION \
  --query 'InternetGateway.InternetGatewayId' \
  --output text)

# Attach to VPC
aws ec2 attach-internet-gateway \
  --internet-gateway-id $IGW_ID \
  --vpc-id $VPC_ID \
  --region $REGION

echo "Internet Gateway: $IGW_ID"
```

### Step 4: Create Public Subnets

```bash
# Public Subnet 1 (us-west-2a)
PUBLIC_SUBNET_1=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.1.0/24 \
  --availability-zone ${REGION}a \
  --region $REGION \
  --query 'Subnet.SubnetId' \
  --output text)

# Public Subnet 2 (us-west-2b)
PUBLIC_SUBNET_2=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.2.0/24 \
  --availability-zone ${REGION}b \
  --region $REGION \
  --query 'Subnet.SubnetId' \
  --output text)

echo "Public Subnet 1: $PUBLIC_SUBNET_1"
echo "Public Subnet 2: $PUBLIC_SUBNET_2"
```

### Step 5: Create Private Subnets

```bash
# Private Subnet 1 (us-west-2a)
PRIVATE_SUBNET_1=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.3.0/24 \
  --availability-zone ${REGION}a \
  --region $REGION \
  --query 'Subnet.SubnetId' \
  --output text)

# Private Subnet 2 (us-west-2b)
PRIVATE_SUBNET_2=$(aws ec2 create-subnet \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.4.0/24 \
  --availability-zone ${REGION}b \
  --region $REGION \
  --query 'Subnet.SubnetId' \
  --output text)

echo "Private Subnet 1: $PRIVATE_SUBNET_1"
echo "Private Subnet 2: $PRIVATE_SUBNET_2"
```

### Step 6: Create Route Table for Public Subnets

```bash
# Create route table
PUBLIC_RT=$(aws ec2 create-route-table \
  --vpc-id $VPC_ID \
  --region $REGION \
  --query 'RouteTable.RouteTableId' \
  --output text)

# Add route to internet
aws ec2 create-route \
  --route-table-id $PUBLIC_RT \
  --destination-cidr-block 0.0.0.0/0 \
  --gateway-id $IGW_ID \
  --region $REGION

# Associate with public subnets
aws ec2 associate-route-table \
  --subnet-id $PUBLIC_SUBNET_1 \
  --route-table-id $PUBLIC_RT \
  --region $REGION

aws ec2 associate-route-table \
  --subnet-id $PUBLIC_SUBNET_2 \
  --route-table-id $PUBLIC_RT \
  --region $REGION
```

### Step 7: Create NAT Gateway (for Private Subnets)

```bash
# Allocate Elastic IP
EIP_ALLOCATION=$(aws ec2 allocate-address \
  --domain vpc \
  --region $REGION \
  --query 'AllocationId' \
  --output text)

# Create NAT Gateway (in public subnet)
NAT_GW_ID=$(aws ec2 create-nat-gateway \
  --subnet-id $PUBLIC_SUBNET_1 \
  --allocation-id $EIP_ALLOCATION \
  --region $REGION \
  --query 'NatGateway.NatGatewayId' \
  --output text)

# Wait for NAT Gateway to be available (takes 3-5 minutes)
echo "Waiting for NAT Gateway..."
aws ec2 wait nat-gateway-available \
  --nat-gateway-ids $NAT_GW_ID \
  --region $REGION

echo "NAT Gateway: $NAT_GW_ID"
```

### Step 8: Create Route Table for Private Subnets

```bash
# Create route table
PRIVATE_RT=$(aws ec2 create-route-table \
  --vpc-id $VPC_ID \
  --region $REGION \
  --query 'RouteTable.RouteTableId' \
  --output text)

# Add route through NAT Gateway
aws ec2 create-route \
  --route-table-id $PRIVATE_RT \
  --destination-cidr-block 0.0.0.0.0/0 \
  --gateway-id $NAT_GW_ID \
  --region $REGION

# Associate with private subnets
aws ec2 associate-route-table \
  --subnet-id $PRIVATE_SUBNET_1 \
  --route-table-id $PRIVATE_RT \
  --region $REGION

aws ec2 associate-route-table \
  --subnet-id $PRIVATE_SUBNET_2 \
  --route-table-id $PRIVATE_RT \
  --region $REGION
```

### Step 9: Create Security Groups

```bash
# ALB Security Group (allows HTTP/HTTPS from internet)
ALB_SG=$(aws ec2 create-security-group \
  --group-name nexus-alb-sg \
  --description "Security group for Nexus ALB" \
  --vpc-id $VPC_ID \
  --region $REGION \
  --query 'GroupId' \
  --output text)

# Allow HTTP
aws ec2 authorize-security-group-ingress \
  --group-id $ALB_SG \
  --protocol tcp \
  --port 80 \
  --cidr 0.0.0.0/0 \
  --region $REGION

# Allow HTTPS
aws ec2 authorize-security-group-ingress \
  --group-id $ALB_SG \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0 \
  --region $REGION

# ECS Tasks Security Group (allows traffic from ALB only)
ECS_SG=$(aws ec2 create-security-group \
  --group-name nexus-ecs-sg \
  --description "Security group for Nexus ECS tasks" \
  --vpc-id $VPC_ID \
  --region $REGION \
  --query 'GroupId' \
  --output text)

# Allow API traffic from ALB
aws ec2 authorize-security-group-ingress \
  --group-id $ECS_SG \
  --protocol tcp \
  --port 8080 \
  --source-group $ALB_SG \
  --region $REGION

# Allow UI traffic from ALB
aws ec2 authorize-security-group-ingress \
  --group-id $ECS_SG \
  --protocol tcp \
  --port 80 \
  --source-group $ALB_SG \
  --region $REGION

echo "ALB Security Group: $ALB_SG"
echo "ECS Security Group: $ECS_SG"
```

### Step 10: Create Application Load Balancer

```bash
# Create ALB
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

# Wait for ALB to be active
aws elbv2 wait load-balancer-available \
  --load-balancer-arns $ALB_ARN \
  --region $REGION

# Get ALB DNS name
ALB_DNS=$(aws elbv2 describe-load-balancers \
  --load-balancer-arns $ALB_ARN \
  --region $REGION \
  --query 'LoadBalancers[0].DNSName' \
  --output text)

echo "ALB ARN: $ALB_ARN"
echo "ALB DNS: $ALB_DNS"
```

---

## Next Steps After Setup

### 1. Create Target Groups

You'll need target groups for your API and UI services:

```bash
# API Target Group
API_TG_ARN=$(aws elbv2 create-target-group \
  --name nexus-api-tg \
  --protocol HTTP \
  --port 8080 \
  --vpc-id $VPC_ID \
  --target-type ip \
  --health-check-path /healthcheck \
  --region $REGION \
  --query 'TargetGroups[0].TargetGroupArn' \
  --output text)

# UI Target Group
UI_TG_ARN=$(aws elbv2 create-target-group \
  --name nexus-ui-tg \
  --protocol HTTP \
  --port 80 \
  --vpc-id $VPC_ID \
  --target-type ip \
  --health-check-path / \
  --region $REGION \
  --query 'TargetGroups[0].TargetGroupArn' \
  --output text)
```

### 2. Configure ALB Listeners

```bash
# Create listener for HTTP (port 80)
aws elbv2 create-listener \
  --load-balancer-arn $ALB_ARN \
  --protocol HTTP \
  --port 80 \
  --default-actions Type=forward,TargetGroupArn=$UI_TG_ARN \
  --region $REGION

# Add rule to forward /api/* to API target group
# (This requires the listener ARN from above)
```

### 3. Update ECS Task Definitions

Update your task definitions with:
- Subnet IDs: `$PRIVATE_SUBNET_1`, `$PRIVATE_SUBNET_2`
- Security Group: `$ECS_SG`

### 4. Update ECS Services

When creating ECS services, use:
- Subnets: Private subnets
- Security Groups: ECS security group
- Load Balancer: ALB with target groups

---

## Cost Considerations

| Resource | Monthly Cost |
|----------|-------------|
| NAT Gateway | ~$32/month + data transfer |
| Application Load Balancer | ~$16/month + LCU charges |
| Elastic IP (for NAT) | Free (when attached to NAT) |
| **Total Additional** | **~$48-60/month** |

**Note**: NAT Gateway is required for private subnets to access the internet (to pull ECR images). If cost is a concern, you could use public subnets for ECS tasks, but this is less secure.

---

## Verification

Check your setup:

```bash
# List VPCs
aws ec2 describe-vpcs --filters "Name=tag:Name,Values=nexus-vpc" --region us-west-2

# List subnets
aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --region us-west-2

# List load balancers
aws elbv2 describe-load-balancers --region us-west-2

# List security groups
aws ec2 describe-security-groups --filters "Name=vpc-id,Values=$VPC_ID" --region us-west-2
```

---

## Cleanup (If Needed)

To delete everything:

```bash
# Delete ALB
aws elbv2 delete-load-balancer --load-balancer-arn $ALB_ARN --region us-west-2

# Delete NAT Gateway
aws ec2 delete-nat-gateway --nat-gateway-id $NAT_GW_ID --region us-west-2

# Release Elastic IP
aws ec2 release-address --allocation-id $EIP_ALLOCATION --region us-west-2

# Delete subnets
aws ec2 delete-subnet --subnet-id $PUBLIC_SUBNET_1 --region us-west-2
aws ec2 delete-subnet --subnet-id $PUBLIC_SUBNET_2 --region us-west-2
aws ec2 delete-subnet --subnet-id $PRIVATE_SUBNET_1 --region us-west-2
aws ec2 delete-subnet --subnet-id $PRIVATE_SUBNET_2 --region us-west-2

# Detach and delete Internet Gateway
aws ec2 detach-internet-gateway --internet-gateway-id $IGW_ID --vpc-id $VPC_ID --region us-west-2
aws ec2 delete-internet-gateway --internet-gateway-id $IGW_ID --region us-west-2

# Delete VPC
aws ec2 delete-vpc --vpc-id $VPC_ID --region us-west-2
```

---

## Troubleshooting

**Issue**: NAT Gateway taking too long
- **Solution**: NAT Gateways take 3-5 minutes to become available. Be patient!

**Issue**: Can't create resources
- **Solution**: Check your AWS account limits and IAM permissions

**Issue**: ALB not accessible
- **Solution**: Verify security groups allow traffic on ports 80/443

**Issue**: ECS tasks can't pull images
- **Solution**: Ensure NAT Gateway is in public subnet and route table is configured

---

Need help? The automated script handles all of this for you!

