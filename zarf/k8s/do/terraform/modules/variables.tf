#
# Required variables.
#

variable "namespace" {
  description = "An arbitrary string that is prefixed to resources to provide uniqueness among resources."
  type        = string
}

#
# Optional variables.
#

variable "kubernetes_version" {
  description = "Version of the Kubernetes cluster to create."
  type        = string
  default     = "1.20.7-do.0"
}

variable "ip_range" {
  description = "IP address range in CIDR notation for the VPC that will be created. Cannot be larger than /16 or smaller than /24."
  type        = string
  default     = "10.0.0.0/16"
}

variable "labels" {
  description = "Labels to apply to the default Kubernetes node pool."
  type        = map(string)
  default     = {}
}

variable "node_count" {
  description = "Number of nodes in the default Kubernetes node pool."
  type        = number
  default     = 3
}

variable "region" {
  description = "Region to place the resources in."
  type        = string
  default     = "nyc3"
}

variable "size" {
  description = "Size of the droplet to use for the default Kubernetes node pool."
  type        = string
  default     = "s-1vcpu-2gb-amd"
}

variable "tags" {
  description = "Tags to apply to the Kubernetes cluster and default node pool"
  type        = list(string)
  default     = []
}
