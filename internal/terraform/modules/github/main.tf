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

# ---------- Repository ----------

resource "github_repository" "project" {
  name        = var.project_name
  description = "Cloud-Native application bootstrapped by CAKD Platform"
  visibility  = "public"
  auto_init   = false

  has_issues   = true
  has_projects = false
  has_wiki     = false
}

# ---------- Branch Protection ----------

resource "github_branch_protection" "main" {
  repository_id = github_repository.project.node_id
  pattern       = "main"

  required_pull_request_reviews {
    required_approving_review_count = 0
  }

  # CI phải pass trước khi merge
  required_status_checks {
    strict = true
    contexts = ["build-and-push"]
  }

  allows_force_pushes = false
}
