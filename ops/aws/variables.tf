variable "region" {
  description = "AWS region for deployment"
  type        = string
  default     = "eu-central-1"
}

variable "github_oidc_provider_arn" {
  description = "OIDC provider ARN for GitHub Actions"
  type        = string
  default     = "arn:aws:iam::615677714887:oidc-provider/token.actions.githubusercontent.com"
}
