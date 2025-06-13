terraform {
  required_providers {
    railway = {
      source = "terraform-community-providers/railway"
    }
  }
}

# Use Railway's managed PostgreSQL template
resource "railway_service" "database" {
  project_id = var.project_id
  name       = "database-${var.environment}"
  
  # Deploy from Railway's PostgreSQL template
  template = "postgres"
}

# Railway's PostgreSQL template automatically provides DATABASE_URL
# We can still set custom environment variables if needed
resource "railway_variable" "postgres_db_name" {
  environment_id = var.environment_id
  service_id     = railway_service.database.id
  name           = "POSTGRES_DB"
  value          = var.db_name
}

resource "railway_variable" "postgres_user" {
  environment_id = var.environment_id
  service_id     = railway_service.database.id
  name           = "POSTGRES_USER"
  value          = var.db_user
}

resource "railway_variable" "postgres_password" {
  environment_id = var.environment_id
  service_id     = railway_service.database.id
  name           = "POSTGRES_PASSWORD"
  value          = var.db_password
}