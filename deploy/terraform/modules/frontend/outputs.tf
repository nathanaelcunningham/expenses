output "service_id" {
  description = "Frontend service ID"
  value       = railway_service.frontend.id
}

output "frontend_url" {
  description = "Frontend application URL"
  value       = "https://${railway_service.frontend.name}.railway.app"
}