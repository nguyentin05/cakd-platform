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

resource "github_repository" "project" {
  name        = var.project_name
  description = "Cloud-Native application created by CAKD Platform"
  visibility  = "public"
  auto_init   = false

  has_issues   = true
  has_projects = false
  has_wiki     = false
}

