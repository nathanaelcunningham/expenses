resource "railway_service" "postgres" {
  name         = "postgres"
  project_id   = var.project_id
  source_image = "ghcr.io/railwayapp-templates/postgres-ssl:16"

  volume = {
    mount_path = "/var/lib/postgresql/data"
    name       = var.volume_name
  }
}

resource "railway_tcp_proxy" "postgres" {
  application_port = 5432
  environment_id   = var.environment_id
  service_id       = railway_service.postgres.id
}

resource "railway_variable_collection" "postgres" {
  environment_id = var.environment_id
  service_id     = railway_service.postgres.id

  variables = [
    {
      name  = "PGDATA"
      value = "${railway_service.postgres.volume.mount_path}/pgdata"
    },
    {
      name  = "PGHOST"
      value = "$${{RAILWAY_PRIVATE_DOMAIN}}"
    },
    {
      name  = "PGPORT"
      value = 5432
    },
    {
      name  = "PGUSER"
      value = var.postgres_user
    },
    {
      name  = "PGPASSWORD"
      value = var.postgres_password
    },
    {
      name  = "PGDATABASE"
      value = var.postgres_db
    },
    {
      name  = "DATABASE_URL"
      value = "postgresql://${var.postgres_user}:${var.postgres_password}@$${{RAILWAY_PRIVATE_DOMAIN}}:5432/${var.postgres_db}"
    },
    {
      name  = "DATABASE_PUBLIC_URL"
      value = "postgresql://${var.postgres_user}:${var.postgres_password}@${railway_tcp_proxy.postgres.domain}:${railway_tcp_proxy.postgres.proxy_port}/${var.postgres_db}"
    },
    {
      name  = "POSTGRES_DB"
      value = var.postgres_db
    },
    {
      name  = "POSTGRES_PASSWORD"
      value = var.postgres_password
    },
    {
      name  = "POSTGRES_USER"
      value = var.postgres_user
    }
  ]
}
