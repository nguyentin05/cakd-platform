variable "project_name" {
  description = "Name of the GitHub repository to create"
  type        = string
}

variable "github_owner" {
  description = "GitHub username or organization that owns the repository"
  type        = string
}

variable "github_token" {
  description = "GitHub Personal Access Token with repo and packages scope"
  type        = string
  sensitive   = true
}
