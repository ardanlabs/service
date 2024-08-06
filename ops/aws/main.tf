module "cluster" {
  source = "./modules/cluster"
  region = var.region
  cluster_name = "eks-cluster"
  cluster_version = "1.30"
}