variable "project_name" {
  description = "Project name used for repository naming"
  type        = string
}

variable "repositories" {
  description = "List of repository names to create"
  type        = list(string)
}
