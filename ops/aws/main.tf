module "repository" {
  source = "./modules/repository"
  region = var.region
  name   = "ardanlabs"
  github_owner = "optiop"
  github_repo = "kubernetes-go-boilerplate"
  github_oidc_provider_arn = var.github_oidc_provider_arn
}

module "cluster" {
  source = "./modules/cluster"
  region = var.region
  cluster_name = "eks-cluster"
  cluster_version = "1.30"
}