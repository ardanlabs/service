variable "region" {
  type = string
}

variable "name" {
  type = string
}

variable "tags" {
  type = map(string)
  default = {}
}

variable "github_owner" {
  type = string
}

variable "github_repo" {
  type = string
}

variable "github_oidc_provider_arn" {
  type = string
}
