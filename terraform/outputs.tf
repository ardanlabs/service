# EKS Cluster Outputs
output "staging_cluster_name" {
  description = "Name of the staging EKS cluster"
  value       = module.eks_staging.cluster_name
}

output "production_cluster_name" {
  description = "Name of the production EKS cluster"
  value       = module.eks_production.cluster_name
}

output "staging_cluster_endpoint" {
  description = "Endpoint for staging EKS cluster"
  value       = module.eks_staging.cluster_endpoint
}

output "production_cluster_endpoint" {
  description = "Endpoint for production EKS cluster"
  value       = module.eks_production.cluster_endpoint
}

# ECR Repository Outputs
output "ecr_repository_urls" {
  description = "URLs of the ECR repositories"
  value       = module.ecr.repository_urls
}

# Load Balancer Outputs
output "staging_url" {
  description = "URL for staging environment"
  value       = "https://${var.staging_domain}"
}

output "production_url" {
  description = "URL for production environment"
  value       = "https://${var.production_domain}"
}

# RDS Outputs (if created)
output "rds_endpoint" {
  description = "RDS endpoint"
  value       = var.create_rds ? module.rds[0].endpoint : null
}

# AWS Account ID
output "aws_account_id" {
  description = "AWS Account ID"
  value       = data.aws_caller_identity.current.account_id
}

# AWS Region
output "aws_region" {
  description = "AWS Region"
  value       = var.aws_region
}
