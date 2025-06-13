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