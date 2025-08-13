# Terraform Infrastructure

This directory contains Terraform configurations for setting up the AWS infrastructure needed for the Ardan Labs service.

## Infrastructure Components

- **VPC** - Virtual Private Cloud with public and private subnets
- **EKS Clusters** - Kubernetes clusters for staging and production
- **ECR Repositories** - Container registries for Docker images
- **RDS** - PostgreSQL database (optional)
- **ALB** - Application Load Balancer for external access

## Prerequisites

1. **Terraform** (>= 1.0)
2. **AWS CLI** configured with appropriate credentials
3. **kubectl** for Kubernetes management

## Quick Start

### 1. Initialize Terraform

```bash
cd terraform
terraform init
```

### 2. Configure Variables

```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values
```

### 3. Plan the Infrastructure

```bash
terraform plan
```

### 4. Apply the Infrastructure

```bash
terraform apply
```

### 5. Get Outputs

```bash
terraform output
```

## Required Secrets for GitHub Actions

After running Terraform, you'll get the following outputs that need to be configured as GitHub secrets:

```bash
# Get AWS Account ID
terraform output aws_account_id

# Get AWS Region
terraform output aws_region

# Get EKS Cluster Names
terraform output staging_cluster_name
terraform output production_cluster_name

# Get URLs
terraform output staging_url
terraform output production_url
```

## GitHub Secrets Configuration

Add these secrets to your GitHub repository (Settings → Secrets and variables → Actions):

- `AWS_ACCOUNT_ID` - Your AWS account ID
- `AWS_REGION` - AWS region (e.g., us-west-2)
- `AWS_ACCESS_KEY_ID` - AWS access key for GitHub Actions
- `AWS_SECRET_ACCESS_KEY` - AWS secret key for GitHub Actions
- `EKS_STAGING_CLUSTER` - Staging EKS cluster name
- `EKS_PRODUCTION_CLUSTER` - Production EKS cluster name
- `STAGING_URL` - Staging environment URL
- `PRODUCTION_URL` - Production environment URL

## Alternative: Manual Setup (No Terraform)

If you don't want to use Terraform, you can manually create the infrastructure:

### 1. Create EKS Clusters

```bash
# Staging cluster
eksctl create cluster --name ardanlabs-service-staging --region us-west-2 --nodes 2

# Production cluster
eksctl create cluster --name ardanlabs-service-production --region us-west-2 --nodes 3
```

### 2. Create ECR Repositories

```bash
aws ecr create-repository --repository-name ardanlabs/service-sales
aws ecr create-repository --repository-name ardanlabs/service-auth
aws ecr create-repository --repository-name ardanlabs/service-metrics
```

### 3. Configure GitHub Secrets

Set the secrets manually based on your AWS account and cluster names.

## Cost Estimation

- **EKS Clusters**: ~$150-300/month (depending on instance types and node count)
- **ECR**: ~$5-10/month (for image storage)
- **ALB**: ~$20-30/month
- **RDS** (optional): ~$50-100/month

## Cleanup

To destroy the infrastructure:

```bash
terraform destroy
```

**Warning**: This will delete all resources including data!
