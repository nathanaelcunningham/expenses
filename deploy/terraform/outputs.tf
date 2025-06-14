output "project_id" {
  description = "Railway project ID"
  value       = railway_project.expenses.id
}

output "project_name" {
  description = "Railway project name"
  value       = local.project.name
}

output "project_url" {
  description = "Railway project URL"
  value       = "https://railway.app/project/${railway_project.expenses.id}"
}

output "database_service_id" {
  description = "Database service ID"
  value       = module.database.service_id
}

output "database_tcp_proxy_domain" {
  description = "Database TCP proxy domain for external connections"
  value       = module.database.tcp_proxy_domain
}

output "database_tcp_proxy_port" {
  description = "Database TCP proxy port for external connections"
  value       = module.database.tcp_proxy_port
}

output "backend_service_id" {
  description = "Backend service ID"
  value       = module.backend.service_id
}

output "backend_url" {
  description = "Backend service URL"
  value       = module.backend.service_url
}

# output "frontend_service_id" {
#   description = "Frontend service ID"
#   value       = module.frontend.service_id
# }

output "deployment_info" {
  description = "Summary of deployment configuration"
  value = {
    project_name       = local.project.name
    database_name      = local.database.database
    repository_url     = local.repository.url
    repository_branch  = local.repository.branch
    backend_directory  = local.directories.backend
    frontend_directory = local.directories.frontend
    environment        = local.tags.environment
    managed_by         = local.tags.managed_by
  }
}

