output repository_url {
  value = { for repo in aws_ecr_repository.repository : repo.name => repo.repository_url }
}
