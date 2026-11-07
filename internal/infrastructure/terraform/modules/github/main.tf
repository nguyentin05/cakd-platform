terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
  }
}

provider "github" {
  owner = var.github_owner
  token = var.github_token
}

# trivy:ignore:AVD-GIT-0001
resource "github_repository" "project" {
  name        = var.project_name
  description = "Cloud-Native application created by CAKD Platform"
  # trivy:ignore:AVD-GIT-0001
  visibility           = "public"
  auto_init            = false
  vulnerability_alerts = true

  has_issues   = true
  has_projects = false
  has_wiki     = false
}

