output "service_id" {
  description = "Database service ID"
  value       = railway_service.database.id
}

output "db_host" {
  description = "Database host"
  value       = railway_service.database.name
}

output "db_port" {
  description = "Database port"
  value       = "5432"
}

output "database_url" {
  description = "Full database connection URL"
  value       = "postgresql://${var.db_user}:${var.db_password}@${railway_service.database.name}:5432/${var.db_name}"
  sensitive   = true
}