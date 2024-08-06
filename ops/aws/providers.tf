terraform {
  required_version = "~>1.3"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.7.0"
    }

    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.9.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.16"
    }
  }

#  backend "s3" {
#    bucket  = "optiop-ar-terraform-state"
#    key     = "optiop/cluster"
#    region  = "eu-central-1"
#    encrypt = true
#  }
}

provider "kubernetes" {
  host                   = module.cluster.cluster_endpoint
  cluster_ca_certificate = base64decode(module.cluster.cluster_certificate_authority_data)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    args = [
      "--region",
      var.region,
      "eks",
      "get-token",
      "--cluster-name",
      module.cluster.cluster_name
    ]
    command = "aws"
  }
}

provider "helm" {
  kubernetes {
    host                   = module.cluster.cluster_endpoint
    cluster_ca_certificate = base64decode(module.cluster.cluster_certificate_authority_data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      args = [
        "--region",
        var.region,
        "eks",
        "get-token",
        "--cluster-name",
        module.cluster.cluster_name
      ]
      command = "aws"
    }
  }
}

provider "aws" {
  region = var.region
}
