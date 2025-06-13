output "project_id" {
  description = "Railway project ID"
  value       = railway_project.expenses.id
}

output "frontend_url" {
  description = "Frontend application URL"
  value       = module.frontend.frontend_url
}

output "backend_url" {
  description = "Backend API URL"
  value       = module.backend.backend_url
}

output "database_url" {
  description = "Database connection URL"
  value       = module.database.database_url
  sensitive   = true
}