terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Data sources
data "aws_caller_identity" "current" {}

# VPC and Networking
module "vpc" {
  source = "./modules/vpc"

  environment = var.environment
  vpc_cidr    = var.vpc_cidr
}

# EKS Clusters
module "eks_staging" {
  source = "./modules/eks"

  cluster_name   = "${var.project_name}-staging"
  environment    = "staging"
  vpc_id         = module.vpc.vpc_id
  subnet_ids     = module.vpc.private_subnet_ids
  instance_types = ["t3.medium"]
  desired_size   = 2
  max_size       = 4
  min_size       = 1
}

module "eks_production" {
  source = "./modules/eks"

  cluster_name   = "${var.project_name}-production"
  environment    = "production"
  vpc_id         = module.vpc.vpc_id
  subnet_ids     = module.vpc.private_subnet_ids
  instance_types = ["t3.large"]
  desired_size   = 3
  max_size       = 6
  min_size       = 2
}

# ECR Repositories
module "ecr" {
  source = "./modules/ecr"

  project_name = var.project_name
  repositories = [
    "sales",
    "auth",
    "metrics"
  ]
}

# RDS Database (Optional - you can use containerized PostgreSQL instead)
module "rds" {
  source = "./modules/rds"
  count  = var.create_rds ? 1 : 0

  environment     = var.environment
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.database_subnet_ids
  security_groups = [module.vpc.database_security_group_id]
}

# Application Load Balancer
module "alb" {
  source = "./modules/alb"

  environment = var.environment
  vpc_id      = module.vpc.vpc_id
  subnet_ids  = module.vpc.public_subnet_ids
}
