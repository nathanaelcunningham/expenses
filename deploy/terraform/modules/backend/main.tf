resource "railway_service" "backend" {
  project_id = var.project_id
  name       = "backend-${var.environment}"
}

resource "railway_deployment" "backend" {
  service_id = railway_service.backend.id
  
  # Use the pre-built Docker image
  image = var.backend_image
  
  variables = {
    DB_HOST     = var.db_host
    DB_PORT     = var.db_port
    DB_USER     = var.db_user
    DB_PASSWORD = var.db_password
    DB_NAME     = var.db_name
  }
}

resource "railway_variable" "db_host" {
  service_id = railway_service.backend.id
  name       = "DB_HOST"
  value      = var.db_host
}

resource "railway_variable" "db_port" {
  service_id = railway_service.backend.id
  name       = "DB_PORT"
  value      = var.db_port
}

resource "railway_variable" "db_user" {
  service_id = railway_service.backend.id
  name       = "DB_USER"
  value      = var.db_user
}

resource "railway_variable" "db_password" {
  service_id = railway_service.backend.id
  name       = "DB_PASSWORD"
  value      = var.db_password
}

resource "railway_variable" "db_name" {
  service_id = railway_service.backend.id
  name       = "DB_NAME"
  value      = var.db_name
}