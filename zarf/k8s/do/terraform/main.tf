terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.9.0"
    }
  }
}

# Configuration provided via environment variables.
provider "digitalocean" {}

# Call the module.
module "k8s_demo" {
  source = "./modules"

  # Required variables.
  namespace = "k8s-demo"

  # Optional variables showing their default value.
  kubernetes_version = "1.20.7-do.0"
  ip_range           = "10.0.0.0/16"
  labels             = {}
  node_count         = 3
  region             = "nyc3"
  size               = "s-1vcpu-2gb-amd"
  tags               = []
}

# Create the `kubeconfig
resource "local_file" "k8s_demo_kubeconfig" {
  sensitive_content = module.k8s_demo.kubernetes_cluster_raw_config
  filename          = "${path.root}/kubeconfig.yml"
  file_permission   = "0644"
}
