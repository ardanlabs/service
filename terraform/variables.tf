variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "ardanlabs-service"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "create_rds" {
  description = "Whether to create RDS database"
  type        = bool
  default     = false
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = ""
}

variable "staging_domain" {
  description = "Staging domain name"
  type        = string
  default     = "staging.example.com"
}

variable "production_domain" {
  description = "Production domain name"
  type        = string
  default     = "app.example.com"
}
