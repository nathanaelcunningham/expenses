resource "railway_service" "database" {
  project_id = var.project_id
  name       = "database-${var.environment}"
}

resource "railway_deployment" "database" {
  service_id = railway_service.database.id
  
  variables = {
    POSTGRES_USER     = var.db_user
    POSTGRES_PASSWORD = var.db_password
    POSTGRES_DB       = var.db_name
  }
}

resource "railway_variable" "postgres_user" {
  service_id = railway_service.database.id
  name       = "POSTGRES_USER"
  value      = var.db_user
}

resource "railway_variable" "postgres_password" {
  service_id = railway_service.database.id
  name       = "POSTGRES_PASSWORD"
  value      = var.db_password
}

resource "railway_variable" "postgres_db" {
  service_id = railway_service.database.id
  name       = "POSTGRES_DB"
  value      = var.db_name
}