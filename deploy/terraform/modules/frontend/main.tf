resource "railway_service" "frontend" {
  name       = "frontend"
  project_id = var.project_id

  source_repo        = var.github_repo
  source_repo_branch = var.github_branch
  root_directory     = var.root_directory
  config_path        = var.config_path
}

resource "railway_variable_collection" "frontend" {
  environment_id = var.environment_id
  service_id     = railway_service.frontend.id

  variables = concat([
    {
      name  = "REACT_APP_API_URL"
      value = var.backend_url
    },
    {
      name  = "NODE_ENV"
      value = "production"
    }
    ], [
    for key, value in var.additional_env_vars : {
      name  = key
      value = value
    }
  ])
}
