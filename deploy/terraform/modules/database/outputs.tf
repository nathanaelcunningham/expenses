output "service_id" {
  description = "PostgreSQL service ID"
  value       = railway_service.postgres.id
}

output "private_domain" {
  description = "Private domain for internal connections"
  value       = "${railway_service.postgres.name}.$${{RAILWAY_PRIVATE_DOMAIN}}"
}

output "tcp_proxy_domain" {
  description = "TCP proxy domain for external connections"
  value       = railway_tcp_proxy.postgres.domain
}

output "tcp_proxy_port" {
  description = "TCP proxy port for external connections"
  value       = railway_tcp_proxy.postgres.proxy_port
}

output "database_url" {
  description = "Database URL for internal connections"
  value       = "postgresql://${var.postgres_user}:${var.postgres_password}@$${{postgres.DATABASE_URL}}:5432/${var.postgres_db}"
  sensitive   = true
}

output "database_public_url" {
  description = "Database URL for external connections"
  value       = "postgresql://${var.postgres_user}:${var.postgres_password}@${railway_tcp_proxy.postgres.domain}:${railway_tcp_proxy.postgres.proxy_port}/${var.postgres_db}"
  sensitive   = true
}

