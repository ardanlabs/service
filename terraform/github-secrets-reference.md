# GitHub Secrets Reference

After running `terraform apply`, use these commands to get the values for your GitHub repository secrets.

## Required GitHub Secrets

### AWS Configuration
```bash
# Get AWS Account ID
terraform output aws_account_id

# Get AWS Region  
terraform output aws_region
```

**Manual values:**
- `AWS_ACCESS_KEY_ID` - Create an IAM user with ECR and EKS permissions
- `AWS_SECRET_ACCESS_KEY` - Secret key for the IAM user

### EKS Cluster Names
```bash
# Get staging cluster name
terraform output staging_cluster_name

# Get production cluster name
terraform output production_cluster_name
```

### Service URLs
```bash
# Get staging URL
terraform output staging_url

# Get production URL
terraform output production_url
```

## IAM User Setup for GitHub Actions

Create an IAM user with these policies:
- `AmazonEC2ContainerRegistryFullAccess`
- `AmazonEKSClusterPolicy`
- `AmazonEKSWorkerNodePolicy`

Or create a custom policy with these permissions:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:*",
        "eks:*",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions"
      ],
      "Resource": "*"
    }
  ]
}
```

## Adding Secrets to GitHub

1. Go to your repository → Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add each secret with the values from above

## Example Secret Values

Based on the current configuration:
- `AWS_ACCOUNT_ID`: [Your AWS Account ID]
- `AWS_REGION`: us-west-2
- `EKS_STAGING_CLUSTER`: ardanlabs-service-staging
- `EKS_PRODUCTION_CLUSTER`: ardanlabs-service-production
- `STAGING_URL`: https://staging.ardanlabs-service.com
- `PRODUCTION_URL`: https://app.ardanlabs-service.com
