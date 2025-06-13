terraform {
  required_providers {
    railway = {
      source = "terraform-community-providers/railway"
    }
  }
}

resource "railway_service" "backend" {
  project_id = var.project_id
  name       = "backend-${var.environment}"
  
  # Use source-based deployment with GitHub repo
  source_repo      = var.github_repo
  source_repo_branch = var.github_branch
  root_directory   = "backend"
}


resource "railway_variable" "db_host" {
  environment_id = var.environment_id
  service_id     = railway_service.backend.id
  name           = "DB_HOST"
  value          = var.db_host
}

resource "railway_variable" "db_port" {
  environment_id = var.environment_id
  service_id     = railway_service.backend.id
  name           = "DB_PORT"
  value          = var.db_port
}

resource "railway_variable" "db_user" {
  environment_id = var.environment_id
  service_id     = railway_service.backend.id
  name           = "DB_USER"
  value          = var.db_user
}

resource "railway_variable" "db_password" {
  environment_id = var.environment_id
  service_id     = railway_service.backend.id
  name           = "DB_PASSWORD"
  value          = var.db_password
}

resource "railway_variable" "db_name" {
  environment_id = var.environment_id
  service_id     = railway_service.backend.id
  name           = "DB_NAME"
  value          = var.db_name
}