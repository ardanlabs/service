#!/bin/bash

# Script to extract Terraform outputs for GitHub secrets
# Run this after: terraform apply

echo "=== GitHub Secrets Values ==="
echo ""

echo "# AWS Configuration"
echo "AWS_ACCOUNT_ID=$(terraform output -raw aws_account_id)"
echo "AWS_REGION=$(terraform output -raw aws_region)"
echo ""

echo "# EKS Cluster Names"
echo "EKS_STAGING_CLUSTER=$(terraform output -raw staging_cluster_name)"
echo "EKS_PRODUCTION_CLUSTER=$(terraform output -raw production_cluster_name)"
echo ""

echo "# Service URLs"
echo "STAGING_URL=$(terraform output -raw staging_url)"
echo "PRODUCTION_URL=$(terraform output -raw production_url)"
echo ""

echo "# Manual Setup Required:"
echo "AWS_ACCESS_KEY_ID=<create IAM user and get access key>"
echo "AWS_SECRET_ACCESS_KEY=<create IAM user and get secret key>"
echo ""

echo "=== Instructions ==="
echo "1. Copy the values above to your GitHub repository secrets"
echo "2. Create an IAM user with ECR and EKS permissions"
echo "3. Add the IAM user's access key and secret key to GitHub secrets"
echo "4. Test your GitHub Actions workflows"
