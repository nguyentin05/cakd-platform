output "repo_full_name" {
  description = "Full name of the repository (owner/name)"
  value       = github_repository.project.full_name
}

output "repo_html_url" {
  description = "URL of the repository"
  value       = github_repository.project.html_url
}

output "repo_clone_url" {
  description = "HTTPS clone URL of the repository"
  value       = github_repository.project.http_clone_url
}
