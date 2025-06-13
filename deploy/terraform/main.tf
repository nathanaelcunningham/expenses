resource "railway_project" "expenses" {
  name = "${var.railway_project_name}-${var.environment}"
}

resource "railway_environment" "expenses" {
  project_id = railway_project.expenses.id
  name       = var.environment
}

module "database" {
  source = "./modules/database"
  
  project_id     = railway_project.expenses.id
  environment_id = railway_environment.expenses.id
  environment    = var.environment
  db_user        = var.db_user
  db_password    = var.db_password
  db_name        = var.db_name
}

module "backend" {
  source = "./modules/backend"
  
  project_id     = railway_project.expenses.id
  environment_id = railway_environment.expenses.id
  environment    = var.environment
  github_repo    = var.github_repo
  github_branch  = var.github_branch
  
  # Database connection - Railway will provide DATABASE_URL automatically
  # But we can still set individual variables if the app expects them
  db_host     = module.database.db_host
  db_port     = module.database.db_port
  db_user     = var.db_user
  db_password = var.db_password
  db_name     = var.db_name
  
  depends_on = [module.database]
}

module "frontend" {
  source = "./modules/frontend"
  
  project_id     = railway_project.expenses.id
  environment_id = railway_environment.expenses.id
  environment    = var.environment
  github_repo    = var.github_repo
  github_branch  = var.github_branch
  backend_url    = module.backend.backend_url
  
  depends_on = [module.backend]
}