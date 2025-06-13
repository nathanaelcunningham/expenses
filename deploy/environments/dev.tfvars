environment = "dev"
railway_project_name = "expenses-app"

# Docker images (use :dev tag for development)
backend_image = "ghcr.io/nathanaelcunningham/expenses/backend:dev"
frontend_image = "ghcr.io/nathanaelcunningham/expenses/frontend:dev"

# Database configuration
db_user = "expenses_dev"
db_name = "expenses_dev"
# db_password should be set via TF_VAR_db_password environment variable