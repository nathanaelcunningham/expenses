locals {
  # Project configuration
  project = {
    name        = var.project_name
    description = var.project_description
    private     = true
  }

  # Database configuration
  database = {
    user        = var.postgres_user
    password    = var.postgres_password
    database    = var.postgres_db
    volume_name = "expenses-data"
  }

  # Repository configuration
  repository = {
    url    = var.github_repo
    branch = var.github_branch
  }

  # Service directories
  directories = {
    backend  = "backend"
    frontend = "frontend"
  }


  # Common tags/labels
  tags = {
    project     = "expenses-app"
    environment = "production"
    managed_by  = "terraform"
  }
}

provider "railway" {
  token = var.railway_token
}

resource "railway_project" "expenses" {
  name        = local.project.name
  description = local.project.description
  private     = local.project.private
}

module "database" {
  source = "./modules/database"

  project_id        = railway_project.expenses.id
  environment_id    = railway_project.expenses.default_environment.id
  postgres_user     = local.database.user
  postgres_password = local.database.password
  postgres_db       = local.database.database
  volume_name       = local.database.volume_name
}

module "backend" {
  source = "./modules/backend"

  project_id     = railway_project.expenses.id
  environment_id = railway_project.expenses.default_environment.id
  github_repo    = local.repository.url
  github_branch  = local.repository.branch
  database_url   = module.database.database_url
  root_directory = local.directories.backend

  depends_on = [module.database]
}

module "frontend" {
  source = "./modules/frontend"

  project_id      = railway_project.expenses.id
  environment_id  = railway_project.expenses.default_environment.id
  github_repo     = local.repository.url
  github_branch   = local.repository.branch
  backend_url     = module.backend.service_url
  root_directory  = local.directories.frontend

  depends_on = [module.backend]
}
