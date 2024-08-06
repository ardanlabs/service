output "region" {
  value = var.region
}

output "cluster_name" {
  description = "Kubernetes Cluster Name"
  value       = module.cluster.cluster_name
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.cluster.cluster_endpoint
}
