terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.9.0"
    }
  }
}

resource "digitalocean_vpc" "main" {
  name     = "${var.namespace}-vpc"
  region   = var.region
  ip_range = var.ip_range

  # Increasing the timeout here to allow time for the Kubernetes resources to be
  # removed from the VPC before the VPC is destroyed.
  # See https://github.com/digitalocean/terraform-provider-digitalocean/issues/446
  # for more details.
  timeouts {
    delete = "10m"
  }
}

resource "digitalocean_kubernetes_cluster" "main" {
  name         = "${var.namespace}-kubernetes-cluster"
  region       = var.region
  version      = var.kubernetes_version
  vpc_uuid     = digitalocean_vpc.main.id
  auto_upgrade = true
  node_pool {
    name       = "${var.namespace}-default-node-pool"
    size       = var.size
    node_count = var.node_count
    auto_scale = false
    labels     = var.labels
    tags       = var.tags
  }
  tags = var.tags
}

output "kubernetes_cluster_raw_config" {
  value     = digitalocean_kubernetes_cluster.main.kube_config[0].raw_config
  sensitive = true
}
