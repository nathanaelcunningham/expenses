output "service_id" {
  description = "Backend service ID"
  value       = railway_service.backend.id
}

output "backend_url" {
  description = "Backend service URL"
  value       = "https://${railway_service.backend.name}.railway.app"
}