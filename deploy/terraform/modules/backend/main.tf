resource "railway_service" "backend" {
  name       = "backend"
  project_id = var.project_id

  source_repo        = var.github_repo
  source_repo_branch = var.github_branch
  root_directory     = var.root_directory
  config_path        = var.config_path
}

resource "railway_variable_collection" "backend" {
  environment_id = var.environment_id
  service_id     = railway_service.backend.id

  variables = concat([
    {
      name  = "DATABASE_URL"
      value = var.database_url
    },
    {
      name  = "PORT"
      value = tostring(var.port)
    },
    {
      name  = "GO_ENV"
      value = "production"
    }
    ], [
    for key, value in var.additional_env_vars : {
      name  = key
      value = value
    }
  ])
}

