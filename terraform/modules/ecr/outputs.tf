output "repository_urls" {
  description = "URLs of the ECR repositories"
  value       = { for k, v in aws_ecr_repository.repositories : k => v.repository_url }
}

output "repository_names" {
  description = "Names of the ECR repositories"
  value       = { for k, v in aws_ecr_repository.repositories : k => v.name }
}
